// Command imagekeys completes migration 0024 using the shared Identifier module.
package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq"
	"github.com/yueli-official/gallery/api/internal/imagekeys"
)

func main() {
	dsn := os.Getenv("GALLERY_DATABASE_URL")
	if dsn == "" {
		fmt.Fprintln(os.Stderr, "GALLERY_DATABASE_URL is required")
		os.Exit(1)
	}
	db, err := sql.Open("postgres", dsn)
	if err == nil {
		defer db.Close()
		var n int
		n, err = imagekeys.Backfill(context.Background(), db)
		if err == nil {
			fmt.Printf("Image public keys ready; backfilled %d images\n", n)
			return
		}
	}
	fmt.Fprintln(os.Stderr, "Image public key migration failed:", err)
	os.Exit(1)
}
