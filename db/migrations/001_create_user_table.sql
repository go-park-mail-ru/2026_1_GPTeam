create table if not exists "user" (
    id int primary key generated always as identity not null,
    username text not null unique,
    password text not null,
    email text not null unique,
    created_at timestamp not null default now(),
    last_login timestamp default null,
    avatar_url text not null,
    balance double precision not null default 0,
    currency text not null default 'RUB',

    constraint username_length check ( length(username) >= 3 and length(username) <= 255),
    constraint password_length check ( length(password) >= 8 ),
    constraint email_is_correct check ( email ~* '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$' ),
    constraint balance_not_negative check ( balance >= 0 ),
    constraint currency_code check (currency in ('RUB', 'USD', 'EUR'))
);

---- create above / drop below ----

drop table if exists "user" cascade;
