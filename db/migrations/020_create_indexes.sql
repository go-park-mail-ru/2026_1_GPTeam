create index concurrently if not exists idx_transaction_user_id on transaction(user_id);
create index concurrently if not exists idx_transaction_account_id on transaction(account_id);
create index concurrently if not exists idx_budget_author_active on budget(author, active);
create index concurrently if not exists idx_budget_category_lookup on budget_category(budget_id, category);

---- create above / drop below ----

drop index concurrently if exists idx_transaction_user_id;
drop index concurrently if exists idx_transaction_account_id;
drop index concurrently if exists idx_budget_author_active;
drop index concurrently if exists idx_budget_category_lookup;
