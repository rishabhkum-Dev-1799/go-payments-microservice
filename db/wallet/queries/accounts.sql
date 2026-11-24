-- name: CreateAccount :one
INSERT INTO accounts (owner_id,name,currency,normal_balance) VALUES ($1,$2,$3,$4) RETURNING *;

-- name: GetAccount :one
SELECT * from accounts WHERE id = $1;

-- name: GetAccountBalance :one
SELECT * from account_balances WHERE account_id = $1;

-- name: ListAccountsByOwner :many
SELECT * from accounts WHERE owner_id = @owner_id::uuid ORDER BY created_at; 