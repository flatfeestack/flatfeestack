-- PostgreSQL migration for golang-migrate

CREATE TABLE IF NOT EXISTS users (
    id                     UUID PRIMARY KEY,
    invited_id             UUID REFERENCES users(id) ON DELETE SET NULL,
    stripe_id              VARCHAR(255),
    stripe_payment_method  VARCHAR(255),
    stripe_last4           VARCHAR(8),
    stripe_client_secret   VARCHAR(64),
    email                  VARCHAR(64) UNIQUE NOT NULL,
    name                   VARCHAR(255),
    image                  BYTEA,
    seats                  INTEGER DEFAULT 1,
    freq                   INTEGER DEFAULT 365,
    created_at             TIMESTAMPTZ NOT NULL,
    multiplier             BOOLEAN DEFAULT FALSE,
    multiplier_daily_limit BIGINT DEFAULT -1
);
CREATE INDEX IF NOT EXISTS users_invited_id_idx ON users(invited_id);

CREATE TABLE IF NOT EXISTS payment_in_event (
    id          UUID PRIMARY KEY,
    external_id UUID NOT NULL,
    user_id     UUID REFERENCES users(id) ON DELETE CASCADE,
    balance     NUMERIC(78),
    currency    VARCHAR(8) NOT NULL,
    status      VARCHAR(16) NOT NULL,
    seats       INTEGER DEFAULT 1,
    freq        INTEGER DEFAULT 365,
    created_at  TIMESTAMPTZ NOT NULL,
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
    source      VARCHAR(255) NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL
);
CREATE INDEX IF NOT EXISTS repo_name_idx ON repo(name);

CREATE TABLE IF NOT EXISTS git_email (
    id           UUID PRIMARY KEY,
    user_id      UUID REFERENCES users(id) ON DELETE CASCADE,
    email        VARCHAR(255) NOT NULL,
    token        VARCHAR(32),
    confirmed_at TIMESTAMPTZ,
    created_at   TIMESTAMPTZ NOT NULL,
    UNIQUE(user_id, email, token)
);
CREATE INDEX IF NOT EXISTS git_email_user_id_idx ON git_email(user_id);

CREATE TABLE IF NOT EXISTS sponsor_event (
    id            UUID PRIMARY KEY,
    repo_id       UUID REFERENCES repo(id) ON DELETE CASCADE,
    user_id       UUID REFERENCES users(id) ON DELETE CASCADE,
    sponsor_at    TIMESTAMPTZ NOT NULL,
    un_sponsor_at TIMESTAMPTZ,
    UNIQUE(repo_id, user_id, sponsor_at)
);
CREATE INDEX IF NOT EXISTS sponsor_event_repo_id_idx ON sponsor_event(repo_id);
CREATE INDEX IF NOT EXISTS sponsor_event_user_id_idx ON sponsor_event(user_id);

CREATE TABLE IF NOT EXISTS trust_event (
    id          UUID PRIMARY KEY,
    repo_id     UUID REFERENCES repo(id) ON DELETE CASCADE,
    trust_at    TIMESTAMPTZ NOT NULL,
    un_trust_at TIMESTAMPTZ,
    UNIQUE(repo_id, trust_at)
);
CREATE INDEX IF NOT EXISTS trust_event_repo_id_idx ON trust_event(repo_id);

CREATE TABLE IF NOT EXISTS multiplier_event (
    id               UUID PRIMARY KEY,
    repo_id          UUID REFERENCES repo(id) ON DELETE CASCADE,
    user_id          UUID REFERENCES users(id) ON DELETE CASCADE,
    multiplier_at    TIMESTAMPTZ NOT NULL,
    un_multiplier_at TIMESTAMPTZ,
    UNIQUE(repo_id, user_id, multiplier_at)
);
CREATE INDEX IF NOT EXISTS multiplier_event_repo_id_idx ON multiplier_event(repo_id);
CREATE INDEX IF NOT EXISTS multiplier_event_user_id_idx ON multiplier_event(user_id);

CREATE TABLE IF NOT EXISTS analysis_request (
    id          UUID PRIMARY KEY,
    repo_id     UUID REFERENCES repo(id) ON DELETE CASCADE,
    date_from   DATE NOT NULL,
    date_to     DATE NOT NULL,
    git_url     VARCHAR(255) NOT NULL,
    received_at TIMESTAMPTZ,
    error       TEXT,
    created_at  TIMESTAMPTZ NOT NULL,
    UNIQUE(repo_id, date_from, date_to)
);
CREATE INDEX IF NOT EXISTS analysis_request_repo_id_idx ON analysis_request(repo_id);

CREATE TABLE IF NOT EXISTS repo_metrics (
    id                         UUID PRIMARY KEY,
    created_at                 TIMESTAMPTZ NOT NULL,
    analysis_request_id        UUID REFERENCES analysis_request(id) ON DELETE CASCADE,
    repo_id                    UUID REFERENCES repo(id) ON DELETE CASCADE,
    
    -- Period definition
    period_start               DATE NOT NULL,
    period_end                 DATE NOT NULL,
    
    -- Contributor data (from analysis_response)
    git_email                  VARCHAR(255),
    git_names                  VARCHAR(255),
    weight                     DOUBLE PRECISION,
    
    -- Incremental metrics (this period only)
    commit_count               INTEGER,
    merge_commit_count         INTEGER,
    external_merge_count       INTEGER,
    branch_merge_count         INTEGER,
    unique_contributors        INTEGER,
    new_contributor_count      INTEGER,
    lines_added                BIGINT,
    lines_deleted              BIGINT,
    files_changed              INTEGER,
    weekend_commit_count       INTEGER,
    
    -- Code quality indicators
    test_file_changes          INTEGER,
    documentation_changes      INTEGER,
    large_commit_count         INTEGER,
    revert_commit_count        INTEGER,
    average_commit_msg_length  INTEGER,
    
    -- Collaboration metrics
    email_domain_count         INTEGER,
    timezone_count             INTEGER,
    co_author_count            INTEGER,
    
    -- Snapshot metrics (point-in-time at period_end)
    sponsor_donation           INTEGER,
    repo_star_count            INTEGER,
    repo_multiplier_count      INTEGER,
    active_ffs_user_count      INTEGER,
    active_contributors        INTEGER,
    total_contributors         INTEGER,
    bus_factor                 INTEGER,
    days_since_last_commit     INTEGER,
    
    UNIQUE(repo_id, period_start, period_end, git_email)
);
CREATE INDEX IF NOT EXISTS repo_metrics_analysis_request_id_idx ON repo_metrics(analysis_request_id);
CREATE INDEX IF NOT EXISTS repo_metrics_repo_id_idx ON repo_metrics(repo_id);
CREATE INDEX IF NOT EXISTS repo_metrics_period_idx ON repo_metrics(period_start, period_end);

CREATE TABLE IF NOT EXISTS daily_contribution (
    id                  UUID PRIMARY KEY,
    user_sponsor_id     UUID REFERENCES users(id) ON DELETE CASCADE,
    user_contributor_id UUID REFERENCES users(id) ON DELETE CASCADE,
    repo_id             UUID REFERENCES repo(id) ON DELETE CASCADE,
    balance             NUMERIC(78),
    currency            VARCHAR(8) NOT NULL,
    day                 DATE NOT NULL,
    created_at          TIMESTAMPTZ NOT NULL,
    claimed_at          TIMESTAMPTZ,
    foundation_payment  BOOLEAN,
    UNIQUE(user_sponsor_id, user_contributor_id, repo_id, day)
);
CREATE INDEX IF NOT EXISTS daily_contribution_user_sponsor_id_idx ON daily_contribution(user_sponsor_id);
CREATE INDEX IF NOT EXISTS daily_contribution_user_contributor_id_idx ON daily_contribution(user_contributor_id);
CREATE INDEX IF NOT EXISTS daily_contribution_repo_id_idx ON daily_contribution(repo_id);

CREATE TABLE IF NOT EXISTS unclaimed (
    id         UUID PRIMARY KEY,
    email      VARCHAR(64),
    repo_id    UUID REFERENCES repo(id) ON DELETE CASCADE,
    balance    NUMERIC(78),
    currency   VARCHAR(8) NOT NULL,
    day        DATE NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    UNIQUE(email, repo_id, currency, day)
);
CREATE INDEX IF NOT EXISTS unclaimed_email_idx ON unclaimed(email);
CREATE INDEX IF NOT EXISTS unclaimed_repo_id_idx ON unclaimed(repo_id);

CREATE TABLE IF NOT EXISTS future_contribution (
    id                 UUID PRIMARY KEY,
    user_sponsor_id    UUID REFERENCES users(id) ON DELETE CASCADE,
    repo_id            UUID REFERENCES repo(id) ON DELETE CASCADE,
    balance            NUMERIC(78),
    currency           VARCHAR(8) NOT NULL,
    day                DATE NOT NULL,
    created_at         TIMESTAMPTZ NOT NULL,
    foundation_payment BOOLEAN,
    UNIQUE(user_sponsor_id, repo_id, currency, day)
);
CREATE INDEX IF NOT EXISTS future_contribution_user_sponsor_id_idx ON future_contribution(user_sponsor_id);
CREATE INDEX IF NOT EXISTS future_contribution_repo_id_idx ON future_contribution(repo_id);

CREATE TABLE IF NOT EXISTS invite (
    id           UUID PRIMARY KEY,
    from_email   VARCHAR(64),
    to_email     VARCHAR(64),
    confirmed_at TIMESTAMPTZ,
    created_at   TIMESTAMPTZ NOT NULL,
    UNIQUE(from_email, to_email)
);
CREATE INDEX IF NOT EXISTS invite_from_email_idx ON invite(from_email);
CREATE INDEX IF NOT EXISTS invite_to_email_idx ON invite(to_email);

CREATE TABLE IF NOT EXISTS user_emails_sent (
    id         UUID PRIMARY KEY,
    user_id    UUID REFERENCES users(id) ON DELETE CASCADE,
    email      VARCHAR(64),
    email_type VARCHAR(64),
    created_at TIMESTAMPTZ NOT NULL
);
CREATE INDEX IF NOT EXISTS user_emails_sent_user_id_idx ON user_emails_sent(user_id);
CREATE INDEX IF NOT EXISTS user_emails_sent_email_type_idx ON user_emails_sent(email_type);
CREATE INDEX IF NOT EXISTS user_emails_sent_email_idx ON user_emails_sent(email);