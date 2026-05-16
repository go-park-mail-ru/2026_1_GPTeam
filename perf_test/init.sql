create table account
(
    id         integer generated always as identity
        primary key,
    name       text                                          not null
        constraint name_length
            check ((length(name) > 0) AND (length(name) <= 255)),
    balance    double precision default 0                    not null
        constraint balance_not_negative
            check (balance >= (0)::double precision),
    currency   currency_code    default 'RUB'::currency_code not null,
    created_at timestamp        default now()                not null,
    updated_at timestamp        default now()                not null,
    deleted_at timestamp,
    constraint updated_at_not_in_past
        check (updated_at >= created_at),
    constraint account_deleted_at_not_in_past
        check ((deleted_at IS NULL) OR (deleted_at >= created_at))
);

create table account_user
(
    id         integer generated always as identity
        primary key,
    account_id integer not null
        references account,
    user_id    integer not null
        references "user",
    unique (account_id, user_id)
);

create table budget
(
    id          integer generated always as identity
        primary key,
    title       text                    not null
        constraint title_length
            check ((length(title) > 0) AND (length(title) <= 255)),
    description text                    not null
        constraint description_length
            check (length(description) > 0),
    created_at  timestamp default now() not null,
    start_at    timestamp default now() not null,
    end_at      timestamp,
    updated_at  timestamp default now() not null,
    actual      double precision        not null
        constraint actual_not_negative
            check (actual >= (0)::double precision),
    target      double precision        not null,
    currency    currency_code           not null,
    author      integer                 not null
        references "user",
    active      boolean   default true  not null,
    constraint start_not_in_past
        check (start_at >= created_at),
    constraint end_after_start
        check ((end_at IS NULL) OR (end_at >= start_at)),
    constraint updated_at_not_in_past
        check (updated_at >= created_at)
);

create table budget_category
(
    id        integer generated always as identity
        primary key,
    budget_id integer       not null
        references budget,
    category  category_type not null,
    unique (budget_id, category)
);

create table jwt
(
    uuid       text                                           not null
        primary key,
    user_id    integer                                        not null
        unique
        references "user",
    expired_at timestamp default (now() + '7 days'::interval) not null
        constraint expired_at_not_in_past
            check (expired_at > now())
);

create table support
(
    id         integer generated always as identity
        primary key,
    user_id    integer                                       not null
        references "user",
    category   text                                          not null
        constraint category_length
            check ((length(category) > 0) AND (length(category) < 255)),
    message    text                                          not null,
    status     support_status default 'OPEN'::support_status not null,
    created_at timestamp      default now()                  not null,
    updated_at timestamp      default now()                  not null,
    deleted    boolean        default false,
    constraint updated_at_not_in_past
        check (updated_at >= created_at)
);

create table transaction
(
    id               integer generated always as identity
        primary key,
    user_id          integer                 not null
        references "user",
    account_id       integer                 not null
        references account
            on delete cascade,
    value            double precision        not null
        constraint value_not_negative
            check (value > (0)::double precision),
    type             transaction_type        not null,
    category         category_type           not null,
    title            text                    not null
        constraint title_length
            check ((length(title) > 0) AND (length(title) <= 255)),
    description      text                    not null
        constraint description_length
            check (length(description) > 0),
    created_at       timestamp default now() not null,
    transaction_date timestamp default now() not null,
    deleted_at       timestamp,
    updated_at       timestamp default now() not null,
    constraint deleted_at_not_in_past
        check ((deleted_at IS NULL) OR (deleted_at >= created_at)),
    constraint updated_at_not_in_past
        check (updated_at >= created_at)
);

create table "user"
(
    id         integer generated always as identity
        primary key,
    username   text                                      not null
        unique
        constraint username_length
            check ((length(username) >= 3) AND (length(username) <= 255)),
    password   text                                      not null
        constraint password_length
            check (length(password) >= 8),
    email      text                                      not null
        unique
        constraint email_is_correct
            check (email ~* '^[A-Za-zа-яёА-ЯЁ0-9._%+-]+@[A-Za-zа-яёА-ЯЁ0-9.-]+\.[A-Za-zа-яёА-ЯЁ]{2,}$'::text),
    created_at timestamp default now()                   not null,
    last_login timestamp,
    avatar_url text      default 'img/default.png'::text not null,
    updated_at timestamp default now()                   not null,
    active     boolean   default true                    not null,
    is_staff   boolean   default false,
    constraint last_login_not_in_past
        check ((last_login IS NULL) OR (last_login >= created_at)),
    constraint update_at_not_in_past
        check (updated_at >= created_at)
);
