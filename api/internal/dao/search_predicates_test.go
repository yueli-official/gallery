package dao

import (
	"strings"
	"testing"

	"github.com/yueli-official/foundation/go/classification"
	"github.com/yueli-official/gallery/api/internal/model"
)

func TestPublicImagePredicatesSearchesContentAndTags(t *testing.T) {
	where, args := publicImagePredicates(
		model.ImageQuery{Search: "海岸"},
		classification.FilterPlan{},
	)

	clause := strings.Join(where, "\n")
	for _, fragment := range []string{
		"i.title ILIKE ?",
		"i.description ILIKE ?",
		"i.alt_text ILIKE ?",
		"search_tag.current_name ILIKE ?",
		"search_tag.current_slug ILIKE ?",
	} {
		if !strings.Contains(clause, fragment) {
			t.Fatalf("search clause does not contain %q:\n%s", fragment, clause)
		}
	}
	if len(args) != 5 {
		t.Fatalf("search args = %d, want 5", len(args))
	}
	for index, value := range args {
		if value != "%海岸%" {
			t.Fatalf("search arg %d = %#v, want %%海岸%%", index, value)
		}
	}
}

func TestPublicImagePredicatesAcceptsTagSlugOrDisplayName(t *testing.T) {
	where, args := publicImagePredicates(
		model.ImageQuery{Tag: "晨光"},
		classification.FilterPlan{},
	)
	clause := strings.Join(where, "\n")
	if !strings.Contains(clause, "t.current_slug = ? OR LOWER(t.current_name) = LOWER(?)") {
		t.Fatalf("tag clause does not accept slug or display name:\n%s", clause)
	}
	if len(args) != 2 || args[0] != "晨光" || args[1] != "晨光" {
		t.Fatalf("tag args = %#v", args)
	}
}
