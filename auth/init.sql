CREATE TABLE IF NOT EXISTS auth (
	email VARCHAR(64) PRIMARY KEY,
	refresh_token VARCHAR(32) NOT NULL,
	email_token VARCHAR(32),
    meta_system TEXT,
    created_at TIMESTAMP NOT NULL
);
