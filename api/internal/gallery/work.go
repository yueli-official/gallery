package gallery

import (
	"context"
	"encoding/json"
	"time"

	"github.com/yueli-official/foundation/go/work"
)

const ClassificationRefreshKind work.Kind = "gallery.classification-refresh"

func WorkDefinition() work.Definition {
	return work.Definition{
		Version: work.DefinitionVersion,
		Queues: []work.QueueDefinition{
			{Key: "projection", Concurrency: 1},
		},
		Kinds: []work.KindDefinition{
			{
				Key: ClassificationRefreshKind, Queue: "projection",
				DefaultAttempts: 5, MaxAttempts: 20, Timeout: 30 * time.Second,
			},
		},
		Retry: work.RetryPolicy{
			BaseDelay: time.Second, MaxDelay: 5 * time.Minute, Jitter: 0.2,
		},
	}
}

func (service *Service) ClassificationRefreshHandler() work.Handler {
	return work.HandlerFunc(func(ctx context.Context, job work.Job, _ work.Progress) (work.Result, error) {
		var payload struct {
			CatalogID string `json:"catalogId"`
			Revision  uint64 `json:"revision"`
			EventType string `json:"eventType"`
		}
		if err := json.Unmarshal(job.Payload, &payload); err != nil {
			return work.Result{}, work.Permanent(err)
		}
		if err := service.RefreshClassificationCatalog(ctx); err != nil {
			return work.Result{}, err
		}
		result, _ := json.Marshal(map[string]any{
			"catalogId": payload.CatalogID, "revision": payload.Revision,
			"eventType": payload.EventType,
		})
		return work.Result{Summary: "classification catalog refreshed", Data: result}, nil
	})
}
