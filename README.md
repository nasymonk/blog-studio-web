# Blog Studio Web

Self-hosted Hugo blog management service with a Vue editor, real Hugo preview, safe blog publishing, backup, audit logs, rollback, WeChat draft publishing, and a Twikoo comment management entry.

**Language:** [中文说明](#中文说明) | [English](#english)

## 中文说明

Blog Studio Web 是一个 Docker 化的自托管 Hugo 博客管理后台，面向个人博客作者。它部署在博客服务器上，通过 Nginx 暴露为 `/studio/`，提供文章编辑、真实 Hugo 预览、安全发布、备份、审计日志、一键回滚、微信公众号草稿发布和 Twikoo 评论入口。

### 功能范围

- 单站点、单管理员。
- 通过挂载目录直接管理 Hugo 站点，不依赖 SSH。
- 发布前自动备份 Page Bundle，并保留最近 5 个版本。
- 发布、回滚、公众号草稿等操作写入审计日志。
- 真实 Hugo 构建预览，前端通过 iframe 查看预览结果。
- 微信公众号发布只保存到草稿箱，不自动群发。
- Twikoo 评论中心只读展示和跳转管理后台，不直接修改 Twikoo 数据。

### 安全默认值

服务启动时必须设置以下环境变量，否则拒绝启动：

- `BLOG_STUDIO_ADMIN_PASSWORD_HASH`
- `BLOG_STUDIO_SESSION_SECRET`

生成管理员密码哈希：

```bash
go run ./cmd/server hash-password 'your-long-password'
```

`BLOG_STUDIO_SESSION_SECRET` 请使用 32 位以上随机字符串。不要提交 `.env`、真实域名、IP、密码、微信公众号密钥或私钥。

### Docker 部署

```bash
cp .env.example .env
docker compose up -d
```

挂载说明：

- `/blog`：Hugo 站点根目录。
- `/data`：Blog Studio 的缓存、备份、审计日志和预览文件。
- 可选 `/twikoo/db.json`：只读 Twikoo 数据，用于评论概览。

Nginx 需要把 `/studio/` 反代到 `blog-studio-web:8080/studio/`，示例见 `deploy/nginx/studio-location.conf`。

### 本地开发

```bash
cd web && npm install && npm run build
cd ..
BLOG_STUDIO_ADMIN_PASSWORD_HASH='...' \
BLOG_STUDIO_SESSION_SECRET='replace-with-at-least-32-random-characters' \
BLOG_STUDIO_BLOG_ROOT="$PWD/testdata/hugo" \
BLOG_STUDIO_DATA_ROOT="$PWD/tmp/data" \
BLOG_STUDIO_STATIC_DIR="$PWD/web/dist" \
go run ./cmd/server
```

打开 `http://localhost:8080/studio/`。

## English

## Security Defaults

The server refuses to start unless both variables are set:

- `BLOG_STUDIO_ADMIN_PASSWORD_HASH`
- `BLOG_STUDIO_SESSION_SECRET`

Generate a password hash:

```bash
go run ./cmd/server hash-password 'your-long-password'
```

Use a random 32+ character session secret. Never commit `.env`, real domains, IPs, passwords, WeChat secrets, or private keys.

## Docker Deployment

```bash
cp .env.example .env
docker compose up -d
```

Mounts:

- `/blog`: Hugo site root.
- `/data`: Blog Studio cache, backups, audit logs, previews.
- optional `/twikoo/db.json`: read-only Twikoo data for comment overview.

Nginx should reverse proxy `/studio/` to `blog-studio-web:8080/studio/`; see `deploy/nginx/studio-location.conf`.

## V1 Scope

- Single Hugo site.
- Single admin user.
- File-system state, no database.
- Blog publish writes Page Bundles locally, creates a backup, runs Hugo, and writes audit logs.
- WeChat publishing saves an article to the Official Account draft box only; final publish is confirmed in WeChat.
- Twikoo management is linked or embedded; Blog Studio does not mutate Twikoo data.

## Local Development

```bash
cd web && npm install && npm run build
cd ..
BLOG_STUDIO_ADMIN_PASSWORD_HASH='...' \
BLOG_STUDIO_SESSION_SECRET='replace-with-at-least-32-random-characters' \
BLOG_STUDIO_BLOG_ROOT="$PWD/testdata/hugo" \
BLOG_STUDIO_DATA_ROOT="$PWD/tmp/data" \
BLOG_STUDIO_STATIC_DIR="$PWD/web/dist" \
go run ./cmd/server
```

Open `http://localhost:8080/studio/`.
