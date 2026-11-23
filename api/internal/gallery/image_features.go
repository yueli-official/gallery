package gallery

import (
	"context"

	"github.com/yueli-official/foundation/go/classification"
	"github.com/yueli-official/gallery/api/internal/galleryerr"
	"github.com/yueli-official/gallery/api/internal/model"
)

const imageFeaturesSlug = "image-features"

// Use the same compiled, revision-bound catalog as submission validation.
// Directory identities and UUID/compact/slug references remain Classification's
// responsibility rather than a second product catalog.
func activeFacetCandidates(catalog *classification.Catalog) ([]classification.CandidateFacet, error) {
	p := catalog.Discover(classification.DiscoverRequest{PolicyKey: "gallery.image.public", CandidateProjection: classification.CandidateProjectionActive})
	facts := p.FactRequest()
	result := p.Complete(classification.DiscoverFacts{CatalogRevision: facts.CatalogRevision, RequestToken: facts.RequestToken})
	if result.Outcome != classification.OutcomeAccepted {
		return nil, classificationValidation(result.Diagnostics)
	}
	return result.Candidates.Facets, nil
}

func (s *Service) classifyImageEdit(ctx context.Context, input *model.AdminImageClassificationInput) error {
	catalog, err := s.classificationCatalog(ctx)
	if err != nil {
		return err
	}
	candidates, err := activeFacetCandidates(catalog)
	if err != nil {
		return err
	}
	owners := make(map[string]string)
	for _, facet := range candidates {
		for _, value := range facet.Values {
			owners[value.ID] = facet.ID
		}
	}
	byFacet := make(map[string][]string)
	for _, id := range input.FacetValueIDs {
		owner := owners[id]
		if owner == "" {
			return galleryerr.Validation("facetValueIds", "choose active facet values")
		}
		byFacet[owner] = append(byFacet[owner], id)
	}
	selections := make([]model.FacetSelection, 0, len(byFacet))
	for facetID, values := range byFacet {
		selections = append(selections, model.FacetSelection{FacetID: facetID, ValueIDs: values})
	}
	write, err := s.classifySubmission(ctx, model.SubmissionInput{
		CategoryIDs: []string{input.PrimaryCategoryID}, PrimaryCategoryID: input.PrimaryCategoryID, Facets: selections,
	})
	if err != nil {
		return err
	}
	input.CatalogRevision = write.CatalogRevision
	input.FacetValueIDs = make([]string, 0, len(write.Facets))
	for _, assignment := range write.Facets {
		input.FacetValueIDs = append(input.FacetValueIDs, assignment.ValueID)
	}
	return nil
}
