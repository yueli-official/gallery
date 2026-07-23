package gallery

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/yueli-official/foundation/go/work"
)

func TestClassificationWorkRefreshesCatalogRevision(t *testing.T) {
	store := validSubmissionStore()
	service := New(store)
	if _, err := service.classificationCatalog(context.Background()); err != nil {
		t.Fatal(err)
	}
	store.classificationSnapshot.Revision = 2

	catalog := work.MustCompile(WorkDefinition())
	backend, err := work.NewMemory(catalog, work.MemoryOptions{})
	if err != nil {
		t.Fatal(err)
	}
	runner, err := work.NewRunner(catalog, backend, map[work.Kind]work.Handler{
		ClassificationRefreshKind: service.ClassificationRefreshHandler(),
	}, work.RunnerOptions{
		WorkerID: "gallery-test", LeaseDuration: 30 * time.Second,
		HeartbeatInterval: 10 * time.Second,
	})
	if err != nil {
		t.Fatal(err)
	}
	payload, _ := json.Marshal(map[string]any{
		"catalogId": "gallery", "revision": 2,
		"eventType": "classification.governance.applied",
	})
	enqueued, err := backend.Enqueue(context.Background(), work.Request{
		Kind: ClassificationRefreshKind, Payload: payload,
		IdempotencyKey: "gallery.classification:gallery:2",
	})
	if err != nil {
		t.Fatal(err)
	}
	if worked, err := runner.RunOnce(context.Background(), "projection", "gallery-test/1"); err != nil || !worked {
		t.Fatalf("run: worked=%v err=%v", worked, err)
	}
	job, err := backend.Get(context.Background(), enqueued.Job.ID)
	if err != nil {
		t.Fatal(err)
	}
	if job.Status != work.StatusSucceeded || service.catalogRevision != 2 {
		t.Fatalf("refresh failed: job=%+v revision=%d", job, service.catalogRevision)
	}
}
