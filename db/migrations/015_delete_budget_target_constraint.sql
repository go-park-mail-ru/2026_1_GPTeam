alter table budget drop constraint if exists target_greater_than_actual;

---- create above / drop below ----

alter table budget add constraint target_greater_than_actual check ( target >= actual );
