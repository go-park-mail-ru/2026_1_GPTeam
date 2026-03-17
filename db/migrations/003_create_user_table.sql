create table if not exists "user" (
    id int primary key generated always as identity,
    username text not null unique,
    password text not null,
    email text not null unique,
    created_at timestamp not null default now(),
    last_login timestamp default null,
    avatar_url text not null default 'img/default.png',
    updated_at timestamp not null default now(),
    active boolean default true,

    constraint username_length check ( length(username) >= 3 and length(username) <= 255),
    constraint password_length check ( length(password) >= 8 ),
    constraint email_is_correct check ( email ~* '^[A-Za-zа-яёА-ЯЁ0-9._%+-]+@[A-Za-zа-яёА-ЯЁ0-9.-]+\.[A-Za-zа-яёА-ЯЁ]{2,}$' ),
    constraint last_login_not_in_past check ( last_login is null or last_login >= created_at),
    constraint update_at_not_in_past check ( updated_at >= created_at )
);

create trigger update_timestamp
    before update on "user"
    for each row execute function new_updated_at();

---- create above / drop below ----

drop table if exists "user" cascade;

drop trigger if exists update_timestamp on "user";
