create or replace function new_updated_at() returns trigger as $$
begin
    new.updated_at = now();
    return new;
end;
$$ language plpgsql;

---- create above / drop below ----

drop function if exists new_updated_at();
