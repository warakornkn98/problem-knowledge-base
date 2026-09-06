-- 0001_init.sql
-- Core schema: users, categories, tags, problems and their child tables.

CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- ─────────────────────────────────────────────────────────────
-- users
-- ─────────────────────────────────────────────────────────────
CREATE TABLE users (
    id            BIGSERIAL PRIMARY KEY,
    username      VARCHAR(50)  NOT NULL UNIQUE,
    email         VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    display_name  VARCHAR(100) NOT NULL,
    role          VARCHAR(20)  NOT NULL DEFAULT 'MEMBER'
                  CHECK (role IN ('ADMIN', 'MEMBER')),
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT now()
);

-- ─────────────────────────────────────────────────────────────
-- problem_categories
-- ─────────────────────────────────────────────────────────────
CREATE TABLE problem_categories (
    id          BIGSERIAL PRIMARY KEY,
    name        VARCHAR(100) NOT NULL UNIQUE,
    slug        VARCHAR(100) NOT NULL UNIQUE,
    description TEXT NOT NULL DEFAULT '',
    color       VARCHAR(20) NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT now()
);

-- ─────────────────────────────────────────────────────────────
-- tags
-- ─────────────────────────────────────────────────────────────
CREATE TABLE tags (
    id         BIGSERIAL PRIMARY KEY,
    name       VARCHAR(50) NOT NULL UNIQUE,
    slug       VARCHAR(50) NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- ─────────────────────────────────────────────────────────────
-- problems
-- ─────────────────────────────────────────────────────────────
CREATE TABLE problems (
    id            BIGSERIAL PRIMARY KEY,
    title         VARCHAR(300) NOT NULL,
    description   TEXT NOT NULL DEFAULT '',
    error_message TEXT NOT NULL DEFAULT '',
    category_id   BIGINT NOT NULL REFERENCES problem_categories(id) ON DELETE RESTRICT,
    severity      VARCHAR(20) NOT NULL DEFAULT 'MEDIUM'
                  CHECK (severity IN ('LOW', 'MEDIUM', 'HIGH', 'CRITICAL')),
    status        VARCHAR(20) NOT NULL DEFAULT 'OPEN'
                  CHECK (status IN ('OPEN', 'INVESTIGATING', 'SOLVED', 'KNOWN')),
    environment   VARCHAR(20) NOT NULL DEFAULT 'PROD'
                  CHECK (environment IN ('LOCAL', 'DEV', 'UAT', 'PROD')),
    project       VARCHAR(150) NOT NULL,
    root_cause    TEXT NOT NULL DEFAULT '',
    solution      TEXT NOT NULL DEFAULT '',
    prevention    TEXT NOT NULL DEFAULT '',
    tags_cached   TEXT NOT NULL DEFAULT '',
    created_by    BIGINT REFERENCES users(id) ON DELETE SET NULL,
    updated_by    BIGINT REFERENCES users(id) ON DELETE SET NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    solved_at     TIMESTAMPTZ
);

CREATE INDEX idx_problems_category    ON problems (category_id);
CREATE INDEX idx_problems_status      ON problems (status);
CREATE INDEX idx_problems_severity    ON problems (severity);
CREATE INDEX idx_problems_environment ON problems (environment);
CREATE INDEX idx_problems_project     ON problems (project);
CREATE INDEX idx_problems_created_at  ON problems (created_at DESC);
CREATE INDEX idx_problems_solved_at   ON problems (solved_at DESC);

-- ─────────────────────────────────────────────────────────────
-- problem_tags  (M:N problems <-> tags)
-- ─────────────────────────────────────────────────────────────
CREATE TABLE problem_tags (
    problem_id BIGINT NOT NULL REFERENCES problems(id) ON DELETE CASCADE,
    tag_id     BIGINT NOT NULL REFERENCES tags(id)     ON DELETE CASCADE,
    PRIMARY KEY (problem_id, tag_id)
);

CREATE INDEX idx_problem_tags_tag ON problem_tags (tag_id);

-- ─────────────────────────────────────────────────────────────
-- problem_steps
-- ─────────────────────────────────────────────────────────────
CREATE TABLE problem_steps (
    id         BIGSERIAL PRIMARY KEY,
    problem_id BIGINT NOT NULL REFERENCES problems(id) ON DELETE CASCADE,
    step_no    INTEGER NOT NULL,
    action     TEXT NOT NULL,
    result     TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (problem_id, step_no)
);

CREATE INDEX idx_problem_steps_problem ON problem_steps (problem_id, step_no);

-- ─────────────────────────────────────────────────────────────
-- problem_related  (self M:N)
-- ─────────────────────────────────────────────────────────────
CREATE TABLE problem_related (
    id                 BIGSERIAL PRIMARY KEY,
    problem_id         BIGINT NOT NULL REFERENCES problems(id) ON DELETE CASCADE,
    related_problem_id BIGINT NOT NULL REFERENCES problems(id) ON DELETE CASCADE,
    relation_type      VARCHAR(20) NOT NULL DEFAULT 'RELATED'
                       CHECK (relation_type IN ('RELATED', 'SIMILAR', 'CAUSED_BY', 'DUPLICATE', 'WORKAROUND')),
    created_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (problem_id, related_problem_id, relation_type),
    CHECK (problem_id <> related_problem_id)
);

CREATE INDEX idx_problem_related_problem ON problem_related (problem_id);
CREATE INDEX idx_problem_related_related ON problem_related (related_problem_id);
