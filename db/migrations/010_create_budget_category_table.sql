create table if not exists budget_category (
    id int primary key generated always as identity,
    budget_id int not null references budget(id),
    category category_type not null,

    unique(budget_id, category)
);

---- create above / drop below ----

drop table if exists budget_category cascade;
