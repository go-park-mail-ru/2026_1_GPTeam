create type currency_code as enum ('RUB', 'USD', 'EUR');

create type transaction_type as enum ('INCOME', 'EXPENSE');

create type category_type as enum ('Зарплата', 'Стипендия', 'Продукты', 'Кафе', 'Транспорт', 'Другое',
    'Коммунальные платежи', 'Интернет', 'Одежда', 'Здоровье', 'Спорт', 'Развлечения', 'Путешествия',
    'Образование', 'Подарки', 'Животные', 'Техника', 'Кредиты и долги');

---- create above / drop below ----

drop type currency_code cascade;

drop type transaction_type cascade;

drop type category_type cascade;
