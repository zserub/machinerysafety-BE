-- query: list_by_risk_id
SELECT c.id, c.risk_id, c.author_id, c.content, c.created_at, c.edited_at, p.person_name
FROM comments c
JOIN persons p ON c.author_id = p.person_id
WHERE c.risk_id = $1
ORDER BY c.created_at ASC

-- query: create
INSERT INTO comments (risk_id, author_id, content)
VALUES ($1, $2, $3)
RETURNING id, created_at

-- query: get_by_id
SELECT c.id, c.risk_id, c.author_id, c.content, c.created_at, c.edited_at, p.person_name
FROM comments c
JOIN persons p ON c.author_id = p.person_id
WHERE c.id = $1
