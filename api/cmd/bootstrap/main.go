// Command bootstrap installs the initial Gallery-owned site records after
// schema migrations. It is idempotent and never overwrites operator changes.
package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"

	_ "github.com/lib/pq"
	"github.com/yueli-official/gallery/api/internal/bootstrap"
)

func main() {
	databaseURL := strings.TrimSpace(os.Getenv("GALLERY_DATABASE_URL"))
	if databaseURL == "" {
		fail("GALLERY_DATABASE_URL is required")
	}
	database, err := sql.Open("postgres", databaseURL)
	if err != nil {
		fail("open database: %v", err)
	}
	defer database.Close()
	ctx := context.Background()
	if err := database.PingContext(ctx); err != nil {
		fail("connect database: %v", err)
	}
	if err := bootstrap.Reconcile(ctx, database); err != nil {
		fail("%v", err)
	}
	fmt.Println("Gallery initial site records are ready")
}

func fail(format string, arguments ...any) {
	fmt.Fprintf(os.Stderr, "bootstrap: "+format+"\n", arguments...)
	os.Exit(1)
}
