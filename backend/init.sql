create table "User"
(
	id uuid not null
		constraint users_pkey
			primary key,
	"stripeId" varchar,
	email varchar
);

alter table "User" owner to postgres;

create unique index users_email_uindex
	on "User" (email);

create table "GitEmail"
(
	email varchar(255) not null,
	uid uuid
		constraint uid
			references "User"
);

alter table "GitEmail" owner to postgres;

create unique index gitemail_email_uindex
	on "GitEmail" (email);

create table "PayOutAddress"
(
	id serial not null
		constraint payoutaddress_pk
			primary key,
	uid uuid
		constraint uid
			references "User",
	address varchar not null,
	"chainId" integer
);

alter table "PayOutAddress" owner to postgres;

create table "Repo"
(
	id uuid not null
		constraint repo_pk
			primary key,
	url varchar not null,
	name varchar not null
);

alter table "Repo" owner to postgres;

create unique index repo_url_uindex
	on "Repo" (url);

create table "GitUserBalance"
(
	id serial not null
		constraint gituserbalance_pk
			primary key,
	"repoId" uuid not null
		constraint repo_id
			references "Repo",
	uid uuid
		constraint uid
			references "User",
	balance integer not null,
	"createdAt" timestamp not null,
	"gitEmail" integer,
	score integer
);

comment on table "GitUserBalance" is 'if the the uid is null, there is no registered user which owns the git_email
';

alter table "GitUserBalance" owner to postgres;

create table "GitUserBalanceEvent"
(
	id serial not null
		constraint gituserbalanceevent_pk
			primary key,
	"gitUserBalanceId" integer not null
		constraint "gitUserBalanceId"
			references "GitUserBalance",
	timestamp timestamp not null,
	type userbalanceevent not null
);

alter table "GitUserBalanceEvent" owner to postgres;

create unique index gituserbalanceevent_id_uindex
	on "GitUserBalanceEvent" (id);

create table "SponsorEvent"
(
	id serial not null
		constraint sponsorevent_pk
			primary key,
	uid uuid not null
		constraint uid
			references "User",
	"repoId" uuid not null
		constraint "repoId"
			references "Repo",
	type sponsorevent not null,
	timestamp timestamp
);

alter table "SponsorEvent" owner to postgres;

create table "RepoBalance"
(
	id serial not null
		constraint repobalance_pk
			primary key,
	"repoId" uuid
		constraint "repoId"
			references "Repo",
	balance integer not null,
	timestamp timestamp
);

alter table "RepoBalance" owner to postgres;

create table "DailyRepoBalance"
(
	id serial not null
		constraint dailyrepobalance_pk
			primary key,
	"repoId" uuid not null
		constraint "repoId"
			references "Repo",
	uid uuid not null
		constraint uid
			references "User",
	"computedAt" timestamp not null,
	balance integer not null
);

alter table "DailyRepoBalance" owner to postgres;

create table "Contribution"
(
	id serial not null
		constraint contribution_pk
			primary key,
	"gitEmail" varchar not null,
	"gitName" varchar not null,
	"computedAt" timestamp not null,
	"fromTimestamp" timestamp not null,
	"toTimestamp" timestamp not null,
	"repoId" uuid not null
		constraint repo
			references "Repo",
	branch varchar
);

alter table "Contribution" owner to postgres;

