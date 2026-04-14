# Rainbow Backend

## Requirements

- Go 1.22+
- MySQL 8.0+

## Configuration

Copy `.env.example` to `.env` and update the database and admin settings.

Required variables:

- `DB_DRIVER=mysql`
- `DB_DSN=root:password@tcp(127.0.0.1:3306)/rainbow?charset=utf8mb4&parseTime=True&loc=Local`
- `JWT_SECRET=replace_with_a_strong_secret`
- `JWT_EXPIRES_IN=7200`
- `ADMIN_USERNAME=admin`
- `ADMIN_PASSWORD=change_me`

## Admin initialization

On startup the service runs database auto-migration and then initializes the admin account from `ADMIN_USERNAME` and `ADMIN_PASSWORD`.

- If the username does not exist, a new row is inserted into the `admins` table.
- The password is stored as a bcrypt hash in `password_hash`.
- If the username already exists and the configured password changes, the stored bcrypt hash is updated on startup.

## Run locally

```bash
go run ./cmd/server
```

## Health check

```bash
curl http://localhost:8080/health
```

## Admin login

```bash
curl -X POST http://localhost:8080/api/admin/login \
  -H 'Content-Type: application/json' \
  -d '{
    "username": "admin",
    "password": "123456"
  }'
```

Successful response:

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "token": "jwt_token_here",
    "expiresIn": 7200
  }
}
```

## Verify

```bash
go test ./...
go build ./...
```
