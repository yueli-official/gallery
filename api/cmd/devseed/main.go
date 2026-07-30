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
	"github.com/yueli-official/gallery/api/internal/bootstrap"
)

//go:embed sql/content.sql
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
	if err := bootstrap.Apply(ctx, transaction); err != nil {
		fatal("%v", err)
	}
	statement, err := seedFiles.ReadFile("sql/content.sql")
	if err != nil {
		fatal("read sql/content.sql: %v", err)
	}
	if _, err := transaction.ExecContext(ctx, string(statement)); err != nil {
		fatal("execute sql/content.sql: %v", err)
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
