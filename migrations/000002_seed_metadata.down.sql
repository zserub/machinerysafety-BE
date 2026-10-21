-- Remove all data seeded by 000002_seed_metadata.up.sql
-- Order matters: child tables first, then parent tables

DELETE FROM role_permissions;
DELETE FROM permissions;
DELETE FROM roles;
DELETE FROM expense_type;
DELETE FROM action_states;
DELETE FROM lines;
DELETE FROM categories;
DELETE FROM risk_levels;
