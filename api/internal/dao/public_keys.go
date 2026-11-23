package dao

import "context"

func (p *PG) ImageIDByPublicKey(ctx context.Context, key string) (string, error) {
	value, err := p.db.GetValue(ctx, `SELECT id::text FROM gallery_images WHERE public_id=?`, key)
	if err != nil {
		return "", err
	}
	return value.String(), nil
}
