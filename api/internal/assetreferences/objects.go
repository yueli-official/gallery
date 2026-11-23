package assetreferences

import "github.com/yueli-official/asset/referencesync"

const imageFactsSQL = `SELECT i.id::text,i.title,'/images/'||i.public_id,
 CASE WHEN i.deleted_at IS NULL AND i.publication_state<>'deleted' THEN i.asset_id::text ELSE '' END FROM gallery_images i`
const submissionFactsSQL = `SELECT s.id::text,s.title,'/submissions/'||s.id::text,
 CASE WHEN s.outcome<>'withdrawn' AND (s.image_id IS NULL OR (i.deleted_at IS NULL AND i.publication_state<>'deleted')) THEN s.asset_id::text ELSE '' END
 FROM gallery_submissions s LEFT JOIN gallery_images i ON i.id=s.image_id`
const classificationNodesSQL = `SELECT 'category' AS kind,id,name,updated_at FROM gallery_categories UNION ALL
 SELECT 'facet',id,name,updated_at FROM gallery_facets UNION ALL
 SELECT 'facet_value',id,name,updated_at FROM gallery_facet_values UNION ALL
 SELECT 'tag',id,current_name AS name,updated_at FROM gallery_tags`
const classificationFactsSQL = `SELECT n.kind||':'||n.id::text, n.name,'/manage/classification',COALESCE(p.asset_id::text,'')
 FROM (` + classificationNodesSQL + `) n LEFT JOIN gallery_classification_presentation p ON n.kind=p.kind AND n.id=p.node_id`
const siteFactsSQL = `SELECT 'site',s.name,'/manage/site',COALESCE(i.asset_id::text,'') FROM gallery_site_settings s
 LEFT JOIN gallery_images i ON i.id=s.sidebar_image_id WHERE s.site_key='gallery'`

func ObjectSource(origins ...string) (*referencesync.SQLObjectSource, error) {
	return referencesync.NewSQLObjectSource([]referencesync.ObjectQuery{
		{Query: referencesync.Query{RefType: "gallery-image", Kind: "asset", SQL: imageFactsSQL},
			OwnerSQL: imageFactsSQL + ` WHERE i.id=$1::uuid`, LockSQL: `SELECT id FROM gallery_images WHERE id=$1::uuid FOR UPDATE`, UpdatedAtSQL: `SELECT updated_at FROM gallery_images WHERE id=$1::uuid`},
		{Query: referencesync.Query{RefType: "gallery-submission-image", Kind: "asset", SQL: submissionFactsSQL},
			OwnerSQL: submissionFactsSQL + ` WHERE s.id=$1::uuid`, LockSQL: `SELECT id FROM gallery_submissions WHERE id=$1::uuid FOR UPDATE`, UpdatedAtSQL: `SELECT updated_at FROM gallery_submissions WHERE id=$1::uuid`},
		{Query: referencesync.Query{RefType: "gallery-classification-cover", Kind: "asset", SQL: classificationFactsSQL},
			OwnerSQL:     classificationFactsSQL + ` WHERE n.kind=split_part($1,':',1) AND n.id=split_part($1,':',2)::uuid`,
			LockSQL:      `SELECT id FROM gallery_classification_catalogs WHERE catalog_key='gallery' AND EXISTS(SELECT 1 FROM (` + classificationNodesSQL + `) n WHERE n.kind=split_part($1,':',1) AND n.id=split_part($1,':',2)::uuid) FOR UPDATE`,
			UpdatedAtSQL: `SELECT GREATEST(n.updated_at,p.updated_at) FROM (` + classificationNodesSQL + `) n LEFT JOIN gallery_classification_presentation p ON n.kind=p.kind AND n.id=p.node_id WHERE n.kind=split_part($1,':',1) AND n.id=split_part($1,':',2)::uuid`},
		{Query: referencesync.Query{RefType: "gallery-site-artwork", Kind: "asset", SQL: siteFactsSQL}, OwnerSQL: siteFactsSQL + ` AND $1='site'`,
			LockSQL: `SELECT site_key FROM gallery_site_settings WHERE site_key='gallery' AND $1='site' FOR UPDATE`, UpdatedAtSQL: `SELECT updated_at FROM gallery_site_settings WHERE site_key='gallery' AND $1='site'`},
	}, origins...)
}
func InitializationSource(origins ...string) (referencesync.InitializationSource, error) {
	source, err := ObjectSource(origins...)
	if err != nil {
		return referencesync.InitializationSource{}, err
	}
	return source.Initialization(), nil
}
