-- ──────────────── SCHEMA & DATA ────────────────

-- 1) AUTHORS
CREATE TABLE authors (
  author_id   SERIAL PRIMARY KEY,
  name        VARCHAR(100) NOT NULL,
  country     VARCHAR(50)
);
-- 2) BOOKS
CREATE TABLE books (
  book_id     SERIAL PRIMARY KEY,
  title       VARCHAR(200) NOT NULL,
  author_id   INTEGER NOT NULL REFERENCES authors(author_id),
  price       NUMERIC(8,2) NOT NULL,
  published   DATE
);
-- 3) CUSTOMERS
CREATE TABLE customers (
  customer_id SERIAL PRIMARY KEY,
  first_name  VARCHAR(50) NOT NULL,
  last_name   VARCHAR(50) NOT NULL,
  email       VARCHAR(100) UNIQUE NOT NULL,
  joined_on   DATE DEFAULT CURRENT_DATE
);
-- 4) ORDERS
CREATE TABLE orders (
  order_id     SERIAL PRIMARY KEY,
  customer_id  INTEGER NOT NULL REFERENCES customers(customer_id),
  order_date   TIMESTAMP NOT NULL DEFAULT NOW()
);
-- 5) ORDER_ITEMS
CREATE TABLE order_items (
  item_id    SERIAL PRIMARY KEY,
  order_id   INTEGER NOT NULL REFERENCES orders(order_id),
  book_id    INTEGER NOT NULL REFERENCES books(book_id),
  quantity   INTEGER NOT NULL CHECK (quantity > 0),
  unit_price NUMERIC(8,2) NOT NULL
);
-- 6) INSERT DATA

INSERT INTO authors (name, country) VALUES
  ('George Orwell',       'United Kingdom'),
  ('Haruki Murakami',     'Japan'),
  ('Isabel Allende',      'Chile'),
  ('Chimamanda Ngozi Adichie', 'Nigeria');

INSERT INTO books (title, author_id, price, published) VALUES
  ('1984',                     1, 9.99,  '1949-06-08'),
  ('Animal Farm',              1, 7.99,  '1945-08-17'),
  ('Kafka on the Shore',       2, 14.50, '2002-09-12'),
  ('One Hundred Years of Solitude', 3, 13.00, '1967-05-30'),
  ('Half of a Yellow Sun',     4, 12.75, '2006-09-12');

INSERT INTO customers (first_name, last_name, email, joined_on) VALUES
  ('Alice',   'Wong',     'alice.wong@example.com',   '2024-01-15'),
  ('Bob',     'Smith',    'bob.smith@example.com',    '2023-11-02'),
  ('Carlos',  'Gonzalez', 'carlos.gonzalez@example.com','2024-03-20'),
  ('Diana',   'Lee',      'diana.lee@example.com',    '2024-04-10');

INSERT INTO orders (customer_id, order_date) VALUES
  (1, '2024-04-01 10:15'),
  (2, '2024-04-02 14:30'),
  (1, '2024-04-05 09:00'),
  (3, '2024-04-06 16:45');

INSERT INTO order_items (order_id, book_id, quantity, unit_price) VALUES
  (1, 1, 1, 9.99),
  (1, 2, 2, 7.99),
  (2, 4, 1, 13.00),
  (2, 5, 1, 12.75),
  (3, 3, 1, 14.50),
  (4, 5, 3, 12.75),
  (4, 2, 1, 7.99);
