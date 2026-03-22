INSERT INTO user_profiles (external_user_id, email, display_name, role)
VALUES
    ('seed_user_default', 'seed.user@example.com', 'Seed User', 'user'),
    ('seed_admin_default', 'seed.admin@example.com', 'Seed Admin', 'admin')
ON CONFLICT (external_user_id) DO UPDATE
SET
    email = EXCLUDED.email,
    display_name = EXCLUDED.display_name,
    role = EXCLUDED.role,
    updated_at = NOW();
