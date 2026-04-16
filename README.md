# Rainbow Backend

## Overview

This project is a Gin + GORM + MySQL backend for:

- H5 public content query by date
- Admin login
- Admin image and audio upload to the current server
- Admin content create, update, delete, and paginated list

The API follows [docs/api-spec.md](docs/api-spec.md) and keeps these rules unchanged:

- `bg_url` is the only background image field
- `tags` is always a string array
- `date` uses `YYYY-MM-DD`
- all responses use a stable JSON envelope

`bg_url` and `music` should now be the public URLs returned by the admin upload APIs:

- `POST /api/admin/upload/image`
- `POST /api/admin/upload/audio`

```json
{
  "code": 0,
  "message": "ok",
  "data": {}
}
```

## Requirements

- Go 1.22+
- MySQL 8.0+
- Bash
- `systemd`
- `nginx`
- `sudo`

## Environment Variables

For local development, copy [.env.example](.env.example) to `.env`.

For dual-environment deployment, use:

- [deploy/env/test.env.example](deploy/env/test.env.example)
- [deploy/env/prod.env.example](deploy/env/prod.env.example)

Supported variables:

- `APP_ENV`: `dev`, `test`, or `prod`
- `HOST`: HTTP listen host
- `PORT`: HTTP listen port
- `DB_DRIVER`: currently `mysql`
- `DB_DSN`: MySQL DSN
- `LOG_ROOT`: log directory for `app.log` and `access.log`
- `JWT_SECRET`: JWT signing secret
- `JWT_EXPIRES_IN`: token expiry in seconds
- `ADMIN_USERNAME`: seeded admin username
- `ADMIN_PASSWORD`: seeded admin password
- `ALLOW_ORIGINS`: comma-separated CORS allowlist, or `*`
- `UPLOAD_ROOT`: upload root directory for the current environment
- `UPLOAD_IMAGE_MAX_SIZE`: max image upload size in bytes, default `52428800`
- `UPLOAD_AUDIO_MAX_SIZE`: max audio upload size in bytes, default `52428800`

On startup the service auto-migrates tables and seeds or updates the admin account from `ADMIN_USERNAME` and `ADMIN_PASSWORD`.

Default upload roots:

- `APP_ENV=dev`: `./uploads/dev`
- `APP_ENV=test`: `/opt/rainbow-backend/uploads/test`
- `APP_ENV=prod`: `/opt/rainbow-backend/uploads/prod`

Default log roots:

- `APP_ENV=dev`: `./logs/dev`
- `APP_ENV=test`: `/opt/rainbow-backend/logs/test`
- `APP_ENV=prod`: `/opt/rainbow-backend/logs/prod`

Generated log files:

- `app.log`: startup, panic recovery, auth, admin operation, upload logs
- `access.log`: HTTP access logs from Gin middleware

## MySQL Setup

### Log in to MySQL

```bash
mysql -uroot -p
```

### Create the two deployment databases

```sql
CREATE DATABASE rainbow_test CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE DATABASE rainbow_prod CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

Example DSNs used by the deployment env files:

```text
root:password@tcp(127.0.0.1:3306)/rainbow_test?charset=utf8mb4&parseTime=True&loc=Local
root:password@tcp(127.0.0.1:3306)/rainbow_prod?charset=utf8mb4&parseTime=True&loc=Local
```

## Local Development

1. Create a local database such as `rainbow`.
2. Copy the environment file:

```bash
cp .env.example .env
```

3. Edit `.env` as needed.
4. Start the server:

```bash
make run
```

or:

```bash
bash ./scripts/start.sh
```

or:

```bash
go run ./cmd/server
```

Typical local settings:

```env
APP_ENV=dev
HOST=0.0.0.0
PORT=8080
LOG_ROOT=./logs/dev
UPLOAD_ROOT=./uploads/dev
```

Local health check:

```bash
curl http://127.0.0.1:8080/health
```

## Dual-Environment Deployment

### Topology

The same public IP serves both environments on different external ports. There is one codebase and one backend binary.

- Test public entry: `http://<public-ip>:18080`
- Prod public entry: `http://<public-ip>:28080`
- Test backend listen: `127.0.0.1:18081`
- Prod backend listen: `127.0.0.1:28081`

Nginx reverse proxy mapping:

- `18080 -> http://127.0.0.1:18081`
- `28080 -> http://127.0.0.1:28081`

Nginx static file mapping:

- `http://<public-ip>:18080/static/images/<filename> -> /opt/rainbow-backend/uploads/test/images/<filename>`
- `http://<public-ip>:18080/static/audio/<filename> -> /opt/rainbow-backend/uploads/test/audio/<filename>`
- `http://<public-ip>:28080/static/images/<filename> -> /opt/rainbow-backend/uploads/prod/images/<filename>`
- `http://<public-ip>:28080/static/audio/<filename> -> /opt/rainbow-backend/uploads/prod/audio/<filename>`

MySQL separation:

- Test database: `rainbow_test`
- Prod database: `rainbow_prod`

Env separation:

- `/opt/rainbow-backend/test.env`
- `/opt/rainbow-backend/prod.env`

systemd separation:

- `rainbow-backend-test.service`
- `rainbow-backend-prod.service`

Deployment assets:

- [deploy/env/test.env.example](deploy/env/test.env.example)
- [deploy/env/prod.env.example](deploy/env/prod.env.example)
- [deploy/systemd/rainbow-backend-test.service](deploy/systemd/rainbow-backend-test.service)
- [deploy/systemd/rainbow-backend-prod.service](deploy/systemd/rainbow-backend-prod.service)
- [deploy/nginx/rainbow-backend-ports.conf](deploy/nginx/rainbow-backend-ports.conf)

### One-Command Deploy And Start

From the repository root:

```bash
bash ./scripts/test.sh
```

```bash
bash ./scripts/prod.sh
```

These scripts:

1. build `./bin/rainbow-backend`
2. ensure `/opt/rainbow-backend/bin` exists
3. copy the latest binary to `/opt/rainbow-backend/bin/rainbow-backend`
4. create `/opt/rainbow-backend/test.env` or `/opt/rainbow-backend/prod.env` only if missing
5. ensure the matching upload directories exist under `/opt/rainbow-backend/uploads/`
6. ensure the matching log directory exists under `/opt/rainbow-backend/logs/`
7. install the matching systemd unit into `/etc/systemd/system/`
8. install the shared Nginx port config into `/etc/nginx/conf.d/`
9. run `systemctl daemon-reload`
10. enable and restart the matching backend service
11. run `nginx -t` and reload Nginx
12. print verification commands and URLs

The scripts do not silently overwrite existing env files.

Before internet-facing use, edit the copied env files and replace:

- database credentials
- JWT secrets
- admin passwords
- `ALLOW_ORIGINS`

## Uploads

### What `bg_url` and `music` mean

- `bg_url` should be the image URL returned by `POST /api/admin/upload/image`
- `music` should be the audio URL returned by `POST /api/admin/upload/audio`
- The content create and update APIs stay unchanged and still accept plain string URLs for these fields

### Authenticated upload endpoints

- `POST /api/admin/upload/image`
- `POST /api/admin/upload/audio`

Both endpoints:

- require `Authorization: Bearer <token>`
- require `multipart/form-data`
- use the form field name `file`
- return the stable JSON envelope with `url`, `filename`, `size`, and `contentType`

Example response:

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "url": "http://<public-ip>:18080/static/images/20260415093015-ab12cd34.png",
    "filename": "20260415093015-ab12cd34.png",
    "size": 12345,
    "contentType": "image/png"
  }
}
```

### Supported formats and size limits

- images: `.jpg`, `.jpeg`, `.png`, `.webp`
- audio: `.mp3`, `.wav`, `.ogg`, `.m4a`
- default image max size: `50 MB`
- default audio max size: `50 MB`
- both limits are configurable with `UPLOAD_IMAGE_MAX_SIZE` and `UPLOAD_AUDIO_MAX_SIZE`

### Storage paths on the current server

- dev images: `./uploads/dev/images/`
- dev audio: `./uploads/dev/audio/`
- test images: `/opt/rainbow-backend/uploads/test/images/`
- test audio: `/opt/rainbow-backend/uploads/test/audio/`
- prod images: `/opt/rainbow-backend/uploads/prod/images/`
- prod audio: `/opt/rainbow-backend/uploads/prod/audio/`

The backend creates missing directories automatically when the first upload is saved.

### Public static URL patterns

- test images: `http://<public-ip>:18080/static/images/<filename>`
- test audio: `http://<public-ip>:18080/static/audio/<filename>`
- prod images: `http://<public-ip>:28080/static/images/<filename>`
- prod audio: `http://<public-ip>:28080/static/audio/<filename>`

The backend builds the returned URL from the current request host and forwarded scheme, so the same codebase works for both `test` and `prod` behind the existing Nginx proxy.

### Nginx static serving

The public Nginx servers now do two things at the same time:

- serve `/static/images/` and `/static/audio/` directly from disk using `alias`
- keep proxying `/api/...` and `/health` requests to the backend unchanged

This is defined in [deploy/nginx/rainbow-backend-ports.conf](deploy/nginx/rainbow-backend-ports.conf).

## Verification

### systemd

```bash
sudo systemctl status rainbow-backend-test.service --no-pager
sudo systemctl status rainbow-backend-prod.service --no-pager
```

```bash
sudo journalctl -u rainbow-backend-test.service -n 100 --no-pager
sudo journalctl -u rainbow-backend-prod.service -n 100 --no-pager
```

### Nginx

```bash
sudo nginx -t
sudo systemctl reload nginx
```

### Health checks

Local backend health checks:

```bash
curl http://127.0.0.1:18081/health
curl http://127.0.0.1:28081/health
```

Public health checks:

```bash
curl http://<public-ip>:18080/health
curl http://<public-ip>:28080/health
```

### Frontend base URLs

Frontend developers should use:

- Test base URL: `http://<public-ip>:18080`
- Prod base URL: `http://<public-ip>:28080`

## Frontend Integration Flow

Expected admin frontend workflow:

1. Login with `POST /api/admin/login` and store the JWT token.
2. Upload a background image with `POST /api/admin/upload/image`.
3. Upload background audio with `POST /api/admin/upload/audio`.
4. Read `data.url` from both upload responses.
5. Submit `POST /api/admin/content` or `PUT /api/admin/content/:id` using those returned URLs as `bg_url` and `music`.
6. Frontend H5 requests `GET /api/public/content?date=YYYY-MM-DD` and uses the stored URLs directly.

Notes for frontend developers:

- Test base URL: `http://<public-ip>:18080`
- Prod base URL: `http://<public-ip>:28080`
- Public content endpoint: `GET /api/public/content?date=YYYY-MM-DD`
- Admin login endpoint: `POST /api/admin/login`
- Admin image upload endpoint: `POST /api/admin/upload/image`
- Admin audio upload endpoint: `POST /api/admin/upload/audio`
- Admin auth header: `Authorization: Bearer <token>`
- CORS note: every frontend origin must be added to `ALLOW_ORIGINS`

Example public API request:

```bash
curl 'http://<public-ip>:18080/api/public/content?date=2026-04-07'
```

Admin login example:

```bash
curl -X POST 'http://<public-ip>:18080/api/admin/login' \
  -H 'Content-Type: application/json' \
  -d '{
    "username": "admin",
    "password": "your_admin_password"
  }'
```

Get a token and call an admin API:

```bash
TOKEN=$(curl -s -X POST 'http://<public-ip>:18080/api/admin/login' \
  -H 'Content-Type: application/json' \
  -d '{
    "username": "admin",
    "password": "your_admin_password"
  }' | jq -r '.data.token')
```

```bash
curl 'http://<public-ip>:18080/api/admin/content?page=1&pageSize=10' \
  -H "Authorization: Bearer ${TOKEN}"
```

## API Examples

Upload image:

```bash
curl -X POST 'http://<public-ip>:18080/api/admin/upload/image' \
  -H "Authorization: Bearer ${TOKEN}" \
  -F 'file=@/absolute/path/to/bg.png'
```

Upload audio:

```bash
curl -X POST 'http://<public-ip>:18080/api/admin/upload/audio' \
  -H "Authorization: Bearer ${TOKEN}" \
  -F 'file=@/absolute/path/to/music.mp3'
```

Upload both files and reuse their returned URLs:

```bash
IMAGE_URL=$(curl -s -X POST 'http://<public-ip>:18080/api/admin/upload/image' \
  -H "Authorization: Bearer ${TOKEN}" \
  -F 'file=@/absolute/path/to/bg.png' | jq -r '.data.url')
```

```bash
AUDIO_URL=$(curl -s -X POST 'http://<public-ip>:18080/api/admin/upload/audio' \
  -H "Authorization: Bearer ${TOKEN}" \
  -F 'file=@/absolute/path/to/music.mp3' | jq -r '.data.url')
```

Create content:

```bash
curl -X POST 'http://<public-ip>:18080/api/admin/content' \
  -H "Authorization: Bearer ${TOKEN}" \
  -H 'Content-Type: application/json' \
  -d '{
    "date": "2026-04-07",
    "text": "今天也要被温柔对待呀",
    "tags": ["心动", "温柔", "春天"],
    "bg_url": "'"${IMAGE_URL}"'",
    "music": "'"${AUDIO_URL}"'"
  }'
```

Update content:

```bash
curl -X PUT 'http://<public-ip>:18080/api/admin/content/1' \
  -H "Authorization: Bearer ${TOKEN}" \
  -H 'Content-Type: application/json' \
  -d '{
    "date": "2026-04-08",
    "text": "今天也要被温柔对待呀",
    "tags": ["浪漫", "春天"],
    "bg_url": "'"${IMAGE_URL}"'",
    "music": "'"${AUDIO_URL}"'"
  }'
```

Delete content:

```bash
curl -X DELETE 'http://<public-ip>:18080/api/admin/content/1' \
  -H "Authorization: Bearer ${TOKEN}"
```

List content:

```bash
curl 'http://<public-ip>:18080/api/admin/content?page=1&pageSize=10' \
  -H "Authorization: Bearer ${TOKEN}"
```

Public content:

```bash
curl 'http://<public-ip>:18080/api/public/content?date=2026-04-07'
```

Open a returned static URL directly:

```bash
curl "${IMAGE_URL}"
curl "${AUDIO_URL}"
```

## Firewall Notes

Open these external ports in both the cloud security group and the local firewall if one is enabled:

- `18080`
- `28080`

The backend ports `18081` and `28081` should remain bound to `127.0.0.1` and should not be opened publicly.

## Validation

```bash
go test ./...
go build ./...
bash -n ./scripts/test.sh
bash -n ./scripts/prod.sh
```
