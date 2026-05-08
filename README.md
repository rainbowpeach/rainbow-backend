# Rainbow Backend

## Overview

This repository is a Gin + GORM + MySQL backend for:

- H5 public content display
- public scene page config queries
- admin login
- scene-aware content CRUD
- host to scene mapping CRUD
- scene page config CRUD
- scene-scoped image and audio upload

The feature uses three business tables:

- `scene_domains`
- `scene_page_configs`
- `content_items`

There is no `scenes` table and no extra scene registry table.

## What The Tables Mean

`scene_domains` maps `host -> scene_code`.

`scene_page_configs` stores scene-level page defaults such as logo, banner, background image URL, default fallback background URL, default music URL, default text, default tags, and colors.

`content_items` stores daily content for each `scene_code + date`.

## How Public H5 Works

Public H5 requests resolve the current scene from the request `Host`.

Resolution flow:

1. read `Host`
2. strip the port if present
3. query `scene_domains`
4. use the mapped `scene_code`
5. fetch `/api/public/scene-page-config`
6. fetch `/api/public/content?date=YYYY-MM-DD`

Fallback rules:

- `logo` comes from `scene_page_configs.logo`
- `banner` comes from `scene_page_configs.banner`
- `GET /api/public/content` returns raw `content_items` data only; H5 merges fallback values from `scene_page_configs`
- background image uses `content_items.bg_url` first when non-empty, then `scene_page_configs.bac_img`, then `scene_page_configs.default_bg_url`, then the existing hardcoded H5 fallback
- music uses `content_items.music` first when non-empty, then `scene_page_configs.default_music`, then the existing hardcoded H5 fallback
- text uses `content_items.text` first when non-empty, then `scene_page_configs.text_default`, then the existing hardcoded H5 fallback
- tags use `content_items.tags` first when non-empty, then `scene_page_configs.tags_default`, then the existing hardcoded H5 fallback
- colors come from `scene_page_configs`

The H5 frontend should use relative API paths only.

## How Admin Works

Admin remains one management system on a fixed admin domain.

The admin frontend should:

- manage host mappings in `scene_domains`
- manage page defaults in `scene_page_configs`
- manage daily content in `content_items`
- upload images through `POST /api/admin/upload/image`
- upload audio through `POST /api/admin/upload/audio`
- save the returned image `data.url` string into `logo`, `banner`, `bac_img`, or `default_bg_url`
- save the returned audio `data.url` string into `default_music`

`logo`, `banner`, `bac_img`, `default_bg_url`, and `default_music` are not uploaded directly through the scene page config APIs. Those APIs store URL strings only.

## Repository State

This repository contains the backend service, deployment scripts, and Nginx/systemd examples.

It does not currently include checked-in H5 frontend source, admin frontend source, or built frontend bundles. The deployment configs expect bundles to be copied to:

- `/opt/rainbow-backend/www/h5`
- `/opt/rainbow-backend/www/admin`

## API Contract

The source of truth is [docs/api-spec.md](docs/api-spec.md).

Key rules:

- `bg_url` is the only background image field for content
- `tags` is always a string array
- `tags_default` is always a string array in responses
- `date` uses `YYYY-MM-DD`
- `scene_code` and `date` are required for admin content create and update; `text`, `tags`, `bg_url`, and `music` are optional
- image and audio fields store URL strings, not binary data
- public content resolves scene from `Host`
- `scene_domains` only stores `host` and `scene_code`
- `scene_page_configs` stores scene-level page defaults

## Requirements

- Go 1.22+
- MySQL 8.0+
- Bash
- `systemd`
- `nginx`
- `sudo`

## Environment Variables

Copy `.env.example` for local development:

```bash
cp .env.example .env
```

Important variables:

- `APP_ENV`: `dev`, `test`, or `prod`
- `HOST`: backend listen host
- `PORT`: backend listen port
- `DB_DRIVER`: currently only `mysql`
- `DB_DSN`: MySQL DSN
- `LOG_ROOT`: directory for `app.log` and `access.log`
- `JWT_SECRET`: JWT signing secret
- `JWT_EXPIRES_IN`: token expiry in seconds
- `ADMIN_USERNAME`: seeded admin username
- `ADMIN_PASSWORD`: seeded admin password
- `ALLOW_ORIGINS`: CORS allowlist
- `UPLOAD_ROOT`: upload root
- `UPLOAD_IMAGE_MAX_SIZE`: image upload size limit
- `UPLOAD_AUDIO_MAX_SIZE`: audio upload size limit
- `ENABLE_PUBLIC_SCENE_OVERRIDE`: optional debug switch for public `scene` query override

Default upload roots:

- `dev`: `./uploads/dev`
- `test`: `/opt/rainbow-backend/uploads/test`
- `prod`: `/opt/rainbow-backend/uploads/prod`

Scene-specific uploads are stored under:

- `./uploads/dev/<scene_code>/images`
- `./uploads/dev/<scene_code>/audio`
- `/opt/rainbow-backend/uploads/test/<scene_code>/images`
- `/opt/rainbow-backend/uploads/test/<scene_code>/audio`
- `/opt/rainbow-backend/uploads/prod/<scene_code>/images`
- `/opt/rainbow-backend/uploads/prod/<scene_code>/audio`

Scene-scoped static paths:

- `/static/<scene_code>/images/<filename>`
- `/static/<scene_code>/audio/<filename>`

## Database Behavior

### `content_items`

`content_items` includes:

- `scene_code`
- unique constraint on `(scene_code, date)`
- optional `text`, `tags`, `bg_url`, and `music`

Migration behavior:

- existing rows with empty or missing `scene_code` are backfilled to `default`
- legacy date-only unique indexes are removed

### `scene_domains`

`scene_domains` is the only host-mapping table used for this feature.

Business fields:

- `host`
- `scene_code`

### `scene_page_configs`

`scene_page_configs` stores:

- `scene_code`
- `logo`
- `banner`
- `bac_img`
- `default_bg_url`
- `default_music`
- `text_default`
- `tags_default`
- `play_button_color`
- `text_default_color`
- `tags_color`
- `tags_bac_color`
- `date_color`

## Local Development

1. Create a local MySQL database such as `rainbow`.
2. Copy `.env.example` to `.env`.
3. Edit `.env`.
4. Start the service:

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
ENABLE_PUBLIC_SCENE_OVERRIDE=false
```

Local health check:

```bash
curl http://127.0.0.1:8080/health
```

## Deployment

### Dual Environment

The repository keeps the existing test/prod split:

- test backend: `127.0.0.1:18081`
- prod backend: `127.0.0.1:28081`
- test public port: `18080`
- prod public port: `28080`

Deployment examples:

- [deploy/env/test.env.example](deploy/env/test.env.example)
- [deploy/env/prod.env.example](deploy/env/prod.env.example)
- [deploy/systemd/rainbow-backend-test.service](deploy/systemd/rainbow-backend-test.service)
- [deploy/systemd/rainbow-backend-prod.service](deploy/systemd/rainbow-backend-prod.service)
- [deploy/nginx/rainbow-backend-ports.conf](deploy/nginx/rainbow-backend-ports.conf)

One-command deployment:

```bash
bash ./scripts/test.sh
```

```bash
bash ./scripts/prod.sh
```

The scripts:

1. build `./bin/rainbow-backend`
2. copy the binary into `/opt/rainbow-backend/bin`
3. install env files if missing
4. ensure upload and log directories exist
5. ensure frontend static roots exist under `/opt/rainbow-backend/www/h5` and `/opt/rainbow-backend/www/admin`
6. install systemd units
7. install the shared Nginx config
8. reload systemd
9. restart the matching backend service
10. validate and reload Nginx

### Nginx Behavior

The provided Nginx examples are designed so that:

- the same H5 bundle can be served under multiple public hosts
- `/api/` preserves the original `Host` header
- `/static/` is proxied to the backend so uploaded files are served from the runtime `UPLOAD_ROOT`
- the backend can resolve `scene_code` from `Host`
- the admin panel remains a single fixed domain
- no scene-specific backend mapping is hardcoded in Go code

Public H5 should use relative API paths such as:

```text
/api/public/content?date=2026-04-07
```

## API Examples

Public content by host:

```bash
curl -H 'Host: love.example.com' \
  'http://127.0.0.1:8080/api/public/content?date=2026-04-07'
```

Public scene page config:

```bash
curl -H 'Host: love.example.com' \
  'http://127.0.0.1:8080/api/public/scene-page-config'
```

Current host mapping:

```bash
curl -H 'Host: love.example.com' \
  'http://127.0.0.1:8080/api/public/scene-domain-mapping'
```

Admin login:

```bash
curl -X POST 'https://admin.example.com/api/admin/login' \
  -H 'Content-Type: application/json' \
  -d '{
    "username": "admin",
    "password": "your_admin_password"
  }'
```

Save the token:

```bash
TOKEN=$(curl -s -X POST 'https://admin.example.com/api/admin/login' \
  -H 'Content-Type: application/json' \
  -d '{
    "username": "admin",
    "password": "your_admin_password"
  }' | jq -r '.data.token')
```

Create content with `scene_code`:

```bash
curl -X POST 'https://admin.example.com/api/admin/content' \
  -H "Authorization: Bearer ${TOKEN}" \
  -H 'Content-Type: application/json' \
  -d '{
    "scene_code": "love",
    "date": "2026-04-07"
  }'
```

Partial content example:

```bash
curl -X POST 'https://admin.example.com/api/admin/content' \
  -H "Authorization: Bearer ${TOKEN}" \
  -H 'Content-Type: application/json' \
  -d '{
    "scene_code": "love",
    "date": "2026-04-07",
    "text": "Today is a good day."
  }'
```

Create scene page config:

```bash
curl -X POST 'https://admin.example.com/api/admin/scene-page-configs' \
  -H "Authorization: Bearer ${TOKEN}" \
  -H 'Content-Type: application/json' \
  -d '{
    "scene_code": "love",
    "logo": "/static/love/images/logo_xxx.png",
    "banner": "/static/love/images/banner_xxx.png",
    "bac_img": "/static/love/images/bg_xxx.jpg",
    "default_bg_url": "/static/love/images/default_bg_xxx.jpg",
    "default_music": "/static/love/audio/default_music_xxx.mp3",
    "text_default": "今天也是值得被温柔对待的一天。",
    "tags_default": ["心动", "温柔", "春天"],
    "play_button_color": "#1a2b3c",
    "text_default_color": "#1a2b3c",
    "tags_color": "#1a2b3c",
    "tags_bac_color": "#ffffff",
    "date_color": "#1a2b3c"
  }'
```

Upload image for a scene:

```bash
curl -X POST 'https://admin.example.com/api/admin/upload/image' \
  -H "Authorization: Bearer ${TOKEN}" \
  -F 'scene_code=love' \
  -F 'file=@/absolute/path/to/logo.png'
```

Upload audio for a scene:

```bash
curl -X POST 'https://admin.example.com/api/admin/upload/audio' \
  -H "Authorization: Bearer ${TOKEN}" \
  -F 'scene_code=love' \
  -F 'file=@/absolute/path/to/music.mp3'
```

## Manual Verification

1. Create a host mapping:

```bash
curl -X POST 'http://127.0.0.1:8080/api/admin/scene-domains' \
  -H "Authorization: Bearer ${TOKEN}" \
  -H 'Content-Type: application/json' \
  -d '{
    "host": "love.dapinsport.cn",
    "scene_code": "love"
  }'
```

2. Upload logo, banner, bac_img, and default background images with `POST /api/admin/upload/image` and copy the returned `data.url` values.

3. Upload default music with `POST /api/admin/upload/audio` and copy the returned `data.url`.

4. Create `scene_page_configs` for `scene_code=love` using those URLs.

5. Create content for `scene_code=love`, `date=2026-04-07`.

6. Request public scene page config through the mapped host:

```bash
curl -H 'Host: love.dapinsport.cn' \
  'http://127.0.0.1:8080/api/public/scene-page-config'
```

7. Request public content through the same host:

```bash
curl -H 'Host: love.dapinsport.cn' \
  'http://127.0.0.1:8080/api/public/content?date=2026-04-07'
```

8. Verify H5 renders the configured logo, banner, fallback background, fallback music, default text, default tags, and configured colors.

9. Verify admin CRUD for scene page configs works.

## Validation Commands

```bash
env GOCACHE=/tmp/go-build-cache GOMODCACHE=/tmp/go-mod-cache go test ./...
```

```bash
env GOCACHE=/tmp/go-build-cache GOMODCACHE=/tmp/go-mod-cache go build ./...
```

```bash
bash -n ./scripts/test.sh
bash -n ./scripts/prod.sh
```
