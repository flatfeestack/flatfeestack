--CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS users (
    id                     UUID PRIMARY KEY,
    invited_id             UUID CONSTRAINT users_id_fk REFERENCES users(id), /*if this is set, then this email pays the amount*/
    stripe_id              VARCHAR(255),
    stripe_payment_method  VARCHAR(255),
    stripe_last4           VARCHAR(8),
    stripe_client_secret   VARCHAR(64),
    email                  VARCHAR(64) UNIQUE NOT NULL,
    name                   VARCHAR(255),
    image                  BYTEA,
    seats                  BIGINT DEFAULT 1,
    freq                   BIGINT DEFAULT 365,
    created_at             TIMESTAMP NOT NULL,
    multiplier             BOOLEAN DEFAULT FALSE, /*Multiplier set*/
    multiplier_daily_limit NUMERIC(78) DEFAULT 100 /*Multiplier Daily Amount*/
);
CREATE INDEX IF NOT EXISTS users_email_idx ON users(email); /*we search for emails*/
CREATE INDEX IF NOT EXISTS users_invited_id_idx ON users(invited_id);

CREATE TABLE IF NOT EXISTS payment_in_event (
    id                  UUID PRIMARY KEY,
    external_id         UUID NOT NULL,
    user_id             UUID CONSTRAINT payment_in_event_user_id_ub_fk REFERENCES users(id),
    balance             NUMERIC(78), /*256 bits, to be precise: 259.5 bits */
    currency            VARCHAR(8) NOT NULL,
    status              VARCHAR(16) NOT NULL,
    seats               BIGINT DEFAULT 1,
    freq                BIGINT DEFAULT 365,
    created_at          TIMESTAMP NOT NULL,
    UNIQUE(status, external_id)
);
CREATE INDEX IF NOT EXISTS payment_in_event_user_id_idx ON payment_in_event(user_id);
CREATE INDEX IF NOT EXISTS payment_in_event_currency_idx ON payment_in_event(currency);
CREATE INDEX IF NOT EXISTS payment_in_event_status_idx ON payment_in_event(status);
CREATE INDEX IF NOT EXISTS payment_in_event_created_at_idx ON payment_in_event(created_at);

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
CREATE INDEX IF NOT EXISTS repo_name_idx ON repo(name);

CREATE TABLE IF NOT EXISTS git_email (
    id           UUID PRIMARY KEY,
    user_id      UUID CONSTRAINT git_email_user_id_fk REFERENCES users(id),
    email        VARCHAR(255) NOT NULL,
    token        VARCHAR(32),
    confirmed_at TIMESTAMP,
    created_at   TIMESTAMP NOT NULL,
    UNIQUE(user_id, email, token)
);
CREATE INDEX IF NOT EXISTS git_email_user_id_idx ON git_email(user_id);

CREATE TABLE IF NOT EXISTS sponsor_event (
    id            UUID PRIMARY KEY,
    repo_id       UUID CONSTRAINT sponsor_event_repo_id_fk REFERENCES repo(id),
    user_id       UUID CONSTRAINT sponsor_event_user_id_fk REFERENCES users(id),
    sponsor_at    TIMESTAMP NOT NULL,
    un_sponsor_at TIMESTAMP,
    UNIQUE(repo_id, user_id, sponsor_at)
);
CREATE INDEX IF NOT EXISTS sponsor_event_repo_id_idx ON sponsor_event(repo_id);
CREATE INDEX IF NOT EXISTS sponsor_event_user_id_idx ON sponsor_event(user_id);

CREATE TABLE IF NOT EXISTS analysis_request (
    id          UUID PRIMARY KEY,
    repo_id     UUID CONSTRAINT analysis_request_repo_id_fk REFERENCES repo(id),
    date_from   DATE NOT NULL,
    date_to     DATE NOT NULL,
    git_url     VARCHAR(255) NOT NULL,
    received_at TIMESTAMP,
    error       TEXT,
    created_at  TIMESTAMP NOT NULL,
    UNIQUE(repo_id, date_from, date_to)
);
CREATE INDEX IF NOT EXISTS analysis_request_repo_id_idx ON analysis_request(repo_id);

CREATE TABLE IF NOT EXISTS analysis_response (
    id                  UUID PRIMARY KEY,
    analysis_request_id UUID CONSTRAINT analysis_response_analysis_request_id_fk REFERENCES analysis_request(id),
    git_email           VARCHAR(255) NOT NULL,
    git_names           VARCHAR(255),
    weight              DOUBLE PRECISION NOT NULL,
    created_at          TIMESTAMP NOT NULL,
    UNIQUE(analysis_request_id, git_email)
);
CREATE INDEX IF NOT EXISTS analysis_response_analysis_request_id_idx ON analysis_response(analysis_request_id);

CREATE TABLE IF NOT EXISTS daily_contribution (
    id                   UUID PRIMARY KEY,
    user_sponsor_id      UUID CONSTRAINT daily_contribution_user_sponsor_id_fk REFERENCES users(id),
    user_contributor_id  UUID CONSTRAINT daily_contribution_user_contributor_id_fk REFERENCES users(id),
    repo_id              UUID CONSTRAINT daily_contribution_repo_id_fk REFERENCES repo(id),
    balance              NUMERIC(78), /*256 bits*/
    currency             VARCHAR(8) NOT NULL,
    day                  DATE NOT NULL,
    created_at           TIMESTAMP NOT NULL,
    claimed_at           TIMESTAMP,
    UNIQUE(user_sponsor_id, user_contributor_id, repo_id, day)
);
CREATE INDEX IF NOT EXISTS daily_contribution_user_sponsor_id_idx ON daily_contribution(user_sponsor_id);
CREATE INDEX IF NOT EXISTS daily_contribution_user_contributor_id_idx ON daily_contribution(user_contributor_id);
CREATE INDEX IF NOT EXISTS daily_contribution_repo_id_idx ON daily_contribution(repo_id);

CREATE TABLE IF NOT EXISTS unclaimed (
    id                   UUID PRIMARY KEY,
    email                VARCHAR(64),
    repo_id              UUID CONSTRAINT unclaimed_repo_id_fk REFERENCES repo(id),
    balance              NUMERIC(78), /*256 bits*/
    currency             VARCHAR(8) NOT NULL,
    day                  DATE NOT NULL,
    created_at           TIMESTAMP NOT NULL,
    UNIQUE(email, repo_id, currency, day)
);
CREATE INDEX IF NOT EXISTS unclaimed_email_idx ON unclaimed(email); /*we do a count on email*/
CREATE INDEX IF NOT EXISTS unclaimed_repo_id_idx ON unclaimed(repo_id);

CREATE TABLE IF NOT EXISTS future_contribution (
    id                  UUID PRIMARY KEY,
    user_sponsor_id     UUID CONSTRAINT future_contribution_user_sponsor_id_fk REFERENCES users(id),
    repo_id             UUID CONSTRAINT future_contribution_repo_id_fk REFERENCES repo(id),
    balance             NUMERIC(78), /*256 bits*/
    currency            VARCHAR(8) NOT NULL,
    day                 DATE NOT NULL,
    created_at          TIMESTAMP NOT NULL,
    UNIQUE(user_sponsor_id, repo_id, currency, day)
);
CREATE INDEX IF NOT EXISTS future_contribution_user_sponsor_id_idx ON future_contribution(user_sponsor_id);
CREATE INDEX IF NOT EXISTS future_contribution_repo_id_idx ON future_contribution(repo_id);

CREATE TABLE IF NOT EXISTS invite (
    id UUID PRIMARY KEY,
    from_email VARCHAR(64),
    to_email VARCHAR(64),
    confirmed_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL,
    UNIQUE (from_email, to_email)
);
CREATE INDEX IF NOT EXISTS invite_from_email_idx ON invite(from_email);
CREATE INDEX IF NOT EXISTS invite_to_email_idx ON invite(to_email);

CREATE TABLE IF NOT EXISTS user_emails_sent (
    id         UUID PRIMARY KEY,
    user_id    UUID CONSTRAINT user_emails_sent_user_id_fk REFERENCES users(id),
    email      VARCHAR(64),
    email_type VARCHAR(64),
    created_at TIMESTAMP NOT NULL
);
CREATE INDEX IF NOT EXISTS user_emails_sent_user_id_idx ON user_emails_sent(user_id);
CREATE INDEX IF NOT EXISTS user_emails_sent_email_type_idx ON user_emails_sent(email_type); /*we do a count on email_type*/
CREATE INDEX IF NOT EXISTS user_emails_sent_email_idx ON user_emails_sent(email); /*we do a count on email*/

CREATE TYPE trust_value_threshold_bound AS (
    lower_bound INTEGER,
    upper_bound INTEGER
);

CREATE TABLE IF NOT EXISTS trust_value_treshold (
    id                    UUID PRIMARY KEY,
    th_contributer_count  range_values trust_value_threshold_bound CHECK ((range_values).lower_bound <= (range_values).upper_bound),
    th_commit_count       range_values trust_value_threshold_bound CHECK ((range_values).lower_bound <= (range_values).upper_bound),
    th_metric3            range_values trust_value_threshold_bound CHECK ((range_values).lower_bound <= (range_values).upper_bound),
    th_metric4            range_values trust_value_threshold_bound CHECK ((range_values).lower_bound <= (range_values).upper_bound),
    th_metric5            range_values trust_value_threshold_bound CHECK ((range_values).lower_bound <= (range_values).upper_bound)
)

CREATE TABLE IF NOT EXISTS trust_value (
    id                  UUID PRIMARY KEY,
    repo_id             UUID CONSTRAINT trust_value_repo_id_fk REFERENCES repo(id),
    created_at          TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    contributer_count   DECIMAL(3,2) CHECK (rating >= 0.00 AND rating <= 2.00),
    commit_count        DECIMAL(3,2) CHECK (rating >= 0.00 AND rating <= 2.00),
    metric_3            DECIMAL(3,2) CHECK (rating >= 0.00 AND rating <= 2.00),
    metric_4            DECIMAL(3,2) CHECK (rating >= 0.00 AND rating <= 2.00),
    metric_5            DECIMAL(3,2) CHECK (rating >= 0.00 AND rating <= 2.00)
);