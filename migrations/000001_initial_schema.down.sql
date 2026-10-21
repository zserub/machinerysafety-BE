-- Drop triggers first
DROP TRIGGER IF EXISTS risks_audit_trigger ON risks;
DROP TRIGGER IF EXISTS persons_audit_trigger ON persons;
DROP FUNCTION IF EXISTS audit_trigger_fn();

-- Drop tables in reverse dependency order
DROP TABLE IF EXISTS audit_log;
DROP TABLE IF EXISTS deadline_history;
DROP TABLE IF EXISTS status_history;
DROP TABLE IF EXISTS comments;
DROP TABLE IF EXISTS photos;
DROP TABLE IF EXISTS risks;
DROP TABLE IF EXISTS preferences;
DROP TABLE IF EXISTS persons;
DROP TABLE IF EXISTS role_permissions;
DROP TABLE IF EXISTS permissions;
DROP TABLE IF EXISTS roles;
DROP TABLE IF EXISTS action_states;
DROP TABLE IF EXISTS risk_levels;
DROP TABLE IF EXISTS lines;
DROP TABLE IF EXISTS expense_type;
DROP TABLE IF EXISTS categories;
