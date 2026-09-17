-- policies
CREATE TABLE IF NOT EXISTS policies (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    mode TEXT NOT NULL DEFAULT 'observe',
    protection_level TEXT NOT NULL DEFAULT 'balanced',
    paranoia_level INT NOT NULL DEFAULT 1,
    inbound_threshold INT NOT NULL DEFAULT 5,
    outbound_threshold INT NOT NULL DEFAULT 4,
    sql_injection BOOLEAN NOT NULL DEFAULT TRUE,
    xss BOOLEAN NOT NULL DEFAULT TRUE,
    rce BOOLEAN NOT NULL DEFAULT TRUE,
    lfi BOOLEAN NOT NULL DEFAULT TRUE,
    scanner BOOLEAN NOT NULL DEFAULT TRUE,
    protocol_attack BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- sites
CREATE TABLE IF NOT EXISTS sites (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    domain TEXT NOT NULL,
    upstream_host TEXT NOT NULL,
    upstream_port INT NOT NULL,
    protocol TEXT NOT NULL DEFAULT 'http',
    protection_mode TEXT NOT NULL DEFAULT 'observe',
    policy_id UUID NULL REFERENCES policies(id) ON DELETE SET NULL,
    node_group_id UUID NULL,
    status TEXT NOT NULL DEFAULT 'enabled',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_sites_domain ON sites(domain);
CREATE INDEX IF NOT EXISTS idx_sites_policy_id ON sites(policy_id);

-- nodes
CREATE TABLE IF NOT EXISTS nodes (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    ip TEXT NOT NULL,
    region TEXT NOT NULL DEFAULT '',
    node_group TEXT NOT NULL DEFAULT '',
    coraza_version TEXT NOT NULL DEFAULT '',
    crs_version TEXT NOT NULL DEFAULT '',
    config_version TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'unknown',
    last_heartbeat TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_nodes_ip ON nodes(ip);

-- rules
CREATE TABLE IF NOT EXISTS rules (
    id UUID PRIMARY KEY,
    rule_id TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    category TEXT NOT NULL DEFAULT '',
    source TEXT NOT NULL DEFAULT 'custom',
    severity TEXT NOT NULL DEFAULT 'medium',
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    scope TEXT NOT NULL DEFAULT 'global',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_rules_rule_id ON rules(rule_id);
CREATE INDEX IF NOT EXISTS idx_rules_category ON rules(category);

-- auth
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY,
    username TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'enabled',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS roles (
    id UUID PRIMARY KEY,
    code TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS permissions (
    id UUID PRIMARY KEY,
    code TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS user_roles (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, role_id)
);

CREATE TABLE IF NOT EXISTS role_permissions (
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    permission_id UUID NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    PRIMARY KEY (role_id, permission_id)
);

-- audit
CREATE TABLE IF NOT EXISTS audit_logs (
    id UUID PRIMARY KEY,
    actor_id UUID NULL,
    actor_name TEXT NOT NULL DEFAULT '',
    method TEXT NOT NULL,
    path TEXT NOT NULL,
    resource TEXT NOT NULL DEFAULT '',
    resource_id TEXT NOT NULL DEFAULT '',
    status_code INT NOT NULL DEFAULT 0,
    ip TEXT NOT NULL DEFAULT '',
    user_agent TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at ON audit_logs(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_audit_logs_actor_id ON audit_logs(actor_id);
