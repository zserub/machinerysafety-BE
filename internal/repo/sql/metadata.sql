-- query: get_risk_levels
SELECT risk_level, severity
FROM risk_levels
ORDER BY severity

-- query: get_statuses
SELECT state, state_id, is_terminal
FROM action_states
ORDER BY state_id

-- query: get_categories
SELECT category_name, id
FROM categories
ORDER BY category_name

-- query: get_expense_types
SELECT type, id
FROM expense_type
ORDER BY id

-- query: get_lines
SELECT line_name, line_id
FROM lines
ORDER BY line_name

-- query: get_responsibles
SELECT username, person_name, person_id
FROM persons
WHERE locked = false AND username IS NOT NULL
ORDER BY person_name

-- query: get_line_id_by_name
SELECT line_id FROM lines WHERE LOWER(line_name) = LOWER($1)

-- query: get_category_id_by_name
SELECT id FROM categories WHERE LOWER(category_name) = LOWER($1)

-- query: get_risk_level_id_by_name
SELECT risk_level_id FROM risk_levels WHERE LOWER(risk_level) = LOWER($1)

-- query: get_expense_type_id_by_name
SELECT id FROM expense_type WHERE LOWER(type) = LOWER($1)

-- query: get_responsible_id_by_username
SELECT person_id FROM persons WHERE LOWER(username) = LOWER($1)

-- query: get_responsible_id_by_name
SELECT person_id FROM persons WHERE LOWER(person_name) = LOWER($1)

-- query: get_status_id_by_name
SELECT state_id FROM action_states WHERE LOWER(state) = LOWER($1)
