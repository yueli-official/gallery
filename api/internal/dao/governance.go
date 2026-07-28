package dao

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/lib/pq"

	"github.com/yueli-official/foundation/go/classification"
	"github.com/yueli-official/gallery/api/internal/galleryerr"
)

type governanceRaw func(string, ...any) *gdb.Model

type governanceImpactRow struct {
	ID               string         `orm:"id"`
	Exists           bool           `orm:"exists"`
	Status           string         `orm:"status"`
	DirectChildIDs   pq.StringArray `orm:"direct_child_ids"`
	AssignmentCount  int64          `orm:"assignment_count"`
	PrimaryCount     int64          `orm:"primary_count"`
	AliasCount       int64          `orm:"alias_count"`
	ReplacementCount int64          `orm:"replacement_count"`
	HistoricalCount  int64          `orm:"historical_count"`
}

func (p *PG) ClassificationGovernanceImpacts(ctx context.Context, requests []classification.ImpactRequest) ([]classification.ReferenceImpact, string, error) {
	if len(requests) == 0 {
		return []classification.ReferenceImpact{}, "", nil
	}
	var impacts []classification.ReferenceImpact
	err := p.db.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		tx = tx.Ctx(ctx)
		if _, err := tx.Ctx(ctx).Exec(`SET TRANSACTION ISOLATION LEVEL REPEATABLE READ, READ ONLY`); err != nil {
			return gerror.Wrap(err, "start gallery governance impact snapshot")
		}
		var err error
		impacts, err = loadClassificationGovernanceImpacts(tx.Raw, requests)
		return err
	})
	if err != nil {
		return nil, "", err
	}
	return impacts, classificationGovernanceImpactToken(requests, impacts), nil
}

func loadClassificationGovernanceImpacts(raw governanceRaw, requests []classification.ImpactRequest) ([]classification.ReferenceImpact, error) {
	impacts := make([]classification.ReferenceImpact, 0, len(requests))
	for _, request := range requests {
		query, err := governanceImpactQuery(request.Kind)
		if err != nil {
			return nil, err
		}
		var row *governanceImpactRow
		if err := raw(query, request.ID).Scan(&row); err != nil {
			return nil, gerror.Wrapf(err, "query gallery %s governance impact", request.Kind)
		}
		if row == nil {
			return nil, gerror.Newf("gallery governance impact query returned no row for %s %s", request.Kind, request.ID)
		}
		impacts = append(impacts, classification.ReferenceImpact{
			Kind: request.Kind, ID: request.ID, Exists: row.Exists, Status: classification.Status(row.Status),
			DirectChildIDs: append([]string(nil), row.DirectChildIDs...), AssignmentCount: row.AssignmentCount,
			PrimaryCount: row.PrimaryCount, AliasCount: row.AliasCount, ReplacementCount: row.ReplacementCount,
			HistoricalCount: row.HistoricalCount,
		})
	}
	return impacts, nil
}

func governanceImpactQuery(kind classification.GovernIdentityKind) (string, error) {
	switch kind {
	case classification.GovernCategory:
		return `
WITH requested AS (SELECT ?::uuid AS id)
SELECT requested.id::text AS id,
       category.id IS NOT NULL AS exists,
       COALESCE(category.status, '') AS status,
       ARRAY(SELECT child.id::text FROM gallery_categories child WHERE child.parent_id = requested.id ORDER BY child.id)::text[] AS direct_child_ids,
       ((SELECT COUNT(*) FROM gallery_submission_category_assignments assignment WHERE assignment.category_id = requested.id) +
        (SELECT COUNT(*) FROM gallery_image_category_assignments assignment WHERE assignment.category_id = requested.id))::bigint AS assignment_count,
       ((SELECT COUNT(*) FROM gallery_submission_primary_categories primary_assignment WHERE primary_assignment.category_id = requested.id) +
        (SELECT COUNT(*) FROM gallery_image_primary_categories primary_assignment WHERE primary_assignment.category_id = requested.id))::bigint AS primary_count,
       0::bigint AS alias_count,
       (SELECT COUNT(*) FROM gallery_categories source WHERE source.replacement_id = requested.id)::bigint AS replacement_count,
       0::bigint AS historical_count
FROM requested
LEFT JOIN gallery_categories category ON category.id = requested.id`, nil
	case classification.GovernFacet:
		return `
WITH requested AS (SELECT ?::uuid AS id)
SELECT requested.id::text AS id,
       facet.id IS NOT NULL AS exists,
       COALESCE(facet.status, '') AS status,
       ARRAY[]::text[] AS direct_child_ids,
       0::bigint AS assignment_count,
       0::bigint AS primary_count,
       0::bigint AS alias_count,
       (SELECT COUNT(*) FROM gallery_facets source WHERE source.replacement_id = requested.id)::bigint AS replacement_count,
       (SELECT COUNT(*) FROM gallery_classification_policy_profiles policy
        WHERE EXISTS (
            SELECT 1 FROM jsonb_array_elements(COALESCE(policy.document->'facets', '[]'::jsonb)) item
            WHERE item->>'facetId' = requested.id::text
        ))::bigint AS historical_count
FROM requested
LEFT JOIN gallery_facets facet ON facet.id = requested.id`, nil
	case classification.GovernFacetValue:
		return `
WITH requested AS (SELECT ?::uuid AS id)
SELECT requested.id::text AS id,
       facet_value.id IS NOT NULL AS exists,
       COALESCE(facet_value.status, '') AS status,
       ARRAY(SELECT child.id::text FROM gallery_facet_values child WHERE child.parent_id = requested.id ORDER BY child.id)::text[] AS direct_child_ids,
       ((SELECT COUNT(*) FROM gallery_submission_facet_assignments assignment WHERE assignment.facet_value_id = requested.id) +
        (SELECT COUNT(*) FROM gallery_image_facet_assignments assignment WHERE assignment.facet_value_id = requested.id))::bigint AS assignment_count,
       0::bigint AS primary_count,
       0::bigint AS alias_count,
       (SELECT COUNT(*) FROM gallery_facet_values source WHERE source.replacement_id = requested.id)::bigint AS replacement_count,
       0::bigint AS historical_count
FROM requested
LEFT JOIN gallery_facet_values facet_value ON facet_value.id = requested.id`, nil
	case classification.GovernTag:
		return `
WITH requested AS (SELECT ?::uuid AS id)
SELECT requested.id::text AS id,
       tag.id IS NOT NULL AS exists,
       COALESCE(tag.status, '') AS status,
       ARRAY[]::text[] AS direct_child_ids,
       ((SELECT COUNT(*) FROM gallery_submission_tag_assignments assignment WHERE assignment.tag_id = requested.id) +
        (SELECT COUNT(*) FROM gallery_image_tag_assignments assignment WHERE assignment.tag_id = requested.id))::bigint AS assignment_count,
       0::bigint AS primary_count,
       (SELECT COUNT(*) FROM gallery_tag_lookup_entries entry WHERE entry.target_tag_id = requested.id AND entry.kind = 'alias')::bigint AS alias_count,
       (SELECT COUNT(*) FROM gallery_tags source WHERE source.replacement_id = requested.id)::bigint AS replacement_count,
       (SELECT COUNT(*) FROM gallery_tag_proposals proposal WHERE proposal.resolved_tag_id = requested.id)::bigint AS historical_count
FROM requested
LEFT JOIN gallery_tags tag ON tag.id = requested.id`, nil
	default:
		return "", fmt.Errorf("unsupported gallery governance identity kind %q", kind)
	}
}

func classificationGovernanceImpactToken(requests []classification.ImpactRequest, impacts []classification.ReferenceImpact) string {
	byKey := make(map[string]classification.ReferenceImpact, len(impacts))
	for _, impact := range impacts {
		children := append([]string(nil), impact.DirectChildIDs...)
		sort.Strings(children)
		impact.DirectChildIDs = children
		byKey[governanceImpactKey(impact.Kind, impact.ID)] = impact
	}
	hash := sha256.New()
	for _, request := range requests {
		impact := byKey[governanceImpactKey(request.Kind, request.ID)]
		writeGovernanceTokenPart(hash,
			string(request.Kind), request.ID, strconv.FormatBool(impact.Exists), string(impact.Status),
			strconv.FormatInt(impact.AssignmentCount, 10), strconv.FormatInt(impact.PrimaryCount, 10),
			strconv.FormatInt(impact.AliasCount, 10), strconv.FormatInt(impact.ReplacementCount, 10),
			strconv.FormatInt(impact.HistoricalCount, 10), strconv.Itoa(len(impact.DirectChildIDs)),
		)
		writeGovernanceTokenPart(hash, impact.DirectChildIDs...)
	}
	return hex.EncodeToString(hash.Sum(nil))
}

func governanceImpactKey(kind classification.GovernIdentityKind, id string) string {
	return string(kind) + "\x00" + id
}

type governanceTokenWriter interface{ Write([]byte) (int, error) }

func writeGovernanceTokenPart(writer governanceTokenWriter, values ...string) {
	for _, value := range values {
		_, _ = writer.Write([]byte(value))
		_, _ = writer.Write([]byte{0})
	}
}

func (p *PG) ExecuteClassificationGovernance(ctx context.Context, operator string, request classification.GovernFactRequest, plan classification.GovernancePlan) (uint64, error) {
	if request.CatalogRevision == 0 || request.CatalogRevision != plan.ExpectedCatalogRevision ||
		request.RequestToken == "" || request.RequestToken != plan.ExpectedRequestToken ||
		(len(request.Impacts) != 0 && plan.ExpectedImpactToken == "") {
		return 0, galleryerr.Conflict("classification_governance")
	}
	var nextRevision uint64
	err := p.db.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		tx = tx.Ctx(ctx)
		if _, err := tx.Ctx(ctx).Exec(`SET TRANSACTION ISOLATION LEVEL REPEATABLE READ`); err != nil {
			return gerror.Wrap(err, "start gallery governance transaction")
		}
		catalog, err := tx.GetOne(`
SELECT id::text AS id, revision
FROM gallery_classification_catalogs
WHERE catalog_key = 'gallery'
FOR UPDATE`)
		if err != nil {
			return gerror.Wrap(err, "lock gallery classification catalog")
		}
		if len(catalog) == 0 || catalog["revision"].Uint64() != plan.ExpectedCatalogRevision {
			return galleryerr.Conflict("classification_revision")
		}
		impacts, err := loadClassificationGovernanceImpacts(tx.Raw, request.Impacts)
		if err != nil {
			return err
		}
		if len(request.Impacts) != 0 && classificationGovernanceImpactToken(request.Impacts, impacts) != plan.ExpectedImpactToken {
			return galleryerr.Conflict("classification_impact")
		}
		for _, step := range plan.Steps {
			if err := executeClassificationGovernanceStep(ctx, tx, step); err != nil {
				return err
			}
		}
		revisionRecord, err := tx.GetOne(`
UPDATE gallery_classification_catalogs
SET revision = revision + 1, updated_at = NOW()
WHERE id = ?::uuid AND revision = ?
RETURNING revision`, catalog["id"].String(), plan.ExpectedCatalogRevision)
		if err != nil {
			return gerror.Wrap(err, "advance gallery classification revision")
		}
		if len(revisionRecord) == 0 {
			return galleryerr.Conflict("classification_revision")
		}
		nextRevision = revisionRecord["revision"].Uint64()
		payload, err := json.Marshal(map[string]any{
			"operator": operator,
			"steps":    plan.Steps,
		})
		if err != nil {
			return gerror.Wrap(err, "encode gallery classification outbox payload")
		}
		if err := p.enqueueClassificationRefresh(
			ctx, tx, catalog["id"].String(), nextRevision,
			"classification.governance.applied", payload,
		); err != nil {
			return gerror.Wrap(err, "enqueue gallery classification refresh")
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	return nextRevision, nil
}

func executeClassificationGovernanceStep(ctx context.Context, tx gdb.TX, step classification.GovernStep) error {
	switch step.Kind {
	case classification.GovernChangeStatus:
		return changeClassificationStatus(ctx, tx, step)
	case classification.GovernMoveChild:
		return moveClassificationChild(ctx, tx, step)
	case classification.GovernMigrateAssignments:
		return migrateClassificationAssignments(ctx, tx, step)
	case classification.GovernMigratePrimary:
		return migrateClassificationPrimary(ctx, tx, step)
	case classification.GovernMigrateAliases:
		return migrateClassificationAliases(ctx, tx, step)
	case classification.GovernSetReplacement:
		return setClassificationReplacement(ctx, tx, step)
	case classification.GovernClearPrimaryAssignments:
		return clearClassificationPrimary(ctx, tx, step)
	case classification.GovernDeleteAssignments:
		return deleteClassificationAssignments(ctx, tx, step)
	case classification.GovernDeleteAliases:
		return deleteClassificationAliases(ctx, tx, step)
	case classification.GovernDeleteReferences:
		return deleteClassificationReferences(ctx, tx, step)
	case classification.GovernDeleteIdentity:
		return deleteClassificationIdentity(ctx, tx, step)
	default:
		return fmt.Errorf("unsupported gallery governance step %q", step.Kind)
	}
}

func changeClassificationStatus(ctx context.Context, tx gdb.TX, step classification.GovernStep) error {
	query, args, err := classificationStatusUpdateStatement(step)
	if err != nil {
		return err
	}
	return execRequired(ctx, tx, query, args...)
}

func classificationStatusUpdateStatement(step classification.GovernStep) (string, []any, error) {
	table, err := classificationIdentityTable(step.IdentityKind)
	if err != nil {
		return "", nil, err
	}
	setClause := "status = ?, updated_at = NOW()"
	if step.IdentityKind != classification.GovernTag && step.Status == classification.StatusActive {
		setClause += ", first_activated_at = COALESCE(first_activated_at, NOW())"
	}
	return `UPDATE ` + table + ` SET ` + setClause + ` WHERE id = ?::uuid AND status <> 'replaced'`, []any{step.Status, step.SourceID}, nil
}

func moveClassificationChild(ctx context.Context, tx gdb.TX, step classification.GovernStep) error {
	var table string
	switch step.IdentityKind {
	case classification.GovernCategory:
		table = "gallery_categories"
	case classification.GovernFacetValue:
		table = "gallery_facet_values"
	default:
		return fmt.Errorf("unsupported gallery child move kind %q", step.IdentityKind)
	}
	return execRequired(ctx, tx, `UPDATE `+table+` SET parent_id = NULLIF(?, '')::uuid, updated_at = NOW() WHERE id = ?::uuid`, step.ParentID, step.SourceID)
}

func migrateClassificationAssignments(ctx context.Context, tx gdb.TX, step classification.GovernStep) error {
	var statements []string
	switch step.IdentityKind {
	case classification.GovernCategory:
		statements = []string{
			`INSERT INTO gallery_submission_category_assignments (submission_id, category_id) SELECT submission_id, ?::uuid FROM gallery_submission_category_assignments WHERE category_id = ?::uuid ON CONFLICT DO NOTHING`,
			`INSERT INTO gallery_image_category_assignments (image_id, category_id) SELECT image_id, ?::uuid FROM gallery_image_category_assignments WHERE category_id = ?::uuid ON CONFLICT DO NOTHING`,
			`DELETE FROM gallery_submission_category_assignments WHERE category_id = ?::uuid`,
			`DELETE FROM gallery_image_category_assignments WHERE category_id = ?::uuid`,
		}
	case classification.GovernFacetValue:
		statements = []string{
			`INSERT INTO gallery_submission_facet_assignments (submission_id, facet_value_id) SELECT submission_id, ?::uuid FROM gallery_submission_facet_assignments WHERE facet_value_id = ?::uuid ON CONFLICT DO NOTHING`,
			`INSERT INTO gallery_image_facet_assignments (image_id, facet_value_id) SELECT image_id, ?::uuid FROM gallery_image_facet_assignments WHERE facet_value_id = ?::uuid ON CONFLICT DO NOTHING`,
			`DELETE FROM gallery_submission_facet_assignments WHERE facet_value_id = ?::uuid`,
			`DELETE FROM gallery_image_facet_assignments WHERE facet_value_id = ?::uuid`,
		}
	case classification.GovernTag:
		statements = []string{
			`INSERT INTO gallery_submission_tag_assignments (submission_id, tag_id) SELECT submission_id, ?::uuid FROM gallery_submission_tag_assignments WHERE tag_id = ?::uuid ON CONFLICT DO NOTHING`,
			`INSERT INTO gallery_image_tag_assignments (image_id, tag_id) SELECT image_id, ?::uuid FROM gallery_image_tag_assignments WHERE tag_id = ?::uuid ON CONFLICT DO NOTHING`,
			`DELETE FROM gallery_submission_tag_assignments WHERE tag_id = ?::uuid`,
			`DELETE FROM gallery_image_tag_assignments WHERE tag_id = ?::uuid`,
		}
	default:
		return fmt.Errorf("unsupported gallery assignment migration kind %q", step.IdentityKind)
	}
	for index, statement := range statements {
		args := []any{step.SourceID}
		if index < 2 {
			args = []any{step.TargetID, step.SourceID}
		}
		if _, err := tx.Ctx(ctx).Exec(statement, args...); err != nil {
			return gerror.Wrap(err, "migrate gallery classification assignments")
		}
	}
	return nil
}

func migrateClassificationPrimary(ctx context.Context, tx gdb.TX, step classification.GovernStep) error {
	if step.IdentityKind != classification.GovernCategory {
		return fmt.Errorf("unsupported gallery primary migration kind %q", step.IdentityKind)
	}
	for _, table := range []string{"gallery_submission_primary_categories", "gallery_image_primary_categories"} {
		if _, err := tx.Ctx(ctx).Exec(`UPDATE `+table+` SET category_id = ?::uuid WHERE category_id = ?::uuid`, step.TargetID, step.SourceID); err != nil {
			return gerror.Wrap(err, "migrate gallery primary category")
		}
	}
	return nil
}

func migrateClassificationAliases(ctx context.Context, tx gdb.TX, step classification.GovernStep) error {
	if step.IdentityKind != classification.GovernTag {
		return fmt.Errorf("unsupported gallery alias migration kind %q", step.IdentityKind)
	}
	_, err := tx.Ctx(ctx).Exec(`
UPDATE gallery_tag_lookup_entries
SET target_tag_id = ?::uuid
WHERE target_tag_id = ?::uuid AND kind = 'alias'`, step.TargetID, step.SourceID)
	return gerror.Wrap(err, "migrate gallery tag aliases")
}

func setClassificationReplacement(ctx context.Context, tx gdb.TX, step classification.GovernStep) error {
	switch step.IdentityKind {
	case classification.GovernCategory, classification.GovernFacetValue:
		table, err := classificationIdentityTable(step.IdentityKind)
		if err != nil {
			return err
		}
		if _, err := tx.Ctx(ctx).Exec(`UPDATE `+table+` SET replacement_id = ?::uuid, updated_at = NOW() WHERE replacement_id = ?::uuid`, step.TargetID, step.SourceID); err != nil {
			return gerror.Wrap(err, "flatten gallery classification replacements")
		}
		return execRequired(ctx, tx, `UPDATE `+table+` SET parent_id = NULL, status = 'replaced', replacement_id = ?::uuid, updated_at = NOW() WHERE id = ?::uuid`, step.TargetID, step.SourceID)
	case classification.GovernTag:
		if _, err := tx.Ctx(ctx).Exec(`UPDATE gallery_tags SET replacement_id = ?::uuid, updated_at = NOW() WHERE replacement_id = ?::uuid`, step.TargetID, step.SourceID); err != nil {
			return gerror.Wrap(err, "flatten gallery tag replacements")
		}
		if _, err := tx.Ctx(ctx).Exec(`UPDATE gallery_tag_lookup_entries SET target_tag_id = ?::uuid WHERE target_tag_id = ?::uuid AND kind IN ('alias', 'replacement')`, step.TargetID, step.SourceID); err != nil {
			return gerror.Wrap(err, "flatten gallery tag lookup replacements")
		}
		if _, err := tx.Ctx(ctx).Exec(`UPDATE gallery_tag_lookup_entries SET target_tag_id = ?::uuid, source_tag_id = ?::uuid, kind = 'replacement' WHERE target_tag_id = ?::uuid AND kind = 'canonical'`, step.TargetID, step.SourceID, step.SourceID); err != nil {
			return gerror.Wrap(err, "replace gallery canonical tag lookup")
		}
		return execRequired(ctx, tx, `UPDATE gallery_tags SET status = 'replaced', replacement_id = ?::uuid, updated_at = NOW() WHERE id = ?::uuid`, step.TargetID, step.SourceID)
	default:
		return fmt.Errorf("unsupported gallery replacement kind %q", step.IdentityKind)
	}
}

func clearClassificationPrimary(ctx context.Context, tx gdb.TX, step classification.GovernStep) error {
	if step.IdentityKind != classification.GovernCategory {
		return fmt.Errorf("unsupported gallery primary deletion kind %q", step.IdentityKind)
	}
	for _, table := range []string{"gallery_submission_primary_categories", "gallery_image_primary_categories"} {
		if _, err := tx.Ctx(ctx).Exec(`DELETE FROM `+table+` WHERE category_id = ?::uuid`, step.SourceID); err != nil {
			return gerror.Wrap(err, "delete gallery primary category assignments")
		}
	}
	return nil
}

func deleteClassificationAssignments(ctx context.Context, tx gdb.TX, step classification.GovernStep) error {
	var statements []string
	switch step.IdentityKind {
	case classification.GovernCategory:
		statements = []string{
			`DELETE FROM gallery_submission_category_assignments WHERE category_id = ?::uuid`,
			`DELETE FROM gallery_image_category_assignments WHERE category_id = ?::uuid`,
		}
	case classification.GovernFacetValue:
		statements = []string{
			`DELETE FROM gallery_submission_facet_assignments WHERE facet_value_id = ?::uuid`,
			`DELETE FROM gallery_image_facet_assignments WHERE facet_value_id = ?::uuid`,
		}
	case classification.GovernTag:
		statements = []string{
			`DELETE FROM gallery_submission_tag_assignments WHERE tag_id = ?::uuid`,
			`DELETE FROM gallery_image_tag_assignments WHERE tag_id = ?::uuid`,
		}
	default:
		return fmt.Errorf("unsupported gallery assignment deletion kind %q", step.IdentityKind)
	}
	for _, statement := range statements {
		if _, err := tx.Ctx(ctx).Exec(statement, step.SourceID); err != nil {
			return gerror.Wrap(err, "delete gallery classification assignments")
		}
	}
	return nil
}

func deleteClassificationAliases(ctx context.Context, tx gdb.TX, step classification.GovernStep) error {
	if step.IdentityKind != classification.GovernTag {
		return fmt.Errorf("unsupported gallery alias deletion kind %q", step.IdentityKind)
	}
	_, err := tx.Ctx(ctx).Exec(`DELETE FROM gallery_tag_lookup_entries WHERE target_tag_id = ?::uuid AND kind = 'alias'`, step.SourceID)
	return gerror.Wrap(err, "delete gallery tag aliases")
}

func deleteClassificationReferences(ctx context.Context, tx gdb.TX, step classification.GovernStep) error {
	switch step.IdentityKind {
	case classification.GovernCategory, classification.GovernFacetValue:
		table, err := classificationIdentityTable(step.IdentityKind)
		if err != nil {
			return err
		}
		_, err = tx.Ctx(ctx).Exec(`DELETE FROM `+table+` WHERE replacement_id = ?::uuid`, step.SourceID)
		return gerror.Wrap(err, "delete gallery classification replacement references")
	case classification.GovernFacet:
		if _, err := tx.Ctx(ctx).Exec(`DELETE FROM gallery_facets WHERE replacement_id = ?::uuid`, step.SourceID); err != nil {
			return gerror.Wrap(err, "delete gallery facet replacement references")
		}
		_, err := tx.Ctx(ctx).Exec(`
UPDATE gallery_classification_policy_profiles
SET document = jsonb_set(
        document,
        '{facets}',
        COALESCE((
            SELECT jsonb_agg(item)
            FROM jsonb_array_elements(COALESCE(document->'facets', '[]'::jsonb)) item
            WHERE item->>'facetId' <> ?
        ), '[]'::jsonb),
        true
    ),
    policy_revision = policy_revision + 1,
    updated_at = NOW()
WHERE EXISTS (
    SELECT 1
    FROM jsonb_array_elements(COALESCE(document->'facets', '[]'::jsonb)) item
    WHERE item->>'facetId' = ?
)`, step.SourceID, step.SourceID)
		return gerror.Wrap(err, "remove gallery facet policy references")
	case classification.GovernTag:
		if _, err := tx.Ctx(ctx).Exec(`DELETE FROM gallery_tag_proposals WHERE resolved_tag_id = ?::uuid`, step.SourceID); err != nil {
			return gerror.Wrap(err, "delete gallery tag proposal references")
		}
		if _, err := tx.Ctx(ctx).Exec(`
DELETE FROM gallery_tag_lookup_entries
WHERE target_tag_id = ?::uuid
   OR source_tag_id = ?::uuid
   OR source_tag_id IN (SELECT id FROM gallery_tags WHERE replacement_id = ?::uuid)`, step.SourceID, step.SourceID, step.SourceID); err != nil {
			return gerror.Wrap(err, "delete gallery tag lookup references")
		}
		_, err := tx.Ctx(ctx).Exec(`DELETE FROM gallery_tags WHERE replacement_id = ?::uuid`, step.SourceID)
		return gerror.Wrap(err, "delete gallery tag replacement references")
	default:
		return fmt.Errorf("unsupported gallery reference deletion kind %q", step.IdentityKind)
	}
}

func deleteClassificationIdentity(ctx context.Context, tx gdb.TX, step classification.GovernStep) error {
	if step.IdentityKind == classification.GovernTag {
		if _, err := tx.Ctx(ctx).Exec(`DELETE FROM gallery_tag_lookup_entries WHERE target_tag_id = ?::uuid OR source_tag_id = ?::uuid`, step.SourceID, step.SourceID); err != nil {
			return gerror.Wrap(err, "delete gallery tag lookup identity")
		}
	}
	table, err := classificationIdentityTable(step.IdentityKind)
	if err != nil {
		return err
	}
	return execRequired(ctx, tx, `DELETE FROM `+table+` WHERE id = ?::uuid`, step.SourceID)
}

func classificationIdentityTable(kind classification.GovernIdentityKind) (string, error) {
	switch kind {
	case classification.GovernCategory:
		return "gallery_categories", nil
	case classification.GovernFacet:
		return "gallery_facets", nil
	case classification.GovernFacetValue:
		return "gallery_facet_values", nil
	case classification.GovernTag:
		return "gallery_tags", nil
	default:
		return "", fmt.Errorf("unsupported gallery classification identity kind %q", kind)
	}
}

func execRequired(ctx context.Context, tx gdb.TX, query string, args ...any) error {
	result, err := tx.Ctx(ctx).Exec(query, args...)
	if err != nil {
		return gerror.Wrap(err, "execute gallery classification governance step")
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return gerror.Wrap(err, "read gallery classification governance result")
	}
	if rows == 0 {
		return galleryerr.Conflict("classification_impact")
	}
	return nil
}
