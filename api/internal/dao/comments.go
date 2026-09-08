package dao

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/yueli-official/gallery/api/internal/gallerycomments"
)

const tComments = "gallery_comments"

var _ gallerycomments.Store = (*PG)(nil)

func (store *PG) PublicImage(ctx context.Context, imageID string) (*gallerycomments.ImageHead, error) {
	var image *gallerycomments.ImageHead
	err := store.db.Model("gallery_images").Ctx(ctx).
		Fields("id, title").
		Where("id", imageID).
		Where("processing_state", "ready").
		WhereIn("review_state", []string{"not_required", "approved"}).
		Where("publication_state", "published").
		Where("safety_state", "safe").
		Where("public_rendition_ready", true).
		Where("deleted_at IS NULL").
		Limit(1).
		Scan(&image)
	return image, err
}

func (store *PG) Comment(ctx context.Context, id string) (*gallerycomments.Comment, error) {
	var comment *gallerycomments.Comment
	err := store.db.Model(tComments).Ctx(ctx).
		Where("id", id).
		Where("deleted_at IS NULL").
		Limit(1).
		Scan(&comment)
	return comment, err
}

func (store *PG) InsertComment(ctx context.Context, comment gallerycomments.Comment) error {
	parent := any(nil)
	if comment.ParentID != "" {
		parent = comment.ParentID
	}
	_, err := store.db.Model(tComments).Ctx(ctx).Data(g.Map{
		"id": comment.ID, "image_id": comment.ImageID, "parent_id": parent,
		"user_key": comment.UserKey, "author_name": comment.AuthorName,
		"author_email": comment.AuthorEmail, "content": comment.Content,
		"status": string(comment.Status), "ip": comment.IP,
		"user_agent": comment.UserAgent, "created_at": comment.CreatedAt,
		"updated_at": comment.CreatedAt,
	}).Insert()
	return err
}

func (store *PG) PublicComments(ctx context.Context, imageID string, ascending bool, limit, offset int) ([]gallerycomments.Comment, int, error) {
	query := store.db.Model(tComments).Ctx(ctx).
		Where("image_id", imageID).
		Where("status", string(gallerycomments.StatusApproved)).
		Where("parent_id IS NULL").
		Where("deleted_at IS NULL")
	total, err := query.Clone().Count()
	if err != nil {
		return nil, 0, err
	}
	ordered := query
	if ascending {
		ordered = ordered.OrderAsc("created_at").OrderAsc("id")
	} else {
		ordered = ordered.OrderDesc("created_at").OrderDesc("id")
	}
	var comments []gallerycomments.Comment
	err = ordered.Limit(limit).Offset(offset).Scan(&comments)
	return comments, total, err
}

func (store *PG) Replies(ctx context.Context, parentIDs []string) (map[string][]gallerycomments.Comment, error) {
	result := map[string][]gallerycomments.Comment{}
	if len(parentIDs) == 0 {
		return result, nil
	}
	var comments []gallerycomments.Comment
	err := store.db.Model(tComments).Ctx(ctx).
		WhereIn("parent_id", parentIDs).
		Where("status", string(gallerycomments.StatusApproved)).
		Where("deleted_at IS NULL").
		OrderAsc("created_at").OrderAsc("id").
		Scan(&comments)
	if err != nil {
		return nil, err
	}
	for _, comment := range comments {
		result[comment.ParentID] = append(result[comment.ParentID], comment)
	}
	return result, nil
}

func (store *PG) AdminComments(ctx context.Context, input gallerycomments.AdminQuery) ([]gallerycomments.AdminComment, int, error) {
	query := store.db.Model(tComments+" c").Ctx(ctx).
		LeftJoin("gallery_images i", "i.id=c.image_id").
		LeftJoin(tComments+" parent", "parent.id=c.parent_id AND parent.image_id=c.image_id AND parent.deleted_at IS NULL").
		Where("c.deleted_at IS NULL")
	if input.Status != "" {
		query = query.Where("c.status", string(input.Status))
	}
	if input.Keyword != "" {
		like := "%" + input.Keyword + "%"
		query = query.Where(
			"(c.content ILIKE ? OR c.author_name ILIKE ? OR c.author_email ILIKE ? OR c.user_key ILIKE ? OR i.title ILIKE ?)",
			like, like, like, like, like,
		)
	}
	total, err := query.Clone().Count()
	if err != nil {
		return nil, 0, err
	}
	ordered := query.Fields("c.*, i.title AS image_title, CASE WHEN i.public_rendition_ready AND i.deleted_at IS NULL THEN i.asset_id::text ELSE NULL END AS image_asset_id, COALESCE(parent.user_key, '') AS parent_user_key, COALESCE(parent.author_name, '') AS parent_author_name, COALESCE(LEFT(parent.content, 160), '') AS parent_content")
	if input.Ascending {
		ordered = ordered.OrderAsc("c.created_at").OrderAsc("c.id")
	} else {
		ordered = ordered.OrderDesc("c.created_at").OrderDesc("c.id")
	}
	var comments []gallerycomments.AdminComment
	err = ordered.Limit(input.Size).Offset((input.Page - 1) * input.Size).Scan(&comments)
	return comments, total, err
}

func (store *PG) SetStatus(ctx context.Context, id string, status gallerycomments.Status) (bool, error) {
	result, err := store.db.Model(tComments).Ctx(ctx).
		Where("id", id).
		Where("deleted_at IS NULL").
		Data(g.Map{"status": string(status), "updated_at": gdb.Raw("NOW()")}).
		Update()
	if err != nil {
		return false, err
	}
	changed, err := result.RowsAffected()
	return changed > 0, err
}

func (store *PG) SoftDelete(ctx context.Context, id string) (bool, error) {
	var changed int64
	err := store.db.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		result, err := tx.Model(tComments).Ctx(ctx).
			Where("(id = ? OR parent_id = ?)", id, id).
			Where("deleted_at IS NULL").
			Data(g.Map{"deleted_at": gdb.Raw("NOW()"), "updated_at": gdb.Raw("NOW()")}).
			Update()
		if err != nil {
			return err
		}
		changed, err = result.RowsAffected()
		return err
	})
	return changed > 0, err
}
