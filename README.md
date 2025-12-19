# Internal Transfers Service

Simple Go HTTP service to create accounts, query balances, and submit internal transfers backed by Postgres.

Setup

1. Ensure Go is installed (1.20+ recommended) and Postgres is running.
2. Create a database and run the migration:

```bash
createdb accounting_db
psql -d accounting_db -f migrations/init.sql
```

3. Set `DATABASE_URL` environment variable. Example:

```bash
export DATABASE_URL="postgres://postgres:postgres@localhost:5432/accounting_db?sslmode=disable"
```

4. Get dependencies and run:

```bash
go run .
```

API

- Create account: `POST /accounts`
  Body:
  ```json
  {"account_id": 123, "initial_balance": "100.23344"}
  ```
  Response: 200 OK on success

  Example:
  ```json
  {"message":"account created successfully"}
  ```

- Get account: `GET /accounts/{id}`
  Response:
  ```json
  {"account_id":"123", "balance":"100.23344"}
  ```

- Create transaction: `POST /transactions`
  Body:
  ```json
  {"source_account_id":123, "destination_account_id":456, "amount":"100.12345"}
  ```
  Response: 200 OK on success

  Example:
  ```json
  {"message":"transaction completed successfully"}
  ```

- List accounts: `GET /accounts`
  Response: 200 OK with JSON array of accounts

  Example:
  ```json
  [{"account_id":123, "balance":"100.23344"}, {"account_id":456, "balance":"200.00"}]
  ```

- List transactions for an account: `GET /accounts/{id}/transactions`
  Response: 200 OK with JSON array of transactions (most recent first)

  Example:
  ```json
  [
    {"id":10, "source_account_id":123, "destination_account_id":456, "amount":"50.00"},
    {"id":9, "source_account_id":456, "destination_account_id":123, "amount":"25.00"}
  ]
  ```
