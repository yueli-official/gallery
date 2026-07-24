package gallerywebhook

import "github.com/yueli-official/foundation/go/webhook"

const (
	ClassificationRevised webhook.EventType = "com.yueli.gallery.classification.revised.v1"
	SubmissionReviewed    webhook.EventType = "com.yueli.gallery.submission.reviewed.v1"
	ImagePublished        webhook.EventType = "com.yueli.gallery.image.published.v1"
)

func Definition(siteSlug string) webhook.Definition {
	return webhook.Definition{
		Version:  webhook.DefinitionVersion,
		Consumer: "gallery",
		Source:   "urn:yueli:gallery:" + siteSlug,
		EventTypes: []webhook.EventTypeDefinition{
			{Type: ClassificationRevised, MaxDataBytes: 64 << 10},
			{Type: SubmissionReviewed, MaxDataBytes: 64 << 10},
			{Type: ImagePublished, MaxDataBytes: 64 << 10},
		},
	}
}
