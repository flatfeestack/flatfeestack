CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS post
(
    id         UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    author     UUID         NOT NULL,
    content    TEXT         NOT NULL,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    open       BOOLEAN      NOT NULL DEFAULT true,
    title      VARCHAR(255) NOT NULL,
    updated_at TIMESTAMPTZ
);

ALTER TABLE post ADD COLUMN IF NOT EXISTS proposal_ids TEXT[] NOT NULL DEFAULT array[]::text[];

CREATE TABLE IF NOT EXISTS comment
(
    id         UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    author     UUID        NOT NULL,
    content    TEXT        NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ,
    post_id    UUID REFERENCES post (id) ON DELETE CASCADE
);