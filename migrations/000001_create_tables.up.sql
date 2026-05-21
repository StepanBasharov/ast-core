CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE IF NOT EXISTS users_cv (
    uuid            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    first_name      TEXT NOT NULL,
    last_name       TEXT NOT NULL,
    cv_title        TEXT NOT NULL,
    specialization  TEXT NOT NULL,
    work_experience INT  NOT NULL DEFAULT 0,
    raw_text        TEXT NOT NULL DEFAULT '',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS skills (
    uuid UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    CONSTRAINT skills_name_unique UNIQUE (name)
);

CREATE TABLE IF NOT EXISTS cv_skills (
    cv_uuid    UUID NOT NULL REFERENCES users_cv(uuid) ON DELETE CASCADE,
    skill_uuid UUID NOT NULL REFERENCES skills(uuid)   ON DELETE CASCADE,
    CONSTRAINT cv_skills_pkey PRIMARY KEY (cv_uuid, skill_uuid)
);

CREATE INDEX IF NOT EXISTS idx_cv_skills_skill_uuid ON cv_skills (skill_uuid);
