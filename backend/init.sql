CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE users (
    id                 UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    stripe_id          VARCHAR(255),
    email              VARCHAR(255) UNIQUE NOT NULL,
    subscription       VARCHAR(255),
    subscription_state VARCHAR(255),
    payout_eth         VARCHAR(255),
    role               VARCHAR(255) DEFAULT 'USER' NOT NULL,
    created_at         TIMESTAMP NOT NULL
);

CREATE TABLE repo (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    orig_id     NUMERIC,
    url         VARCHAR(255) UNIQUE NOT NULL,
    name        VARCHAR(255)        NOT NULL,
    description TEXT,
    created_at  TIMESTAMP NOT NULL
);

CREATE TABLE git_email (
    id         UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id    UUID CONSTRAINT fk_user_id_gub REFERENCES users (id),
    email      VARCHAR(255) UNIQUE NOT NULL,
    created_at TIMESTAMP NOT NULL
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
    repo_id    UUID CONSTRAINT fk_repo_id_ar REFERENCES repo (id),
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

