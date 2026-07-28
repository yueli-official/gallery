package dao_test

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	_ "github.com/gogf/gf/contrib/drivers/pgsql/v2"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/lib/pq"

	"github.com/yueli-official/foundation/go/classification"
	"github.com/yueli-official/gallery/api/internal/dao"
)

func TestPostgreSQL18ClassificationGovernanceRoundTrip(t *testing.T) {
	host := os.Getenv("GALLERY_PG_HOST")
	if host == "" {
		t.Skip("set GALLERY_PG_HOST to run the Gallery PostgreSQL 18.4 integration test")
	}
	port := integrationEnv("GALLERY_PG_PORT", "5432")
	user := integrationEnv("GALLERY_PG_USER", "postgres")
	password := os.Getenv("GALLERY_PG_PASS")
	adminDSN := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=postgres sslmode=disable", host, port, user, password)
	admin, err := sql.Open("postgres", adminDSN)
	if err != nil {
		t.Fatal(err)
	}
	defer admin.Close()

	database := fmt.Sprintf("gallery_classification_test_%d", time.Now().UnixNano())
	if _, err := admin.Exec(`CREATE DATABASE ` + pq.QuoteIdentifier(database)); err != nil {
		t.Fatal(err)
	}
	defer func() {
		_, _ = admin.Exec(`SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = $1`, database)
		_, _ = admin.Exec(`DROP DATABASE IF EXISTS ` + pq.QuoteIdentifier(database))
	}()

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, database)
	sqlDB, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer sqlDB.Close()
	var serverVersion int
	if err := sqlDB.QueryRow(`SELECT current_setting('server_version_num')::int`).Scan(&serverVersion); err != nil {
		t.Fatal(err)
	}
	if serverVersion < 180004 {
		t.Fatalf("PostgreSQL %d is too old; Gallery requires 18.4", serverVersion)
	}
	up, err := os.ReadFile("../../manifest/sql/migrations/0001_init.up.sql")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := sqlDB.Exec(string(up)); err != nil {
		t.Fatalf("apply Gallery migration: %v", err)
	}

	var catalogID, categoryID string
	if err := sqlDB.QueryRow(`INSERT INTO gallery_classification_catalogs (catalog_key) VALUES ('gallery') RETURNING id::text`).Scan(&catalogID); err != nil {
		t.Fatal(err)
	}
	if err := sqlDB.QueryRow(`
INSERT INTO gallery_categories (catalog_id, slug, name, status, first_activated_at)
VALUES ($1::uuid, 'wallpaper', '壁纸', 'active', NOW())
RETURNING id::text`, catalogID).Scan(&categoryID); err != nil {
		t.Fatal(err)
	}
	if _, err := sqlDB.Exec(`
INSERT INTO gallery_classification_policy_profiles (catalog_id, policy_key, schema_version, policy_revision, document)
VALUES ($1::uuid, 'gallery.image.public', 1, 1, '{"category":{},"facets":[],"tags":{"unknown":"propose"},"discovery":{"defaultSort":"editorial"}}'::jsonb)`, catalogID); err != nil {
		t.Fatal(err)
	}

	databaseHandle, err := gdb.New(gdb.ConfigNode{Type: "pgsql", Host: host, Port: port, User: user, Pass: password, Name: database})
	if err != nil {
		t.Fatal(err)
	}
	defer databaseHandle.Close(context.Background())
	store := dao.NewPG(databaseHandle)
	snapshot, err := store.ClassificationSnapshot(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	compiled := classification.Compile(snapshot)
	if compiled.Outcome != classification.OutcomeAccepted {
		t.Fatalf("compile = %#v", compiled)
	}
	preparation := compiled.Catalog.Govern(classification.GovernRequest{
		PolicyKey: "gallery.image.public",
		Command: classification.GovernCommand{
			Operation: classification.GovernSetStatus,
			Kind:      classification.GovernCategory,
			ID:        categoryID,
			Status:    classification.StatusInactive,
		},
	})
	factRequest := preparation.FactRequest()
	impacts, impactToken, err := store.ClassificationGovernanceImpacts(context.Background(), factRequest.Impacts)
	if err != nil {
		t.Fatal(err)
	}
	result := preparation.Complete(classification.GovernFacts{
		CatalogRevision: factRequest.CatalogRevision,
		RequestToken:    factRequest.RequestToken,
		FreshnessToken:  impactToken,
		Impacts:         impacts,
	})
	if result.Outcome != classification.OutcomePlanned {
		t.Fatalf("governance result = %#v", result)
	}

	listener := pq.NewListener(dsn, 100*time.Millisecond, time.Second, nil)
	if err := listener.Listen("classification_catalog_changed"); err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	revision, err := store.ExecuteClassificationGovernance(context.Background(), "integration-test", factRequest, result.Plan)
	if err != nil {
		t.Fatal(err)
	}
	if revision != 2 {
		t.Fatalf("revision = %d, want 2", revision)
	}
	select {
	case notification := <-listener.Notify:
		if notification == nil || strings.TrimSpace(notification.Extra) != "gallery:2" {
			t.Fatalf("notification = %#v", notification)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for classification revision notification")
	}
	var status string
	if err := sqlDB.QueryRow(`SELECT status FROM gallery_categories WHERE id = $1::uuid`, categoryID).Scan(&status); err != nil {
		t.Fatal(err)
	}
	if status != "inactive" {
		t.Fatalf("category status = %q", status)
	}
	var outboxCount int
	if err := sqlDB.QueryRow(`SELECT COUNT(*) FROM gallery_classification_outbox WHERE revision = 2`).Scan(&outboxCount); err != nil {
		t.Fatal(err)
	}
	if outboxCount != 1 {
		t.Fatalf("outbox count = %d", outboxCount)
	}
}

func integrationEnv(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}
