-- Attention: any word that contains public

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE users (
    id                    UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    sponsor_id            UUID CONSTRAINT fk_sponsor_id_users REFERENCES users (id),
    invited_email         VARCHAR(64), /*if this is set, then this email pays the amount*/
    stripe_id             VARCHAR(255),
    stripe_payment_method VARCHAR(255),
    stripe_last4          VARCHAR(4),
    payment_cycle_id      UUID, /*CONSTRAINT fk_payment_cycle_id_u REFERENCES payment_cycle (id)*/
    email                 VARCHAR(64) UNIQUE NOT NULL,
    name                  VARCHAR(255),
    image                 BYTEA,
    created_at            TIMESTAMP NOT NULL
);

CREATE TABLE payment_cycle (
    id         UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id    UUID CONSTRAINT fk_user_id_pc REFERENCES users (id),
    seats      BIGINT DEFAULT 0,
    freq       BIGINT DEFAULT 365,
    days_left  BIGINT NOT NULL,
    created_at TIMESTAMP NOT NULL
);
ALTER TABLE users ADD CONSTRAINT fk_payment_cycle_id_u FOREIGN KEY (payment_cycle_id) REFERENCES payment_cycle (id);

CREATE TABLE invoice (
     id                     UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
     nowpayments_invoice_id BIGINT NOT NULL,
     payment_cycle_id       UUID CONSTRAINT fk_payment_cycle_id_ub REFERENCES payment_cycle (id),
     payment_id             BIGINT,
     price_amount           BIGINT NOT NULL,
     price_currency         VARCHAR(16) NOT NULL,
     pay_amount             BIGINT,
     pay_currency           VARCHAR(16) NOT NULL,
     actually_paid          BIGINT,
     outcome_amount         BIGINT,
     outcome_currency       VARCHAR(16),
     payment_status         VARCHAR(16),
     freq                   BIGINT NOT NULL,
     invoice_url            TEXT,
     created_at             TIMESTAMP NOT NULL,
     last_update            TIMESTAMP NULL
);

CREATE TABLE daily_payment (
  id                UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  payment_cycle_id  UUID CONSTRAINT fk_payment_cycle_id_ub REFERENCES payment_cycle (id),
  currency          VARCHAR(16) NOT NULL,
  amount            BIGINT NOT NULL,
  days_left         BIGINT NOT NULL,
  last_update       TIMESTAMP NOT NULL
);

CREATE table wallet_address(
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id     UUID CONSTRAINT fk_user_id_duc REFERENCES users (id),
    currency    VARCHAR(16) NOT NULL,
    address  	VARCHAR(255),
    is_deleted	BOOLEAN
);

CREATE UNIQUE INDEX wallet_address_index ON wallet_address(address);

CREATE TABLE user_balances (
    id               UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    payment_cycle_id UUID CONSTRAINT fk_payment_cycle_id_ub REFERENCES payment_cycle (id),
    user_id          UUID CONSTRAINT fk_user_id_ub REFERENCES users (id),
    from_user_id     UUID CONSTRAINT fk_from_user_id_ub REFERENCES users (id),
    balance          BIGINT,
    balance_type     VARCHAR(16) NOT NULL,
    currency         VARCHAR(16) NOT NULL,
    day              DATE DEFAULT to_date('1970', 'YYYY') NOT NULL,
    created_at       TIMESTAMP NOT NULL
);
CREATE UNIQUE INDEX user_balances_index ON user_balances (
    payment_cycle_id,
    user_id,
    balance_type,
    currency,
    day
) where (balance_type != 'SPONSOR');

CREATE TABLE user_emails_sent (
    id         UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id    UUID CONSTRAINT fk_user_id_ub REFERENCES users (id),
    email_type VARCHAR(64),
    created_at TIMESTAMP NOT NULL
);

CREATE TABLE repo (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    orig_id     NUMERIC,
    url         VARCHAR(255) UNIQUE NOT NULL,
    git_url     VARCHAR(255) UNIQUE NOT NULL,
    branch      VARCHAR(16) NOT NULL,
    name        VARCHAR(255) NOT NULL,
    description TEXT,
    tags        BYTEA,
    score       NUMERIC,
    source      VARCHAR(255) NOT NULL,
    created_at  TIMESTAMP NOT NULL
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
    git_email           VARCHAR(255) NOT NULL,
    git_name            VARCHAR(255),
    weight              DOUBLE PRECISION NOT NULL,
    created_at          TIMESTAMP NOT NULL
);
CREATE UNIQUE INDEX analysis_response_index ON analysis_response(analysis_request_id, git_email);

CREATE TABLE daily_repo_balance (
    id         UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    repo_id    UUID CONSTRAINT fk_repo_id_drb REFERENCES repo (id),
    balance    BIGINT NOT NULL,
    day        DATE NOT NULL,
    currency   VARCHAR(16) NOT NULL,
    created_at TIMESTAMP NOT NULL
);
CREATE UNIQUE INDEX daily_repo_balance_index ON daily_repo_balance(repo_id, day, currency);

CREATE TABLE daily_repo_weight (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    repo_id     UUID CONSTRAINT fk_repo_id_drw REFERENCES repo (id),
    weight      DOUBLE PRECISION NOT NULL,
    day         DATE NOT NULL,
    created_at  TIMESTAMP NOT NULL
);
CREATE UNIQUE INDEX daily_repo_weight_index ON daily_repo_weight(repo_id, day);

CREATE TABLE daily_email_payout (
    id         UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email      VARCHAR(255) NOT NULL,
    balance    BIGINT NOT NULL,
    day        DATE NOT NULL,
    currency   VARCHAR(16) NOT NULL,
    created_at TIMESTAMP NOT NULL
);
CREATE UNIQUE INDEX daily_email_payout_index ON daily_email_payout(email, day, currency);

CREATE TABLE daily_user_payout (
    id         UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id    UUID CONSTRAINT fk_user_id_dup REFERENCES users (id),
    balance    BIGINT NOT NULL,
    day        DATE NOT NULL,
    currency   VARCHAR(16) NOT NULL,
    created_at TIMESTAMP NOT NULL
);
CREATE UNIQUE INDEX daily_user_payout_index ON daily_user_payout(user_id, day, currency);

CREATE TABLE daily_future_leftover (
    id         UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    repo_id    UUID CONSTRAINT fk_repo_id_dfl REFERENCES repo (id),
    balance    BIGINT NOT NULL,
    day        DATE NOT NULL,
    currency   VARCHAR(16) NOT NULL,
    created_at TIMESTAMP NOT NULL
);
CREATE UNIQUE INDEX daily_future_leftover_index ON daily_future_leftover(repo_id, day, currency);

CREATE table daily_user_contribution(
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id             UUID CONSTRAINT fk_user_id_duc REFERENCES users (id),
    repo_id             UUID CONSTRAINT fk_repo_id_duc REFERENCES repo (id),
    contributor_email   VARCHAR(255),
    contributor_name    VARCHAR(255),
    contributor_weight  DOUBLE PRECISION,
    contributor_user_id UUID CONSTRAINT fk_contributor_user_id_duc REFERENCES users (id) ,
    balance             BIGINT,
    balance_repo        BIGINT,
    day                 DATE NOT NULL,
    currency            VARCHAR(16) NOT NULL,
    created_at          TIMESTAMP NOT NULL
);
CREATE UNIQUE INDEX daily_user_contribution_index ON daily_user_contribution(user_id, repo_id, contributor_email, day, currency);

CREATE TABLE payout_request (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id             UUID CONSTRAINT fk_user_id_pc REFERENCES users (id),
    batch_id            UUID NOT NULL,
    currency            VARCHAR(16) NOT NULL,
    exchange_rate       NUMERIC,
    tea                 BIGINT NOT NULL,
    address             TEXT NOT NULL,
    created_at          TIMESTAMP NOT NULL
);
CREATE TABLE payout_response (
    id         UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    batch_id   UUID UNIQUE NOT NULL,
    tx_hash    VARCHAR(66) NOT NULL,
    error      TEXT,
    created_at TIMESTAMP NOT NULL
);

CREATE TABLE payout_response_details (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    payout_response_id  UUID CONSTRAINT fk_payout_response_id_pres REFERENCES payout_response (id),
    currency            VARCHAR(16) NOT NULL,
    nano_tea            BIGINT NOT NULL,
    smart_contract_tea  NUMERIC NOT NULL,
    address             VARCHAR(42),
    created_at          TIMESTAMP NOT NULL
);

CREATE TABLE invite (
    email VARCHAR(64),
    invite_email VARCHAR(64),
    confirmed_at TIMESTAMP,
    freq BIGINT NOT NULL,
    created_at TIMESTAMP NOT NULL,
    PRIMARY KEY(email, invite_email)
);