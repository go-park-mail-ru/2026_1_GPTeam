# Оптимизация работы СУБД
Для проведения нагрузочного тестирования выбрана сущность "**транзакции**",
так как действия с ними вызывают пересчёт денег пользователя. При этом транзакции создаются часто.

Для проведения нагрузочного тестирования выбран инструмент https://github.com/tsenart/vegeta.

Для запуска тестов были созданы скрипты на go в папке `perf_test`.

На момент тестирования на сервере были **отключены** некоторые функции, например, nginx, rate limiter и т.д. 

В тесте на создание транзакций создаётся 100000 сущностей с целевым RPS = 200.

В тесте на чтение транзакций получаются 100000 сущностей с сервера с целевым RPS = 200.

## Результаты:
- Запись
```
Results:
  Success rate:  24.20%
  Total requests: 100000
  Total duration:      8m50.0024228s
  Actual rate:   200.00 req/s
  Throughput:    45.66 req/s
Latencies:
  Min:  78.5698ms
  Mean: 24.422331685s
  50th: 22.199301379s
  90th: 30.000573905s
  Max:  30.7123487s
  Bytes (In/Out): 13 / 46
  Status codes: map[0:75526 200:24199 400:275]
```

- Чтение
```
Results:
  Success rate:  25.27%
  Total requests: 100000
  Total duration:      8m20.0090641s
  Actual rate:   200.00 req/s
  Throughput:    50.55 req/s
Latencies:
  Min:  6.0161ms
  Mean: 9.623109ms
  50th: 8.5017ms
  90th: 10.204491ms
  Max:  556.0159ms
  Bytes (In/Out): 127 / 0
  Status codes: map[200:25275 404:74725]
```

Очевидны проблемы:
- большое количество запросов превышают лимит vegeta (статус код 0).
- средняя длительность запросов на запись 24 секунд, а 90 процентиль уже выходит за лимиты vegeta.

Вывод: сервис не справляется с нагрузкой в 200 rps. 

## Оптимизации
- Чтение

Запрос `select user_id, account_id, value, type, category, title, description, created_at, transaction_date, updated_at from transaction where id = $1 and deleted_at is null;`

Запрос построен оптимально.

План запроса тоже оптимальный (количество записей в таблице 400к+):
```
Index Scan using transaction_pkey on transaction  (cost=0.42..8.44 rows=1 width=86) (actual time=0.032..0.033 rows=1 loops=1)
  Index Cond: (id = 1000)
  Filter: (deleted_at IS NULL)
Planning Time: 0.193 ms
Execution Time: 0.058 ms
```
При помощи индекса, который по умолчанию ставится на PK, бд быстро находит нужную транзакцию.
Значит, оптимизировать в чтении нечего.
- Запись

Создание транзакции состоит из нескольких запросов, объединённых в бд-ю транзакцию:

Создание записи в таблице `insert into transaction (user_id, account_id, value, type, category, title, description, transaction_date) values ($1, $2, $3, $4, $5, $6, $7, $8) returning id;`

Изменение баланса `update account set balance = balance + (case when $1 = 'INCOME' then $2 else -1 * $2 end) where id = $3;`

Изменение бюджета `update budget set actual = greatest(0, actual + (case when $1 = 'INCOME' then $2 else -1 * $2 end)) where author = $3 and active = true and exists(select 1 from budget_category where budget_id = budget.id and category = $4);`

Планы выполнения этих запросов:
```
Insert on transaction  (cost=0.00..0.03 rows=1 width=128) (actual time=0.473..0.474 rows=1 loops=1)
  ->  Result  (cost=0.00..0.03 rows=1 width=128) (actual time=0.033..0.034 rows=1 loops=1)
Planning Time: 0.027 ms
Trigger for constraint transaction_user_id_fkey: time=35.626 calls=1
Trigger for constraint transaction_account_id_fkey: time=9.516 calls=1
Execution Time: 45.754 ms
```
```
Update on account  (cost=0.00..8.06 rows=0 width=0) (actual time=177.810..177.811 rows=0 loops=1)
  ->  Seq Scan on account  (cost=0.00..8.06 rows=1 width=14) (actual time=0.020..0.022 rows=1 loops=1)
        Filter: (id = 1)
        Rows Removed by Filter: 4
Planning Time: 0.148 ms
Trigger update_timestamp: time=0.443 calls=1
Execution Time: 187.291 ms
```
```
Update on budget  (cost=0.15..26.26 rows=0 width=0) (actual time=0.007..0.009 rows=0 loops=1)
  ->  Nested Loop  (cost=0.15..26.26 rows=1 width=20) (actual time=0.006..0.007 rows=0 loops=1)
        ->  Seq Scan on budget  (cost=0.00..16.62 rows=1 width=18) (actual time=0.006..0.006 rows=0 loops=1)
              Filter: (active AND (author = 1))
        ->  Index Scan using budget_category_budget_id_category_key on budget_category  (cost=0.15..8.17 rows=1 width=10) (never executed)
              Index Cond: ((budget_id = budget.id) AND (category = 'Зарплата'::category_type))
Planning Time: 11.904 ms
Execution Time: 0.356 ms

```

Первый запрос выглядит оптимально.
Можно провести мини оптимизацию, добавив индексы для ускорения references проверок.

Во втором запросе постгрес, скорее всего, выбрал seq scan из-за того, что таблица account состоит из всего 5 записей,
значит, добавление индексов не поможет из-за того, что таблица просто маленькая.

Третий запрос можно оптимизировать, добавив дополнительные индексы:
- на таблицу бюджета индекс с автором и статусом.
- на таблицу связи бюджета и категории, чтобы можно было читать все данные из таблицы индексов.
