create table if not exists category (
    id int primary key generated always as identity,
    title text not null,
    description text not null,

    constraint title_length check ( length(title) > 0 and length(title) <= 255 ),
    constraint description_length check ( length(description) > 0 )
);

---- create above / drop below ----

drop table if exists category cascade;
