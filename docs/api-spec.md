# 彩虹屁项目 API 规格说明

## 1. 文档概述

本文档定义“彩虹屁”项目的前后端接口规范，用于支持：

- H5 落地页内容展示
- 后台管理系统内容维护

本文档是前后端联调、后端开发、接口测试的统一依据。

---

## 2. 基本约定

### 2.1 数据传输格式

- 请求体与响应体统一使用 `application/json`
- 字符编码统一为 `UTF-8`

### 2.2 时间格式

所有日期字段统一使用以下格式：

```text
YYYY-MM-DD
```

例如：

```text
2026-04-07
```

### 2.3 通用响应结构

后端建议统一返回如下 JSON 结构：

```json
{
  "code": 0,
  "message": "ok",
  "data": {}
}
```

字段说明：

| 字段名 | 类型 | 说明 |
|---|---|---|
| code | number | 业务状态码，`0` 表示成功，非 `0` 表示失败 |
| message | string | 响应消息 |
| data | object/null | 具体返回数据 |

### 2.5 通用错误码建议

| code | message | 说明 |
|---|---|---|
| 0 | ok | 请求成功 |
| 40001 | invalid params | 请求参数错误 |
| 40002 | invalid date format | 日期格式错误 |
| 40003 | content not found | 内容不存在 |
| 40004 | unauthorized | 未登录或认证失败 |
| 40005 | forbidden | 无权限访问 |
| 40006 | duplicate date | 该日期内容已存在 |
| 50000 | internal server error | 服务器内部错误 |

---

## 3. 数据模型定义

### 3.1 内容对象 ContentItem

```json
{
  "id": 1,
  "date": "2026-04-07",
  "text": "111",
  "tags": ["心动", "温柔", "春天"],
  "bg_url": "xxx",
  "music": "xxx",
  "createdAt": "2026-04-07",
  "updatedAt": "2026-04-07"
}
```

字段说明：

| 字段名 | 类型 | 必填 | 说明 |
|---|---|---:|---|
| id | number | 是 | 内容主键 ID |
| date | string | 是 | 对应展示日期，格式为 `YYYY-MM-DD` |
| text | string | 是 | 页面文案展示内容 |
| tags | array[string] | 是 | 标签数组，前端渲染时每个 tag 前需加 `#` |
| bg_url | string | 是 | 页面背景图片地址 |
| music | string | 是 | 页面背景音乐地址 |
| createdAt | string | 否 | 数据创建日期 |
| updatedAt | string | 否 | 数据修改日期 |

---

## 4. 认证说明

后台管理接口需要登录认证。

建议认证方式如下：

- 登录成功后返回 `token`
- 后续后台接口通过请求头携带认证信息：

```http
Authorization: Bearer <token>
```

用户名和密码由后台预先创建，不提供注册接口。

---

## 5. H5 落地页接口

### 5.1 获取指定日期内容

#### 接口说明

根据日期获取 H5 页面展示内容。

#### 请求信息

- **请求路径**：`GET /api/public/content`
- **请求方式**：`GET`

#### Query 参数

| 参数名 | 类型 | 必填 | 说明 |
|---|---|---:|---|
| date | string | 是 | 日期，格式 `YYYY-MM-DD` |

#### 请求示例

```http
GET /api/public/content?date=2026-04-07
```

#### 成功响应示例

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "id": 1,
    "date": "2026-04-07",
    "text": "111",
    "tags": ["心动", "温柔", "春天"],
    "bg_url": "xxx",
    "music": "xxx",
    "createdAt": "2026-04-07",
    "updatedAt": "2026-04-07"
  }
}
```

#### 失败响应示例

```json
{
  "code": 40003,
  "message": "content not found",
  "data": null
}
```

#### 业务说明

- 前端拿到 `tags` 后，展示时每个 tag 前添加 `#`
- 若指定日期无内容，返回“数据不存在”
- 建议一个日期只对应一条内容记录

---

## 6. 后台管理接口

### 6.1 后台登录

#### 接口说明

后台管理员登录接口。

#### 请求信息

- **请求路径**：`POST /api/admin/login`
- **请求方式**：`POST`

#### 请求体参数

| 参数名 | 类型 | 必填 | 说明 |
|---|---|---:|---|
| username | string | 是 | 后台账号 |
| password | string | 是 | 后台密码 |

#### 请求示例

```json
{
  "username": "admin",
  "password": "123456"
}
```

#### 成功响应示例

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

#### 失败响应示例

```json
{
  "code": 40004,
  "message": "unauthorized",
  "data": null
}
```

---

### 6.2 新增内容

#### 接口说明

新增一条页面内容数据。

#### 请求信息

- **请求路径**：`POST /api/admin/content`
- **请求方式**：`POST`
- **是否鉴权**：是

#### 请求头

```http
Authorization: Bearer <token>
Content-Type: application/json
```

#### 请求体参数

| 参数名 | 类型 | 必填 | 说明 |
|---|---|---:|---|
| date | string | 是 | 日期，格式 `YYYY-MM-DD` |
| text | string | 是 | 页面文案展示 |
| tags | array[string] | 是 | 标签数组 |
| bg_url | string | 是 | 背景图片地址 |
| music | string | 是 | 背景音乐地址 |

#### 请求示例

```json
{
  "date": "2026-04-07",
  "text": "111",
  "tags": ["甜", "恋爱"],
  "bg_url": "xxx",
  "music": "xxx"
}
```

#### 成功响应示例

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "id": 1
  }
}
```

#### 失败响应示例

```json
{
  "code": 40006,
  "message": "duplicate date",
  "data": null
}
```

#### 业务说明

- `date` 建议全局唯一
- `tags` 必须为字符串数组
- 新增时统一使用 `bg_url`

---

### 6.3 修改内容

#### 接口说明

根据 ID 修改一条页面内容数据。

#### 请求信息

- **请求路径**：`PUT /api/admin/content/:id`
- **请求方式**：`PUT`
- **是否鉴权**：是

#### Path 参数

| 参数名 | 类型 | 必填 | 说明 |
|---|---|---:|---|
| id | number | 是 | 内容 ID |

#### 请求体参数

| 参数名 | 类型 | 必填 | 说明 |
|---|---|---:|---|
| date | string | 是 | 日期，格式 `YYYY-MM-DD` |
| text | string | 是 | 页面文案展示 |
| tags | array[string] | 是 | 标签数组 |
| bg_url | string | 是 | 背景图片地址 |
| music | string | 是 | 背景音乐地址 |

#### 请求示例

```http
PUT /api/admin/content/1
```

```json
{
  "date": "2026-04-07",
  "text": "111",
  "tags": ["甜", "恋爱"],
  "bg_url": "xxx",
  "music": "xxx"
}
```

#### 成功响应示例

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "id": 1
  }
}
```

#### 失败响应示例

```json
{
  "code": 40003,
  "message": "content not found",
  "data": null
}
```

---

### 6.4 删除内容

#### 接口说明

根据 ID 删除一条页面内容数据。

#### 请求信息

- **请求路径**：`DELETE /api/admin/content/:id`
- **请求方式**：`DELETE`
- **是否鉴权**：是

#### Path 参数

| 参数名 | 类型 | 必填 | 说明 |
|---|---|---:|---|
| id | number | 是 | 内容 ID |

#### 请求示例

```http
DELETE /api/admin/content/1
Authorization: Bearer <token>
```

#### 成功响应示例

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "id": 1
  }
}
```

#### 失败响应示例

```json
{
  "code": 40003,
  "message": "content not found",
  "data": null
}
```

---

### 6.5 分页获取内容列表

#### 接口说明

后台分页获取内容列表，用于管理页面展示。

#### 请求信息

- **请求路径**：`GET /api/admin/content`
- **请求方式**：`GET`
- **是否鉴权**：是

#### Query 参数

| 参数名 | 类型 | 必填 | 说明 |
|---|---|---:|---|
| page | number | 是 | 页码，从 1 开始 |
| pageSize | number | 是 | 每页数量 |

#### 请求示例

```http
GET /api/admin/content?page=1&pageSize=10
Authorization: Bearer <token>
```

#### 成功响应示例

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "list": [
      {
        "id": 1,
        "date": "2026-04-07",
        "text": "111",
        "tags": ["浪漫"],
        "bg_url": "xxx",
        "music": "xxx"
      }
    ],
    "total": 1,
    "page": 1,
    "pageSize": 10
  }
}
```

#### 字段说明

| 字段名 | 类型 | 说明 |
|---|---|---|
| list | array | 当前页数据列表 |
| total | number | 总记录数 |
| page | number | 当前页码 |
| pageSize | number | 当前分页大小 |

---

### 6.6 上传背景图片

#### 接口说明

后台上传图片文件，返回可直接写入 `bg_url` 的公开访问地址。

#### 请求信息

- **请求路径**：`POST /api/admin/upload/image`
- **请求方式**：`POST`
- **是否鉴权**：是
- **Content-Type**：`multipart/form-data`

#### Form 参数

| 参数名 | 类型 | 必填 | 说明 |
|---|---|---:|---|
| file | file | 是 | 图片文件 |

#### 请求示例

```http
POST /api/admin/upload/image
Authorization: Bearer <token>
Content-Type: multipart/form-data
```

#### 成功响应示例

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

#### 业务说明

- 支持格式：`.jpg`、`.jpeg`、`.png`、`.webp`
- 默认最大文件大小：`50 MB`
- 返回的 `url` 可直接作为内容新增或修改接口中的 `bg_url`

---

### 6.7 上传背景音乐

#### 接口说明

后台上传音频文件，返回可直接写入 `music` 的公开访问地址。

#### 请求信息

- **请求路径**：`POST /api/admin/upload/audio`
- **请求方式**：`POST`
- **是否鉴权**：是
- **Content-Type**：`multipart/form-data`

#### Form 参数

| 参数名 | 类型 | 必填 | 说明 |
|---|---|---:|---|
| file | file | 是 | 音频文件 |

#### 请求示例

```http
POST /api/admin/upload/audio
Authorization: Bearer <token>
Content-Type: multipart/form-data
```

#### 成功响应示例

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "url": "http://<public-ip>:18080/static/audio/20260415093015-ab12cd34.mp3",
    "filename": "20260415093015-ab12cd34.mp3",
    "size": 456789,
    "contentType": "audio/mpeg"
  }
}
```

#### 业务说明

- 支持格式：`.mp3`、`.wav`、`.ogg`、`.m4a`
- 默认最大文件大小：`50 MB`
- 返回的 `url` 可直接作为内容新增或修改接口中的 `music`

---

## 7. 参数校验规则

### 7.1 date

- 必填
- 格式必须为 `YYYY-MM-DD`
- 示例：`2026-04-07`

### 7.2 text

- 必填
- 类型为字符串
- 不允许为空字符串

### 7.3 tags

- 必填
- 类型为字符串数组
- 至少包含 1 个标签
- 示例：`["心动", "温柔", "春天"]`

### 7.4 bg_url

- 必填
- 类型为字符串
- 建议为完整可访问的图片 URL 或资源地址

### 7.5 music

- 必填
- 类型为字符串
- 建议为完整可访问的音频 URL 或资源地址

### 7.6 page / pageSize

- 必填
- `page >= 1`
- `pageSize >= 1`
- 建议 `pageSize <= 100`

### 7.7 upload file

- `file` 为必填
- 不允许空文件
- 文件扩展名和文件内容类型都必须在允许范围内
- 图片默认大小限制为 `50 MB`
- 音频默认大小限制为 `50 MB`
- 上传成功后返回的公开 URL 可直接写入 `bg_url` 或 `music`

---

## 8. 业务约束建议

1. 一个 `date` 只允许存在一条数据，便于 H5 按日期直接读取  
2. 后台用户名和密码由系统初始化，不对外开放注册  
3. 所有后台接口都需要鉴权  
4. 后端统一字段命名为 `bg_url`  
5. 建议对后台接口增加操作日志，便于排查问题  
6. 建议上线时对图片地址和音频地址进行格式校验  

---

## 9. 接口清单总览

| 模块 | 接口 | 方法 | 是否鉴权 | 说明 |
|---|---|---|---:|---|
| H5 落地页 | `/api/public/content` | GET | 否 | 按日期获取页面内容 |
| 后台管理 | `/api/admin/login` | POST | 否 | 管理员登录 |
| 后台管理 | `/api/admin/content` | POST | 是 | 新增内容 |
| 后台管理 | `/api/admin/content/:id` | PUT | 是 | 修改内容 |
| 后台管理 | `/api/admin/content/:id` | DELETE | 是 | 删除内容 |
| 后台管理 | `/api/admin/content` | GET | 是 | 分页获取内容列表 |

---

## 10. 联调示例

### 10.1 新增一条数据

请求：

```http
POST /api/admin/content
Authorization: Bearer <token>
Content-Type: application/json
```

```json
{
  "date": "2026-04-07",
  "text": "今天也要被温柔对待呀",
  "tags": ["心动", "温柔", "春天"],
  "bg_url": "https://example.com/bg.jpg",
  "music": "https://example.com/music.mp3"
}
```

响应：

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "id": 1
  }
}
```

### 10.2 H5 按日期读取

请求：

```http
GET /api/public/content?date=2026-04-07
```

响应：

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "id": 1,
    "date": "2026-04-07",
    "text": "今天也要被温柔对待呀",
    "tags": ["心动", "温柔", "春天"],
    "bg_url": "https://example.com/bg.jpg",
    "music": "https://example.com/music.mp3",
    "createdAt": "2026-04-07",
    "updatedAt": "2026-04-07"
  }
}
```

---

## 11. 版本说明

当前版本：`v1`

如后续接口发生变更，应同步更新本文档，并以本文档作为联调基准。
