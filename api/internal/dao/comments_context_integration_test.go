package dao_test

import (
	"context"
	"strings"
	"testing"

	"github.com/yueli-official/gallery/api/internal/gallerycomments"
)

func TestAdminCommentContextSurvivesFiltering(t *testing.T) {
	f := newGalleryPG18Fixture(t)
	const imageID = "01990000-0000-7000-8a00-000000000091"
	const parentID = "01990000-0000-7000-8c00-000000000091"
	const replyID = "01990000-0000-7000-8c00-000000000092"
	if _, err := f.SQL.Exec(`INSERT INTO gallery_images (id,asset_id,title,alt_text,width,height,processing_state,review_state,publication_state,safety_state,public_rendition_ready) VALUES ($1,'01990000-0000-7000-8b00-000000000091','Context test','Context test',1600,900,'ready','approved','draft','safe',TRUE)`, imageID); err != nil {
		t.Fatal(err)
	}
	if _, err := f.SQL.Exec(`INSERT INTO gallery_comments (id,image_id,author_name,content,status) VALUES ($1,$2,'原作者',$3,'approved')`, parentID, imageID, strings.Repeat("原", 180)); err != nil {
		t.Fatal(err)
	}
	if _, err := f.SQL.Exec(`INSERT INTO gallery_comments (id,image_id,parent_id,author_name,content,status) VALUES ($1,$2,$3,'回复者','唯一回复搜索词','pending')`, replyID, imageID, parentID); err != nil {
		t.Fatal(err)
	}
	query := gallerycomments.AdminQuery{Keyword: "唯一回复搜索词", Status: gallerycomments.StatusPending, Page: 1, Size: 1}
	items, total, err := f.Store.AdminComments(context.Background(), query)
	if err != nil || total != 1 || len(items) != 1 {
		t.Fatalf("filtered reply: total=%d items=%v err=%v", total, items, err)
	}
	if items[0].ParentAuthorName != "原作者" || items[0].ParentContent != strings.Repeat("原", 160) {
		t.Fatalf("missing or unbounded parent context: %+v", items[0])
	}
	if _, err := f.SQL.Exec(`UPDATE gallery_comments SET deleted_at=NOW() WHERE id=$1`, parentID); err != nil {
		t.Fatal(err)
	}
	items, total, err = f.Store.AdminComments(context.Background(), query)
	if err != nil || total != 1 || len(items) != 1 {
		t.Fatalf("deleted parent must retain reply: total=%d err=%v", total, err)
	}
	if items[0].ParentAuthorName != "" || items[0].ParentContent != "" {
		t.Fatal("deleted parent context must not be returned")
	}
}
