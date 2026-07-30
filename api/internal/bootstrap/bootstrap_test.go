package bootstrap

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"
)

type recordingExecer struct {
	statement string
	err       error
}

func (executor *recordingExecer) ExecContext(
	_ context.Context,
	statement string,
	_ ...any,
) (sql.Result, error) {
	executor.statement = statement
	return nil, executor.err
}

func TestApplyUsesInsertOnlyBootstrap(t *testing.T) {
	executor := &recordingExecer{}
	if err := Apply(context.Background(), executor); err != nil {
		t.Fatal(err)
	}
	for _, required := range []string{
		"gallery_site_settings",
		"gallery_home_sections",
		"gallery_classification_catalogs",
		"gallery_classification_policy_profiles",
	} {
		if !strings.Contains(executor.statement, required) {
			t.Fatalf("bootstrap SQL is missing %s", required)
		}
	}
	if strings.Contains(executor.statement, "DO UPDATE") {
		t.Fatal("bootstrap SQL must not overwrite operator changes")
	}
	if strings.Contains(strings.ToLower(executor.statement), "insert into assets") {
		t.Fatal("bootstrap SQL crosses the Asset database boundary")
	}
}

func TestApplyWrapsExecutionFailure(t *testing.T) {
	executor := &recordingExecer{err: errors.New("database unavailable")}
	if err := Apply(context.Background(), executor); err == nil ||
		!strings.Contains(err.Error(), "database unavailable") {
		t.Fatalf("Apply() error = %v", err)
	}
}
