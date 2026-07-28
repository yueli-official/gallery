// Command devseed reconciles Gallery-owned local development fixtures.
package main

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"os"
	"strings"

	_ "github.com/lib/pq"
)

//go:embed sql/*.sql
var seedFiles embed.FS

func main() {
	databaseURL := strings.TrimSpace(os.Getenv("GALLERY_DATABASE_URL"))
	if databaseURL == "" {
		fatal("GALLERY_DATABASE_URL is required")
	}
	database, err := sql.Open("postgres", databaseURL)
	if err != nil {
		fatal("open database: %v", err)
	}
	defer database.Close()

	ctx := context.Background()
	if err := database.PingContext(ctx); err != nil {
		fatal("connect database: %v", err)
	}
	transaction, err := database.BeginTx(ctx, nil)
	if err != nil {
		fatal("begin transaction: %v", err)
	}
	defer transaction.Rollback()
	if _, err := transaction.ExecContext(ctx, `SELECT pg_advisory_xact_lock(hashtext('gallery:devseed'))`); err != nil {
		fatal("lock development seed: %v", err)
	}
	for _, path := range []string{"sql/site.sql", "sql/content.sql"} {
		statement, err := seedFiles.ReadFile(path)
		if err != nil {
			fatal("read %s: %v", path, err)
		}
		if _, err := transaction.ExecContext(ctx, string(statement)); err != nil {
			fatal("execute %s: %v", path, err)
		}
	}
	if err := transaction.Commit(); err != nil {
		fatal("commit development seed: %v", err)
	}
	fmt.Println("Gallery 开发夹具已对账：站点分类、128 张图片、48 条投稿、20 个处理单和 4 个专题")
}

func fatal(format string, arguments ...any) {
	fmt.Fprintf(os.Stderr, "devseed: "+format+"\n", arguments...)
	os.Exit(1)
}
