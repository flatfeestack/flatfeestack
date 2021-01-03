CREATE TABLE users (
    id                 UUID PRIMARY KEY,
    stripe_id          VARCHAR(255),
    email              VARCHAR(255) UNIQUE NOT NULL,
    subscription       VARCHAR(255),
    subscription_state VARCHAR(255),
    payout_eth         VARCHAR(255)
);

CREATE TABLE repo (
    id          UUID PRIMARY KEY,
    orig_id     INTEGER,
    orig_from   VARCHAR(255),
    url         VARCHAR(255) UNIQUE NOT NULL,
    name        VARCHAR(255)        NOT NULL,
    description TEXT
);

CREATE TABLE git_email (
    id      UUID PRIMARY KEY,
    user_id UUID CONSTRAINT fk_user_id_gub REFERENCES users (id),
    email   VARCHAR(255) UNIQUE NOT NULL
);

-- if the the uid is null, there is no registered user which owns the git_email
CREATE TABLE git_user (
    id         UUID PRIMARY KEY,
    repo_id    UUID CONSTRAINT fk_repo_id_gub REFERENCES repo (id),
    user_id    UUID CONSTRAINT fk_user_id_gub REFERENCES users (id),
    email      VARCHAR(255) UNIQUE                 NOT NULL,
    balance    NUMERIC                             NOT NULL,
    score      NUMERIC                             NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE TABLE git_user_event (
    id          UUID PRIMARY KEY,
    git_user_id UUID CONSTRAINT fk_git_user_id_gu REFERENCES git_user (id),
    event_type  SMALLINT                            NOT NULL,
    created_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE TABLE sponsor_event (
    id         UUID PRIMARY KEY,
    repo_id    UUID CONSTRAINT fk_repo_id_se REFERENCES repo (id),
    user_id    UUID CONSTRAINT fk_user_id_se REFERENCES users (id),
    event_type SMALLINT                            NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE TABLE repo_balance (
    id         UUID PRIMARY KEY,
    repo_id    UUID CONSTRAINT fk_repo_id_rb REFERENCES repo (id),
    balance    NUMERIC                             NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE TABLE daily_repo_balance (
    id         UUID PRIMARY KEY,
    repo_id    UUID CONSTRAINT fk_repo_id_drb REFERENCES repo (id),
    balance    NUMERIC                             NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE TABLE payments (
    id        UUID PRIMARY KEY,
    user_id   UUID CONSTRAINT fk_user_id_p REFERENCES users (id),
    date_from TIMESTAMP NOT NULL,
    date_to   TIMESTAMP,
    sub       TEXT,
    amount    INTEGER
);

CREATE TABLE analysis_request (
    id        UUID PRIMARY KEY,
    repo_id   UUID CONSTRAINT fk_repo_id_ar REFERENCES repo (id),
    date_from TIMESTAMP NOT NULL,
    date_to   TIMESTAMP NOT NULL,
    branch    TEXT
);

CREATE TABLE contribution (
    id                  UUID PRIMARY KEY,
    analysis_request_id UUID CONSTRAINT fk_analysis_request_id_c REFERENCES analysis_request (id) ON DELETE CASCADE,
    git_email           VARCHAR(255)     NOT NULL,
    computed_at         TIMESTAMP        NOT NULL,
    weight              DOUBLE PRECISION NOT NULL
);

CREATE TABLE repo_balance_event (
    id         UUID PRIMARY KEY,
    repo_id    UUID CONSTRAINT fk_repo_id_rbe REFERENCES repo (id) ON DELETE CASCADE,
    event_type SMALLINT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE TABLE exchange (
    id         UUID PRIMARY KEY,
    amount     NUMERIC,
    chain_id   VARCHAR(255)                        NOT NULL,
    price      NUMERIC,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE TABLE payouts (
    id          UUID PRIMARY KEY,
    user_id     UUID CONSTRAINT fk_user_id_pay REFERENCES users (id) ON DELETE CASCADE,
    exchange_id UUID CONSTRAINT fk_exchange_id_pay REFERENCES exchange (id) ON DELETE CASCADE,
    amount      NUMERIC,
    paid        TIMESTAMP,
    created_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);
