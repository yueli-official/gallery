package gallery

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/yueli-official/foundation/go/identifier"
)

func PublicID(value string) string {
	id, err := identifier.Parse(strings.TrimSpace(value))
	if err != nil {
		return value
	}
	return base64.RawURLEncoding.EncodeToString(id[:])
}

func DatabaseID(value string) (string, error) {
	value = strings.TrimSpace(value)
	if id, err := identifier.Parse(value); err == nil {
		return id.String(), nil
	}
	bytes, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil || len(bytes) != 16 {
		return "", fmt.Errorf("invalid public UUID")
	}
	id := identifier.UUID([16]byte(bytes))
	if _, err := identifier.Parse(id.String()); err != nil {
		return "", fmt.Errorf("invalid public UUID: %w", err)
	}
	return id.String(), nil
}
