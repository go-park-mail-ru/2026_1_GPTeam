create type currency_code as enum ('RUB', 'USD', 'EUR');

create type transaction_type as enum ('INCOME', 'EXPENSE');

---- create above / drop below ----

drop type currency_code cascade;

drop type transaction_type cascade;
