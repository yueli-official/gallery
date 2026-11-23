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
	"github.com/yueli-official/foundation/go/identifier"
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

	if _, err := transaction.ExecContext(ctx, `LOCK TABLE gallery_images IN SHARE ROW EXCLUSIVE MODE; CREATE TEMP TABLE gallery_seed_image_keys(id UUID PRIMARY KEY, public_id TEXT UNIQUE NOT NULL) ON COMMIT DROP`); err != nil {
		fatal("prepare image keys: %v", err)
	}
	for i := 1; i <= 128; i++ {
		id := fmt.Sprintf("019b1000-0000-7000-9000-%012d", i)
		var existing sql.NullString
		err := transaction.QueryRowContext(ctx, `SELECT public_id FROM gallery_images WHERE id=$1::uuid`, id).Scan(&existing)
		if err != nil && err != sql.ErrNoRows {
			fatal("read image key: %v", err)
		}
		if existing.Valid {
			if _, err := transaction.ExecContext(ctx, `INSERT INTO gallery_seed_image_keys VALUES($1::uuid,$2)`, id, existing.String); err != nil {
				fatal("preserve image key: %v", err)
			}
			continue
		}
		_, err = identifier.Allocate(ctx, identifier.CompactURLV1, func(ctx context.Context, key identifier.Key) (identifier.ClaimResult, error) {
			result, err := transaction.ExecContext(ctx, `INSERT INTO gallery_seed_image_keys SELECT $1::uuid,$2 WHERE NOT EXISTS(SELECT 1 FROM gallery_images WHERE public_id=$2) ON CONFLICT(public_id) DO NOTHING`, id, key.String())
			if err != nil {
				return 0, err
			}
			n, err := result.RowsAffected()
			if err != nil {
				return 0, err
			}
			if n == 0 {
				return identifier.Collision, nil
			}
			return identifier.Claimed, nil
		})
		if err != nil {
			fatal("allocate image key: %v", err)
		}
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
	fmt.Println("Gallery 开发夹具已对账：站点分类、128 张图片、6 条评论、48 条投稿、20 个处理单和 4 个专题")
}

func fatal(format string, arguments ...any) {
	fmt.Fprintf(os.Stderr, "devseed: "+format+"\n", arguments...)
	os.Exit(1)
}
