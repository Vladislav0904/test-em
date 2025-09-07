CREATE TABLE IF NOT EXISTS subscriptions
(
    id           uuid PRIMARY KEY,
    service_name TEXT NOT NULL,
    price        INT  NOT NULL,
    user_id      uuid NOT NULL,
    start_date   DATE NOT NULL,
    end_date     DATE NULL
);