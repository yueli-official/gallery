// Command asset-references performs an explicitly requested maintenance cutover.
// Normal API startup never calls this initializer.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/yueli-official/gallery/api/internal/assetreferences"
	"github.com/yueli-official/gallery/api/internal/runtime"
	"os"
)

func main() {
	initialize := flag.Bool("initialize", false, "initialize current business objects after stopping the old writer")
	archive := flag.String("archive-retired", "", "archive confirmed retired indexes for an initialization run even though Work history stays")
	actor := flag.String("actor", "", "trusted maintenance operator recorded with this run")
	flag.Parse()
	if *initialize == (*archive != "") || *actor == "" {
		fmt.Fprintln(os.Stderr, "choose -initialize or -archive-retired RUN_ID, and supply -actor")
		os.Exit(2)
	}
	if err := runtime.EnableEnvironmentConfig(); err != nil {
		panic(err)
	}
	ctx := context.Background()
	db, err := runtime.OpenDefaultPostgres(ctx)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	module, _, err := assetreferences.New(ctx, db)
	if err != nil {
		panic(err)
	}
	if *archive != "" {
		count, err := module.ArchiveRetired(ctx, *archive)
		if err != nil {
			panic(err)
		}
		if err = json.NewEncoder(os.Stdout).Encode(map[string]any{"initializationId": *archive, "actor": *actor, "archived": count}); err != nil {
			panic(err)
		}
		return
	}
	source, err := assetreferences.InitializationSource(g.Cfg().MustGet(ctx, "gallery.siteUrl").String(), os.Getenv("ASSET_PUBLIC_ORIGIN"))
	if err != nil {
		panic(err)
	}
	run, err := module.Initialize(ctx, *actor, source)
	if err != nil {
		panic(err)
	}
	if err = json.NewEncoder(os.Stdout).Encode(run); err != nil {
		panic(err)
	}
}
