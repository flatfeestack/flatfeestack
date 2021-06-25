--22.05.2021

CREATE table daily_user_contribution(
                                        id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
                                        user_id             UUID CONSTRAINT fk_user_id_duc REFERENCES users (id),
                                        repo_id             UUID CONSTRAINT fk_repo_id_duc REFERENCES repo (id),
                                        contributor_email   VARCHAR(255),
                                        contributor_weight  DOUBLE PRECISION,
                                        contributor_user_id UUID CONSTRAINT fk_contributor_user_id_duc REFERENCES users (id) ,
                                        balance             BIGINT,
                                        balance_repo        BIGINT,
                                        day                 DATE NOT NULL,
                                        created_at          TIMESTAMP NOT NULL
);
CREATE UNIQUE INDEX daily_user_contribution_index ON daily_user_contribution(user_id, repo_id, contributor_email, day);

CREATE UNIQUE INDEX analysis_response_index ON analysis_response(analysis_request_id, git_email);

--25.06.2021
alter table analysis_response add git_name VARCHAR(255);
alter table users add invited_email VARCHAR(64);
