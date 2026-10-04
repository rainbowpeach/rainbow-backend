# Rainbow Backend API Spec


## 1. Overview

This backend serves:

- public H5 content queries
- public scene page config queries
- admin login
- scene-aware content CRUD
- host to scene mapping CRUD
- scene page config CRUD
- image and audio upload

For this design, the business feature relies on three business tables:

- `scene_domains`
- `scene_page_configs`
- `content_items`

There is no `scenes` table in this version.

Public requests resolve `scene_code` dynamically from the incoming HTTP `Host` header by querying `scene_domains`.

Example mappings:

- `love.example.com` -> `love`
- `sweet.example.com` -> `sweet`

## 2. Common Rules

### 2.1 Content Type

- request and response bodies use `application/json`
- character encoding uses `UTF-8`

### 2.2 Date Format

All `date` fields use:

```text
YYYY-MM-DD
```

Example:

```text
2026-04-07
```

### 2.3 Response Envelope

All APIs return the stable JSON envelope:

```json
{
  "code": 0,
  "message": "ok",
  "data": {}
}
```

### 2.4 Field Rules

- use `bg_url` as the only background image field
- do not use `bg`
- `tags` must be a string array
- `tags_default` must be a string array in API responses
- `date` must use `YYYY-MM-DD`
- for admin content create, `scene_code` and `date` are required; `text`, `tags`, `image_urls`, `bg_url`, and `music` are optional
- for admin content update, `scene_code` and `date` stay required in the current API; `text`, `tags`, `image_urls`, `bg_url`, and `music` are optional
- `scene_code` is required for scene page config create and update
- use `tags_default`, not `tags_dafult`
- `logo`, `banner`, `bac_img`, `default_bg_url`, `default_music`, `video_url`, and `avatar_url` are URL fields stored as strings
- `logo`, `banner`, `bac_img`, `default_bg_url`, `default_music`, `video_url`, and `avatar_url` must not store binary media data

### 2.5 Suggested Error Codes

| code | message | meaning |
|---|---|---|
| 0 | ok | success |
| 40001 | invalid params | request parameters are invalid |
| 40002 | invalid date format | `date` format is invalid |
| 40003 | content not found | content record not found |
| 40004 | unauthorized | login required or token invalid |
| 40005 | forbidden | no permission |
| 40006 | duplicate date | duplicate `(scene_code, date)` content |
| 40007 | duplicate host | duplicate `scene_domains.host` |
| 40009 | scene domain not found | host mapping not found |
| 40010 | duplicate scene_code | duplicate `scene_page_configs.scene_code` |
| 40011 | scene page config not found | page config record not found |
| 40012 | user not found | authenticated user record not found |
| 50000 | internal server error | server error |

## 3. Data Models

### 3.1 ContentItem

```json
{
  "id": 1,
  "scene_code": "love",
  "date": "2026-04-07",
  "text": "Today is a good day.",
  "tags": ["warm", "spring", "joy"],
  "image_urls": [
    "https://love.example.com/static/love/images/detail-1.png",
    "https://love.example.com/static/love/images/detail-2.png"
  ],
  "bg_url": "https://love.example.com/static/love/images/demo.png",
  "music": "https://love.example.com/static/love/audio/demo.mp3",
  "createdAt": "2026-04-07",
  "updatedAt": "2026-04-07"
}
```

| field | type | required | meaning |
|---|---|---:|---|
| id | number | yes | content ID |
| scene_code | string | yes | scene identifier |
| date | string | yes | display date |
| text | string | no | content text |
| tags | array[string] | no | tag list |
| image_urls | array[string] | no | extra image URL list |
| bg_url | string | no | background image URL |
| music | string | no | background music URL |
| createdAt | string | no | creation date |
| updatedAt | string | no | update date |

Constraints:

- `scene_code` is required
- `(scene_code, date)` must be unique
- empty `text`, `tags`, `image_urls`, `bg_url`, and `music` values are allowed
- existing legacy rows are backfilled to `scene_code=default`

### 3.2 SceneDomain

```json
{
  "host": "love.example.com",
  "scene_code": "love"
}
```

Rules:

- `host` uniquely maps to one `scene_code`
- `host` must be unique
- `scene_code` must be non-empty
- `scene_domains` is the only source of truth for host to scene mapping

### 3.3 ScenePageConfig

```json
{
  "scene_code": "love",
  "logo": "/static/love/images/logo_xxx.png",
  "banner": "/static/love/images/banner_xxx.png",
  "bac_img": "/static/love/images/bg_xxx.jpg",
  "default_bg_url": "/static/love/images/default_bg_xxx.jpg",
  "default_music": "/static/love/audio/default_music_xxx.mp3",
  "text_default": "今天也是值得被温柔对待的一天。",
  "tags_default": ["心动", "温柔", "春天"],
  "image_urls": ["/static/love/images/1.png", "/static/love/images/2.png"],
  "video_url": "/static/love/videos/intro.mp4",
  "avatar_url": "/static/love/avatars/avatar.png",
  "play_button_color": "#1a2b3c",
  "text_default_color": "#1a2b3c",
  "tags_color": "#1a2b3c",
  "tags_bac_color": "#ffffff",
  "date_color": "#1a2b3c"
}
```

Suggested MySQL fields:

- `scene_code VARCHAR(128) PRIMARY KEY`
- `logo VARCHAR(1024)`
- `banner VARCHAR(1024)`
- `bac_img VARCHAR(1024)`
- `default_bg_url VARCHAR(1024)`
- `default_music VARCHAR(1024)`
- `text_default TEXT`
- `tags_default JSON`
- `image_urls JSON`
- `video_url VARCHAR(1024)`
- `avatar_url VARCHAR(1024)`
- `play_button_color VARCHAR(32)`
- `text_default_color VARCHAR(32)`
- `tags_color VARCHAR(32)`
- `tags_bac_color VARCHAR(32)`
- `date_color VARCHAR(32)`
- `created_at DATETIME`
- `updated_at DATETIME`

Rules:

- `scene_code` is required and unique
- `scene_code` is the primary key
- `logo`, `banner`, `bac_img`, `default_bg_url`, `default_music`, `video_url`, and `avatar_url` are URL fields stored as strings
- the database stores URL/path strings only, never image, video, or audio binary data
- `tags_default` is stored as JSON in MySQL and returned as `array[string]`
- `image_urls` is stored as JSON in MySQL and returned as `array[string]`
- color fields must be valid hex color strings when non-empty
- scene page config image, video, audio, and avatar fields are populated through the upload APIs before being saved

## 4. Auth

Admin APIs require:

```http
Authorization: Bearer <token>
```

Only `POST /api/admin/login` is public under `/api/admin`.

### 4.1 WeChat Mini Program Login

- Path: `POST /api/user/login`
- Frontend URL: `https://dapinsport.cn/api/user/login`
- Auth required: no

Request body:

```json
{
  "code": "code_from_wx.login"
}
```

The backend calls the configured WeChat `jscode2session` endpoint with `appid`, `secret`, `js_code`, and `grant_type=authorization_code`, then reads `openid`. `session_key` is not returned. A user is found or created by unique `openid`; the backend generates a secure random token, saves it with `token_expire_at` and `last_login_at`, and returns the token.

Success response:

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "token": "random_token",
    "expiresIn": 2592000
  }
}
```

User-protected routes use:

```http
Authorization: Bearer <token>
```

The backend looks up `users.token`, checks `token_expire_at`, and provides the user ID to the handler. Missing, unknown, or expired tokens return HTTP `401` with code `40004`.

### 4.2 User Profile

All user profile APIs use the following frontend URL prefix:

```text
https://dapinsport.cn
```

All profile APIs require:

```http
Authorization: Bearer <token>
```

The authenticated user is identified by the token. The frontend must not send
`user_id` to select another user.

#### Get current user profile

- Path: `GET /api/user/profile`
- URL: `https://dapinsport.cn/api/user/profile`
- Auth required: yes

Success response:

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "user_id": 1,
    "scene_code": ["curry", "love"],
    "nickname": "小明",
    "avatar_url": "https://dapinsport.cn/static/user/avatars/20260919010101-ab12cd34.png",
    "birthday": "1990-01-02",
    "gender": 1,
    "occupation": 2
  }
}
```

When an optional value has not been set, `scene_code` is `[]`, text fields
are empty strings, and `birthday`, `gender`, and `occupation` are `null`.

#### Create/save or update profile fields

- Path: `POST /api/user/profile`
- URL: `https://dapinsport.cn/api/user/profile`
- Auth required: yes

- Path: `PUT /api/user/profile`
- URL: `https://dapinsport.cn/api/user/profile`
- Auth required: yes

Both methods use partial-update semantics. Every request field is optional.
Fields not included in the request remain unchanged.

Request body:

```json
{
  "scene_code": ["curry", "love"],
  "nickname": "小明",
  "avatar_url": "https://dapinsport.cn/static/user/avatars/avatar.png",
  "birthday": "1990-01-02",
  "gender": 1,
  "occupation": 2
}
```

Field rules:

| field | JSON type | required | meaning |
|---|---|---:|---|
| `scene_code` | array[string] | no | user scene code list; each value must match the scene code format |
| `nickname` | string | no | nickname, up to 128 characters |
| `avatar_url` | string | no | URL returned by the avatar upload API, or another valid HTTP(S)/static URL |
| `birthday` | string | no | date in `YYYY-MM-DD` format |
| `gender` | integer | no | business-defined gender code |
| `occupation` | integer | no | business-defined occupation code |

To clear a field, send `null` where supported, an empty string for text/date
fields, or an empty array for `scene_code`.

#### Clear profile fields

- Path: `DELETE /api/user/profile`
- URL: `https://dapinsport.cn/api/user/profile`
- Auth required: yes

This clears the six profile fields but does not delete the `users` row,
`openid`, `user_id`, or the current login account. The response is the cleared
profile.

### 4.3 User Avatar Upload

- Path: `POST /api/user/upload/avatar`
- URL: `https://dapinsport.cn/api/user/upload/avatar`
- Auth required: yes
- Content-Type: `multipart/form-data`
- File field: `file`

Example:

```bash
curl -X POST 'https://dapinsport.cn/api/user/upload/avatar' \
  -H 'Authorization: Bearer <token>' \
  -F 'file=@avatar.png'
```

Success response:

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "url": "https://dapinsport.cn/static/user/avatars/20260919010101-ab12cd34.png",
    "filename": "20260919010101-ab12cd34.png",
    "size": 12345,
    "contentType": "image/png"
  }
}
```

The upload API only generates and returns the avatar URL. The frontend should
then send that URL as `avatar_url` to `POST` or `PUT /api/user/profile`.
Supported image formats follow the existing image upload rules: JPG, JPEG,
PNG, and WebP.

## 5. Public APIs

### 5.1 Get Public Content

- Path: `GET /api/public/content`
- Purpose: get content by `date`; scene is resolved from `Host`

Query parameters:

| name | type | required | meaning |
|---|---|---:|---|
| date | string | yes | content date |
| scene | string | no | debug override only when backend config enables it |

Host resolution behavior:

1. read the request `Host`
2. strip the port if present
3. query `scene_domains.host`
4. if a mapping exists, use its `scene_code`
5. if no mapping exists, return a clear not-configured error

Example request:

```http
GET /api/public/content?date=2026-04-07
Host: love.example.com
```

Success example:

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "id": 1,
    "scene_code": "love",
    "date": "2026-04-07",
    "text": "Today is a good day.",
    "tags": ["warm", "spring", "joy"],
    "image_urls": [
      "https://love.example.com/static/love/images/detail-1.png",
      "https://love.example.com/static/love/images/detail-2.png"
    ],
    "bg_url": "https://love.example.com/static/love/images/demo.png",
    "music": "https://love.example.com/static/love/audio/demo.mp3",
    "createdAt": "2026-04-07",
    "updatedAt": "2026-04-07"
  }
}
```

Host not configured example:

```json
{
  "code": 40009,
  "message": "scene not configured",
  "data": null
}
```

### 5.2 Get Current Scene-Domain Mapping

- Path: `GET /api/public/scene-domain-mapping`
- Purpose: inspect the current host resolution result

Example request:

```http
GET /api/public/scene-domain-mapping
Host: sweet.example.com
```

Success example:

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "host": "sweet.example.com",
    "scene_code": "sweet"
  }
}
```

### 5.3 Get Current Scene Page Config

- Path: `GET /api/public/scene-page-config`
- Purpose: resolve `scene_code` from `Host` and return the scene page config

Query parameters:

| name | type | required | meaning |
|---|---|---:|---|
| scene | string | no | debug override only when backend config enables it |

Behavior:

1. read `Host` from the request
2. if `ENABLE_PUBLIC_SCENE_OVERRIDE=true` and `scene` is present, use that `scene_code`
3. otherwise resolve `scene_code` through `scene_domains`
4. query `scene_page_configs` by `scene_code`
5. return the current scene page config
6. if no scene mapping exists, return a clear JSON error
7. if scene mapping exists but no page config exists, return a clear JSON not-found error

Success example:

```json
{
  "code": 0,
  "message": "ok",
  "data": {
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
  }
}
```

Scene page config not found example:

```json
{
  "code": 40011,
  "message": "scene page config not found",
  "data": null
}
```

### 5.4 Public Fallback Logic For H5

H5 should request both:

- `GET /api/public/scene-page-config`
- `GET /api/public/content?date=YYYY-MM-DD`

`GET /api/public/content` returns the raw `content_items` row for the date. It does not merge `scene_page_configs` fallback values into `text`, `tags`, `bg_url`, or `music`.

Rendering rules:

1. `logo` comes from `scene_page_configs.logo`
2. `banner` comes from `scene_page_configs.banner`
3. background image uses `content_items.bg_url` first when non-empty, then `scene_page_configs.bac_img`, then `scene_page_configs.default_bg_url`, then the existing hardcoded H5 fallback
4. music uses `content_items.music` first when non-empty, then `scene_page_configs.default_music`, then the existing hardcoded H5 fallback
5. text uses `content_items.text` first when non-empty, then `scene_page_configs.text_default`, then the existing hardcoded H5 fallback
6. tags use `content_items.tags` first when non-empty, then `scene_page_configs.tags_default`, then the existing hardcoded H5 fallback
7. play button color comes from `scene_page_configs.play_button_color`
8. default text color comes from `scene_page_configs.text_default_color`
9. tags font color comes from `scene_page_configs.tags_color`
10. tags background color comes from `scene_page_configs.tags_bac_color`
11. date color comes from `scene_page_configs.date_color`

### 5.5 Explicit Scene Query APIs

These APIs are separate from `/api/public/*`.

- frontend URL prefix: `https://dapinsport.cn`
- authentication: not required
- `scene_code` is supplied explicitly by the client
- these APIs do not read `Host`, query `scene_domains`, or perform Host to
  `scene_code` resolution

#### Get Scene Page Config By Scene Code

- Path: `GET /api/query/scene-page-config`
- URL: `https://dapinsport.cn/api/query/scene-page-config`

Query parameters:

| name | type | required | meaning |
|---|---|---:|---|
| scene_code | string | yes | scene identifier |

Example:

```http
GET https://dapinsport.cn/api/query/scene-page-config?scene_code=curry
```

The backend queries `scene_page_configs` by `scene_code`.

#### Get Content By Scene Code And Date

- Path: `GET /api/query/content`
- URL: `https://dapinsport.cn/api/query/content`

Query parameters:

| name | type | required | meaning |
|---|---|---:|---|
| scene_code | string | yes | scene identifier |
| date | string | yes | content date in `YYYY-MM-DD` |

Example:

```http
GET https://dapinsport.cn/api/query/content?scene_code=curry&date=2026-09-19
```

The backend queries `content_items` by the exact `(scene_code, date)` pair.

Both endpoints use the standard response envelope. Invalid parameters return
HTTP `400`; a missing scene page config returns HTTP `404` with code `40011`;
a missing content record returns HTTP `404` with code `40003`.

## 6. Admin APIs

### 6.1 Admin Login

- Path: `POST /api/admin/login`

Request body:

```json
{
  "username": "admin",
  "password": "123456"
}
```

Success response:

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

### 6.2 Create Content

- Path: `POST /api/admin/content`
- Auth required: yes

Field requirements:

| field | required | notes |
|---|---:|---|
| scene_code | yes | scene identifier |
| date | yes | must use `YYYY-MM-DD` |
| text | no | empty or missing value falls back on H5 |
| tags | no | if provided, must be an array of strings; `[]` is allowed |
| image_urls | no | if provided, must be an array of strings; `[]` is allowed |
| bg_url | no | empty string is allowed |
| music | no | empty string is allowed |

Request body:

```json
{
  "scene_code": "love",
  "date": "2026-04-07"
}
```

Partial create example:

```json
{
  "scene_code": "love",
  "date": "2026-04-07",
  "text": "Today is a good day."
}
```

### 6.3 Update Content

- Path: `PUT /api/admin/content/:id`
- Auth required: yes

Request body is the same as create. In the current API, `scene_code` and `date` are still required for update, while `text`, `tags`, `image_urls`, `bg_url`, and `music` stay optional.

Empty optional fields on create or update use the public H5 fallback order from `scene_page_configs` values before the existing hardcoded H5 fallback.

### 6.4 Delete Content

- Path: `DELETE /api/admin/content/:id`
- Auth required: yes

### 6.5 List Content

- Path: `GET /api/admin/content`
- Auth required: yes

Query parameters:

| name | type | required | meaning |
|---|---|---:|---|
| page | number | yes | page number, starts at 1 |
| pageSize | number | yes | page size |
| scene | string | no | filter by `scene_code` |
| date | string | no | filter by date |

### 6.6 List Scene-Domain Mappings

- Path: `GET /api/admin/scene-domains`
- Auth required: yes

Query parameters:

| name | type | required | meaning |
|---|---|---:|---|
| page | number | no | default `1` |
| pageSize | number | no | default `10`, max `100` |
| host | string | no | fuzzy host filter |
| scene | string | no | exact `scene_code` filter |

### 6.7 Get Scene-Domain Mapping Detail

- Path: `GET /api/admin/scene-domains/:host`
- Auth required: yes

### 6.8 Create Scene-Domain Mapping

- Path: `POST /api/admin/scene-domains`
- Auth required: yes

Request body:

```json
{
  "host": "love.example.com",
  "scene_code": "love"
}
```

### 6.9 Update Scene-Domain Mapping

- Path: `PUT /api/admin/scene-domains/:host`
- Auth required: yes

Request body:

```json
{
  "host": "sweet.example.com",
  "scene_code": "sweet"
}
```

Notes:

- request body `host` is optional on update
- if `host` is omitted or blank, the service uses the `:host` path value

### 6.10 Delete Scene-Domain Mapping

- Path: `DELETE /api/admin/scene-domains/:host`
- Auth required: yes

### 6.11 List Scene Page Configs

- Path: `GET /api/admin/scene-page-configs`
- Auth required: yes

Query parameters:

| name | type | required | meaning |
|---|---|---:|---|
| page | number | no | default `1` |
| pageSize | number | no | default `10`, max `100` |
| scene | string | no | exact `scene_code` filter |

Success response example:

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "list": [
      {
        "scene_code": "love",
        "logo": "/static/love/images/logo_xxx.png",
        "banner": "/static/love/images/banner_xxx.png",
        "bac_img": "/static/love/images/bg_xxx.jpg",
        "default_bg_url": "/static/love/images/default_bg_xxx.jpg",
        "default_music": "/static/love/audio/default_music_xxx.mp3",
        "text_default": "今天也是值得被温柔对待的一天。",
        "tags_default": ["心动", "温柔", "春天"],
        "image_urls": [
          "/static/love/images/1.png",
          "/static/love/images/2.png"
        ],
        "play_button_color": "#1a2b3c",
        "text_default_color": "#1a2b3c",
        "tags_color": "#1a2b3c",
        "tags_bac_color": "#ffffff",
        "date_color": "#1a2b3c"
      }
    ],
    "total": 1,
    "page": 1,
    "pageSize": 10
  }
}
```

### 6.12 Get Scene Page Config Detail

- Path: `GET /api/admin/scene-page-configs/:scene_code`
- Auth required: yes

### 6.13 Create Scene Page Config

- Path: `POST /api/admin/scene-page-configs`
- Auth required: yes

Request body:

```json
{
  "scene_code": "love",
  "logo": "/static/love/images/logo_xxx.png",
  "banner": "/static/love/images/banner_xxx.png",
  "bac_img": "/static/love/images/bg_xxx.jpg",
  "default_bg_url": "/static/love/images/default_bg_xxx.jpg",
  "default_music": "/static/love/audio/default_music_xxx.mp3",
  "text_default": "今天也是值得被温柔对待的一天。",
  "tags_default": ["心动", "温柔", "春天"],
  "image_urls": ["/static/love/images/1.png", "/static/love/images/2.png"],
  "play_button_color": "#1a2b3c",
  "text_default_color": "#1a2b3c",
  "tags_color": "#1a2b3c",
  "tags_bac_color": "#ffffff",
  "date_color": "#1a2b3c"
}
```

Validation:

- `scene_code` is required
- `tags_default` must be a string array
- `image_urls` must be a string array
- color fields must be valid hex colors when non-empty
- `logo`, `banner`, `bac_img`, `default_bg_url`, and `default_music` must be URL strings or static paths
- create rejects duplicate `scene_code`
- image binary data is not accepted in this API
- audio binary data is not accepted in this API

### 6.14 Update Scene Page Config

- Path: `PUT /api/admin/scene-page-configs/:scene_code`
- Auth required: yes

Request body is the same as create.

Notes:

- the path `scene_code` identifies the record
- if the body includes `scene_code`, it must match the path value
- update returns a clear not-found error if the record does not exist
- `default_bg_url` should be populated through `POST /api/admin/upload/image`
- `default_music` should be populated through `POST /api/admin/upload/audio`
- `video_url` should be populated through `POST /api/admin/upload/video`
- `avatar_url` should be populated through `POST /api/admin/upload/avatar`

### 6.15 Delete Scene Page Config

- Path: `DELETE /api/admin/scene-page-configs/:scene_code`
- Auth required: yes

Delete returns a clear not-found error if the record does not exist.

### 6.16 Upload Image

- Path: `POST /api/admin/upload/image`
- Auth required: yes
- Content-Type: `multipart/form-data`

Form fields:

| name | type | required | meaning |
|---|---|---:|---|
| file | file | yes | image file |
| scene_code | string | no | upload scene, defaults to `default` |

Example:

```bash
curl -X POST 'https://admin.example.com/api/admin/upload/image' \
  -H "Authorization: Bearer ${TOKEN}" \
  -F 'scene_code=love' \
  -F 'file=@/absolute/path/to/logo.png'
```

Success response:

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "url": "/static/love/images/20260415093015-ab12cd34.png",
    "filename": "20260415093015-ab12cd34.png",
    "size": 12345,
    "contentType": "image/png"
  }
}
```

The returned `data.url` value is what admin should save into:

- `scene_page_configs.logo`
- `scene_page_configs.banner`
- `scene_page_configs.bac_img`
- `scene_page_configs.default_bg_url`
- `content_items.bg_url`

Storage layout:

- dev: `./uploads/dev/<scene_code>/images/`
- test: `/opt/rainbow-backend/uploads/test/<scene_code>/images/`
- prod: `/opt/rainbow-backend/uploads/prod/<scene_code>/images/`

### 6.17 Upload Audio

- Path: `POST /api/admin/upload/audio`
- Auth required: yes
- Content-Type: `multipart/form-data`

Form fields:

| name | type | required | meaning |
|---|---|---:|---|
| file | file | yes | audio file |
| scene_code | string | no | upload scene, defaults to `default` |

The returned `data.url` value is what admin should save into:

- `scene_page_configs.default_music`
- `content_items.music`

### 6.18 Upload Video

- Path: `POST /api/admin/upload/video`
- Auth required: yes
- Content-Type: `multipart/form-data`

Form fields:

| name | type | required | meaning |
|---|---|---:|---|
| file | file | yes | video file |
| scene_code | string | no | upload scene, defaults to `default` |

Supported video containers are MP4, WebM, and QuickTime/MOV.

Example:

```bash
curl -X POST 'https://admin.example.com/api/admin/upload/video' \
  -H "Authorization: Bearer ${TOKEN}" \
  -F 'scene_code=love' \
  -F 'file=@/absolute/path/to/intro.mp4'
```

The returned `data.url` value should be saved into `scene_page_configs.video_url`.

Storage layout:

- dev: `./uploads/dev/<scene_code>/videos/`
- test: `/opt/rainbow-backend/uploads/test/<scene_code>/videos/`
- prod: `/opt/rainbow-backend/uploads/prod/<scene_code>/videos/`

### 6.19 Upload Scene Avatar

- Path: `POST /api/admin/upload/avatar`
- Auth required: yes
- Content-Type: `multipart/form-data`

Form fields:

| name | type | required | meaning |
|---|---|---:|---|
| file | file | yes | avatar image |
| scene_code | string | no | upload scene, defaults to `default` |

Example:

```bash
curl -X POST 'https://admin.example.com/api/admin/upload/avatar' \
  -H "Authorization: Bearer ${TOKEN}" \
  -F 'scene_code=love' \
  -F 'file=@/absolute/path/to/avatar.png'
```

The returned `data.url` value should be saved into `scene_page_configs.avatar_url`.

Storage layout:

- dev: `./uploads/dev/<scene_code>/avatars/`
- test: `/opt/rainbow-backend/uploads/test/<scene_code>/avatars/`
- prod: `/opt/rainbow-backend/uploads/prod/<scene_code>/avatars/`

## 7. Frontend Integration Notes

### 7.1 H5 Frontend

The H5 frontend should:

- use relative API paths only
- request `/api/public/scene-page-config`
- request `/api/public/content?date=YYYY-MM-DD`
- keep current date query behavior
- keep current music play behavior
- hide logo or banner gracefully when empty
- background fallback order: `content_items.bg_url`, `scene_page_configs.bac_img`, `scene_page_configs.default_bg_url`, then the existing hardcoded fallback
- music fallback order: `content_items.music`, `scene_page_configs.default_music`, then the existing hardcoded fallback

### 7.2 Admin Frontend

The admin panel domain remains unchanged.

The admin frontend should:

- add a dedicated scene page config management page or section
- list configs
- create config
- edit config
- delete config
- upload `logo`, `banner`, and `bac_img` through `POST /api/admin/upload/image`
- upload `default_bg_url` through `POST /api/admin/upload/image`
- upload `default_music` through `POST /api/admin/upload/audio`
- upload `video_url` through `POST /api/admin/upload/video`
- upload `avatar_url` through `POST /api/admin/upload/avatar`
- read returned `data.url` and save that string into the form
- allow manual editing of the URL if needed
- show image preview for `default_bg_url` when practical
- provide audio preview/playback for `default_music` when practical
- provide video preview for `video_url` when practical
- show avatar preview for `avatar_url` when practical
- edit `tags_default` as an array of strings
- use color inputs or text inputs for hex color strings

## 8. Manual Verification Targets

1. Create a host mapping through `POST /api/admin/scene-domains`.
2. Upload a logo image through `POST /api/admin/upload/image` with `scene_code=love`.
3. Upload a banner image through `POST /api/admin/upload/image` with `scene_code=love`.
4. Upload a background image through `POST /api/admin/upload/image` with `scene_code=love`.
5. Upload a default background image through `POST /api/admin/upload/image` with `scene_code=love`.
6. Upload a default music file through `POST /api/admin/upload/audio` with `scene_code=love`.
7. Upload a scene video through `POST /api/admin/upload/video` with `scene_code=love`.
8. Upload a scene avatar through `POST /api/admin/upload/avatar` with `scene_code=love`.
9. Create `scene_page_configs` for `scene_code=love` using the returned upload URLs.
10. Create content for `scene_code=love`, `date=2026-04-07`.
11. Request public scene page config with `Host: love.dapinsport.cn`.
12. Request public content with `Host: love.dapinsport.cn`.
13. Verify H5 renders the configured video and avatar alongside the existing page defaults.
14. Verify admin CRUD for scene page configs.
