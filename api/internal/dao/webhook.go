package dao

import (
	"context"
	"encoding/json"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/yueli-official/foundation/go/webhook"

	"platform/products/gallery/api/internal/gallerywebhook"
	"platform/products/gallery/api/internal/model"
)

func (p *PG) publishSubmissionReviewedTx(
	ctx context.Context,
	tx gdb.TX,
	submission *model.Submission,
	decision string,
) error {
	if p.webhooks == nil || submission == nil {
		return nil
	}
	now := time.Now().UTC()
	payload, err := json.Marshal(struct {
		SubmissionID string `json:"submissionId"`
		ImageID      string `json:"imageId,omitempty"`
		Decision     string `json:"decision"`
		Outcome      string `json:"outcome"`
	}{
		SubmissionID: submission.ID, ImageID: submission.ImageID,
		Decision: decision, Outcome: submission.Outcome,
	})
	if err != nil {
		return err
	}
	if _, err := p.webhooks.PublishTx(ctx, tx.GetSqlTX(), webhook.EventCommand{
		Type: gallerywebhook.SubmissionReviewed, Subject: "submission/" + submission.ID,
		Data: payload, OccurredAt: now,
		IdempotencyKey: "gallery:submission:" + submission.ID + ":reviewed:v1",
	}); err != nil {
		return err
	}
	if submission.Outcome != "published" || submission.ImageID == "" {
		return nil
	}
	imagePayload, err := json.Marshal(struct {
		ImageID            string `json:"imageId"`
		OriginSubmissionID string `json:"originSubmissionId"`
	}{
		ImageID: submission.ImageID, OriginSubmissionID: submission.ID,
	})
	if err != nil {
		return err
	}
	_, err = p.webhooks.PublishTx(ctx, tx.GetSqlTX(), webhook.EventCommand{
		Type: gallerywebhook.ImagePublished, Subject: "image/" + submission.ImageID,
		Data: imagePayload, OccurredAt: now,
		IdempotencyKey: "gallery:image:" + submission.ImageID + ":published:v1",
	})
	return err
}
