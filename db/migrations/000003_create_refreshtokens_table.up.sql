CREATE TABLE IF NOT EXISTS refresh_tokens(
    id serial PRIMARY KEY,
    user_id CHAR(36) NOT NULL,
    hashed_refresh_token CHAR(64) NOT NULL,
    status VARCHAR(10)  NOT NULL DEFAULT 'Valid',
    created_at TIMESTAMP NOT NULL,
    expired_at TIMESTAMP NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
)