-- SQLite does NOT automatically create indexes for foreign key columns, PostgreSQL does,
-- so we create an index also for foreign key columns if they do not exist

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
    multiplier_daily_limit BIGINT DEFAULT -1 /*Multiplier Daily Amount*/
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

CREATE TABLE IF NOT EXISTS trust_event (
    id            UUID PRIMARY KEY,
    repo_id       UUID CONSTRAINT trust_event_repo_id_fk REFERENCES repo(id),
    trust_at      TIMESTAMP NOT NULL,
    un_trust_at   TIMESTAMP,
    UNIQUE(repo_id, trust_at)
);
CREATE INDEX IF NOT EXISTS trust_event_repo_id_idx ON trust_event(repo_id);

CREATE TABLE IF NOT EXISTS multiplier_event (
    id               UUID PRIMARY KEY,
    repo_id          UUID CONSTRAINT multiplier_event_repo_id_fk REFERENCES repo(id),
    user_id          UUID CONSTRAINT multiplier_event_user_id_fk REFERENCES users(id),
    multiplier_at    TIMESTAMP NOT NULL,
    un_multiplier_at TIMESTAMP,
    UNIQUE(repo_id, user_id, multiplier_at)
);
CREATE INDEX IF NOT EXISTS multiplier_event_repo_id_idx ON multiplier_event(repo_id);
CREATE INDEX IF NOT EXISTS multiplier_event_user_id_idx ON multiplier_event(user_id);

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
    repo_id             UUID CONSTRAINT analysis_request_repo_id_fk REFERENCES repo(id),
    git_email           VARCHAR(255) NOT NULL,
    git_names           VARCHAR(255),
    weight              DOUBLE PRECISION NOT NULL,
    created_at          TIMESTAMP NOT NULL,
    UNIQUE(analysis_request_id, git_email)
);
CREATE INDEX IF NOT EXISTS analysis_response_analysis_request_id_idx ON analysis_response(analysis_request_id);
CREATE INDEX IF NOT EXISTS analysis_response_repo_id_idx ON analysis_response(repo_id);

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
    foundation_payment   BOOLEAN,
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
    foundation_payment  BOOLEAN,
    UNIQUE(user_sponsor_id, repo_id, currency, day)
);
CREATE INDEX IF NOT EXISTS future_contribution_user_sponsor_id_idx ON future_contribution(user_sponsor_id);
CREATE INDEX IF NOT EXISTS future_contribution_repo_id_idx ON future_contribution(repo_id);

CREATE TABLE IF NOT EXISTS invite (
    id           UUID PRIMARY KEY,
    from_email   VARCHAR(64),
    to_email     VARCHAR(64),
    confirmed_at TIMESTAMP,
    created_at   TIMESTAMP NOT NULL,
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
CREATE INDEX IF NOT EXISTS user_emails_sent_user_id_idx ON multiplier_event(user_id);
CREATE INDEX IF NOT EXISTS user_emails_sent_email_type_idx ON user_emails_sent(email_type); /*we do a count on email_type*/
CREATE INDEX IF NOT EXISTS user_emails_sent_email_idx ON user_emails_sent(email); /*we do a count on email*/

CREATE TABLE IF NOT EXISTS repo_health_metrics (
    id                    UUID PRIMARY KEY,
    created_at            TIMESTAMP NOT NULL,
    repo_id               UUID CONSTRAINT trust_value_repo_id_fk REFERENCES repo(id),
    contributor_count     NUMERIC(78),
    commit_count          NUMERIC(78),
    sponsor_donation      NUMERIC(78),
    repo_star_count       NUMERIC(78),
    repo_multiplier_count NUMERIC(78),
    active_ffs_user_count NUMERIC(78)
);
CREATE INDEX IF NOT EXISTS repo_health_metrics_repo_id_idx ON daily_contribution(repo_id);

CREATE TABLE IF NOT EXISTS config (
    id         VARCHAR(32) PRIMARY KEY,
    created_at TIMESTAMP,
    values     JSON
);
-- set initial values for threshold
INSERT INTO config (id, created_at, values)
VALUES ('repo_health_threshold',CURRENT_TIMESTAMP,'{
    "contributors":     {"lower": 4,  "upper": 13},
    "commits":          {"lower": 40, "upper": 130},
    "sponsor_donation": {"lower": 1,  "upper": 10},
    "repo_stars":       {"lower": 5,  "upper": 20},
    "repo_multiplier":  {"lower": 1,  "upper": 5},
    "active_ffs_users": {"lower": 1,  "upper": 10}}')
ON CONFLICT (id) DO NOTHING;
