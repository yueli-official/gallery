package gallerycomments

import (
	"os"
	"strings"
	"testing"
)

func TestCommentsMigrationOwnsImageThreadsAndModerationState(t *testing.T) {
	content, err := os.ReadFile("../../manifest/sql/migrations/0010_comments.up.sql")
	if err != nil {
		t.Fatal(err)
	}
	text := string(content)
	for _, expected := range []string{
		"CREATE TABLE gallery_comments",
		"REFERENCES gallery_images(id) ON DELETE CASCADE",
		"REFERENCES gallery_comments(image_id, id) ON DELETE CASCADE",
		"pending", "approved", "spam", "trash",
	} {
		if !strings.Contains(text, expected) {
			t.Fatalf("comments migration is missing %q", expected)
		}
	}
}
