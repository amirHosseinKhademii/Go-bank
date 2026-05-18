-- name: CreateUser :one
INSERT INTO users (username, hashed_password, full_name, email)
VALUES ($1, $2, $3, $4)
RETURNING username, full_name, email, password_changed_at, created_at;

-- name: GetUser :one
SELECT username, full_name, email, password_changed_at, created_at FROM users WHERE username = $1;



