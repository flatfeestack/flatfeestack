--CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS users (
    id                    UUID PRIMARY KEY,
    invited_id            UUID CONSTRAINT fk_users_id_u REFERENCES users (id), /*if this is set, then this email pays the amount*/
    stripe_id             VARCHAR(255),
    stripe_payment_method VARCHAR(255),
    stripe_last4          VARCHAR(8),
    email                 VARCHAR(64) UNIQUE NOT NULL,
    name                  VARCHAR(255),
    image                 BYTEA,
    seats                 BIGINT DEFAULT 1,
    freq                  BIGINT DEFAULT 365,
    created_at            TIMESTAMP NOT NULL
);
CREATE INDEX IF NOT EXISTS users_email_index ON users(email);
CREATE INDEX IF NOT EXISTS users_invited_id_index ON users(invited_id);

CREATE TABLE IF NOT EXISTS payment_event (
    id                  UUID PRIMARY KEY,
    user_id             UUID CONSTRAINT fk_user_id_ub REFERENCES users (id),
    balance             NUMERIC(78), /*256 bits, to be precise: 259.5 bits */
    daily_spending      NUMERIC(78), /*256 bits, to be precise: 259.5 bits */
    currency            VARCHAR(8) NOT NULL,
    current_status      VARCHAR(16) NOT NULL,
    seats               BIGINT DEFAULT 1,
    freq                BIGINT DEFAULT 365,
    created_at          TIMESTAMP NOT NULL
);
CREATE INDEX IF NOT EXISTS payment_event_user_id_index ON payment_event(user_id);

CREATE TABLE IF NOT EXISTS payment_status (
    id               UUID PRIMARY KEY,
    payment_event_id UUID CONSTRAINT fk_payment_event_id REFERENCES payment_event(id),
    status           VARCHAR(16) NOT NULL
);
CREATE INDEX IF NOT EXISTS payment_status_payment_event_id_index ON payment_status(payment_event_id);

CREATE TABLE IF NOT EXISTS repo (
    id          UUID PRIMARY KEY,
    git_url     VARCHAR(255) UNIQUE NOT NULL,
    name        VARCHAR(255) NOT NULL,
    description TEXT,
    url         VARCHAR(255),
    score       NUMERIC,
    source      VARCHAR(255) NOT NULL,
    created_at  TIMESTAMP NOT NULL
);
CREATE INDEX IF NOT EXISTS repo_name_index ON repo(name);

CREATE TABLE IF NOT EXISTS git_email (
    id           UUID PRIMARY KEY,
    user_id      UUID CONSTRAINT fk_user_id_gub REFERENCES users (id),
    email        VARCHAR(255) UNIQUE NOT NULL,
    token        VARCHAR(32),
    confirmed_at TIMESTAMP,
    created_at   TIMESTAMP NOT NULL
);
CREATE UNIQUE INDEX IF NOT EXISTS git_email_index ON git_email(user_id, email, token);
CREATE INDEX IF NOT EXISTS git_email_user_id_index ON git_email(user_id);

CREATE TABLE IF NOT EXISTS sponsor_event (
    id            UUID PRIMARY KEY,
    repo_id       UUID CONSTRAINT fk_repo_id_se REFERENCES repo (id),
    user_id       UUID CONSTRAINT fk_user_id_se REFERENCES users (id),
    sponsor_at    TIMESTAMP NOT NULL,
    un_sponsor_at TIMESTAMP
);
CREATE UNIQUE INDEX IF NOT EXISTS sponsor_event_index ON sponsor_event(repo_id, user_id, sponsor_at);
CREATE INDEX IF NOT EXISTS sponsor_event_repo_id_index ON sponsor_event(repo_id);
CREATE INDEX IF NOT EXISTS sponsor_event_user_id_index ON sponsor_event(user_id);

CREATE TABLE IF NOT EXISTS analysis_request (
    id          UUID PRIMARY KEY,
    repo_id     UUID CONSTRAINT fk_repo_id_req REFERENCES repo (id),
    date_from   DATE NOT NULL,
    date_to     DATE NOT NULL,
    git_url     VARCHAR(255) NOT NULL,
    received_at TIMESTAMP,
    error       TEXT,
    created_at  TIMESTAMP NOT NULL
);
CREATE UNIQUE INDEX IF NOT EXISTS analysis_request_index ON analysis_request(repo_id, date_from, date_to);
CREATE INDEX IF NOT EXISTS analysis_request_repo_id_index ON analysis_request(repo_id);

CREATE TABLE IF NOT EXISTS analysis_response (
    id                  UUID PRIMARY KEY,
    analysis_request_id UUID CONSTRAINT fk_analysis_request_id_c REFERENCES analysis_request (id),
    git_email           VARCHAR(255) NOT NULL,
    git_names           VARCHAR(255),
    weight              DOUBLE PRECISION NOT NULL,
    created_at          TIMESTAMP NOT NULL
);
CREATE UNIQUE INDEX IF NOT EXISTS analysis_response_index ON analysis_response(analysis_request_id, git_email);
CREATE INDEX IF NOT EXISTS analysis_response_analysis_request_id_index ON analysis_response(analysis_request_id);

CREATE TABLE IF NOT EXISTS daily_contribution (
    id                   UUID PRIMARY KEY,
    user_sponsor_id      UUID CONSTRAINT fk_user_sponsor_id_dc REFERENCES users (id),
    user_contributor_id  UUID CONSTRAINT fk_user_contributor_id_dc REFERENCES users (id),
    repo_id              UUID CONSTRAINT fk_repo_id_dc REFERENCES repo (id),
    balance              NUMERIC(78), /*256 bits*/
    currency             VARCHAR(8) NOT NULL,
    day                  DATE NOT NULL,
    created_at           TIMESTAMP NOT NULL
);
CREATE UNIQUE INDEX IF NOT EXISTS daily_contribution_index ON daily_contribution(user_sponsor_id, user_contributor_id, repo_id, currency, day);
CREATE INDEX IF NOT EXISTS daily_contribution_user_sponsor_id_index ON daily_contribution(user_sponsor_id);
CREATE INDEX IF NOT EXISTS daily_contribution_user_contributor_id_index ON daily_contribution(user_contributor_id);

CREATE TABLE IF NOT EXISTS unclaimed (
    id                   UUID PRIMARY KEY,
    email                VARCHAR(64),
    repo_id              UUID CONSTRAINT fk_repo_id_dc REFERENCES repo (id),
    balance              NUMERIC(78), /*256 bits*/
    currency             VARCHAR(8) NOT NULL,
    day                  DATE NOT NULL,
    created_at           TIMESTAMP NOT NULL
);
CREATE UNIQUE INDEX IF NOT EXISTS unclaimed_index ON unclaimed(email, repo_id, currency);
CREATE INDEX IF NOT EXISTS unclaimed_email_index ON unclaimed(email); /*we do a count on email*/
CREATE INDEX IF NOT EXISTS unclaimed_repo_id_index ON unclaimed(repo_id);

CREATE TABLE IF NOT EXISTS future_contribution (
    id                  UUID PRIMARY KEY,
    user_sponsor_id     UUID CONSTRAINT fk_user_id_fc REFERENCES users (id),
    repo_id             UUID CONSTRAINT fk_repo_id_fc REFERENCES repo (id),
    balance             NUMERIC(78), /*256 bits*/
    currency            VARCHAR(8) NOT NULL,
    day                 DATE NOT NULL,
    created_at          TIMESTAMP NOT NULL
);
CREATE UNIQUE INDEX IF NOT EXISTS future_contribution_index ON future_contribution(user_sponsor_id, repo_id, currency, day);
CREATE INDEX IF NOT EXISTS future_contribution_user_sponsor_id_index ON future_contribution(user_sponsor_id);
CREATE INDEX IF NOT EXISTS future_contribution_repo_id_index ON future_contribution(repo_id);

CREATE TABLE IF NOT EXISTS invite (
    id UUID PRIMARY KEY,
    from_email VARCHAR(64),
    to_email VARCHAR(64),
    confirmed_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL,
    UNIQUE (from_email, to_email)
);
CREATE INDEX IF NOT EXISTS invite_from_email_index ON invite(from_email);
CREATE INDEX IF NOT EXISTS invite_to_email_index ON invite(to_email);

CREATE TABLE IF NOT EXISTS user_emails_sent (
    id         UUID PRIMARY KEY,
    user_id    UUID CONSTRAINT fk_user_id_ub REFERENCES users (id),
    email      VARCHAR(64),
    email_type VARCHAR(64),
    created_at TIMESTAMP NOT NULL
);
CREATE INDEX IF NOT EXISTS user_emails_sent_user_id_index ON user_emails_sent(user_id);
CREATE INDEX IF NOT EXISTS user_emails_sent_email_typeu_index ON user_emails_sent(email_type); /*we do a count on email_type*/
CREATE INDEX IF NOT EXISTS user_emails_sent_email_index ON user_emails_sent(email); /*we do a count on email*/
