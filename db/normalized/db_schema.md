```mermaid
erDiagram    
    User {
        id int "PK"
        username text "NOT NULL"
        password text "NOT NULL"
        email text "NOT NULL"
        created_at datetime "NOT NULL"
        last_login datetime
        avatar_url text "NOT NULL"
        balance double "NOT NULL"
        currency text "NOT NULL"
    }
    
    JWT {
        id int "PK"
        user_id int "FK"
        expired_at datetime "NOT NULL"
    }
    
    Budget {
        id int "PK"
        title text "NOT NULL"
        description text "NOT NULL"
        created_at datetime "NOT NULL"
        start_at datetime "NOT NULL"
        end_at datetime
        actual double "NOT NULL"
        target double "NOT NULL"
        currency text "NOT NULL"
        author int "FK"
    }
    
    User ||--|{ JWT: id
    User ||--|{ Budget: id
```

Или [ссылка](https://dbdiagram.io/d/GPTeam-690d0ed46735e11170a2094f) на dbdiagram.
