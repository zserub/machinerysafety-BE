-- query: get_by_id
SELECT p.person_id, p.person_name, p.person_email, p.username, p.pw_hash, p.role_id, p.locked, r.role_name
FROM persons p
JOIN roles r ON p.role_id = r.role_id
WHERE p.person_id = $1

-- query: get_by_username
SELECT p.person_id, p.person_name, p.person_email, p.username, p.pw_hash, p.role_id, p.locked, r.role_name
FROM persons p
JOIN roles r ON p.role_id = r.role_id
WHERE p.username = $1

-- query: get_permissions
SELECT p.permission_id, p.resource, p.action
FROM permissions p
JOIN role_permissions rp ON p.permission_id = rp.permission_id
WHERE rp.role_id = $1

-- query: create_initial_admin
INSERT INTO persons (person_name, username, pw_hash, role_id)
SELECT $1::varchar(30), $1::varchar(30), $2, 77
WHERE NOT EXISTS (SELECT 1 FROM persons)
RETURNING person_id
