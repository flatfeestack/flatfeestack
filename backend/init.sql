DO $$ BEGIN
    create type userbalanceevent as enum ('PAY_USER', 'PAY_REPO');
    alter type userbalanceevent owner to postgres;
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

DO $$ BEGIN
    create type sponsorevent as enum ('SPONSOR', 'UNSPONSOR');
    alter type sponsorevent owner to postgres;
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;




CREATE TABLE IF NOT EXISTS "user"
(
    id uuid not null
        constraint users_pkey
            primary key,
    stripe_id varchar,
    email varchar,
    username varchar not null,
    subscription varchar,
    subscription_state varchar
);

alter table "user" owner to postgres;



CREATE TABLE IF NOT EXISTS git_email
(
    email varchar(255) not null,
    uid uuid
        constraint uid
            references "user"
);

alter table git_email owner to postgres;



CREATE TABLE IF NOT EXISTS pay_out_address
(
    id serial not null
        constraint payoutaddress_pk
            primary key,
    uid uuid
        constraint uid
            references "user",
    address varchar not null,
    chain_id integer
);

alter table pay_out_address owner to postgres;

CREATE TABLE IF NOT EXISTS repo
(
    id uuid not null
        constraint repo_pk
            primary key,
    url varchar not null,
    name varchar not null
);

alter table repo owner to postgres;


CREATE TABLE IF NOT EXISTS git_user_balance
(
    id serial not null
        constraint gituserbalance_pk
            primary key,
    repo_id uuid not null
        constraint repo_id
            references repo,
    uid uuid
        constraint uid
            references "user",
    balance integer not null,
    created_at timestamp not null,
    git_email integer,
    score integer
);

comment on table git_user_balance is 'if the the uid is null, there is no registered user which owns the git_email
';

alter table git_user_balance owner to postgres;

CREATE TABLE IF NOT EXISTS git_userbalanceevent
(
    id serial not null
        constraint gituserbalanceevent_pk
            primary key,
    git_user_balance_id integer not null
        constraint "gitUserBalanceId"
            references git_user_balance,
    timestamp timestamp not null,
    type userbalanceevent not null
);

alter table git_userbalanceevent owner to postgres;



CREATE TABLE IF NOT EXISTS sponsor_event
(
    id serial not null
        constraint sponsor_event_pk
            primary key,
    uid uuid not null
        constraint uid
            references "user",
    repo_id uuid not null
        constraint "repoId"
            references repo,
    type sponsorevent not null,
    timestamp bigint not null
);

alter table sponsor_event owner to postgres;

CREATE TABLE IF NOT EXISTS repo_balance
(
    id serial not null
        constraint repobalance_pk
            primary key,
    repo_id uuid
        constraint "repoId"
            references repo,
    balance integer not null,
    timestamp timestamp
);

alter table repo_balance owner to postgres;

CREATE TABLE IF NOT EXISTS daily_repo_balance
(
    id serial not null
        constraint dailyrepobalance_pk
            primary key,
    repo_id uuid not null
        constraint "repoId"
            references repo,
    uid uuid not null
        constraint uid
            references "user",
    computed_at date not null,
    balance integer not null
);

alter table daily_repo_balance owner to postgres;

CREATE TABLE IF NOT EXISTS contribution
(
    id serial not null
        constraint contribution_pk
            primary key,
    git_email varchar not null,
    git_name varchar not null,
    computed_at timestamp not null,
    from_timestamp timestamp not null,
    to_timestamp timestamp not null,
    repo_id uuid not null
        constraint repo
            references repo,
    branch varchar
);

alter table contribution owner to postgres;


DO $$ BEGIN
    create unique index users_email_uindex
        on "user" (email);
EXCEPTION
    WHEN duplicate_table THEN null;
END $$;

DO $$ BEGIN
    create unique index gitemail_email_uindex
    on git_email (email);
EXCEPTION
    WHEN duplicate_table THEN null;
END $$;

DO $$ BEGIN
    create unique index repo_url_uindex
    on repo (url);
EXCEPTION
    WHEN duplicate_table THEN null;
END $$;

DO $$ BEGIN
    create unique index gituserbalanceevent_id_uindex
    on git_userbalanceevent (id);
EXCEPTION
    WHEN duplicate_table THEN null;
END $$;
