create table if not exists account (
    id int primary key generated always as identity,
    name text not null,
    balance double precision not null default 0,
    currency currency_code not null default 'RUB',
    created_at timestamp not null default now(),
    updated_at timestamp not null default now(),

    constraint name_length check ( length(name) > 0 and length(name) <= 255 ),
    constraint balance_not_negative check ( balance >= 0 ),
    constraint updated_at_not_in_past check ( updated_at >= created_at )
);

create trigger update_timestamp
    before update on account
    for each row execute function new_updated_at();

---- create above / drop below ----

drop table if exists account cascade;

drop trigger if exists update_timestamp on account;
