-- 0002_fulltext.sql
-- Full text search support for problems.
--
-- search_vector is a STORED generated column so it stays in sync with the row
-- automatically (no triggers). Tag names are folded in through the denormalised
-- problems.tags_cached column, which the repository layer keeps up to date
-- whenever a problem's tags change.
--
-- Weights:  A = title
--           B = error_message, tags
--           C = description, root_cause, solution
--           D = prevention

ALTER TABLE problems
    ADD COLUMN search_vector tsvector
    GENERATED ALWAYS AS (
        setweight(to_tsvector('english', coalesce(title, '')),         'A') ||
        setweight(to_tsvector('english', coalesce(error_message, '')),  'B') ||
        setweight(to_tsvector('english', coalesce(tags_cached, '')),    'B') ||
        setweight(to_tsvector('english', coalesce(description, '')),    'C') ||
        setweight(to_tsvector('english', coalesce(root_cause, '')),     'C') ||
        setweight(to_tsvector('english', coalesce(solution, '')),       'C') ||
        setweight(to_tsvector('english', coalesce(prevention, '')),     'D')
    ) STORED;

CREATE INDEX idx_problems_search_vector ON problems USING GIN (search_vector);

-- Trigram indexes for fuzzy "I can't remember the exact wording" matching.
CREATE INDEX idx_problems_title_trgm ON problems USING GIN (title gin_trgm_ops);
CREATE INDEX idx_problems_error_trgm ON problems USING GIN (error_message gin_trgm_ops);
