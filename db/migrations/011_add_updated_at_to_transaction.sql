alter table transaction add column if not exists updated_at timestamp not null default now() constraint updated_at_not_in_past check ( updated_at >= created_at );

create trigger update_timestamp
    before update on transaction
    for each row execute function new_updated_at();

---- create above / drop below ----

alter table transaction drop column if exists updated_at;

drop trigger if exists update_timestamp on transaction;
