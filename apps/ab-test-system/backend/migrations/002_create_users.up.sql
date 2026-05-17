CREATE TYPE user_role AS ENUM ('admin', 'member', 'viewer');

CREATE TABLE users (
    id           UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    firebase_uid TEXT        NOT NULL UNIQUE,
    org_id       UUID        REFERENCES organizations(id) ON DELETE SET NULL,
    email        TEXT        NOT NULL UNIQUE,
    name         TEXT        NOT NULL,
    photo_url    TEXT,
    role         user_role   NOT NULL DEFAULT 'member',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_users_firebase_uid ON users(firebase_uid);
CREATE INDEX idx_users_org_id       ON users(org_id);
