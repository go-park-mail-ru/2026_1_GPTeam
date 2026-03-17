create table if not exists budget (
    id int primary key generated always as identity,
    title text not null,
    description text not null,
    created_at timestamp not null default now(),
    start_at timestamp not null default now(),
    end_at timestamp default null,
    updated_at timestamp not null default now(),
    actual double precision not null,
    target double precision not null,
    currency currency_code not null,
    author int not null references "user"(id),
    active boolean default true,

    constraint title_length check ( length(title) > 0 and length(title) <= 255 ),
    constraint description_length check ( length(description) > 0 ),
    constraint start_not_in_past check ( start_at >= created_at ),
    constraint end_after_start check ( end_at is null or end_at > start_at ),
    constraint updated_at_not_in_past check ( updated_at >= created_at ),
    constraint actual_not_negative check ( actual >= 0 ),
    constraint target_greater_than_actual check ( target > actual )
);

create trigger update_timestamp
    before update on budget
    for each row execute function new_updated_at();

---- create above / drop below ----

drop table if exists budget cascade;

drop trigger if exists update_timestamp on budget;
