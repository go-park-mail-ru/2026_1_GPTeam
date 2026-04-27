alter table "user" add column if not exists is_staff boolean default false;

---- create above / drop below ----

alter table "user" drop column if exists is_staff;
