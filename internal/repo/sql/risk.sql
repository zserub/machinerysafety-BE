-- query: count_risks
SELECT COUNT(*) FROM risks r

-- query: list_risks_base
SELECT r.id, r.risk_type, r.linked_risk_id, r.creation_date, r.machines,
       r.line_id, r.category_id, r.danger_description, r.risk_level_id,
       r.action, r.responsible_id, r.moc_number, r.deadline, r.action_state_id,
       r.implementation_date, r.entry_modified_at, r.created_by_id, r.committed_at,
       r.cost, r.expense_type_id,
       l.line_name, c.category_name, rl.risk_level, ast.state,
       resp.person_name, cb.person_name, et.type
FROM risks r
JOIN lines l ON r.line_id = l.line_id
LEFT JOIN categories c ON r.category_id = c.id
JOIN risk_levels rl ON r.risk_level_id = rl.risk_level_id
JOIN action_states ast ON r.action_state_id = ast.state_id
LEFT JOIN persons resp ON r.responsible_id = resp.person_id
LEFT JOIN persons cb ON r.created_by_id = cb.person_id
LEFT JOIN expense_type et ON r.expense_type_id = et.id

-- query: get_by_id
SELECT r.id, r.risk_type, r.linked_risk_id, r.creation_date, r.machines,
       r.line_id, r.category_id, r.danger_description, r.risk_level_id,
       r.action, r.responsible_id, r.moc_number, r.deadline, r.action_state_id,
       r.implementation_date, r.entry_modified_at, r.created_by_id, r.committed_at,
       r.cost, r.expense_type_id,
       l.line_name, c.category_name, rl.risk_level, ast.state,
       resp.person_name, cb.person_name, et.type
FROM risks r
JOIN lines l ON r.line_id = l.line_id
LEFT JOIN categories c ON r.category_id = c.id
JOIN risk_levels rl ON r.risk_level_id = rl.risk_level_id
JOIN action_states ast ON r.action_state_id = ast.state_id
LEFT JOIN persons resp ON r.responsible_id = resp.person_id
LEFT JOIN persons cb ON r.created_by_id = cb.person_id
LEFT JOIN expense_type et ON r.expense_type_id = et.id
WHERE r.id = $1

-- query: get_photos
SELECT risk_id, photo_number, file_path, uploaded_at
FROM photos
WHERE risk_id = $1
ORDER BY photo_number

-- query: get_status_history
SELECT sh.id, sh.risk_id, sh.old_status, sh.new_status, sh.reason,
       sh.changed_by, sh.changed_at,
       old_st.state, new_st.state, p.person_name
FROM status_history sh
JOIN action_states old_st ON sh.old_status = old_st.state_id
JOIN action_states new_st ON sh.new_status = new_st.state_id
JOIN persons p ON sh.changed_by = p.person_id
WHERE sh.risk_id = $1
ORDER BY sh.changed_at DESC

-- query: get_deadline_history
SELECT dh.id, dh.risk_id, dh.old_deadline, dh.new_deadline, dh.reason,
       dh.changed_by, dh.changed_at, p.person_name
FROM deadline_history dh
JOIN persons p ON dh.changed_by = p.person_id
WHERE dh.risk_id = $1
ORDER BY dh.changed_at DESC

-- query: get_linked_risk_summary
SELECT r.id, r.machines, l.line_name, r.danger_description, r.risk_type
FROM risks r
JOIN lines l ON r.line_id = l.line_id
WHERE r.id = $1

-- query: create
INSERT INTO risks (
    risk_type, linked_risk_id, machines, line_id, category_id,
    danger_description, risk_level_id, action, responsible_id, moc_number,
    deadline, action_state_id, created_by_id, cost, expense_type_id
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15
) RETURNING id, creation_date, entry_modified_at

-- query: update
UPDATE risks SET
    machines = $2, line_id = $3, category_id = $4, danger_description = $5,
    risk_level_id = $6, action = $7, responsible_id = $8, moc_number = $9,
    deadline = $10, action_state_id = $11, implementation_date = $12,
    cost = $13, expense_type_id = $14, linked_risk_id = $15,
    entry_modified_at = NOW()
WHERE id = $1
RETURNING entry_modified_at

-- query: delete
DELETE FROM risks WHERE id = $1

-- query: delete_status_history
DELETE FROM status_history WHERE risk_id = $1

-- query: delete_deadline_history
DELETE FROM deadline_history WHERE risk_id = $1

-- query: delete_comments
DELETE FROM comments WHERE risk_id = $1

-- query: delete_photos
DELETE FROM photos WHERE risk_id = $1

-- query: clear_linked_risk_references
UPDATE risks SET linked_risk_id = NULL WHERE linked_risk_id = $1

-- query: get_risk_type
SELECT risk_type FROM risks WHERE id = $1

-- query: get_current_status
SELECT action_state_id FROM risks WHERE id = $1

-- query: get_current_deadline
SELECT deadline FROM risks WHERE id = $1

-- query: insert_status_history
INSERT INTO status_history (risk_id, old_status, new_status, reason, changed_by)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, changed_at

-- query: insert_deadline_history
INSERT INTO deadline_history (risk_id, old_deadline, new_deadline, reason, changed_by)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, changed_at
