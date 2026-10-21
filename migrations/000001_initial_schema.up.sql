-- Reference Tables (no dependencies)

CREATE TABLE categories (
    id SERIAL PRIMARY KEY,
    category_name TEXT NOT NULL UNIQUE
);

CREATE TABLE expense_type (
    id SERIAL PRIMARY KEY,
    type TEXT NOT NULL UNIQUE
);

CREATE TABLE lines (
    line_id SMALLSERIAL PRIMARY KEY,
    line_name TEXT NOT NULL UNIQUE
);

CREATE TABLE risk_levels (
    risk_level_id SMALLSERIAL PRIMARY KEY,
    risk_level TEXT NOT NULL UNIQUE,
    severity INT NOT NULL CHECK (severity BETWEEN 1 AND 5)
);

CREATE TABLE action_states (
    state_id SMALLSERIAL PRIMARY KEY,
    state TEXT NOT NULL UNIQUE,
    is_terminal BOOLEAN NOT NULL DEFAULT FALSE
);

-- RBAC Tables

CREATE TABLE roles (
    role_id SMALLSERIAL PRIMARY KEY,
    role_name TEXT NOT NULL UNIQUE
);

CREATE TABLE permissions (
    permission_id SMALLSERIAL PRIMARY KEY,
    resource TEXT NOT NULL,
    action TEXT NOT NULL,
    UNIQUE (resource, action)
);

CREATE TABLE role_permissions (
    role_id SMALLINT NOT NULL REFERENCES roles(role_id) ON DELETE CASCADE,
    permission_id SMALLINT NOT NULL REFERENCES permissions(permission_id) ON DELETE CASCADE,
    PRIMARY KEY (role_id, permission_id)
);

-- Persons & Preferences

CREATE TABLE persons (
    person_id SMALLSERIAL PRIMARY KEY,
    person_name TEXT,
    person_email TEXT,
    username VARCHAR(30) UNIQUE,
    pw_hash TEXT,
    role_id SMALLINT NOT NULL DEFAULT 1 REFERENCES roles(role_id),
    locked BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE TABLE preferences (
    person_id SMALLINT PRIMARY KEY REFERENCES persons(person_id) ON DELETE CASCADE,
    language CHAR(2) NOT NULL DEFAULT 'en',
    theme BOOLEAN NOT NULL DEFAULT FALSE,
    active_columns JSONB
);

-- Main Risks Table

CREATE TABLE risks (
    id BIGSERIAL PRIMARY KEY,
    risk_type INT NOT NULL DEFAULT 1 CHECK (risk_type IN (1, 2)),
    linked_risk_id BIGINT REFERENCES risks(id) ON DELETE SET NULL,
    creation_date TIMESTAMP NOT NULL DEFAULT NOW(),
    machines TEXT NOT NULL,
    line_id SMALLINT NOT NULL REFERENCES lines(line_id),
    category_id INT REFERENCES categories(id),
    danger_description TEXT NOT NULL,
    risk_level_id SMALLINT NOT NULL REFERENCES risk_levels(risk_level_id),
    action TEXT,
    responsible_id SMALLINT REFERENCES persons(person_id),
    moc_number INT,
    deadline DATE,
    action_state_id SMALLINT NOT NULL REFERENCES action_states(state_id),
    implementation_date DATE,
    entry_modified_at TIMESTAMP NOT NULL DEFAULT NOW(),
    created_by_id SMALLINT NOT NULL REFERENCES persons(person_id),
    committed_at TIMESTAMP,
    cost NUMERIC(10, 2),
    expense_type_id INT REFERENCES expense_type(id)
);

-- Indexes for risks
CREATE INDEX idx_risks_line_id ON risks(line_id);
CREATE INDEX idx_risks_category_id ON risks(category_id);
CREATE INDEX idx_risks_action_state_id ON risks(action_state_id);
CREATE INDEX idx_risks_responsible_id ON risks(responsible_id);
CREATE INDEX idx_risks_created_by_id ON risks(created_by_id);
CREATE INDEX idx_risks_risk_type ON risks(risk_type);
CREATE INDEX idx_risks_deadline ON risks(deadline);
CREATE INDEX idx_risks_creation_date ON risks(creation_date);

-- Child Tables

CREATE TABLE photos (
    risk_id BIGINT NOT NULL REFERENCES risks(id) ON DELETE CASCADE,
    photo_number INT NOT NULL,
    file_path TEXT NOT NULL,
    uploaded_at TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (risk_id, photo_number)
);

CREATE TABLE comments (
    id BIGSERIAL PRIMARY KEY,
    risk_id BIGINT NOT NULL REFERENCES risks(id) ON DELETE CASCADE,
    author_id SMALLINT NOT NULL REFERENCES persons(person_id),
    content TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    edited_at TIMESTAMP
);

CREATE INDEX idx_comments_risk_id ON comments(risk_id);
CREATE INDEX idx_comments_author_id ON comments(author_id);

-- History Tables

CREATE TABLE status_history (
    id BIGSERIAL PRIMARY KEY,
    risk_id BIGINT NOT NULL REFERENCES risks(id) ON DELETE CASCADE,
    old_status SMALLINT NOT NULL REFERENCES action_states(state_id),
    new_status SMALLINT NOT NULL REFERENCES action_states(state_id),
    reason TEXT,
    changed_by SMALLINT NOT NULL REFERENCES persons(person_id),
    changed_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_status_history_risk_id ON status_history(risk_id);

CREATE TABLE deadline_history (
    id BIGSERIAL PRIMARY KEY,
    risk_id BIGINT NOT NULL REFERENCES risks(id) ON DELETE CASCADE,
    old_deadline DATE NOT NULL,
    new_deadline DATE NOT NULL,
    reason TEXT NOT NULL,
    changed_by SMALLINT NOT NULL REFERENCES persons(person_id),
    changed_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_deadline_history_risk_id ON deadline_history(risk_id);

-- Audit Log Table

CREATE TABLE audit_log (
    id BIGSERIAL PRIMARY KEY,
    table_name TEXT NOT NULL,
    record_id BIGINT NOT NULL,
    action TEXT NOT NULL CHECK (action IN ('INSERT', 'UPDATE', 'DELETE')),
    old_values JSONB,
    new_values JSONB,
    changed_by SMALLINT REFERENCES persons(person_id),
    changed_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_audit_log_table_name ON audit_log(table_name);
CREATE INDEX idx_audit_log_record_id ON audit_log(record_id);
CREATE INDEX idx_audit_log_changed_at ON audit_log(changed_at);

-- Audit Trigger Function

CREATE OR REPLACE FUNCTION audit_trigger_fn()
RETURNS TRIGGER AS $$
DECLARE
    v_row JSONB;
    v_id  BIGINT;
BEGIN
    -- Use JSON extraction to find the PK regardless of column name (id vs person_id)
    v_row := CASE TG_OP
        WHEN 'DELETE' THEN row_to_json(OLD)::JSONB
        ELSE               row_to_json(NEW)::JSONB
    END;
    v_id := COALESCE(
        (v_row->>'id')::BIGINT,
        (v_row->>'person_id')::BIGINT
    );
    INSERT INTO audit_log (table_name, record_id, action, old_values, new_values, changed_by, changed_at)
    VALUES (
        TG_TABLE_NAME,
        v_id,
        TG_OP,
        CASE WHEN TG_OP IN ('UPDATE', 'DELETE') THEN row_to_json(OLD) END,
        CASE WHEN TG_OP IN ('INSERT', 'UPDATE') THEN row_to_json(NEW) END,
        NULLIF(current_setting('app.current_user_id', true), '')::SMALLINT,
        NOW()
    );
    RETURN COALESCE(NEW, OLD);
END;
$$ LANGUAGE plpgsql;

-- Attach audit triggers to risks and persons only
CREATE TRIGGER risks_audit_trigger
    AFTER INSERT OR UPDATE OR DELETE ON risks
    FOR EACH ROW EXECUTE FUNCTION audit_trigger_fn();

CREATE TRIGGER persons_audit_trigger
    AFTER INSERT OR UPDATE OR DELETE ON persons
    FOR EACH ROW EXECUTE FUNCTION audit_trigger_fn();

