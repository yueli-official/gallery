ALTER TABLE gallery_classification_presentation ADD COLUMN updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW();
