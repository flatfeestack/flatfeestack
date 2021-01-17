CREATE TABLE daily_repo_hours (
    user_id    UUID CONSTRAINT fk_user_id_drh REFERENCES users (id),
    repo_hours INTEGER                             NOT NULL,
    day        DATE                                NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);
CREATE UNIQUE INDEX daily_repo_hours_index ON daily_repo_hours(user_id, day);

CREATE TABLE daily_repo_balance (
    repo_id    UUID CONSTRAINT fk_repo_id_drb REFERENCES repo (id),
    balance    BIGINT                              NOT NULL,
    day        DATE                                NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);
CREATE UNIQUE INDEX daily_repo_balance_index ON daily_repo_balance(repo_id, day);

CREATE TABLE weekly_repo_balance (
    repo_id    UUID CONSTRAINT fk_repo_id_gu REFERENCES repo (id),
    balance    BIGINT                              NOT NULL,
    day        DATE                                NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);
CREATE UNIQUE INDEX weekly_repo_balance_index ON weekly_repo_balance(repo_id, day);

CREATE TABLE weekly_email_payout (
    email      VARCHAR(255)                        NOT NULL,
    balance    BIGINT                              NOT NULL,
    day        DATE                                NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);
CREATE UNIQUE INDEX weekly_email_payout_index ON weekly_email_payout(email, day);

CREATE TABLE monthly_repo_weight (
    repo_id    UUID CONSTRAINT fk_repo_id_gu REFERENCES repo (id),
    weight     DOUBLE PRECISION                    NOT NULL,
    day        DATE                                NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);
CREATE UNIQUE INDEX monthly_repo_weight_index ON monthly_repo_weight(repo_id, day);

CREATE TABLE monthly_repo_balance (
    repo_id    UUID CONSTRAINT fk_repo_id_gu REFERENCES repo (id),
    balance    BIGINT                              NOT NULL,
    day        DATE                                NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);
CREATE UNIQUE INDEX monthly_repo_balance_index ON monthly_repo_balance(repo_id, day);

CREATE TABLE monthly_user_payout (
    user_id    UUID CONSTRAINT fk_user_id_mup REFERENCES users (id),
    balance    BIGINT                              NOT NULL,
    day        DATE                                NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);
CREATE UNIQUE INDEX monthly_user_payout_index ON monthly_user_payout(user_id, day);

CREATE TABLE monthly_future_leftover (
    repo_id    UUID CONSTRAINT fk_repo_id_gu REFERENCES repo (id),
    balance    BIGINT                              NOT NULL,
    day        DATE                                NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);
CREATE UNIQUE INDEX monthly_future_leftover_index ON monthly_future_leftover(repo_id, day);
