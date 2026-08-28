// Package identityclient adapts Identity public profiles to Gallery presentation seams.
package identityclient

import (
	"context"
	"net/http"
	"net/url"
	"strings"

	foundationhttpclient "github.com/yueli-official/foundation/go/httpclient"
	"github.com/yueli-official/gallery/api/internal/gallerycomments"
)

type publicUser struct {
	UserKey     string `json:"userKey"`
	DisplayName string `json:"displayName"`
	Handle      string `json:"handle"`
	Avatar      *struct {
		MediaKey string `json:"mediaKey"`
	} `json:"avatar"`
}

type Comments struct {
	base   string
	client *http.Client
}

func NewComments(baseURL string, client *http.Client) *Comments {
	if client == nil {
		client = http.DefaultClient
	}
	return &Comments{base: strings.TrimRight(baseURL, "/"), client: client}
}

func (client *Comments) Resolve(ctx context.Context, userKeys []string) map[string]gallerycomments.PublicProfile {
	result := map[string]gallerycomments.PublicProfile{}
	unique := make([]string, 0, len(userKeys))
	seen := map[string]bool{}
	for _, userKey := range userKeys {
		userKey = strings.TrimSpace(userKey)
		if userKey != "" && !seen[userKey] {
			seen[userKey] = true
			unique = append(unique, userKey)
		}
	}
	if len(unique) == 0 || client.base == "" {
		return result
	}
	query := url.Values{"ids": {strings.Join(unique, ",")}}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, client.base+"/api/v1/users?"+query.Encode(), nil)
	if err != nil {
		return result
	}
	response, err := client.client.Do(request)
	if err != nil {
		return result
	}
	defer response.Body.Close()
	payload, err := foundationhttpclient.DecodeJSON[struct {
		Users []publicUser `json:"users"`
	}](response, foundationhttpclient.Limits{})
	if err != nil {
		return result
	}
	for _, user := range payload.Users {
		if user.UserKey == "" {
			continue
		}
		name := strings.TrimSpace(user.DisplayName)
		if name == "" {
			name = strings.TrimSpace(user.Handle)
		}
		profile := gallerycomments.PublicProfile{UserKey: user.UserKey, DisplayName: name}
		if user.Avatar != nil {
			profile.AvatarMediaKey = user.Avatar.MediaKey
		}
		result[user.UserKey] = profile
	}
	return result
}
