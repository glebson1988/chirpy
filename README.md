# Chirpy

Chirpy is a small Go REST API for a microblogging service. It supports user accounts, authenticated chirp creation, refresh tokens, and an admin-only reset endpoint. It also integrates a Polka webhook to upgrade users to Chirpy Red.

## Why care?

- A compact, real-world example of a JWT + refresh-token auth flow.
- Clean, testable handlers with explicit authorization checks.
- A good reference for designing and documenting REST endpoints in Go.

## Features

- Users: create, update (authenticated), login
- Chirps: create (authenticated), list (filter + sort), get, delete (author-only)
- Auth: access tokens (JWT), refresh tokens, revoke
- Webhooks: Polka `user.upgraded` to grant Chirpy Red
- Admin: reset users (dev only), metrics endpoint

## Requirements

- Go (see `go.mod`)
- PostgreSQL
- `sqlc` (for generating DB code)

## Setup

1) Install dependencies

```bash
go mod download
```

2) Create a `.env` file in the project root (example keys):

```bash
DB_URL=postgres://user:password@localhost:5432/chirpy?sslmode=disable
PLATFORM=dev
BEARER_TOKEN=your_jwt_secret
POLKA_KEY=your_polka_api_key
```

3) Run database migrations (example using goose if you have it installed):

```bash
goose -dir sql/schema postgres "$DB_URL" up
```

4) Generate DB code:

```bash
sqlc generate
```

5) Run the server:

```bash
go run .
```

Server listens on `:8080`.

## API overview

### Health & Admin

- `GET /api/healthz` → `200 OK`
- `GET /admin/metrics` → HTML metrics
- `POST /admin/reset` → `200 OK` (only when `PLATFORM=dev`)

### Users

- `POST /api/users`
  - Body: `{ "email": "...", "password": "..." }`
  - Response: user resource

- `PUT /api/users` (authenticated)
  - Header: `Authorization: Bearer <access_token>`
  - Body: `{ "email": "...", "password": "..." }`
  - Response: updated user resource

- `POST /api/login`
  - Body: `{ "email": "...", "password": "..." }`
  - Response: user + access token + refresh token

User resource shape:

```json
{
  "id": "uuid",
  "created_at": "RFC3339",
  "updated_at": "RFC3339",
  "email": "user@example.com",
  "is_chirpy_red": false
}
```

### Chirps

- `POST /api/chirps` (authenticated)
  - Header: `Authorization: Bearer <access_token>`
  - Body: `{ "body": "..." }`
  - Response: chirp resource

- `GET /api/chirps`
  - Optional query params:
    - `author_id=<uuid>` → filter by author
    - `sort=asc|desc` → sort by `created_at` (default `asc`)
  - Response: list of chirps

- `GET /api/chirps/{chirpID}`
  - Response: chirp resource

- `DELETE /api/chirps/{chirpID}` (authenticated)
  - Header: `Authorization: Bearer <access_token>`
  - Only the author can delete
  - Response: `204 No Content`

Chirp resource shape:

```json
{
  "id": "uuid",
  "created_at": "RFC3339",
  "updated_at": "RFC3339",
  "body": "text",
  "user_id": "uuid"
}
```

### Tokens

- `POST /api/refresh`
  - Header: `Authorization: Bearer <refresh_token>`
  - Response: `{ "token": "<new_access_token>" }`

- `POST /api/revoke`
  - Header: `Authorization: Bearer <refresh_token>`
  - Response: `204 No Content`

### Polka webhooks

Polka is a fictional payment provider used for this project.

- `POST /api/polka/webhooks`
  - Header: `Authorization: ApiKey <POLKA_KEY>`
  - Body:
    ```json
    { "event": "user.upgraded", "data": { "user_id": "uuid" } }
    ```
  - `204 No Content` on success or ignored events
  - `404` if user not found

## Running tests

```bash
go test ./...
```
