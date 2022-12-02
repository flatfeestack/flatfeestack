CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS payment_cycle_in (
    id         UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    seats      BIGINT DEFAULT 1,
    freq       BIGINT DEFAULT 365,
    created_at TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS payment_cycle_out (
    id         UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS users (
    id                    UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    invited_id            UUID CONSTRAINT fk_users_id_u REFERENCES users (id), /*if this is set, then this email pays the amount*/
    stripe_id             VARCHAR(255),
    stripe_payment_method VARCHAR(255),
    stripe_last4          VARCHAR(8),
    payment_cycle_in_id   UUID CONSTRAINT fk_payment_cycle_in_id_u REFERENCES payment_cycle_in (id),
    payment_cycle_out_id  UUID CONSTRAINT fk_payment_cycle_out_id_u REFERENCES payment_cycle_out (id) NOT NULL,
    email                 VARCHAR(64) UNIQUE NOT NULL,
    name                  VARCHAR(255),
    image                 BYTEA,
    created_at            TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS user_balances (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    payment_cycle_in_id UUID CONSTRAINT fk_payment_cycle_in_id_ub REFERENCES payment_cycle_in (id),
    user_id             UUID CONSTRAINT fk_user_id_ub REFERENCES users (id),
    from_user_id        UUID CONSTRAINT fk_from_user_id_ub REFERENCES users (id),
    balance             NUMERIC(78), /*256 bits*/
    split               NUMERIC(78), /*256 bits*/
    currency            VARCHAR(8) NOT NULL,
    balance_type        VARCHAR(16) NOT NULL,
    created_at          TIMESTAMP NOT NULL
);
CREATE UNIQUE INDEX IF NOT EXISTS user_balances_index ON user_balances (payment_cycle_in_id, user_id, balance_type, currency, split);

CREATE TABLE IF NOT EXISTS repo (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
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
    id           UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id      UUID CONSTRAINT fk_user_id_gub REFERENCES users (id),
    email        VARCHAR(255) UNIQUE NOT NULL,
    token        VARCHAR(32),
    confirmed_at TIMESTAMP,
    created_at   TIMESTAMP NOT NULL
);
CREATE UNIQUE INDEX IF NOT EXISTS git_email_index ON git_email(user_id, email, token);
CREATE INDEX IF NOT EXISTS git_email_email_index ON git_email(email); /*we do a count on email*/

CREATE TABLE IF NOT EXISTS sponsor_event (
    id            UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    repo_id       UUID CONSTRAINT fk_repo_id_se REFERENCES repo (id),
    user_id       UUID CONSTRAINT fk_user_id_se REFERENCES users (id),
    sponsor_at    TIMESTAMP NOT NULL,
    un_sponsor_at TIMESTAMP
);
CREATE UNIQUE INDEX IF NOT EXISTS sponsor_event_index ON sponsor_event(repo_id, user_id, sponsor_at);

CREATE TABLE IF NOT EXISTS analysis_request (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    repo_id     UUID CONSTRAINT fk_repo_id_req REFERENCES repo (id),
    date_from   DATE NOT NULL,
    date_to     DATE NOT NULL,
    git_url     VARCHAR(255) NOT NULL,
    received_at TIMESTAMP,
    error       TEXT,
    created_at  TIMESTAMP NOT NULL
);
CREATE UNIQUE INDEX IF NOT EXISTS analysis_request_index ON analysis_request(repo_id, date_from, date_to);

CREATE TABLE IF NOT EXISTS analysis_response (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    analysis_request_id UUID CONSTRAINT fk_analysis_request_id_c REFERENCES analysis_request (id),
    git_email           VARCHAR(255) NOT NULL,
    git_names           VARCHAR(255)[],
    weight              DOUBLE PRECISION NOT NULL,
    created_at          TIMESTAMP NOT NULL
);
CREATE UNIQUE INDEX IF NOT EXISTS analysis_response_index ON analysis_response(analysis_request_id, git_email);

CREATE TABLE IF NOT EXISTS daily_contribution (
    id                   UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_sponsor_id      UUID CONSTRAINT fk_user_sponsor_id_dc REFERENCES users (id),
    user_contributor_id  UUID CONSTRAINT fk_user_contributor_id_dc REFERENCES users (id),
    repo_id              UUID CONSTRAINT fk_repo_id_dc REFERENCES repo (id),
    payment_cycle_in_id  UUID CONSTRAINT fk_payment_cycle_in_id_dc REFERENCES payment_cycle_in (id) NOT NULL,
    payment_cycle_out_id UUID CONSTRAINT fk_payment_cycle_out_id_dc REFERENCES payment_cycle_out (id) NOT NULL,
    balance              NUMERIC(78), /*256 bits*/
    currency             VARCHAR(8) NOT NULL,
    day                  DATE NOT NULL,
    created_at           TIMESTAMP NOT NULL
);
CREATE UNIQUE INDEX IF NOT EXISTS daily_contribution_index ON daily_contribution(user_sponsor_id, user_contributor_id, repo_id, currency, day);

CREATE TABLE IF NOT EXISTS unclaimed (
    id                   UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email                VARCHAR(64),
    repo_id              UUID CONSTRAINT fk_repo_id_dc REFERENCES repo (id),
    balance              NUMERIC(78), /*256 bits*/
    currency             VARCHAR(8) NOT NULL,
    day                  DATE NOT NULL,
    created_at           TIMESTAMP NOT NULL
);
CREATE UNIQUE INDEX IF NOT EXISTS unclaimed_index ON unclaimed(email, repo_id, currency);
CREATE INDEX IF NOT EXISTS unclaimed_email_index ON unclaimed(email); /*we do a count on email*/

CREATE TABLE IF NOT EXISTS future_contribution (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id             UUID CONSTRAINT fk_user_id_fc REFERENCES users (id),
    repo_id             UUID CONSTRAINT fk_repo_id_fc REFERENCES repo (id),
    payment_cycle_in_id UUID CONSTRAINT fk_payment_cycle_in_id_fc REFERENCES payment_cycle_in (id),
    balance             NUMERIC(78), /*256 bits*/
    currency            VARCHAR(8) NOT NULL,
    day                 DATE NOT NULL,
    created_at          TIMESTAMP NOT NULL
);
CREATE UNIQUE INDEX IF NOT EXISTS future_contribution_index ON future_contribution(user_id, repo_id, currency, day);

CREATE TABLE IF NOT EXISTS wallet_address (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id     UUID CONSTRAINT fk_user_id_duc REFERENCES users (id),
    currency    VARCHAR(8) NOT NULL,
    address  	VARCHAR(255),
    is_deleted	BOOLEAN
);
CREATE UNIQUE INDEX IF NOT EXISTS wallet_address_index ON wallet_address(address);

CREATE TABLE IF NOT EXISTS invite (
    email VARCHAR(64),
    invite_email VARCHAR(64),
    confirmed_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL,
    PRIMARY KEY(email, invite_email)
);

CREATE TABLE IF NOT EXISTS user_emails_sent (
    id         UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id    UUID CONSTRAINT fk_user_id_ub REFERENCES users (id),
    email      VARCHAR(64),
    email_type VARCHAR(64),
    created_at TIMESTAMP NOT NULL
);
CREATE INDEX IF NOT EXISTS user_emails_sent_index1 ON user_emails_sent(email_type); /*we do a count on email_type*/
CREATE INDEX IF NOT EXISTS user_emails_sent_index2 ON user_emails_sent(email); /*we do a count on email*/
