CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE users (
    id                 UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    sponsor_id         UUID CONSTRAINT fk_user_id_uid REFERENCES users (id),
    stripe_id          VARCHAR(255),
    email              VARCHAR(255) UNIQUE NOT NULL,
    name               VARCHAR(255),
    image              BYTEA,
    subscription       VARCHAR(255),
    subscription_state VARCHAR(255),
    payout_eth         VARCHAR(255),
    seats              INTEGER,
    token              VARCHAR(64) NOT NULL,
    role               VARCHAR(3) DEFAULT 'USR' NOT NULL,
    created_at         TIMESTAMP NOT NULL
);

CREATE TABLE repo (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    orig_id     NUMERIC,
    url         VARCHAR(255) UNIQUE NOT NULL,
    git_url     VARCHAR(255) UNIQUE NOT NULL,
    branch      VARCHAR(16)         NOT NULL,
    name        VARCHAR(255)        NOT NULL,
    description TEXT,
    tags        BYTEA,
    score       NUMERIC,
    source      VARCHAR(255)        NOT NULL,
    created_at  TIMESTAMP           NOT NULL
);

CREATE TABLE git_email (
    id           UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id      UUID CONSTRAINT fk_user_id_gub REFERENCES users (id),
    email        VARCHAR(255) UNIQUE NOT NULL,
    token        VARCHAR(32),
    confirmed_at TIMESTAMP,
    created_at   TIMESTAMP NOT NULL
);

CREATE TABLE sponsor_event (
    id           UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    repo_id      UUID CONSTRAINT fk_repo_id_se REFERENCES repo (id),
    user_id      UUID CONSTRAINT fk_user_id_se REFERENCES users (id),
    sponsor_at   TIMESTAMP NOT NULL,
    unsponsor_at TIMESTAMP DEFAULT to_date('9999', 'YYYY') NOT NULL
);
CREATE UNIQUE INDEX sponsor_event_index ON sponsor_event(repo_id, user_id, sponsor_at);

CREATE TABLE payments (
    id         UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id    UUID CONSTRAINT fk_user_id_p REFERENCES users (id),
    date_from  DATE NOT NULL,
    date_to    DATE,
    sub        TEXT,
    amount     BIGINT,
    created_at TIMESTAMP NOT NULL
);

CREATE TABLE analysis_request (
    id         UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    repo_id    UUID CONSTRAINT fk_repo_id_req REFERENCES repo (id),
    date_from  DATE NOT NULL,
    date_to    DATE NOT NULL,
    branch     TEXT,
    created_at TIMESTAMP NOT NULL
);
CREATE UNIQUE INDEX analysis_request_index ON analysis_request(repo_id, date_from, date_to);

CREATE TABLE analysis_response (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    analysis_request_id UUID CONSTRAINT fk_analysis_request_id_c REFERENCES analysis_request (id),
    git_email           VARCHAR(255)     NOT NULL,
    weight              DOUBLE PRECISION NOT NULL,
    created_at          TIMESTAMP NOT NULL
);

CREATE TABLE daily_repo_hours (
    id         UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id    UUID CONSTRAINT fk_user_id_drh REFERENCES users (id),
    repo_hours INTEGER                             NOT NULL,
    day        DATE                                NOT NULL,
    created_at TIMESTAMP                           NOT NULL
);
CREATE UNIQUE INDEX daily_repo_hours_index ON daily_repo_hours(user_id, day);

CREATE TABLE daily_repo_balance (
    id         UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    repo_id    UUID CONSTRAINT fk_repo_id_drb REFERENCES repo (id),
    balance    BIGINT                              NOT NULL,
    day        DATE                                NOT NULL,
    created_at TIMESTAMP                           NOT NULL
);
CREATE UNIQUE INDEX daily_repo_balance_index ON daily_repo_balance(repo_id, day);

CREATE TABLE daily_repo_weight (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    repo_id     UUID CONSTRAINT fk_repo_id_drw REFERENCES repo (id),
    weight      DOUBLE PRECISION                    NOT NULL,
    day         DATE                                NOT NULL,
    created_at  TIMESTAMP                           NOT NULL
);
CREATE UNIQUE INDEX daily_repo_weight_index ON daily_repo_weight(repo_id, day);

CREATE TABLE daily_email_payout (
    id         UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email      VARCHAR(255)                        NOT NULL,
    balance    BIGINT                              NOT NULL,
    day        DATE                                NOT NULL,
    created_at TIMESTAMP                           NOT NULL
);
CREATE UNIQUE INDEX daily_email_payout_index ON daily_email_payout(email, day);

CREATE TABLE daily_user_payout (
    id         UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id    UUID CONSTRAINT fk_user_id_dup REFERENCES users (id),
    balance    BIGINT                              NOT NULL,
    day        DATE                                NOT NULL,
    created_at TIMESTAMP                           NOT NULL
);
CREATE UNIQUE INDEX daily_user_payout_index ON daily_user_payout(user_id, day);

CREATE TABLE daily_future_leftover (
    id         UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    repo_id    UUID CONSTRAINT fk_repo_id_dfl REFERENCES repo (id),
    balance    BIGINT                              NOT NULL,
    day        DATE                                NOT NULL,
    created_at TIMESTAMP                           NOT NULL
);
CREATE UNIQUE INDEX daily_future_leftover_index ON daily_future_leftover(repo_id, day);

CREATE TABLE payouts_request (
    id                     UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    daily_user_payout_id   UUID CONSTRAINT fk_daily_user_payout_id_pay REFERENCES daily_user_payout (id),
    batch_id               UUID                   NOT NULL,
    exchange_rate          NUMERIC                NOT NULL,
    created_at             TIMESTAMP              NOT NULL
);
CREATE INDEX payouts_index ON payouts_request(batch_id);

CREATE TABLE payouts_response (
    id         UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    batch_id   UUID UNIQUE NOT NULL,
    tx_hash    VARCHAR(66),
    error      TEXT,
    created_at TIMESTAMP NOT NULL
);

CREATE TABLE payouts_response_details (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    payouts_response_id UUID CONSTRAINT fk_payouts_response_id_pres REFERENCES payouts_response (id),
    address             VARCHAR(42),
    balance_wei         NUMERIC NOT NULL,
    created_at          TIMESTAMP NOT NULL
);
