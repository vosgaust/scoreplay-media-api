CREATE TABLE tags (
    id         UUID        PRIMARY KEY,
    name       TEXT        NOT NULL,
    created_at TIMESTAMPTZ NOT NULL
);

CREATE UNIQUE INDEX tags_name_lower_key ON tags (lower(name));

CREATE TABLE media (
    id           UUID        PRIMARY KEY,
    name         TEXT        NOT NULL,

    storage_key  TEXT        NOT NULL UNIQUE,
    type         TEXT        NOT NULL,

    created_at   TIMESTAMPTZ NOT NULL
);

CREATE TABLE media_tags (
    media_id UUID NOT NULL REFERENCES media (id) ON DELETE CASCADE,
    tag_id   UUID NOT NULL REFERENCES tags (id) ON DELETE RESTRICT,

    PRIMARY KEY (media_id, tag_id)
);

CREATE INDEX idx_media_tags_tag_id ON media_tags (tag_id, media_id);
