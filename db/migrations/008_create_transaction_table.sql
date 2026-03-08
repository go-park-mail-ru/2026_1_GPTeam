create table if not exists transaction (
    id int primary key generated always as identity,
    user_id int not null references "user"(id),
    account_id int not null references account(id),
    value double precision not null,
    type text not null,
    category category_type not null,
    description text not null,
    created_at timestamp not null default now(),
    transaction_date timestamp not null default now(),

    constraint value_not_negative check ( value > 0 ),
    constraint type_is_correct check ( type in ('in', 'out') ),
    constraint description_length check ( length(description) > 0 )
);

---- create above / drop below ----

drop table if exists transaction cascade;
