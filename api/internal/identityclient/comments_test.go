package identityclient

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCommentsResolvePublicProfilesInOneBatch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/api/v1/users" || request.URL.Query().Get("ids") != "TestA123,TestB234" {
			t.Fatalf("request = %s", request.URL.RequestURI())
		}
		writer.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(writer).Encode(map[string]any{"users": []map[string]any{
			{"userKey": "TestA123", "displayName": "测试用户", "avatar": map[string]any{"mediaKey": "identity/avatar"}},
			{"userKey": "TestB234", "handle": "reader"},
		}})
	}))
	defer server.Close()

	profiles := NewComments(server.URL, server.Client()).Resolve(context.Background(), []string{"TestA123", "TestA123", "", "TestB234"})
	if len(profiles) != 2 || profiles["TestA123"].DisplayName != "测试用户" || profiles["TestA123"].AvatarMediaKey != "identity/avatar" || profiles["TestB234"].DisplayName != "reader" {
		t.Fatalf("profiles = %#v", profiles)
	}
}
