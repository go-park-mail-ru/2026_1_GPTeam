alter table transaction
    drop constraint if exists transaction_account_id_fkey,
    add constraint transaction_account_id_fkey
        foreign key (account_id) references account(id) on delete cascade;

---- create above / drop below ----

alter table transaction
    drop constraint if exists transaction_account_id_fkey,
    add constraint transaction_account_id_fkey
        foreign key (account_id) references account(id);
