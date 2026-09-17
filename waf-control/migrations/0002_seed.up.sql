-- Seed permissions and admin role (no user password here — seeded in Go on boot)
INSERT INTO permissions (id, code, name) VALUES
    ('a1000001-0000-4000-8000-000000000001', 'site:read', '站点查看'),
    ('a1000001-0000-4000-8000-000000000002', 'site:write', '站点管理'),
    ('a1000001-0000-4000-8000-000000000003', 'node:read', '节点查看'),
    ('a1000001-0000-4000-8000-000000000004', 'node:write', '节点管理'),
    ('a1000001-0000-4000-8000-000000000005', 'policy:read', '策略查看'),
    ('a1000001-0000-4000-8000-000000000006', 'policy:write', '策略管理'),
    ('a1000001-0000-4000-8000-000000000007', 'rule:read', '规则查看'),
    ('a1000001-0000-4000-8000-000000000008', 'rule:write', '规则管理'),
    ('a1000001-0000-4000-8000-000000000009', 'user:read', '用户查看'),
    ('a1000001-0000-4000-8000-00000000000a', 'user:write', '用户管理'),
    ('a1000001-0000-4000-8000-00000000000b', 'audit:read', '审计查看')
ON CONFLICT (code) DO NOTHING;

INSERT INTO roles (id, code, name, created_at) VALUES
    ('b1000001-0000-4000-8000-000000000001', 'admin', '系统管理员', NOW())
ON CONFLICT (code) DO NOTHING;

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
CROSS JOIN permissions p
WHERE r.code = 'admin'
ON CONFLICT DO NOTHING;
