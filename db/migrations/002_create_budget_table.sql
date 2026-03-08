create table if not exists budget (
    id int primary key generated always as identity not null,
    title text not null,
    description text not null,
    created_at timestamp not null default now(),
    start_at timestamp not null default now(),
    end_at timestamp default null,
    actual double precision not null,
    target double precision not null,
    currency text not null,
    author int not null references "user"(id)

        constraint title_length check ( length(title) > 0 and length(title) <= 255 ),
    constraint description_length check ( length(description) > 0 ),
    constraint start_not_in_past check ( start_at >= created_at ),
    constraint end_after_start check ( end_at is null or end_at > start_at ),
    constraint actual_not_negative check ( actual >= 0 ),
    constraint target_greater_then_actual check ( target > actual ),
    constraint currency_code check (currency in ('RUB', 'USD', 'EUR'))
);

---- create above / drop below ----

drop table if exists budget cascade;
