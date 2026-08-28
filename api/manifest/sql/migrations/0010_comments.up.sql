CREATE TABLE gallery_comments (
    id UUID PRIMARY KEY,
    image_id UUID NOT NULL REFERENCES gallery_images(id) ON DELETE CASCADE,
    parent_id UUID,
    user_key TEXT NOT NULL DEFAULT '',
    author_name TEXT NOT NULL DEFAULT '',
    author_email TEXT NOT NULL DEFAULT '',
    content TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'approved', 'spam', 'trash')),
    ip TEXT NOT NULL DEFAULT '',
    user_agent TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    UNIQUE (image_id, id),
    FOREIGN KEY (image_id, parent_id)
        REFERENCES gallery_comments(image_id, id) ON DELETE CASCADE,
    CONSTRAINT gallery_comment_parent_check CHECK (parent_id IS NULL OR parent_id <> id)
);

CREATE INDEX gallery_comments_public_threads_idx
    ON gallery_comments (image_id, created_at DESC, id DESC)
    WHERE status = 'approved' AND parent_id IS NULL AND deleted_at IS NULL;

CREATE INDEX gallery_comments_public_replies_idx
    ON gallery_comments (parent_id, created_at, id)
    WHERE status = 'approved' AND deleted_at IS NULL;

CREATE INDEX gallery_comments_moderation_idx
    ON gallery_comments (status, created_at DESC, id DESC)
    WHERE deleted_at IS NULL;
