package gallerycomments

import (
	"context"
	"testing"

	galleryservice "github.com/yueli-official/gallery/api/internal/gallery"
)

const (
	imageID  = "019b1000-0000-7000-9000-000000000001"
	topID    = "019b5000-0000-7000-9000-000000000001"
	replyID  = "019b5000-0000-7000-9000-000000000002"
	nestedID = "019b5000-0000-7000-9000-000000000003"
)

type fakeStore struct {
	publicImage *ImageHead
	comments    map[string]Comment
	inserted    []Comment
	status      Status
	deleted     string
	ascending   bool
}

func newFakeStore() *fakeStore {
	return &fakeStore{
		publicImage: &ImageHead{ID: imageID, Title: "海岸晨光"},
		comments: map[string]Comment{
			topID:    {ID: topID, ImageID: imageID, Content: "顶层", Status: StatusApproved},
			replyID:  {ID: replyID, ImageID: imageID, ParentID: topID, Content: "回复", Status: StatusApproved},
			nestedID: {ID: nestedID, ImageID: imageID, ParentID: topID, Content: "旧回复", Status: StatusApproved},
		},
	}
}

func (store *fakeStore) PublicImage(context.Context, string) (*ImageHead, error) {
	return store.publicImage, nil
}
func (store *fakeStore) Comment(_ context.Context, id string) (*Comment, error) {
	comment, ok := store.comments[id]
	if !ok {
		return nil, nil
	}
	return &comment, nil
}
func (store *fakeStore) InsertComment(_ context.Context, comment Comment) error {
	store.inserted = append(store.inserted, comment)
	store.comments[comment.ID] = comment
	return nil
}
func (store *fakeStore) PublicComments(_ context.Context, _ string, ascending bool, _, _ int) ([]Comment, int, error) {
	store.ascending = ascending
	return []Comment{store.comments[topID]}, 1, nil
}
func (store *fakeStore) Replies(context.Context, []string) (map[string][]Comment, error) {
	return map[string][]Comment{topID: {store.comments[replyID]}}, nil
}
func (store *fakeStore) AdminComments(context.Context, AdminQuery) ([]AdminComment, int, error) {
	return []AdminComment{{Comment: store.comments[topID], ImageTitle: "海岸晨光"}}, 1, nil
}
func (store *fakeStore) SetStatus(_ context.Context, _ string, status Status) (bool, error) {
	store.status = status
	return true, nil
}
func (store *fakeStore) SoftDelete(_ context.Context, id string) (bool, error) {
	store.deleted = id
	return true, nil
}

func TestCreateCommentProjectsMemberAndAnonymousStates(t *testing.T) {
	store := newFakeStore()
	module := New(store, nil)

	member, pending, err := module.Create(context.Background(), Actor{UserKey: "TestA123"}, imageID, CreateInput{Content: "  很喜欢这张图  "})
	if err != nil {
		t.Fatal(err)
	}
	if pending || member.Status != StatusApproved || member.UserKey != "TestA123" || member.Content != "很喜欢这张图" {
		t.Fatalf("member comment = %#v, pending=%v", member, pending)
	}

	anonymous, pending, err := module.Create(context.Background(), Actor{}, imageID, CreateInput{Content: "匿名留言", AuthorName: "  访客  "})
	if err != nil {
		t.Fatal(err)
	}
	if !pending || anonymous.Status != StatusPending || anonymous.AuthorName != "访客" {
		t.Fatalf("anonymous comment = %#v, pending=%v", anonymous, pending)
	}
}

func TestReplyToReplyFlattensToTopLevelThread(t *testing.T) {
	store := newFakeStore()
	module := New(store, nil)

	created, _, err := module.Create(context.Background(), Actor{UserKey: "TestA123"}, imageID, CreateInput{
		Content: "继续回复", ParentID: replyID,
	})
	if err != nil {
		t.Fatal(err)
	}
	wantParent := galleryservice.PublicID(topID)
	if created.ParentID != wantParent {
		t.Fatalf("parent = %q, want top-level %q", created.ParentID, wantParent)
	}
}

func TestListPublicBuildsTwoLevelThreads(t *testing.T) {
	store := newFakeStore()
	module := New(store, nil)
	page, err := module.ListPublic(context.Background(), imageID, true, 1, 20)
	if err != nil {
		t.Fatal(err)
	}
	if page.Total != 1 || len(page.Items) != 1 || len(page.Items[0].Replies) != 1 {
		t.Fatalf("page = %#v", page)
	}
	if !store.ascending {
		t.Fatal("public sort order was not passed to the persistence adapter")
	}
	if _, err := module.ListPublic(context.Background(), imageID, false, 1, 20); err != nil {
		t.Fatal(err)
	}
	if store.ascending {
		t.Fatal("descending public sort order was not passed to the persistence adapter")
	}
}

func TestCommentValidationAndModeration(t *testing.T) {
	store := newFakeStore()
	module := New(store, nil)

	if _, _, err := module.Create(context.Background(), Actor{}, imageID, CreateInput{Content: "匿名但没名字"}); err == nil {
		t.Fatal("anonymous comment without a name must fail")
	}
	store.publicImage = nil
	if _, _, err := module.Create(context.Background(), Actor{UserKey: "TestA123"}, imageID, CreateInput{Content: "隐藏图片"}); err == nil {
		t.Fatal("non-public image must fail")
	}
	store.publicImage = &ImageHead{ID: imageID, Title: "海岸晨光"}
	if _, err := module.SetStatus(context.Background(), topID, StatusPending); err == nil {
		t.Fatal("moderation cannot move a comment back to pending")
	}
	if _, err := module.SetStatus(context.Background(), topID, StatusSpam); err != nil {
		t.Fatal(err)
	}
	if store.status != StatusSpam {
		t.Fatalf("status = %q", store.status)
	}
	if err := module.Delete(context.Background(), topID); err != nil {
		t.Fatal(err)
	}
	if store.deleted != topID {
		t.Fatalf("deleted = %q", store.deleted)
	}
}
