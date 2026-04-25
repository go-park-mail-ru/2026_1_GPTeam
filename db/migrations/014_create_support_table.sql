create type support_status as enum ('OPEN', 'CLOSED', 'IN_WORK');

create table if not exists support (
    id int primary key generated always as identity,
    user_id  int not null references "user"(id),
    category text not null,
    message text not null,
    status support_status not null default 'OPEN',
    created_at timestamp not null default now(),
    updated_at timestamp not null default now(),
    deleted boolean default false,

    constraint category_length check ( length(category) > 0 and length(category) < 255 ),
    constraint updated_at_not_in_past check ( updated_at >= created_at )
);

create trigger update_timestamp
    before update on support
    for each row execute function new_updated_at();

---- create above / drop below ----

drop table if exists support cascade;

drop type support_status cascade;

drop trigger if exists update_timestamp on support;
