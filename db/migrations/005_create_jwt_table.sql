create table if not exists jwt (
    uuid text primary key,
    user_id int not null references "user"(id),
    expired_at timestamp not null default now() + interval '7 day',

    constraint expired_at_not_in_past check ( expired_at > now() )
);

---- create above / drop below ----

drop table if exists jwt cascade;
