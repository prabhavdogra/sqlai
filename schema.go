package main

var schema = `CREATE TABLE authors (
	author_id   SERIAL PRIMARY KEY,
	name        VARCHAR(100) NOT NULL,
	country     VARCHAR(50)
  );
  CREATE TABLE books (
	book_id     SERIAL PRIMARY KEY,
	title       VARCHAR(200) NOT NULL,
	author_id   INTEGER NOT NULL REFERENCES authors(author_id),
	price       NUMERIC(8,2) NOT NULL,
	published   DATE
  );
  CREATE TABLE customers (
	customer_id SERIAL PRIMARY KEY,
	first_name  VARCHAR(50) NOT NULL,
	last_name   VARCHAR(50) NOT NULL,
	email       VARCHAR(100) UNIQUE NOT NULL,
	joined_on   DATE DEFAULT CURRENT_DATE
  );
  CREATE TABLE orders (
	order_id     SERIAL PRIMARY KEY,
	customer_id  INTEGER NOT NULL REFERENCES customers(customer_id),
	order_date   TIMESTAMP NOT NULL DEFAULT NOW()
  );
  CREATE TABLE order_items (
	item_id    SERIAL PRIMARY KEY,
	order_id   INTEGER NOT NULL REFERENCES orders(order_id),
	book_id    INTEGER NOT NULL REFERENCES books(book_id),
	quantity   INTEGER NOT NULL CHECK (quantity > 0),
	unit_price NUMERIC(8,2) NOT NULL
  );`
