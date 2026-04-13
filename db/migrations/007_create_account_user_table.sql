create table if not exists account_user (
    id int primary key generated always as identity,
    account_id int not null references account(id),
    user_id int not null references "user"(id),

    unique(account_id, user_id)
);

---- create above / drop below ----

drop table if exists account_user cascade;
