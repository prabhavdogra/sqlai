### Setup
- Install Go
- Run `go mod tidy`
- Make sure your Docker has atleast 5.5 GB of allocated memory
- Run the docker containers using `docker-compose up -d`

### Check Everything is working as expected
- Check the Ollama is working as expected using the curl `curl http://localhost:11434/`
- Check the database is working and the data is present by running the following commands in the container
    - `psql -U postgres -d postgres`
    - `SELECT * FROM students;`
- To login into SQL cli: `PGPASSWORD=mysecretpassword psql -h db -p 5432 -U postgres -d postgres`
- Query:
```
curl -X POST http://localhost:8080/query \
    -H "Content-Type: application/json" \
    -d '{
        "query": "Generate a SQL query that answers the question \"Show me all students\". This query will run on a database whose schema is represented in this string: CREATE TABLE students ( id SERIAL PRIMARY KEY, name VARCHAR(100), age INTEGER );"
    }'
```
- Query Ollama directly:
```
curl -X POST https://204e-2a09-bac1-36a0-1b8-00-39-138.ngrok-free.app/api/generate \
  -H "Content-Type: application/json" \
  -d '{
    "model": "sqlcoder",
    "prompt": "Generate a SQL query that answers the question \"Show me all students\". This query will run on a database whose schema is represented in this string: CREATE TABLE students ( id SERIAL PRIMARY KEY, name VARCHAR(100), age INTEGER );",
    "stream": false,
    "max_tokens": 100
  }'
```

```
clojure -M:run
yarn build
```

curl -X POST 'https://204e-2a09-bac1-36a0-1b8-00-39-138.ngrok-free.app/api/generate' \
  -H 'Content-Type: application/json' \
  -d '{
    "model": "sqlcoder",
    "prompt": "Generate a SQL query that answers the question \"Show me all authors\". This query will run on a database whose schema is represented in this string:\n\n\nCREATE TABLE authors (\n  author_id   SERIAL PRIMARY KEY,\n  name        VARCHAR(100) NOT NULL,\n  country     VARCHAR(50)\n);\n\n\nCREATE TABLE books (\n  book_id     SERIAL PRIMARY KEY,\n  title       VARCHAR(200) NOT NULL,\n  author_id   INTEGER NOT NULL REFERENCES authors(author_id),\n  price       NUMERIC(8,2) NOT NULL,\n  published   DATE\n);\n\n\nCREATE TABLE customers (\n  customer_id SERIAL PRIMARY KEY,\n  first_name  VARCHAR(50) NOT NULL,\n  last_name   VARCHAR(50) NOT NULL,\n  email       VARCHAR(100) UNIQUE NOT NULL,\n  joined_on   DATE DEFAULT CURRENT_DATE\n);\n\n\nCREATE TABLE orders (\n  order_id     SERIAL PRIMARY KEY,\n  customer_id  INTEGER NOT NULL REFERENCES customers(customer_id),\n  order_date   TIMESTAMP NOT NULL DEFAULT NOW()\n);\n\n-- 5) ORDER_ITEMS\nCREATE TABLE order_items (\n  item_id    SERIAL PRIMARY KEY,\n  order_id   INTEGER NOT NULL REFERENCES orders(order_id),\n  book_id    INTEGER NOT NULL REFERENCES books(book_id),\n  quantity   INTEGER NOT NULL CHECK (quantity > 0),\n  unit_price NUMERIC(8,2) NOT NULL\n);",  
    "stream": false,
    "max_tokens": 1000
}'

```
test@gmail.com
testing@123
```

While linking postgres:
- host: localhost