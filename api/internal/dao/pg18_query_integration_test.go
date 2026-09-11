package dao_test

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	_ "github.com/gogf/gf/contrib/drivers/pgsql/v2"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/lib/pq"

	"github.com/yueli-official/foundation/go/classification"
	"github.com/yueli-official/gallery/api/internal/dao"
	"github.com/yueli-official/gallery/api/internal/model"
)

type galleryPG18Fixture struct {
	SQL   *sql.DB
	Store *dao.PG
}

func newGalleryPG18Fixture(t *testing.T) galleryPG18Fixture {
	t.Helper()
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
	t.Cleanup(func() { _ = admin.Close() })
	var serverVersion int
	if err := admin.QueryRow(`SELECT current_setting('server_version_num')::int`).Scan(&serverVersion); err != nil {
		t.Fatal(err)
	}
	if serverVersion < 180004 {
		t.Fatalf("PostgreSQL %d is too old; Gallery requires 18.4", serverVersion)
	}

	database := fmt.Sprintf("gallery_classification_query_test_%d", time.Now().UnixNano())
	if _, err := admin.Exec(`CREATE DATABASE ` + pq.QuoteIdentifier(database)); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, _ = admin.Exec(`SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = $1`, database)
		_, _ = admin.Exec(`DROP DATABASE IF EXISTS ` + pq.QuoteIdentifier(database))
	})

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, database)
	sqlDB, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	migrations, err := filepath.Glob("../../manifest/sql/migrations/*.up.sql")
	if err != nil {
		t.Fatal(err)
	}
	for _, migration := range migrations {
		up, readErr := os.ReadFile(migration)
		if readErr != nil {
			t.Fatal(readErr)
		}
		if _, execErr := sqlDB.Exec(string(up)); execErr != nil {
			t.Fatalf("apply Gallery migration %s: %v", filepath.Base(migration), execErr)
		}
	}
	databaseHandle, err := gdb.New(gdb.ConfigNode{Type: "pgsql", Host: host, Port: port, User: user, Pass: password, Name: database})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = databaseHandle.Close(context.Background()) })
	return galleryPG18Fixture{SQL: sqlDB, Store: dao.NewPG(databaseHandle)}
}

func TestPostgreSQL18ContextualCategoryCountsPreserveOtherFacetsAndCountDistinctImages(t *testing.T) {
	fixture := newGalleryPG18Fixture(t)
	if _, err := fixture.SQL.Exec(`
INSERT INTO gallery_classification_catalogs (id, catalog_key) VALUES ('01990000-0000-7000-8000-000000000001', 'gallery');
INSERT INTO gallery_categories (id, catalog_id, parent_id, slug, name, status, first_activated_at) VALUES
('01990000-0000-7000-8100-000000000001', '01990000-0000-7000-8000-000000000001', NULL, 'visual', '视觉', 'active', NOW()),
('01990000-0000-7000-8100-000000000002', '01990000-0000-7000-8000-000000000001', '01990000-0000-7000-8100-000000000001', 'wallpaper', '壁纸', 'active', NOW()),
('01990000-0000-7000-8100-000000000003', '01990000-0000-7000-8000-000000000001', NULL, 'illustration', '插画', 'active', NOW());
INSERT INTO gallery_facets (id, catalog_id, slug, name, status, first_activated_at)
VALUES ('01990000-0000-7000-8200-000000000001', '01990000-0000-7000-8000-000000000001', 'scene', '场景', 'active', NOW());
INSERT INTO gallery_facet_values (id, catalog_id, facet_id, slug, name, status, first_activated_at) VALUES
('01990000-0000-7000-8300-000000000001', '01990000-0000-7000-8000-000000000001', '01990000-0000-7000-8200-000000000001', 'nature', '自然', 'active', NOW()),
('01990000-0000-7000-8300-000000000002', '01990000-0000-7000-8000-000000000001', '01990000-0000-7000-8200-000000000001', 'people', '人物', 'active', NOW());
INSERT INTO gallery_images (id, asset_id, title, alt_text, width, height, processing_state, review_state, publication_state, safety_state, public_rendition_ready, published_at) VALUES
('01990000-0000-7000-8400-000000000001', '01990000-0000-7000-8500-000000000001', '自然壁纸', '自然壁纸', 1600, 900, 'ready', 'approved', 'published', 'safe', TRUE, NOW()),
('01990000-0000-7000-8400-000000000002', '01990000-0000-7000-8500-000000000002', '人物壁纸', '人物壁纸', 1600, 900, 'ready', 'approved', 'published', 'safe', TRUE, NOW()),
('01990000-0000-7000-8400-000000000003', '01990000-0000-7000-8500-000000000003', '自然插画', '自然插画', 1600, 900, 'ready', 'approved', 'published', 'safe', TRUE, NOW()),
('01990000-0000-7000-8400-000000000004', '01990000-0000-7000-8500-000000000004', '隐藏插画', '隐藏插画', 1600, 900, 'ready', 'approved', 'hidden', 'safe', TRUE, NULL);
INSERT INTO gallery_image_category_assignments (image_id, category_id) VALUES
('01990000-0000-7000-8400-000000000001', '01990000-0000-7000-8100-000000000001'),
('01990000-0000-7000-8400-000000000001', '01990000-0000-7000-8100-000000000002'),
('01990000-0000-7000-8400-000000000002', '01990000-0000-7000-8100-000000000002'),
('01990000-0000-7000-8400-000000000003', '01990000-0000-7000-8100-000000000003'),
('01990000-0000-7000-8400-000000000004', '01990000-0000-7000-8100-000000000003');
INSERT INTO gallery_image_facet_assignments (image_id, facet_value_id) VALUES
('01990000-0000-7000-8400-000000000001', '01990000-0000-7000-8300-000000000001'),
('01990000-0000-7000-8400-000000000002', '01990000-0000-7000-8300-000000000002'),
('01990000-0000-7000-8400-000000000003', '01990000-0000-7000-8300-000000000001'),
('01990000-0000-7000-8400-000000000004', '01990000-0000-7000-8300-000000000001');`); err != nil {
		t.Fatal(err)
	}

	groups, freshness, err := fixture.Store.ClassificationCandidateCounts(context.Background(), model.ImageQuery{}, []classification.CandidateCountGroupRequest{{
		Kind: classification.FilterGroupCategory,
		OtherFilters: classification.FilterPlan{Groups: []classification.FilterGroup{{
			Kind: classification.FilterGroupFacet, OwnerID: "01990000-0000-7000-8200-000000000001",
			ValueIDs: []string{"01990000-0000-7000-8300-000000000001"},
		}}},
		Candidates: []classification.CandidateBucketRequest{
			{ValueID: "01990000-0000-7000-8100-000000000001", MatchingIDs: []string{"01990000-0000-7000-8100-000000000001", "01990000-0000-7000-8100-000000000002"}},
			{ValueID: "01990000-0000-7000-8100-000000000003", MatchingIDs: []string{"01990000-0000-7000-8100-000000000003"}},
		},
	}})
	if err != nil {
		t.Fatal(err)
	}
	want := []classification.CandidateCountGroup{{
		Kind: classification.FilterGroupCategory,
		Counts: []classification.CandidateCount{
			{ValueID: "01990000-0000-7000-8100-000000000001", Count: 1},
			{ValueID: "01990000-0000-7000-8100-000000000003", Count: 1},
		},
	}}
	if freshness == "" || !reflect.DeepEqual(groups, want) {
		t.Fatalf("groups = %#v, freshness = %q", groups, freshness)
	}
}

func TestPostgreSQL18RelatedImagesRanksAndExplainsSharedSignals(t *testing.T) {
	fixture := newGalleryPG18Fixture(t)
	if _, err := fixture.SQL.Exec(`
INSERT INTO gallery_classification_catalogs (id, catalog_key) VALUES ('01990000-0000-7000-8000-000000000001', 'gallery');
INSERT INTO gallery_categories (id, catalog_id, slug, name, status, first_activated_at) VALUES
('01990000-0000-7000-8100-000000000001', '01990000-0000-7000-8000-000000000001', 'wallpaper', '壁纸', 'active', NOW()),
('01990000-0000-7000-8100-000000000002', '01990000-0000-7000-8000-000000000001', 'illustration', '插画', 'active', NOW());
INSERT INTO gallery_facets (id, catalog_id, slug, name, status, first_activated_at)
VALUES ('01990000-0000-7000-8200-000000000001', '01990000-0000-7000-8000-000000000001', 'scene', '场景', 'active', NOW());
INSERT INTO gallery_facet_values (id, catalog_id, facet_id, slug, name, status, first_activated_at)
VALUES ('01990000-0000-7000-8300-000000000001', '01990000-0000-7000-8000-000000000001', '01990000-0000-7000-8200-000000000001', 'nature', '自然', 'active', NOW());
INSERT INTO gallery_images (id, asset_id, title, alt_text, width, height, processing_state, review_state, publication_state, safety_state, public_rendition_ready, published_at) VALUES
('01990000-0000-7000-8400-000000000001', '01990000-0000-7000-8500-000000000001', '来源图片', '来源图片', 1600, 900, 'ready', 'approved', 'published', 'safe', TRUE, NOW()),
('01990000-0000-7000-8400-000000000002', '01990000-0000-7000-8500-000000000002', '同类图片', '同类图片', 1600, 900, 'ready', 'approved', 'published', 'safe', TRUE, NOW()),
('01990000-0000-7000-8400-000000000003', '01990000-0000-7000-8500-000000000003', '同属性图片', '同属性图片', 1600, 900, 'ready', 'approved', 'published', 'safe', TRUE, NOW()),
('01990000-0000-7000-8400-000000000004', '01990000-0000-7000-8500-000000000004', '隐藏图片', '隐藏图片', 1600, 900, 'ready', 'approved', 'hidden', 'safe', TRUE, NULL);
INSERT INTO gallery_image_category_assignments (image_id, category_id) VALUES
('01990000-0000-7000-8400-000000000001', '01990000-0000-7000-8100-000000000001'),
('01990000-0000-7000-8400-000000000002', '01990000-0000-7000-8100-000000000001'),
('01990000-0000-7000-8400-000000000003', '01990000-0000-7000-8100-000000000002'),
('01990000-0000-7000-8400-000000000004', '01990000-0000-7000-8100-000000000001');
INSERT INTO gallery_image_primary_categories (image_id, category_id) VALUES
('01990000-0000-7000-8400-000000000001', '01990000-0000-7000-8100-000000000001'),
('01990000-0000-7000-8400-000000000002', '01990000-0000-7000-8100-000000000001'),
('01990000-0000-7000-8400-000000000003', '01990000-0000-7000-8100-000000000002'),
('01990000-0000-7000-8400-000000000004', '01990000-0000-7000-8100-000000000001');
INSERT INTO gallery_image_facet_assignments (image_id, facet_value_id) VALUES
('01990000-0000-7000-8400-000000000001', '01990000-0000-7000-8300-000000000001'),
('01990000-0000-7000-8400-000000000003', '01990000-0000-7000-8300-000000000001');`); err != nil {
		t.Fatal(err)
	}

	values, err := fixture.Store.RelatedImages(context.Background(), "01990000-0000-7000-8400-000000000001", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(values) != 2 || values[0].ID != "01990000-0000-7000-8400-000000000002" || values[1].ID != "01990000-0000-7000-8400-000000000003" {
		t.Fatalf("unexpected related ranking: %#v", values)
	}
	if len(values[0].Reasons) == 0 || values[0].Reasons[0].Kind != "primary_category" {
		t.Fatalf("same primary category must be explained: %#v", values[0].Reasons)
	}
	if len(values[1].Reasons) == 0 || values[1].Reasons[0].Kind != "facet" {
		t.Fatalf("shared facet must be explained: %#v", values[1].Reasons)
	}
}

func TestPostgreSQL18EditorialCollectionUpdateCoverAndManualOrder(t *testing.T) {
	fixture := newGalleryPG18Fixture(t)
	if _, err := fixture.SQL.Exec(`
INSERT INTO gallery_images (id, asset_id, title, alt_text, width, height, processing_state, review_state, publication_state, safety_state, public_rendition_ready, published_at) VALUES
('01990000-0000-7000-8400-000000000001', '01990000-0000-7000-8500-000000000001', '第一张', '第一张', 1600, 900, 'ready', 'approved', 'published', 'safe', TRUE, NOW()),
('01990000-0000-7000-8400-000000000002', '01990000-0000-7000-8500-000000000002', '第二张', '第二张', 1600, 900, 'ready', 'approved', 'published', 'safe', TRUE, NOW());
INSERT INTO gallery_collections (id, kind, resource_kind, owner_kind, owner_id, visibility, name, description)
VALUES ('01990000-0000-7000-8700-000000000001', 'gallery.editorial', 'gallery.image', 'site', 'gallery', 'private', '旧专题', '旧说明');
INSERT INTO gallery_collection_editorial (collection_id, slug)
VALUES ('01990000-0000-7000-8700-000000000001', 'old-slug');
INSERT INTO gallery_collection_members (collection_id, image_id, manual_position) VALUES
('01990000-0000-7000-8700-000000000001', '01990000-0000-7000-8400-000000000001', 1),
('01990000-0000-7000-8700-000000000001', '01990000-0000-7000-8400-000000000002', 0);`); err != nil {
		t.Fatal(err)
	}

	updated, err := fixture.Store.UpdateEditorialCollection(context.Background(), "01990000-0000-7000-8700-000000000001", model.EditorialCollectionUpdateInput{
		Version: 1, Name: "夜色", Description: "蓝调影像", Slug: "night-colors", Visibility: "public",
		CoverImageID: "01990000-0000-7000-8400-000000000002", SEOTitle: "夜色专题", SEODescription: "夜色图片精选",
	})
	if err != nil {
		t.Fatal(err)
	}
	if updated.Version != 2 || updated.CoverImageID != "01990000-0000-7000-8400-000000000002" || updated.SEOTitle != "夜色专题" {
		t.Fatalf("unexpected updated collection: %#v", updated)
	}
	reordered, err := fixture.Store.ReorderEditorialMembers(context.Background(), updated.ID, updated.Version, []string{
		"01990000-0000-7000-8400-000000000001", "01990000-0000-7000-8400-000000000002",
	})
	if err != nil {
		t.Fatal(err)
	}
	if reordered.Version != 3 {
		t.Fatalf("reorder must bump version: %#v", reordered)
	}
	if _, err := fixture.Store.UpdateEditorialCollection(context.Background(), reordered.ID, model.EditorialCollectionUpdateInput{
		Version: 2, Name: "过期写入", Slug: "stale-write", Visibility: "public",
	}); err == nil {
		t.Fatal("stale expectedVersion must be rejected")
	}
	detail, err := fixture.Store.CollectionDetail(context.Background(), reordered.ID, 1, 12)
	if err != nil {
		t.Fatal(err)
	}
	if len(detail.Images) != 2 || detail.Images[0].ID != "01990000-0000-7000-8400-000000000001" {
		t.Fatalf("manual order was not persisted: %#v", detail.Images)
	}
	if _, err := fixture.SQL.Exec(`
INSERT INTO gallery_images (id, asset_id, title, alt_text, width, height, processing_state, review_state, publication_state, safety_state, public_rendition_ready) VALUES
('01990000-0000-7000-8400-000000000003', '01990000-0000-7000-8500-000000000003', '隐藏成员', '隐藏成员', 1600, 900, 'ready', 'approved', 'hidden', 'safe', TRUE);
INSERT INTO gallery_collection_members (collection_id, image_id, manual_position)
VALUES ('01990000-0000-7000-8700-000000000001', '01990000-0000-7000-8400-000000000003', 2);`); err != nil {
		t.Fatal(err)
	}
	publicDetail, err := fixture.Store.PublicCollection(context.Background(), "night-colors", 1, 12)
	if err != nil {
		t.Fatal(err)
	}
	if publicDetail.ItemCount != 2 || len(publicDetail.Images) != 2 {
		t.Fatalf("hidden members must not affect the public collection: %#v", publicDetail)
	}
}

func TestPostgreSQL18PersonalWorkspaceFiltersAndPaginates(t *testing.T) {
	fixture := newGalleryPG18Fixture(t)
	if _, err := fixture.SQL.Exec(`
INSERT INTO gallery_submissions (id, subject_kind, subject_id, asset_id, title, alt_text, processing_state, review_state, safety_state, outcome, created_at) VALUES
('01990000-0000-7000-8800-000000000001', 'user', 'user-1', '01990000-0000-7000-8900-000000000001', '等待项', '等待项', 'queued', 'pending', 'pending', 'pending', NOW() - INTERVAL '2 hours'),
('01990000-0000-7000-8800-000000000002', 'user', 'user-1', '01990000-0000-7000-8900-000000000002', '已发布项', '已发布项', 'ready', 'approved', 'safe', 'published', NOW() - INTERVAL '1 hour'),
('01990000-0000-7000-8800-000000000003', 'user', 'other-user', '01990000-0000-7000-8900-000000000003', '他人的投稿', '他人的投稿', 'ready', 'approved', 'safe', 'published', NOW());
INSERT INTO gallery_images (id, asset_id, title, alt_text, width, height, processing_state, review_state, publication_state, safety_state, public_rendition_ready, published_at) VALUES
('01990000-0000-7000-8a00-000000000001', '01990000-0000-7000-8b00-000000000001', 'Beta', 'Beta', 1600, 900, 'ready', 'approved', 'published', 'safe', TRUE, NOW()),
('01990000-0000-7000-8a00-000000000002', '01990000-0000-7000-8b00-000000000002', 'Alpha', 'Alpha', 1600, 900, 'ready', 'approved', 'published', 'safe', TRUE, NOW()),
('01990000-0000-7000-8a00-000000000003', '01990000-0000-7000-8b00-000000000003', 'Hidden', 'Hidden', 1600, 900, 'ready', 'approved', 'hidden', 'safe', TRUE, NULL);
INSERT INTO gallery_collections (id, kind, resource_kind, owner_kind, owner_id, visibility, name)
VALUES ('01990000-0000-7000-8c00-000000000001', 'gallery.favorites', 'gallery.image', 'user', 'user-1', 'private', '我的收藏');
INSERT INTO gallery_collection_members (collection_id, image_id, added_at) VALUES
('01990000-0000-7000-8c00-000000000001', '01990000-0000-7000-8a00-000000000001', NOW() - INTERVAL '2 hours'),
('01990000-0000-7000-8c00-000000000001', '01990000-0000-7000-8a00-000000000002', NOW() - INTERVAL '1 hour'),
('01990000-0000-7000-8c00-000000000001', '01990000-0000-7000-8a00-000000000003', NOW());
INSERT INTO gallery_cases (id, image_id, kind, status, reason, description, resolution_note, created_at) VALUES
('01990000-0000-7000-8d00-000000000001', '01990000-0000-7000-8a00-000000000001', 'report', 'open', '版权争议', '等待权利证明', '', NOW() - INTERVAL '2 hours'),
('01990000-0000-7000-8d00-000000000002', '01990000-0000-7000-8a00-000000000002', 'source_correction', 'resolved', '来源需要修正', '原链接失效', '已替换来源', NOW() - INTERVAL '1 hour');`); err != nil {
		t.Fatal(err)
	}

	submissions, total, err := fixture.Store.MySubmissions(context.Background(), model.Subject{Kind: "user", ID: "user-1"}, model.MySubmissionQuery{
		Page: 1, PageSize: 20, Outcome: "published", ProcessingState: "ready", ReviewState: "approved",
	})
	if err != nil {
		t.Fatal(err)
	}
	if total != 1 || len(submissions) != 1 || submissions[0].Title != "已发布项" {
		t.Fatalf("filtered submissions = total %d values %#v", total, submissions)
	}

	adminSubmissions, adminSubmissionTotal, err := fixture.Store.ReviewQueue(context.Background(), model.AdminSubmissionQuery{
		Page: 1, PageSize: 1, SortBy: "createdAt", SortOrder: "asc", ProcessingState: "ready", ReviewState: "approved", Outcome: "published",
	})
	if err != nil {
		t.Fatal(err)
	}
	if adminSubmissionTotal != 2 || len(adminSubmissions) != 1 || adminSubmissions[0].Title != "已发布项" {
		t.Fatalf("admin submission queue = total %d values %#v", adminSubmissionTotal, adminSubmissions)
	}

	adminCases, adminCaseTotal, err := fixture.Store.Cases(context.Background(), model.AdminCaseQuery{
		Page: 1, PageSize: 20, SortBy: "createdAt", SortOrder: "asc", Status: "open", Kind: "report", Search: "权利",
	})
	if err != nil {
		t.Fatal(err)
	}
	if adminCaseTotal != 1 || len(adminCases) != 1 || adminCases[0].Reason != "版权争议" || adminCases[0].UpdatedAt == nil {
		t.Fatalf("admin case queue = total %d values %#v", adminCaseTotal, adminCases)
	}
	caseVersion := adminCases[0].UpdatedAt.Format(time.RFC3339Nano)
	if _, err := fixture.Store.ResolveCase(context.Background(), "operator-1", "01990000-0000-7000-8d00-000000000001", model.CaseResolutionInput{
		ExpectedUpdatedAt: caseVersion, Status: "reviewing",
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := fixture.Store.ResolveCase(context.Background(), "operator-2", "01990000-0000-7000-8d00-000000000001", model.CaseResolutionInput{
		ExpectedUpdatedAt: caseVersion, Status: "resolved", Note: "stale conclusion",
	}); err == nil {
		t.Fatal("stale case resolution must be rejected")
	}

	favorites, err := fixture.Store.FavoritesDetail(context.Background(), "01990000-0000-7000-8c00-000000000001", 1, 1, "title_asc")
	if err != nil {
		t.Fatal(err)
	}
	if favorites.ItemCount != 2 || len(favorites.Images) != 1 || favorites.Images[0].Title != "Alpha" {
		t.Fatalf("favorite page = %#v", favorites)
	}
	second, err := fixture.Store.FavoritesDetail(context.Background(), "01990000-0000-7000-8c00-000000000001", 2, 1, "title_asc")
	if err != nil {
		t.Fatal(err)
	}
	if len(second.Images) != 1 || second.Images[0].Title != "Beta" {
		t.Fatalf("second favorite page = %#v", second)
	}

	adminImages, adminTotal, err := fixture.Store.AdminImages(context.Background(), model.AdminImageQuery{
		Page: 1, PageSize: 20, SortBy: "updatedAt", SortOrder: "desc", PublicationState: "hidden", SafetyState: "safe",
	})
	if err != nil {
		t.Fatal(err)
	}
	if adminTotal != 1 || len(adminImages) != 1 || adminImages[0].Title != "Hidden" {
		t.Fatalf("admin lifecycle filter = total %d values %#v", adminTotal, adminImages)
	}
	if adminImages[0].UpdatedAt == nil {
		t.Fatal("admin image must carry an optimistic concurrency timestamp")
	}
	expectedUpdatedAt := adminImages[0].UpdatedAt.Format(time.RFC3339Nano)
	updated, err := fixture.Store.UpdateAdminImage(context.Background(), "01990000-0000-7000-8a00-000000000003", model.AdminImageUpdateInput{
		ExpectedUpdatedAt: expectedUpdatedAt,
		Title:             "Hidden revised",
		Description:       "curated from the lifecycle workbench",
		AltText:           "Hidden revised",
		SourceURL:         "https://example.com/hidden",
	})
	if err != nil {
		t.Fatal(err)
	}
	if updated.Title != "Hidden revised" || updated.SourceURL != "https://example.com/hidden" {
		t.Fatalf("updated admin image = %#v", updated)
	}
	if _, err := fixture.Store.UpdateAdminImage(context.Background(), "01990000-0000-7000-8a00-000000000003", model.AdminImageUpdateInput{
		ExpectedUpdatedAt: expectedUpdatedAt, Title: "stale write", AltText: "stale write",
	}); err == nil {
		t.Fatal("stale admin image update must be rejected")
	}
}

func TestPostgreSQL18TagKeysetCursorBreaksCaseInsensitiveNameTiesWithoutGaps(t *testing.T) {
	fixture := newGalleryPG18Fixture(t)
	if _, err := fixture.SQL.Exec(`
INSERT INTO gallery_classification_catalogs (id, catalog_key) VALUES ('01990000-0000-7000-8000-000000000001', 'gallery');
INSERT INTO gallery_tags (id, catalog_id, current_name, current_slug) VALUES
('01990000-0000-7000-8600-000000000001', '01990000-0000-7000-8000-000000000001', 'Alpha', 'alpha-upper'),
('01990000-0000-7000-8600-000000000002', '01990000-0000-7000-8000-000000000001', 'alpha', 'alpha-lower'),
('01990000-0000-7000-8600-000000000003', '01990000-0000-7000-8000-000000000001', 'Beta', 'beta'),
('01990000-0000-7000-8600-000000000004', '01990000-0000-7000-8000-000000000001', 'Gamma', 'gamma');`); err != nil {
		t.Fatal(err)
	}

	first, hasMore, err := fixture.Store.ClassificationTags(context.Background(), model.ClassificationTagCursor{}, 2)
	if err != nil {
		t.Fatal(err)
	}
	if !hasMore || len(first) != 2 || first[0].ID != "01990000-0000-7000-8600-000000000001" || first[1].ID != "01990000-0000-7000-8600-000000000002" {
		t.Fatalf("first page = %#v, hasMore = %v", first, hasMore)
	}
	second, hasMore, err := fixture.Store.ClassificationTags(context.Background(), model.ClassificationTagCursor{
		Name: first[1].Name,
		ID:   first[1].ID,
	}, 2)
	if err != nil {
		t.Fatal(err)
	}
	if hasMore || len(second) != 2 || second[0].ID != "01990000-0000-7000-8600-000000000003" || second[1].ID != "01990000-0000-7000-8600-000000000004" {
		t.Fatalf("second page = %#v, hasMore = %v", second, hasMore)
	}
}

func TestPostgreSQL18TagLookupReturnsCanonicalAndMissingInputsInRequestOrder(t *testing.T) {
	fixture := newGalleryPG18Fixture(t)
	if _, err := fixture.SQL.Exec(`
INSERT INTO gallery_classification_catalogs (id, catalog_key) VALUES ('01990000-0000-7000-8000-000000000001', 'gallery');
INSERT INTO gallery_tags (id, catalog_id, current_name, current_slug)
VALUES ('01990000-0000-7000-8600-000000000001', '01990000-0000-7000-8000-000000000001', 'Alpha', 'alpha');
INSERT INTO gallery_tag_lookup_entries (catalog_id, lookup_key, target_tag_id, kind, display_value)
VALUES ('01990000-0000-7000-8000-000000000001', 'alpha', '01990000-0000-7000-8600-000000000001', 'canonical', 'Alpha');`); err != nil {
		t.Fatal(err)
	}

	matches, freshness, err := fixture.Store.ClassificationTagMatches(context.Background(), []classification.TagLookupRequest{
		{LookupKey: "alpha", DisplayValue: "Alpha"},
		{LookupKey: "missing", DisplayValue: "Missing"},
	})
	if err != nil {
		t.Fatal(err)
	}
	want := []classification.TagMatch{
		{LookupKey: "alpha", Kind: classification.TagMatchCanonical, TagID: "01990000-0000-7000-8600-000000000001"},
		{LookupKey: "missing", Kind: classification.TagMatchNotFound},
	}
	if freshness == "" || !reflect.DeepEqual(matches, want) {
		t.Fatalf("matches = %#v, freshness = %q", matches, freshness)
	}
}

func TestPostgreSQL18CollectableAcceptsMultipleImageIDsAsOneArrayParameter(t *testing.T) {
	fixture := newGalleryPG18Fixture(t)
	if _, err := fixture.SQL.Exec(`
INSERT INTO gallery_images (id, asset_id, title, alt_text, width, height, processing_state, review_state, publication_state, safety_state, public_rendition_ready, published_at) VALUES
('01990000-0000-7000-8400-000000000001', '01990000-0000-7000-8500-000000000001', '公开一', '公开一', 1600, 900, 'ready', 'approved', 'published', 'safe', TRUE, NOW()),
('01990000-0000-7000-8400-000000000002', '01990000-0000-7000-8500-000000000002', '公开二', '公开二', 1600, 900, 'ready', 'approved', 'published', 'safe', TRUE, NOW()),
('01990000-0000-7000-8400-000000000003', '01990000-0000-7000-8500-000000000003', '隐藏', '隐藏', 1600, 900, 'ready', 'approved', 'hidden', 'safe', TRUE, NULL);`); err != nil {
		t.Fatal(err)
	}

	got, err := fixture.Store.Collectable(context.Background(), []string{
		"01990000-0000-7000-8400-000000000001",
		"01990000-0000-7000-8400-000000000002",
		"01990000-0000-7000-8400-000000000003",
	})
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]bool{
		"01990000-0000-7000-8400-000000000001": true,
		"01990000-0000-7000-8400-000000000002": true,
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("collectable = %#v, want %#v", got, want)
	}
}

func TestPostgreSQL18TagKeysetPlanUsesCursorIndexAtScale(t *testing.T) {
	fixture := newGalleryPG18Fixture(t)
	if _, err := fixture.SQL.Exec(`INSERT INTO gallery_classification_catalogs (id, catalog_key) VALUES ('01990000-0000-7000-8000-000000000001', 'gallery')`); err != nil {
		t.Fatal(err)
	}
	if _, err := fixture.SQL.Exec(`
INSERT INTO gallery_tags (id, catalog_id, current_name, current_slug)
SELECT seed.id, '01990000-0000-7000-8000-000000000001',
       'Tag ' || LPAD(seed.ordinality::text, 5, '0'),
       'tag-' || LPAD(seed.ordinality::text, 5, '0')
FROM unnest($1::uuid[]) WITH ORDINALITY AS seed(id, ordinality)`, pq.Array(fixtureIdentifiers("tag-plan", 20000))); err != nil {
		t.Fatal(err)
	}
	if _, err := fixture.SQL.Exec(`ANALYZE gallery_tags`); err != nil {
		t.Fatal(err)
	}
	var cursorID string
	if err := fixture.SQL.QueryRow(`SELECT id::text FROM gallery_tags WHERE current_slug = 'tag-10000'`).Scan(&cursorID); err != nil {
		t.Fatal(err)
	}
	rows, err := fixture.SQL.Query(`
EXPLAIN (ANALYZE, BUFFERS, FORMAT TEXT)
SELECT tag.id::text AS id, tag.current_name AS name, tag.current_slug AS slug, tag.status,
       COALESCE(tag.replacement_id::text, '') AS replacement_id,
       ((SELECT COUNT(*) FROM gallery_submission_tag_assignments assignment WHERE assignment.tag_id = tag.id) +
        (SELECT COUNT(*) FROM gallery_image_tag_assignments assignment WHERE assignment.tag_id = tag.id))::bigint AS assignment_count,
       (SELECT COUNT(*) FROM gallery_tag_lookup_entries entry WHERE entry.target_tag_id = tag.id AND entry.kind = 'alias')::bigint AS alias_count
FROM gallery_tags tag
WHERE (LOWER(tag.current_name), tag.id) > (LOWER($1), $2::uuid)
ORDER BY LOWER(tag.current_name), tag.id
LIMIT 41`, "Tag 10000", cursorID)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	var lines []string
	for rows.Next() {
		var line string
		if err := rows.Scan(&line); err != nil {
			t.Fatal(err)
		}
		lines = append(lines, line)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	plan := strings.Join(lines, "\n")
	t.Logf("Tag keyset EXPLAIN baseline:\n%s", plan)
	if !strings.Contains(plan, "gallery_tags_name_cursor_idx") {
		t.Fatalf("Tag keyset plan is not cursor-index backed:\n%s", plan)
	}
}

func TestPostgreSQL18ContextualCountPlanUsesAssignmentIndexesAtScale(t *testing.T) {
	fixture := newGalleryPG18Fixture(t)
	tx, err := fixture.SQL.Begin()
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()
	if _, err := tx.Exec(`
INSERT INTO gallery_classification_catalogs (id, catalog_key) VALUES ('01990000-0000-7000-8000-000000000001', 'gallery');
CREATE TEMP TABLE seed_categories (ordinal integer PRIMARY KEY, id uuid NOT NULL) ON COMMIT DROP;`); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.Exec(`
INSERT INTO seed_categories
SELECT ordinality::integer, id FROM unnest($1::uuid[]) WITH ORDINALITY AS seed(id, ordinality)`,
		pq.Array(fixtureIdentifiers("context-category", 100))); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.Exec(`
INSERT INTO gallery_categories (id, catalog_id, slug, name, status, first_activated_at)
SELECT id, '01990000-0000-7000-8000-000000000001', 'category-' || ordinal, '分类 ' || ordinal, 'active', NOW()
FROM seed_categories;
INSERT INTO gallery_facets (id, catalog_id, slug, name, status, first_activated_at)
VALUES ('01990000-0000-7000-8200-000000000001', '01990000-0000-7000-8000-000000000001', 'scene', '场景', 'active', NOW());
CREATE TEMP TABLE seed_facet_values (ordinal integer PRIMARY KEY, id uuid NOT NULL) ON COMMIT DROP;`); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.Exec(`
INSERT INTO seed_facet_values
SELECT ordinality::integer, id FROM unnest($1::uuid[]) WITH ORDINALITY AS seed(id, ordinality)`,
		pq.Array(fixtureIdentifiers("context-facet-value", 10))); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.Exec(`
INSERT INTO gallery_facet_values (id, catalog_id, facet_id, slug, name, status, first_activated_at)
SELECT id, '01990000-0000-7000-8000-000000000001', '01990000-0000-7000-8200-000000000001',
       'scene-' || ordinal, '场景 ' || ordinal, 'active', NOW()
FROM seed_facet_values;
CREATE TEMP TABLE seed_images (ordinal integer PRIMARY KEY, id uuid NOT NULL, asset_id uuid NOT NULL) ON COMMIT DROP;`); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.Exec(`
INSERT INTO seed_images
SELECT ordinality::integer, id, asset_id
FROM unnest($1::uuid[], $2::uuid[]) WITH ORDINALITY AS seed(id, asset_id, ordinality)`,
		pq.Array(fixtureIdentifiers("context-image", 12000)),
		pq.Array(fixtureIdentifiers("context-asset", 12000))); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.Exec(`
INSERT INTO gallery_images (id, asset_id, title, alt_text, width, height, processing_state, review_state, publication_state, safety_state, public_rendition_ready, published_at)
SELECT id, asset_id, '图片 ' || ordinal, '图片 ' || ordinal, 1600, 900, 'ready', 'approved', 'published', 'safe', TRUE,
       NOW() - ordinal * INTERVAL '1 second'
FROM seed_images;
INSERT INTO gallery_image_category_assignments (image_id, category_id)
SELECT image.id, category.id
FROM seed_images image
JOIN seed_categories category ON category.ordinal = ((image.ordinal - 1) % 100) + 1;
INSERT INTO gallery_image_facet_assignments (image_id, facet_value_id)
SELECT image.id, value.id
FROM seed_images image
JOIN seed_facet_values value ON value.ordinal = ((image.ordinal - 1) % 10) + 1;
ANALYZE gallery_images;
ANALYZE gallery_image_category_assignments;
ANALYZE gallery_image_facet_assignments;`); err != nil {
		t.Fatal(err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}
	categoryOne := fixtureIdentifier("context-category", 1)
	categoryTwo := fixtureIdentifier("context-category", 2)
	facetValue := fixtureIdentifier("context-facet-value", 1)
	payload := fmt.Sprintf(`[{"value_id":%q,"matching_ids":[%q]},{"value_id":%q,"matching_ids":[%q]}]`, categoryOne, categoryOne, categoryTwo, categoryTwo)
	rows, err := fixture.SQL.Query(`
EXPLAIN (ANALYZE, BUFFERS, FORMAT TEXT)
WITH candidates AS (
    SELECT value_id::uuid AS value_id,
           ARRAY(SELECT item::uuid FROM jsonb_array_elements_text(matching_ids) AS elements(item))::uuid[] AS matching_ids
    FROM jsonb_to_recordset($1::jsonb) AS candidate(value_id text, matching_ids jsonb)
)
SELECT candidate.value_id::text AS value_id,
       (
           SELECT COUNT(DISTINCT i.id)::bigint
           FROM gallery_images i
           WHERE i.processing_state = 'ready'
             AND i.review_state IN ('not_required', 'approved')
             AND i.publication_state = 'published'
             AND i.safety_state = 'safe'
             AND i.public_rendition_ready
             AND EXISTS (
                 SELECT 1
                 FROM gallery_image_facet_assignments facet_assignment
                 JOIN gallery_facet_values facet_value ON facet_value.id = facet_assignment.facet_value_id
                 WHERE facet_assignment.image_id = i.id
                   AND facet_value.facet_id = $2::uuid
                   AND facet_assignment.facet_value_id = ANY($3::uuid[])
             )
             AND EXISTS (
                 SELECT 1
                 FROM gallery_image_category_assignments assignment
                 WHERE assignment.image_id = i.id
                   AND assignment.category_id = ANY(candidate.matching_ids)
             )
       ) AS object_count
FROM candidates candidate
ORDER BY candidate.value_id`, payload, "01990000-0000-7000-8200-000000000001", "{"+facetValue+"}")
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	var lines []string
	for rows.Next() {
		var line string
		if err := rows.Scan(&line); err != nil {
			t.Fatal(err)
		}
		lines = append(lines, line)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	plan := strings.Join(lines, "\n")
	t.Logf("Contextual count EXPLAIN baseline:\n%s", plan)
	if !strings.Contains(plan, "gallery_image_facet_filter_idx") {
		t.Fatalf("contextual count plan does not use the Facet reverse index:\n%s", plan)
	}
	if !strings.Contains(plan, "gallery_image_category_filter_idx") && !strings.Contains(plan, "gallery_image_category_assignments_pkey") {
		t.Fatalf("contextual count plan does not use an index-backed Category join:\n%s", plan)
	}
}

func TestPostgreSQL18EventsAggregateDailyAndRankingSnapshotsRefreshOnce(t *testing.T) {
	fixture := newGalleryPG18Fixture(t)
	if _, err := fixture.SQL.Exec(`
INSERT INTO gallery_images (id, asset_id, title, alt_text, width, height, processing_state, review_state, publication_state, safety_state, public_rendition_ready, published_at) VALUES
('01990000-0000-7000-8400-000000000001', '01990000-0000-7000-8500-000000000001', '第一张', '第一张', 1600, 900, 'ready', 'approved', 'published', 'safe', TRUE, NOW() - INTERVAL '1 hour'),
('01990000-0000-7000-8400-000000000002', '01990000-0000-7000-8500-000000000002', '第二张', '第二张', 1600, 900, 'ready', 'approved', 'published', 'safe', TRUE, NOW());`); err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	subject := model.Subject{Kind: "guest", ID: "guest-1"}
	for _, event := range []model.EventInput{
		{Type: "qualified_view", SessionKey: "session-1"},
		{Type: "qualified_view", SessionKey: "session-1"},
		{Type: "favorite"},
		{Type: "unfavorite"},
		{Type: "share"},
	} {
		if err := fixture.Store.RecordEvent(ctx, subject, "01990000-0000-7000-8400-000000000001", event); err != nil {
			t.Fatal(err)
		}
	}
	var views, visitors, favorites, shares int64
	if err := fixture.SQL.QueryRow(`
SELECT qualified_views, unique_visitors, favorites, shares
FROM gallery_image_metrics_daily
WHERE image_id = '01990000-0000-7000-8400-000000000001'
ORDER BY metric_date DESC LIMIT 1`).Scan(&views, &visitors, &favorites, &shares); err != nil {
		t.Fatal(err)
	}
	if views != 2 || visitors != 1 || favorites != 0 || shares != 1 {
		t.Fatalf("aggregated metrics = views %d visitors %d favorites %d shares %d", views, visitors, favorites, shares)
	}

	if _, err := fixture.SQL.Exec(`DELETE FROM gallery_ranking_snapshots`); err != nil {
		t.Fatal(err)
	}
	var wait sync.WaitGroup
	errorsSeen := make(chan error, 8)
	for range 8 {
		wait.Add(1)
		go func() {
			defer wait.Done()
			_, rankingErr := fixture.Store.Ranking(ctx, "trending", "7d", 60)
			errorsSeen <- rankingErr
		}()
	}
	wait.Wait()
	close(errorsSeen)
	for rankingErr := range errorsSeen {
		if rankingErr != nil {
			t.Fatal(rankingErr)
		}
	}
	var snapshotCount int
	if err := fixture.SQL.QueryRow(`SELECT COUNT(*) FROM gallery_ranking_snapshots`).Scan(&snapshotCount); err != nil {
		t.Fatal(err)
	}
	if snapshotCount != 1 {
		t.Fatalf("concurrent cache miss created %d snapshots, want 1", snapshotCount)
	}
	first, err := fixture.Store.Ranking(ctx, "trending", "7d", 60)
	if err != nil {
		t.Fatal(err)
	}
	if first.Generated == nil || len(first.Images) != 2 || first.Images[0].ID != "01990000-0000-7000-8400-000000000001" {
		t.Fatalf("first ranking = %#v", first)
	}
	if _, err := fixture.SQL.Exec(`
INSERT INTO gallery_image_metrics_daily (image_id, metric_date, qualified_views)
VALUES ('01990000-0000-7000-8400-000000000002', CURRENT_DATE, 100)
ON CONFLICT (image_id, metric_date) DO UPDATE SET qualified_views = 100`); err != nil {
		t.Fatal(err)
	}
	cached, err := fixture.Store.Ranking(ctx, "trending", "7d", 60)
	if err != nil {
		t.Fatal(err)
	}
	if cached.Images[0].ID != first.Images[0].ID || !cached.Generated.Equal(first.Generated) {
		t.Fatalf("fresh snapshot was not reused: first %#v cached %#v", first, cached)
	}
	if _, err := fixture.SQL.Exec(`UPDATE gallery_ranking_snapshots SET expires_at = NOW() - INTERVAL '1 second'`); err != nil {
		t.Fatal(err)
	}
	refreshed, err := fixture.Store.Ranking(ctx, "trending", "7d", 60)
	if err != nil {
		t.Fatal(err)
	}
	if len(refreshed.Images) != 2 || refreshed.Images[0].ID != "01990000-0000-7000-8400-000000000002" {
		t.Fatalf("expired snapshot did not refresh: %#v", refreshed)
	}
	if _, err := fixture.SQL.Exec(`UPDATE gallery_images SET publication_state = 'hidden', published_at = NULL WHERE id = '01990000-0000-7000-8400-000000000002'`); err != nil {
		t.Fatal(err)
	}
	withoutHidden, err := fixture.Store.Ranking(ctx, "trending", "7d", 60)
	if err != nil {
		t.Fatal(err)
	}
	if len(withoutHidden.Images) != 1 || withoutHidden.Images[0].ID != "01990000-0000-7000-8400-000000000001" {
		t.Fatalf("cached snapshot exposed an ineligible image: %#v", withoutHidden)
	}
	if _, err := fixture.SQL.Exec(`
UPDATE gallery_images SET publication_state = 'published', published_at = NOW() WHERE id = '01990000-0000-7000-8400-000000000002';
UPDATE gallery_ranking_snapshots SET expires_at = NOW() - INTERVAL '1 second';
CREATE FUNCTION reject_gallery_snapshot_refresh() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    RAISE EXCEPTION 'simulated snapshot writer outage';
END $$;
CREATE TRIGGER reject_gallery_snapshot_refresh
BEFORE INSERT ON gallery_ranking_snapshots
FOR EACH STATEMENT EXECUTE FUNCTION reject_gallery_snapshot_refresh();`); err != nil {
		t.Fatal(err)
	}
	fallback, err := fixture.Store.Ranking(ctx, "trending", "7d", 60)
	if err != nil {
		t.Fatal(err)
	}
	if fallback.Generated == nil || !fallback.Generated.Equal(refreshed.Generated) || len(fallback.Images) != 2 || fallback.Images[0].ID != "01990000-0000-7000-8400-000000000002" {
		t.Fatalf("ranking did not degrade to the latest expired snapshot: %#v", fallback)
	}
}

func TestPostgreSQL18CatalogScalePlansStayWithinBudgetAndOrderFromIndexes(t *testing.T) {
	fixture := newGalleryPG18Fixture(t)
	if _, err := fixture.SQL.Exec(`
INSERT INTO gallery_images (
    id, asset_id, title, description, alt_text, width, height,
    processing_state, review_state, publication_state, safety_state,
    public_rendition_ready, published_at, updated_at
)
SELECT seed.id, seed.asset_id,
       CASE WHEN seed.ordinality % 997 = 0 THEN 'Needle city ' || seed.ordinality ELSE 'Catalog image ' || seed.ordinality END,
       CASE WHEN seed.ordinality % 991 = 0 THEN 'rare needle description' ELSE 'curated visual' END,
       'Catalog image ' || seed.ordinality,
       1600, 900, 'ready', 'approved', 'published', 'safe', TRUE,
       NOW() - seed.ordinality * INTERVAL '1 second', NOW() - seed.ordinality * INTERVAL '1 second'
FROM unnest($1::uuid[], $2::uuid[]) WITH ORDINALITY AS seed(id, asset_id, ordinality)`,
		pq.Array(fixtureIdentifiers("catalog-image", 12000)),
		pq.Array(fixtureIdentifiers("catalog-asset", 12000))); err != nil {
		t.Fatal(err)
	}
	if _, err := fixture.SQL.Exec(`ANALYZE gallery_images`); err != nil {
		t.Fatal(err)
	}

	searchPlan := explainText(t, fixture.SQL, `
SELECT i.id
FROM gallery_images i
WHERE i.processing_state = 'ready'
  AND i.review_state IN ('not_required', 'approved')
  AND i.publication_state = 'published'
  AND i.safety_state = 'safe'
  AND i.public_rendition_ready
  AND (i.title ILIKE '%needle%' OR i.description ILIKE '%needle%')
ORDER BY i.published_at DESC, i.id DESC
LIMIT 40`)
	usesBothTrigramIndexes := strings.Contains(searchPlan, "gallery_images_public_title_search_idx") &&
		strings.Contains(searchPlan, "gallery_images_public_description_search_idx")
	if !usesBothTrigramIndexes && !strings.Contains(searchPlan, "Seq Scan on gallery_images") {
		t.Fatalf("public search uses neither the expected trigram indexes nor the measured small-catalog sequential baseline:\n%s", searchPlan)
	}
	// At 12k rows PostgreSQL may correctly prefer a ~300-buffer sequential scan
	// over two bitmap indexes. Keep that crossover explicit and tightly budgeted;
	// larger catalogs are expected to choose the trigram BitmapOr path.
	assertExecutionBudget(t, "public search", searchPlan, 100)

	titlePlan := explainText(t, fixture.SQL, `
SELECT i.id
FROM gallery_images i
WHERE i.processing_state = 'ready'
  AND i.review_state IN ('not_required', 'approved')
  AND i.publication_state = 'published'
  AND i.safety_state = 'safe'
  AND i.public_rendition_ready
ORDER BY LOWER(i.title), i.id
LIMIT 40`)
	if !strings.Contains(titlePlan, "gallery_images_public_title_idx") {
		t.Fatalf("public title sort is not index backed:\n%s", titlePlan)
	}
	assertExecutionBudget(t, "public title sort", titlePlan, 500)

	adminPlan := explainText(t, fixture.SQL, `
SELECT i.id
FROM gallery_images i
WHERE i.publication_state = 'published'
  AND i.processing_state = 'ready'
  AND i.review_state = 'approved'
  AND i.safety_state = 'safe'
ORDER BY i.updated_at DESC, i.id DESC
LIMIT 60`)
	if !strings.Contains(adminPlan, "gallery_images_admin_lifecycle_idx") {
		t.Fatalf("admin lifecycle query is not index backed:\n%s", adminPlan)
	}
	assertExecutionBudget(t, "admin lifecycle", adminPlan, 500)

	if _, err := fixture.SQL.Exec(`
INSERT INTO gallery_ranking_snapshots (id, ranking_kind, window_key, generated_at, expires_at)
VALUES ('01990000-0000-7000-8e00-000000000001', 'trending', '7d', NOW(), NOW() + INTERVAL '5 minutes');
INSERT INTO gallery_ranking_entries (snapshot_id, image_id, rank, score)
SELECT '01990000-0000-7000-8e00-000000000001', id,
       ROW_NUMBER() OVER (ORDER BY published_at DESC, id DESC),
       1
FROM gallery_images;
ANALYZE gallery_ranking_entries;`); err != nil {
		t.Fatal(err)
	}
	rankingPlan := explainText(t, fixture.SQL, `
SELECT image_id
FROM gallery_ranking_entries
WHERE snapshot_id = '01990000-0000-7000-8e00-000000000001'
ORDER BY rank
LIMIT 60`)
	if !strings.Contains(rankingPlan, "gallery_ranking_entries_snapshot_id_rank_key") {
		t.Fatalf("ranking cache read is not index backed:\n%s", rankingPlan)
	}
	assertExecutionBudget(t, "ranking snapshot", rankingPlan, 500)
}

func explainText(t *testing.T, db *sql.DB, query string) string {
	t.Helper()
	rows, err := db.Query("EXPLAIN (ANALYZE, BUFFERS, FORMAT TEXT) " + query)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	var lines []string
	for rows.Next() {
		var line string
		if err := rows.Scan(&line); err != nil {
			t.Fatal(err)
		}
		lines = append(lines, line)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	plan := strings.Join(lines, "\n")
	t.Log(plan)
	return plan
}

var executionTimePattern = regexp.MustCompile(`Execution Time: ([0-9.]+) ms`)

func assertExecutionBudget(t *testing.T, name, plan string, milliseconds float64) {
	t.Helper()
	match := executionTimePattern.FindStringSubmatch(plan)
	if len(match) != 2 {
		t.Fatalf("%s plan has no execution timing:\n%s", name, plan)
	}
	value, err := strconv.ParseFloat(match[1], 64)
	if err != nil {
		t.Fatal(err)
	}
	if value > milliseconds {
		t.Fatalf("%s execution %.3fms exceeds %.0fms budget:\n%s", name, value, milliseconds, plan)
	}
}

func TestPostgreSQL18ImagePublicationAndHideAreAtomic(t *testing.T) {
	f := newGalleryPG18Fixture(t)
	ctx := context.Background()
	id := "01990000-0000-7000-8a00-000000000091"
	if _, err := f.SQL.Exec(`INSERT INTO gallery_images (id,asset_id,title,alt_text,width,height,processing_state,review_state,publication_state,safety_state,public_rendition_ready) VALUES ($1,'01990000-0000-7000-8b00-000000000091','Publication test','Publication test',1600,900,'ready','approved','draft','safe',FALSE)`, id); err != nil {
		t.Fatal(err)
	}
	current, err := f.Store.AdminImage(ctx, id)
	if err != nil {
		t.Fatal(err)
	}
	input := model.AdminImageUpdateInput{ExpectedUpdatedAt: current.UpdatedAt.Format(time.RFC3339Nano), Title: current.Title, AltText: current.AltText, PublicationState: "published", PublicationReady: true, Operator: "admin-test"}
	published, err := f.Store.UpdateAdminImage(ctx, id, input)
	if err != nil {
		t.Fatal(err)
	}
	if published.PublicationState != "published" || !published.PublicRenditionReady {
		t.Fatal("image was not published")
	}
	public, err := f.Store.Image(ctx, id, "")
	if err != nil || public == nil {
		t.Fatalf("published image missing: %v", err)
	}
	if _, err = f.Store.UpdateAdminImage(ctx, id, input); err == nil {
		t.Fatal("stale publication accepted")
	}
	input.ExpectedUpdatedAt = published.UpdatedAt.Format(time.RFC3339Nano)
	input.PublicationState = "hidden"
	input.PublicationReady = false
	hidden, err := f.Store.UpdateAdminImage(ctx, id, input)
	if err != nil {
		t.Fatal(err)
	}
	public, err = f.Store.Image(ctx, id, "")
	if err != nil || public != nil {
		t.Fatalf("hidden image still public: %v", err)
	}
	var auditCount int
	if err = f.SQL.QueryRow(`SELECT count(*) FROM gallery_cases WHERE image_id=$1 AND operator_sub='admin-test'`, id).Scan(&auditCount); err != nil || auditCount != 1 {
		t.Fatal("missing hide audit", err)
	}
	for _, safety := range []string{"blocked", "uncertain", "pending"} {
		if _, err = f.SQL.Exec(`UPDATE gallery_images SET safety_state=$2 WHERE id=$1`, id, safety); err != nil {
			t.Fatal(err)
		}
		input.ExpectedUpdatedAt = hidden.UpdatedAt.Format(time.RFC3339Nano)
		input.PublicationState = "published"
		input.Title = "Must not persist"
		if _, err = f.Store.UpdateAdminImage(ctx, id, input); err == nil {
			t.Fatal("unsafe publication accepted")
		}
		got, err := f.Store.AdminImage(ctx, id)
		if err != nil || got.Title == input.Title || got.PublicationState != "hidden" {
			t.Fatal("rejected publication partially updated metadata", err)
		}
	}
}
