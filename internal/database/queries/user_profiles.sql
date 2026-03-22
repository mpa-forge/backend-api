-- name: GetUserProfileByClerkUserID :one
SELECT
    id,
    clerk_user_id,
    email,
    display_name,
    role,
    created_at,
    updated_at
FROM user_profiles
WHERE clerk_user_id = sqlc.arg(clerk_user_id)
LIMIT 1;

-- name: UpsertUserProfile :one
INSERT INTO user_profiles (
    clerk_user_id,
    email,
    display_name,
    role
)
VALUES (
    sqlc.arg(clerk_user_id),
    sqlc.arg(email),
    sqlc.arg(display_name),
    sqlc.arg(role)
)
ON CONFLICT (clerk_user_id) DO UPDATE
SET
    email = EXCLUDED.email,
    display_name = EXCLUDED.display_name,
    role = EXCLUDED.role,
    updated_at = NOW()
RETURNING
    id,
    clerk_user_id,
    email,
    display_name,
    role,
    created_at,
    updated_at;
