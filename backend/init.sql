CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE users (
    id                    UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    invited_email         VARCHAR(64), /*if this is set, then this email pays the amount*/
    stripe_id             VARCHAR(255),
    stripe_payment_method VARCHAR(255),
    stripe_last4          VARCHAR(8),
    payment_cycle_id      UUID, /*CONSTRAINT fk_payment_cycle_id_u REFERENCES payment_cycle (id)*/
    email                 VARCHAR(64) UNIQUE NOT NULL,
    name                  VARCHAR(255),
    image                 BYTEA,
    created_at            TIMESTAMP NOT NULL
);

CREATE TABLE payment_cycle (
    id         UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id    UUID CONSTRAINT fk_user_id_pc REFERENCES users (id),
    seats      BIGINT DEFAULT 1,
    freq       BIGINT DEFAULT 365,
    created_at TIMESTAMP NOT NULL
);
ALTER TABLE users ADD CONSTRAINT fk_payment_cycle_id_u FOREIGN KEY (payment_cycle_id) REFERENCES payment_cycle (id);


CREATE TABLE user_balances (
    id               UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    payment_cycle_id UUID CONSTRAINT fk_payment_cycle_id_ub REFERENCES payment_cycle (id),
    user_id          UUID CONSTRAINT fk_user_id_ub REFERENCES users (id),
    from_user_id     UUID CONSTRAINT fk_from_user_id_ub REFERENCES users (id),
    balance          NUMERIC(78), /*256 bits*/
    split            NUMERIC(78), /*256 bits*/
    currency         VARCHAR(8) NOT NULL,
    balance_type     VARCHAR(16) NOT NULL,
    created_at       TIMESTAMP NOT NULL
);
CREATE UNIQUE INDEX user_balances_index ON user_balances (payment_cycle_id, user_id, balance_type, currency, split);

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
CREATE UNIQUE INDEX git_email_index ON git_email(user_id, email, token);

CREATE TABLE sponsor_event (
    id            UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    repo_id       UUID CONSTRAINT fk_repo_id_se REFERENCES repo (id),
    user_id       UUID CONSTRAINT fk_user_id_se REFERENCES users (id),
    sponsor_at    TIMESTAMP NOT NULL,
    un_sponsor_at TIMESTAMP DEFAULT to_date('9999', 'YYYY') NOT NULL
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

CREATE TABLE daily_contribution (
    id                   UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id              UUID CONSTRAINT fk_user_id_dc REFERENCES users (id),
    user_id_git          UUID CONSTRAINT fk_user_id_git_dc REFERENCES users (id),
    repo_id              UUID CONSTRAINT fk_repo_id_dc REFERENCES repo (id),
    payment_cycle_id     UUID CONSTRAINT fk_payment_cycle_id_dc REFERENCES payment_cycle (id) NOT NULL,
    payment_cycle_id_git UUID CONSTRAINT fk_payment_cycle_id_git_dc REFERENCES payment_cycle (id),
    balance              NUMERIC(78), /*256 bits*/
    currency             VARCHAR(8) NOT NULL,
    day                  DATE NOT NULL,
    created_at           TIMESTAMP NOT NULL
);
CREATE UNIQUE INDEX daily_contribution_index ON daily_contribution(user_id, user_id_git, repo_id, currency, day);

CREATE TABLE future_contribution (
    id               UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id          UUID CONSTRAINT fk_user_id_fc REFERENCES users (id),
    repo_id          UUID CONSTRAINT fk_repo_id_fc REFERENCES repo (id),
    payment_cycle_id UUID CONSTRAINT fk_payment_cycle_id_fc REFERENCES payment_cycle (id),
    balance          NUMERIC(78), /*256 bits*/
    currency         VARCHAR(8) NOT NULL,
    day              DATE NOT NULL,
    created_at       TIMESTAMP NOT NULL
);
CREATE UNIQUE INDEX future_contribution_index ON future_contribution(user_id, repo_id, currency, day);


CREATE TABLE payout_request (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id             UUID CONSTRAINT fk_user_id_pc REFERENCES users (id),
    batch_id            UUID NOT NULL,
    currency            VARCHAR(8) NOT NULL,
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
    currency            VARCHAR(8) NOT NULL,
    nano_tea            BIGINT NOT NULL,
    smart_contract_tea  NUMERIC NOT NULL,
    address             VARCHAR(42),
    created_at          TIMESTAMP NOT NULL
);

CREATE table wallet_address(
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id     UUID CONSTRAINT fk_user_id_duc REFERENCES users (id),
    currency    VARCHAR(8) NOT NULL,
    address  	VARCHAR(255),
    is_deleted	BOOLEAN
);
CREATE UNIQUE INDEX wallet_address_index ON wallet_address(address);

CREATE TABLE invite (
    email VARCHAR(64),
    invite_email VARCHAR(64),
    confirmed_at TIMESTAMP,
    freq BIGINT NOT NULL,
    created_at TIMESTAMP NOT NULL,
    PRIMARY KEY(email, invite_email)
);

CREATE TABLE user_emails_sent (
    id         UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id    UUID CONSTRAINT fk_user_id_ub REFERENCES users (id),
    email_type VARCHAR(64),
    created_at TIMESTAMP NOT NULL
);
CREATE INDEX user_emails_sent_index ON user_emails_sent(email_type); /*we do a count on email_type*/
