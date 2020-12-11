DO $$ BEGIN
    create type userbalanceevent as enum ('PAY_USER', 'PAY_REPO');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

DO $$ BEGIN
    create type sponsorevent as enum ('SPONSOR', 'UNSPONSOR');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

DO $$ BEGIN
    create type repobalanceevent as enum ('DISTRIBUTED');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

create table if not exists "user"
(
    id                 uuid    not null
        constraint users_pkey
            primary key,
    stripe_id          varchar,
    email              varchar,
    username           varchar not null,
    subscription       varchar,
    subscription_state varchar
);

create unique index if not exists users_email_uindex
    on "user" (email);

create table if not exists git_email
(
    email varchar(255) not null,
    uid   uuid
        constraint uid
            references "user"
);

create unique index if not exists gitemail_email_uindex
    on git_email (email);

create table if not exists pay_out_address
(
    id       serial  not null
        constraint payoutaddress_pk
            primary key,
    uid      uuid
        constraint uid
            references "user",
    address  varchar not null,
    chain_id integer
);

create table if not exists repo
(
    id          integer not null
        constraint repo_pk
            primary key,
    url         varchar not null,
    name        varchar not null,
    description varchar
);

create unique index if not exists repo_url_uindex
    on repo (url);

create table if not exists git_user_balance
(
    id         serial                  not null
        constraint gituserbalance_pk
            primary key,
    repo_id    integer                 not null
        constraint repo_id
            references repo,
    uid        uuid
        constraint uid
            references "user",
    balance    numeric                 not null,
    created_at timestamp default now() not null,
    git_email  varchar                 not null,
    score      numeric                 not null
);

comment on table git_user_balance is 'if the the uid is null, there is no registered user which owns the git_email ';

create table if not exists git_userbalanceevent
(
    id                  serial                  not null
        constraint gituserbalanceevent_pk
            primary key,
    git_user_balance_id integer                 not null
        constraint "gitUserBalanceId"
            references git_user_balance,
    timestamp           timestamp default now() not null,
    type                userbalanceevent        not null
);

create unique index if not exists gituserbalanceevent_id_uindex
    on git_userbalanceevent (id);

create table if not exists sponsor_event
(
    id        serial       not null
        constraint sponsor_event_pk
            primary key,
    uid       uuid         not null
        constraint uid
            references "user",
    repo_id   integer      not null
        constraint "repoId"
            references repo,
    type      sponsorevent not null,
    timestamp bigint       not null
);

create table if not exists repo_balance
(
    id        serial  not null
        constraint repobalance_pk
            primary key,
    repo_id   integer not null
        constraint "repoId"
            references repo,
    balance   numeric not null,
    timestamp timestamp
);

create table if not exists daily_repo_balance
(
    id        serial  not null
        constraint dailyrepobalance_pk
            primary key,
    repo_id   integer not null
        constraint "repoId"
            references repo,
    timestamp date    not null,
    balance   numeric not null
);

create table if not exists payments
(
    id     serial    not null
        constraint payments_pk
            primary key,
    uid    uuid      not null
        constraint "user"
            references "user",
    "from" timestamp not null,
    "to"   timestamp,
    sub    text,
    amount integer
);

create table if not exists analysis_request
(
    id      uuid      not null
        constraint analysis_request_pk
            primary key,
    "from"  timestamp not null,
    "to"    timestamp not null,
    repo_id integer   not null
        constraint analysis_request_repo_id_fk
            references repo,
    branch  text
);

create table if not exists contribution
(
    id               serial    not null
        constraint contribution_pk
            primary key,
    git_email        varchar   not null,
    computed_at      timestamp not null,
    weight           real      not null,
    analysis_request uuid      not null
        constraint contribution_analysis_request_fkey
            references analysis_request
            on delete cascade
);

create table if not exists payouts
(
    id        serial not null
        constraint payouts_pkey
            primary key,
    uid       uuid   not null
        constraint payouts_uid_fkey
            references "user"
            on delete cascade,
    amount    numeric,
    fulfilled boolean default false
);

create table if not exists repo_balance_event
(
    id           serial not null
        constraint repo_balance_event_pkey
            primary key,
    repo_balance integer
        constraint repo_balance_event_repo_balance_fkey
            references repo_balance
            on delete cascade,
    type         repobalanceevent,
    timestamp    timestamp default now()
);

create or replace view sponsored_repos(repo_id, user_id) as
SELECT r.id AS repo_id,
       u.id AS user_id
FROM (SELECT sponsor_event.uid,
             sponsor_event.repo_id,
             max(sponsor_event."timestamp") AS "timestamp"
      FROM sponsor_event
      GROUP BY sponsor_event.uid, sponsor_event.repo_id) latest
         JOIN sponsor_event s
              ON latest.uid = s.uid AND latest.repo_id = s.repo_id AND latest."timestamp" = s."timestamp"
         JOIN repo r ON r.id = s.repo_id
         JOIN "user" u ON u.id = s.uid
WHERE s.type = 'SPONSOR'::sponsorevent;

create or replace function countsponsoredrepos(uid uuid) returns integer
    language sql
as
$$
select count(repo_id)
from sponsored_repos
where user_id = uid
$$;

create or replace function days_in_month(date) returns integer
    language sql
as
$$
select CAST(to_char(date_trunc('month', $1) + interval '1 month' - date_trunc('month', $1), 'dd') as integer)
$$;

create or replace function get_sponsored_repos_at(date)
    returns TABLE
            (
                repo_id integer,
                user_id uuid
            )
    language sql
as
$$
SELECT r.id AS repo_id,
       u.id AS user_id
FROM ((((SELECT sponsor_event.uid,
                sponsor_event.repo_id,
                max(sponsor_event."timestamp") AS "timestamp"
         FROM sponsor_event
         WHERE to_timestamp(sponsor_event."timestamp")::date <= $1
         GROUP BY sponsor_event.uid, sponsor_event.repo_id) latest
    JOIN sponsor_event s ON (((latest.uid = s.uid) AND (latest.repo_id = s.repo_id) AND
                              (latest."timestamp" = s."timestamp"))))
    JOIN repo r ON ((r.id = s.repo_id)))
         JOIN "user" u ON ((u.id = s.uid)))
WHERE (s.type = 'SPONSOR'::sponsorevent);
$$;

create or replace function countsponsoredrepos_at(uid uuid, d date) returns integer
    language sql
as
$$
select count(repo_id)
from get_sponsored_repos_at(d)
where user_id = uid
$$;

create or replace function get_monthly_repo_balance_at(date)
    returns TABLE
            (
                repoid          integer,
                sponsor_amounts double precision
            )
    language sql
as
$$
SELECT daily_balance.repo_id, SUM(daily_balance.sponsor_amounts) as sponsor_amounts
FROM generate_series(
             (SELECT date_trunc('MONTH', $1)),
             (SELECT (date_trunc('MONTH', $1) + INTERVAL '1 MONTH - 1 day')::DATE), '1 day') d,
     get_daily_repo_balance_at(d::date) daily_balance
GROUP BY daily_balance.repo_id;
$$;

create or replace function get_last_payment(uuid)
    returns TABLE
            (
                id     integer,
                uid    uuid,
                "from" timestamp without time zone,
                "to"   timestamp without time zone,
                sub    text,
                amount integer
            )
    language sql
as
$$
SELECT p.*
FROM (SELECT MAX("to") as "to", uid
      FROM payments
      WHERE "to" >= now()
        AND uid = $1
      GROUP BY uid) as latest
         JOIN payments p ON p.to = latest.to AND p.uid = latest.uid;
$$;

create or replace function get_monthly_sponsor_amount(uuid) returns integer
    language sql
as
$$
SELECT amount / (EXTRACT(YEAR FROM age) * 12 + EXTRACT(MONTH FROM age))::int AS monthly_amount
FROM (SELECT age("to", "from"), uid, amount from get_last_payment($1)) as p;
$$;

create or replace function get_payment_at(uuid, date)
    returns TABLE
            (
                id     integer,
                uid    uuid,
                "from" timestamp without time zone,
                "to"   timestamp without time zone,
                sub    text,
                amount integer
            )
    language sql
as
$$
SELECT p.*
FROM (SELECT MAX("to") as "to", uid
      FROM payments
      WHERE "to"::DATE >= $2
        AND "from"::DATE <= $2
        AND uid = $1
      GROUP BY uid) as latest
         JOIN payments p ON p.to = latest.to AND p.uid = latest.uid;
$$;

create or replace function get_monthly_sponsor_amount_at(uuid, timestamp without time zone) returns integer
    language sql
as
$$
SELECT amount / (EXTRACT(YEAR FROM age) * 12 + EXTRACT(MONTH FROM age))::int AS monthly_amount
FROM (SELECT age("to", "from"), uid, amount from get_payment_at($1, $2::DATE)) as p;
$$;

create or replace function get_daily_repo_balance()
    returns TABLE
            (
                repo_id integer,
                balance double precision
            )
    language sql
as
$$
SELECT repo_id,
       ROUND(((get_monthly_sponsor_amount(user_id)::FLOAT / days_in_month(now()::date)) /
              (countsponsoredrepos(user_id))::FLOAT)::NUMERIC, 3) as sponsor_amounts
FROM sponsored_repos
$$;

create or replace function get_unclaimed_balances()
    returns TABLE
            (
                id        integer,
                uid       uuid,
                git_email character varying,
                balance   numeric
            )
    language sql
as
$$
(SELECT b.id, b.uid, b.git_email, b.balance
 FROM git_user_balance b
 WHERE NOT EXISTS
     (SELECT * FROM git_userbalanceevent e WHERE b.id = e.git_user_balance_id)
   AND uid IS NOT NULL)
$$;

create or replace function get_expired_balances()
    returns TABLE
            (
                id      integer,
                uid     uuid,
                repo_id integer,
                balance numeric
            )
    language sql
as
$$
(SELECT b.id, b.uid, b.repo_id, b.balance
 FROM git_user_balance b
 WHERE NOT EXISTS
     (SELECT * FROM git_userbalanceevent e WHERE b.id = e.git_user_balance_id)
   AND uid IS NULL
   AND created_at < now() - '6 MONTHS'::interval)
$$;

create or replace function get_daily_repo_balance_at(date)
    returns TABLE
            (
                repo_id integer,
                balance double precision
            )
    language sql
as
$$
SELECT repo_id,
       ROUND(((get_monthly_sponsor_amount_at(user_id, $1)::FLOAT / days_in_month($1)) /
              (countsponsoredrepos_at(user_id, $1))::FLOAT)::NUMERIC, 3) as sponsor_amounts
FROM get_sponsored_repos_at($1)
WHERE get_monthly_sponsor_amount_at(user_id, $1) is not NULL
$$;

create or replace function latest_repo_balance_at(date)
    returns TABLE
            (
                id          integer,
                repo_id     integer,
                "timestamp" date,
                balance     numeric
            )
    language sql
as
$$
SELECT drb.*
FROM (SELECT MAX(id) as id FROM daily_repo_balance WHERE timestamp = now()::date GROUP BY repo_id) as latest
         JOIN daily_repo_balance drb ON latest.id = drb.id

$$;

create or replace function check_daily_repo_balance_insert() returns trigger
    language plpgsql
as
$$
BEGIN
    IF (Select count(id) FROM daily_repo_balance WHERE repo_id = NEW.repo_id and "timestamp" = NEW."timestamp") > 0 THEN
        RAISE EXCEPTION 'Already inserted daily repo balance for this day';
    END IF;
    RETURN NEW;
END;
$$;

create trigger check_daily_repo_balance_insert
    before insert or update
    on daily_repo_balance
    for each row
execute procedure check_daily_repo_balance_insert();

