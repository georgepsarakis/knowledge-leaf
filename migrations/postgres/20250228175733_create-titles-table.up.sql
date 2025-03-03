CREATE TABLE wikipedia_titles (
  id VARCHAR(255) PRIMARY KEY,
  title VARCHAR(255) NOT NULL,
  numeric_id BIGSERIAL,
  created_at
      TIMESTAMP WITH TIME ZONE DEFAULT
      CURRENT_TIMESTAMP NOT NULL,
  updated_at
      TIMESTAMP WITH TIME ZONE DEFAULT
      CURRENT_TIMESTAMP NOT NULL
);

CREATE UNIQUE INDEX idx_wk_titles_title
    ON wikipedia_titles USING btree (title);

CREATE INDEX idx_wk_titles_numeric_id
    ON wikipedia_titles USING btree (numeric_id);