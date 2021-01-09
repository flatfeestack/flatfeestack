CREATE TABLE users (
    id                 UUID PRIMARY KEY,
    stripe_id          VARCHAR(255),
    email              VARCHAR(255) UNIQUE NOT NULL,
    subscription       VARCHAR(255),
    subscription_state VARCHAR(255),
    payout_eth         VARCHAR(255),
    created_at         TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE TABLE repo (
    id          UUID PRIMARY KEY,
    orig_id     NUMERIC,
    url         VARCHAR(255) UNIQUE NOT NULL,
    name        VARCHAR(255)        NOT NULL,
    description TEXT,
    created_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE TABLE git_email (
    id         UUID PRIMARY KEY,
    user_id    UUID CONSTRAINT fk_user_id_gub REFERENCES users (id),
    email      VARCHAR(255) UNIQUE NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE TABLE sponsor_event (
    id         UUID PRIMARY KEY,
    repo_id    UUID CONSTRAINT fk_repo_id_se REFERENCES repo (id),
    user_id    UUID CONSTRAINT fk_user_id_se REFERENCES users (id),
    event_type SMALLINT                            NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE UNIQUE INDEX sponsor_event_index ON sponsor_event(repo_id, user_id, created_at);

CREATE TABLE payments (
    id         UUID PRIMARY KEY,
    user_id    UUID CONSTRAINT fk_user_id_p REFERENCES users (id),
    date_from  TIMESTAMP NOT NULL,
    date_to    TIMESTAMP,
    sub        TEXT,
    amount     BIGINT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE TABLE analysis_request (
    id         UUID PRIMARY KEY,
    repo_id    UUID CONSTRAINT fk_repo_id_ar REFERENCES repo (id),
    date_from  TIMESTAMP NOT NULL,
    date_to    TIMESTAMP NOT NULL,
    branch     TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE TABLE exchange (
    id         UUID PRIMARY KEY,
    chain      VARCHAR(255) NOT NULL,
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
