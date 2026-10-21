-- query: get_by_person_id
SELECT person_id, language, theme, active_columns
FROM preferences
WHERE person_id = $1

-- query: update
UPDATE preferences
SET language = COALESCE($2, language), theme = COALESCE($3, theme)
WHERE person_id = $1
RETURNING person_id, language, theme, active_columns

-- query: update_columns
UPDATE preferences
SET active_columns = $2
WHERE person_id = $1
RETURNING person_id, language, theme, active_columns

-- query: update_with_columns
UPDATE preferences
SET language = COALESCE($2, language), theme = COALESCE($3, theme), active_columns = COALESCE($4, active_columns)
WHERE person_id = $1
RETURNING person_id, language, theme, active_columns

-- query: create_default
INSERT INTO preferences (person_id, language, theme)
VALUES ($1, 'en', false)
RETURNING person_id, language, theme, active_columns
