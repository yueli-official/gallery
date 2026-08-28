package dao_test

import (
	"context"
	"testing"
	"time"

	"github.com/yueli-official/gallery/api/internal/gallerycomments"
)

func TestPostgreSQL18CommentThreadsAndModeration(t *testing.T) {
	fixture := newGalleryPG18Fixture(t)
	ctx := context.Background()
	const imageID = "01990000-0000-7000-8e00-000000000001"
	if _, err := fixture.SQL.Exec(`
INSERT INTO gallery_images (
  id, asset_id, title, alt_text, width, height, processing_state,
  review_state, publication_state, safety_state, public_rendition_ready, published_at
) VALUES (
  $1, '01990000-0000-7000-8f00-000000000001', '评论验收图片', '评论验收图片', 1200, 800,
  'ready', 'approved', 'published', 'safe', TRUE, NOW()
)`, imageID); err != nil {
		t.Fatal(err)
	}
	top := gallerycomments.Comment{
		ID: "01990000-0000-7000-8e00-000000000002", ImageID: imageID,
		UserKey: "TestA123", Content: "顶层评论", Status: gallerycomments.StatusApproved,
		CreatedAt: time.Date(2026, 7, 11, 8, 0, 0, 0, time.UTC),
	}
	reply := gallerycomments.Comment{
		ID: "01990000-0000-7000-8e00-000000000003", ImageID: imageID, ParentID: top.ID,
		AuthorName: "访客", Content: "回复", Status: gallerycomments.StatusApproved,
		CreatedAt: time.Date(2026, 7, 11, 8, 5, 0, 0, time.UTC),
	}
	if err := fixture.Store.InsertComment(ctx, top); err != nil {
		t.Fatal(err)
	}
	if err := fixture.Store.InsertComment(ctx, reply); err != nil {
		t.Fatal(err)
	}
	comments, total, err := fixture.Store.PublicComments(ctx, imageID, true, 20, 0)
	if err != nil || total != 1 || len(comments) != 1 {
		t.Fatalf("public comments = %#v total=%d err=%v", comments, total, err)
	}
	replies, err := fixture.Store.Replies(ctx, []string{top.ID})
	if err != nil || len(replies[top.ID]) != 1 {
		t.Fatalf("replies = %#v err=%v", replies, err)
	}
	admin, adminTotal, err := fixture.Store.AdminComments(ctx, gallerycomments.AdminQuery{
		Status: gallerycomments.StatusApproved, Keyword: "验收图片", Page: 1, Size: 20,
	})
	if err != nil || adminTotal != 2 || len(admin) != 2 || admin[0].ImageTitle != "评论验收图片" {
		t.Fatalf("admin comments = %#v total=%d err=%v", admin, adminTotal, err)
	}
	if changed, err := fixture.Store.SetStatus(ctx, top.ID, gallerycomments.StatusSpam); err != nil || !changed {
		t.Fatalf("set status changed=%v err=%v", changed, err)
	}
	if deleted, err := fixture.Store.SoftDelete(ctx, top.ID); err != nil || !deleted {
		t.Fatalf("delete changed=%v err=%v", deleted, err)
	}
	if comment, err := fixture.Store.Comment(ctx, reply.ID); err != nil || comment != nil {
		t.Fatalf("reply after cascade soft delete = %#v err=%v", comment, err)
	}
}
