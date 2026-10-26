package assetreferences

import (
	"context"
	"database/sql"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/yueli-official/asset/referencesync"
	"github.com/yueli-official/foundation/go/work"
	"github.com/yueli-official/gallery/api/internal/appconfig"
	"github.com/yueli-official/gallery/api/internal/assetclient"
	"os"
	"time"
)

func New(ctx context.Context, db *sql.DB) (*referencesync.Module, *work.Runner, error) {
	cfg := appconfig.LoadAssetClient(ctx)
	client, err := (referencesync.Config{BaseURL: cfg.BaseURL, TokenURL: cfg.TokenURL, ClientID: cfg.ClientID, ClientSecret: cfg.ClientSecret}).NewClient(150 * time.Millisecond)
	if err != nil {
		return nil, nil, err
	}
	port, err := assetclient.NewHTTP(assetclient.Config{BaseURL: cfg.BaseURL, TokenURL: cfg.TokenURL, ClientID: cfg.ClientID, ClientSecret: cfg.ClientSecret, Scope: cfg.Scope, SiteKey: appconfig.AssetNamespace(ctx), HTTPClient: client.HTTP})
	if err != nil {
		return nil, nil, err
	}
	module, runner, err := referencesync.NewPostgres(ctx, db, "gallery:asset-references", client, work.RunnerOptions{OnError: func(err error) { g.Log().Error(ctx, "asset reference worker:", err) }})
	if err != nil {
		return nil, nil, err
	}
	module.Source, err = ObjectSource(g.Cfg().MustGet(ctx, "gallery.siteUrl").String(), os.Getenv("ASSET_PUBLIC_ORIGIN"))
	module.AfterApply = PublicationHook(db, port)
	return module, runner, err
}
