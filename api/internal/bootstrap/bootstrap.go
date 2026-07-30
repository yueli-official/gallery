// Package bootstrap installs the initial Gallery-owned site and classification
// records without overwriting later operator changes.
package bootstrap

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
)

//go:embed site.sql
var siteSQL string

type Execer interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
}

func Apply(ctx context.Context, executor Execer) error {
	if executor == nil {
		return fmt.Errorf("gallery bootstrap executor is nil")
	}
	if _, err := executor.ExecContext(ctx, siteSQL); err != nil {
		return fmt.Errorf("apply Gallery site bootstrap: %w", err)
	}
	return nil
}

func Reconcile(ctx context.Context, database *sql.DB) error {
	if database == nil {
		return fmt.Errorf("gallery bootstrap database is nil")
	}
	transaction, err := database.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin Gallery bootstrap: %w", err)
	}
	defer transaction.Rollback()
	if _, err := transaction.ExecContext(
		ctx,
		`SELECT pg_advisory_xact_lock(hashtext('gallery:bootstrap'))`,
	); err != nil {
		return fmt.Errorf("lock Gallery bootstrap: %w", err)
	}
	if err := Apply(ctx, transaction); err != nil {
		return err
	}
	if err := transaction.Commit(); err != nil {
		return fmt.Errorf("commit Gallery bootstrap: %w", err)
	}
	return nil
}
