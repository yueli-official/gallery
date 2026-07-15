package classificationwatcher

import (
	"net/url"
	"testing"

	"github.com/gogf/gf/v2/database/gdb"
)

func TestConnectionURLPreservesEscapedCredentialsAndExtraTLSOptions(t *testing.T) {
	raw, err := connectionURL(&gdb.ConfigNode{
		Host: "db.internal", Port: "5544", User: "gallery user", Pass: "p@ss:word", Name: "gallery_main",
		Extra: "sslmode=verify-full&sslrootcert=%2Fcerts%2Froot.pem",
	})
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := url.Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	password, _ := parsed.User.Password()
	if parsed.User.Username() != "gallery user" || password != "p@ss:word" {
		t.Fatalf("credentials were not preserved: %q", raw)
	}
	if parsed.Host != "db.internal:5544" || parsed.Path != "/gallery_main" {
		t.Fatalf("endpoint = %q %q", parsed.Host, parsed.Path)
	}
	if parsed.Query().Get("sslmode") != "verify-full" || parsed.Query().Get("sslrootcert") != "/certs/root.pem" {
		t.Fatalf("query = %#v", parsed.Query())
	}
}

func TestRelevantPayloadOnlyAcceptsGalleryRevisionNotifications(t *testing.T) {
	for _, payload := range []string{"gallery:2", "gallery:18446744073709551615"} {
		if !relevantPayload(payload) {
			t.Fatalf("expected %q to be relevant", payload)
		}
	}
	for _, payload := range []string{"", "gallery", "gallery:0", "gallery:-1", "gallery:next", "blog:2"} {
		if relevantPayload(payload) {
			t.Fatalf("expected %q to be ignored", payload)
		}
	}
}
