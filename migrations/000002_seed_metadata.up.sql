-- =============================================================================
-- Seed: Reference / Lookup Data
-- All tables use explicit IDs to match the production system.
-- Sequences are reset after explicit-ID inserts.
-- =============================================================================

-- action_states
-- 0=UNKNOWN, 1=OPEN, 2=ASSIGNED, 3=IN_PROGRESS,
-- 4=PENDING_VERIFICATION, 5=REJECTED, 7=CLOSED (terminal)
INSERT INTO action_states (state_id, state, is_terminal) VALUES
    (0, 'UNKNOWN',              FALSE),
    (1, 'OPEN',                 FALSE),
    (2, 'ASSIGNED',             FALSE),
    (3, 'IN_PROGRESS',          FALSE),
    (4, 'PENDING_VERIFICATION', FALSE),
    (5, 'REJECTED',             FALSE),
    (7, 'CLOSED',               TRUE)
ON CONFLICT DO NOTHING;
SELECT setval('action_states_state_id_seq', 7);

-- expense_type: 1=CAPEX, 2=Maintenance, 3=OPEX
INSERT INTO expense_type (id, type) VALUES
    (1, 'CAPEX'),
    (2, 'Maintenance'),
    (3, 'OPEX')
ON CONFLICT DO NOTHING;
SELECT setval('expense_type_id_seq', 3);

-- roles: 1=visitor, 10=Responsible, 20=Manager, 77=Admin
INSERT INTO roles (role_id, role_name) VALUES
    (1,  'visitor'),
    (10, 'Responsible'),
    (20, 'Manager'),
    (77, 'Admin')
ON CONFLICT DO NOTHING;
SELECT setval('roles_role_id_seq', 77);

-- permissions (IDs 1-9)
INSERT INTO permissions (permission_id, resource, action) VALUES
    (1, 'dashboard',           'view'),
    (2, 'entries',             'view'),
    (3, 'entries',             'create'),
    (4, 'entries',             'edit'),
    (5, 'entries',             'delete'),
    (6, 'statistics',          'view'),
    (7, 'supervisor_dashboard','view'),
    (8, 'settings',            'view'),
    (9, 'settings',            'edit')
ON CONFLICT DO NOTHING;
SELECT setval('permissions_permission_id_seq', 9);

-- role_permissions
-- visitor (1): dashboard:view, entries:view, statistics:view
INSERT INTO role_permissions (role_id, permission_id) VALUES
    (1, 1),  -- dashboard:view
    (1, 2),  -- entries:view
    (1, 6)   -- statistics:view
ON CONFLICT DO NOTHING;

-- Responsible (10): dashboard:view, entries:view/create/edit, statistics:view
INSERT INTO role_permissions (role_id, permission_id) VALUES
    (10, 1),  -- dashboard:view
    (10, 2),  -- entries:view
    (10, 3),  -- entries:create
    (10, 4),  -- entries:edit
    (10, 6)   -- statistics:view
ON CONFLICT DO NOTHING;

-- Manager (20): all of Responsible + entries:delete, supervisor_dashboard:view
INSERT INTO role_permissions (role_id, permission_id) VALUES
    (20, 1),  -- dashboard:view
    (20, 2),  -- entries:view
    (20, 3),  -- entries:create
    (20, 4),  -- entries:edit
    (20, 5),  -- entries:delete
    (20, 6),  -- statistics:view
    (20, 7)   -- supervisor_dashboard:view
ON CONFLICT DO NOTHING;

-- Admin (77): all permissions
INSERT INTO role_permissions (role_id, permission_id) VALUES
    (77, 1),
    (77, 2),
    (77, 3),
    (77, 4),
    (77, 5),
    (77, 6),
    (77, 7),
    (77, 8),
    (77, 9)
ON CONFLICT DO NOTHING;

-- risk_levels
INSERT INTO risk_levels (risk_level_id, risk_level, severity) VALUES
    (1, 'urgent', 1),
    (2, 'high',   2),
    (3, 'medium', 3),
    (4, 'low',    4)
ON CONFLICT DO NOTHING;
SELECT setval('risk_levels_risk_level_id_seq', 4);

-- categories
INSERT INTO categories (id, category_name) VALUES
    (1, 'mechanical'),
    (2, 'electrical'),
    (3, 'fall'),
    (4, 'emergency'),
    (5, 'lockout'),
    (6, 'chemical'),
    (7, 'ergonomic'),
    (8, 'fire'),
    (9, 'other')
ON CONFLICT DO NOTHING;
SELECT setval('categories_id_seq', 9);

-- lines
INSERT INTO lines (line_id, line_name) VALUES
    (1, 'Production Line A'),
    (2, 'Production Line B'),
    (3, 'Production Line C'),
    (4, 'Assembly Line 1'),
    (5, 'Assembly Line 2'),
    (6, 'Packaging Line')
ON CONFLICT DO NOTHING;
SELECT setval('lines_line_id_seq', 6);
