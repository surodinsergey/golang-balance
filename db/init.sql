CREATE TABLE IF NOT EXISTS users (
                                     id SERIAL PRIMARY KEY,
                                     firstname varchar(80) NOT NULL,
    lastname varchar(80) NOT NULL
    );

CREATE TABLE IF NOT EXISTS balances (
                                        id SERIAL PRIMARY KEY,
                                        sum varchar(80) NOT NULL,
    user_id INTEGER REFERENCES users (Id)
    );

INSERT INTO users (
    firstname,
    lastname
)
VALUES
    ('Василий', 'Шорохов'),
    ('Андрей', 'Боголюбов'),
    ('Иван', 'Меньшов');

INSERT INTO balances (
    sum,
    user_id
)
VALUES
    ('100', 1),
    ('1000', 2),
    ('99999', 3);