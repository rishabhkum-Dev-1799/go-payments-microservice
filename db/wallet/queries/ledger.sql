-- name: CreateTransaction :one
INSERT INTO transactions (description) VALUES ($1) RETURNING *;

-- name: CreateEntry :one 
INSERT INTO entries (transaction_id,account_id,direction,amount,currency)
VALUES ($1,$2,$3,$4,$5)
RETURNING *;

-- name: ListEntriesByAccount :many
SELECT * from entries 
WHERE account_id = @account_id
    AND id < @before_id
ORDER BY id DESC
LIMIT @page_size;