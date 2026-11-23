package dao

import (
	"context"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/yueli-official/gallery/api/internal/model"
)

func (p *PG) Case(ctx context.Context, id string) (*model.Case, error) {
	record, err := p.db.GetOne(ctx, `SELECT id::text, COALESCE(image_id::text, '') AS image_id,
 COALESCE((SELECT i.public_id FROM gallery_images i WHERE i.id = image_id), '') AS image_public_id,
 COALESCE(submission_id::text, '') AS submission_id, kind, status, reason, description,
 COALESCE(proposed_source_url, '') AS proposed_source_url, COALESCE(proposed_metadata::text, '') AS proposed_metadata_json,
 resolution_note, created_at, updated_at FROM gallery_cases WHERE id = ?::uuid`, id)
	if err != nil {
		return nil, gerror.Wrap(err, "read gallery case")
	}
	if len(record) == 0 {
		return nil, nil
	}
	return recordAs[model.Case](record)
}
