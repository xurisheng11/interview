# Railway 部署指南

## 前置准备

1. 代码推到 GitHub（后端和前端在同一个 repo 里就行）
2. 注册 [Railway](https://railway.app)，用 GitHub 账号登录

---

## 第一步：部署后端

### 1. 新建项目
- 进入 Railway Dashboard → **New Project**
- 选择 **Deploy from GitHub repo**
- 选择你的仓库，设置 **Root Directory** 为 `api`

### 2. 添加 Redis
- 在同一个项目里点 **New** → **Database** → **Add Redis**
- Railway 会自动注入 `REDIS_URL` 环境变量

### 3. 配置后端环境变量
在后端 Service → **Variables** 标签页，添加以下变量：

```
SERVER_PORT=8080
GIN_MODE=release
JWT_SECRET=你自己设置一个长字符串
JWT_EXPIRE_DAYS=7
REDIS_HOST=${{Redis.RAILWAY_PRIVATE_DOMAIN}}
REDIS_PORT=6379
REDIS_PASSWORD=${{Redis.REDISPASSWORD}}
REDIS_DB=0
DEEPSEEK_API_KEY=你的DeepSeek密钥
DEEPSEEK_BASE_URL=https://api.deepseek.com
AI_DAILY_LIMIT=10
INTERVIEW_QUESTION_COUNT=10
```

> **注意**：`${{Redis.RAILWAY_PRIVATE_DOMAIN}}` 是 Railway 内部变量引用语法，直接填入即可自动解析。

### 4. 获取后端域名
- Service → **Settings** → **Networking** → **Generate Domain**
- 记下这个域名，如 `https://your-api.up.railway.app`

---

## 第二步：部署前端

### 1. 新建 Service
- 在同一个 Railway 项目里点 **New** → **GitHub Repo**
- 选同一个仓库，设置 **Root Directory** 为 `ui`

### 2. 配置前端环境变量
在前端 Service → **Variables** 标签页，添加：

```
VUE_APP_API_BASE_URL=https://your-api.up.railway.app/api/v1
```

把 `your-api.up.railway.app` 换成第一步拿到的后端域名。

### 3. 生成前端域名
- Service → **Settings** → **Networking** → **Generate Domain**
- 这就是你分享给别人的访问地址

---

## 第三步：配置 CORS（重要）

后端需要允许前端域名跨域。在后端环境变量里加：

```
FRONTEND_URL=https://your-frontend.up.railway.app
```

然后检查 `api/middleware/cors.go`，确保允许了前端域名（见下方说明）。

---

## 访问流程

```
用户浏览器
    → 前端 (Railway)  https://your-frontend.up.railway.app
    → 后端 API (Railway)  https://your-api.up.railway.app/api/v1
    → Redis (Railway 内网)
    → DeepSeek API (外网)
```

---

## 免费额度说明

Railway 免费套餐每月 $5 额度，轻量使用完全够：
- 后端服务：~$1-2/月
- Redis：~$0.5/月  
- 前端静态服务：~$0.5/月

超出后会暂停服务，不会扣费。

---

## 常见问题

**Q: 部署失败看不到日志怎么办？**  
A: Service → **Deployments** → 点击失败的部署 → 查看 Build/Deploy 日志

**Q: Redis 连不上？**  
A: 确认环境变量里用的是 `RAILWAY_PRIVATE_DOMAIN` 而不是公网地址，内网通信更快也更稳定

**Q: 前端请求 API 跨域报错？**  
A: 检查 `VUE_APP_API_BASE_URL` 是否填了完整的 https 地址，检查后端 CORS 配置

---

# VPS 自建部署（内地 38 元秒杀档适用）

适用场景：小程序正式发布。一台轻量服务器跑全套：Go + Redis + MySQL + Nginx + 自动 HTTPS。

## 0. 购买与秒杀攻略

- **腾讯云**：云产品限时秒杀频道（cloud.tencent.com/act），新用户 38 元/年起，配置 2核2G 起步
- **阿里云**：特惠频道每日 10 点 / 15 点限量抢，2核2G 38 元/年
- 新用户认定：身份证维度未购买过云服务；**秒杀价通常仅首年**，续费约 99~188 元/年，下单前看清续费价
- 镜像选 **Ubuntu 22.04**，地域选内地（上海/广州）延迟最低
- 2核2G 能跑但余量小，建议加 2G swap；预算允许直接上 2核4G

## 1. 域名与备案（内地节点必须）

1. 买域名（.com/.cn，新用户首年几块钱活动常见）
2. 云解析加 A 记录：`api.你的域名.com → 服务器公网 IP`
3. **ICP 备案**：云厂商控制台 → 备案系统提交（免费），约 2~3 周；备案期间可先用 IP+HTTP 内测
4. 备案完成后，微信公众平台 → 开发管理 → 服务器域名 → 添加 `https://api.你的域名.com`

## 2. 服务器初始化（一次性）

```bash
# 安装 Docker（官方一键脚本）
curl -fsSL https://get.docker.com | sh

# 克隆代码
git clone https://github.com/xurisheng11/interview.git
cd interview/deploy

# 配置环境变量
cp .env.example .env
vi .env   # 改掉所有 please-change-me，填入 DEEPSEEK_API_KEY / WX_SECRET
```

## 3. 启动与首次证书申请

```bash
# 启动全套（首次构建镜像约 3-5 分钟）
docker compose up -d --build

# 替换 nginx.conf 中的 api.example.com 为你的域名后，申请 Let's Encrypt 证书：
docker compose run --rm certbot certonly --webroot -w /var/www/certbot \
  -d api.你的域名.com --email 你的邮箱 --agree-tos --no-eff-email

# 按 nginx.conf 内注释启用 443 server 块（80 端口改 301），然后重载：
docker compose exec nginx nginx -s reload
```

certbot 容器已带自动续期循环（每 12 小时检查），无需额外 cron。

## 4. 验证与小程序切换

```bash
curl https://api.你的域名.com/api/v1/questions?page=1&pageSize=1   # 应返回 JSON
```

1. `miniprogram/src/app.js` 的 `apiBaseUrl` 改为 `https://api.你的域名.com/api/v1`
2. 开发者工具**取消勾选**"不校验合法域名"，完整回归测试（这才是线上真实环境）
3. 上传代码 → 公众平台设体验版 → 提交审核 → 发布

## 5. 运维备忘

- **备份**（数据在容器卷里）：
  ```bash
  docker exec interview-mysql mysqldump -uroot -p interview_sim > backup_$(date +%F).sql
  docker run --rm -v interview_redis-data:/data -v $PWD:/backup alpine tar czf /backup/redis_$(date +%F).tgz -C /data .
  ```
- **看日志**：`docker compose logs -f backend`
- **重启**：`docker compose restart`；**升级代码**：`git pull && docker compose up -d --build`
- **端口只放行 80/443/22**（云控制台安全组），Redis/MySQL 不对外暴露

---

# 微信云托管 0 元部署（体验月推荐）

适用场景：不想买服务器、无信用卡，一个月内走通「部署 → HTTPS → 合法域名 → 体验版真机」全流程。
2025 年 2 月起每个小程序账号可创建 **1 个免费云环境**，未发布上线阶段免费；发布后需 19.9 元/月或按量。

## 方案说明

云托管只有容器服务，没有独立 MySQL/Redis，故采用 all-in-one 单容器：
Ubuntu 22.04 + MySQL 8 + Redis 7 + Go 后端（`deploy/Dockerfile.cloudrun` + `deploy/cloudrun-start.sh`）。
后端启动时自动建库、自动执行迁移脚本，无需手动导表。

## 第 1 步：开通免费云环境

1. 打开 https://cloud.weixin.qq.com/ → 微信扫码登录（选择你的小程序账号）
2. 「环境」→ 创建环境 → 选择**免费云环境**（每个账号限 1 个）
3. 进入环境 → 云托管 → 同意开通

## 第 2 步：打包代码

本地项目根目录执行：

```powershell
powershell -ExecutionPolicy Bypass -File deploy\pack-cloudrun.ps1
```

产物 `deploy\cloudrun-pack.zip`（Dockerfile 位于 zip 根目录，.env 已被排除，不会泄露密钥）。

## 第 3 步：创建服务与版本

1. 云托管控制台 → 服务管理 → **新建服务**（名称如 `interview-api`，权限选「公网访问/所有用户可访问」）
2. 进入服务 → **新建版本** → 上传 `cloudrun-pack.zip` → 云端构建
3. 版本设置：
   - 监听端口：**8080**
   - 实例规格：最小档即可（CPU 0.25~0.5 核、内存 ≥1G 为宜）
   - 最小实例数：0（省钱，冷启动约 30 秒）；介意首请求慢就设 1
4. **环境变量**（服务设置 → 环境变量）：

```
WX_APP_ID=你的小程序AppID
WX_SECRET=你的小程序Secret
DEEPSEEK_API_KEY=你的DeepSeek密钥
JWT_SECRET=一串足够长的随机字符
MYSQL_PASSWORD=一串随机密码
ADMIN_USERNAME=admin
ADMIN_PASSWORD=强密码
```

5. 等构建部署完成（首次约 5-10 分钟），服务详情页确认「运行中」

## 第 4 步：开通公网访问，拿默认域名

1. 服务设置 → **公网访问** → 开启，获得默认域名（形如 `https://xxx.ap-shanghai.app.tcloudbase.com`）
2. 浏览器访问 `https://xxx.../api/v1/questions?page=1&pageSize=1` 应返回 JSON
   - 注意 all-in-one 容器首次启动要初始化 MySQL，若 502/超时等 1-2 分钟再试
3. > 官方提示：默认公网域名仅供测试，勿用于正式生产（体验月无所谓，正式发布前切 VPS 或绑备案域名）

## 第 5 步：小程序配合法域名

1. mp.weixin.qq.com → 开发管理 → 开发设置 → **服务器域名**
2. request 合法域名 + **uploadFile 合法域名** 都加上默认域名（简历上传走 wx.uploadFile）
3. `miniprogram/src/app.js` 的 `apiBaseUrl` 改为 `https://xxx.../api/v1`
4. 开发者工具 → **上传** → 公众平台设体验版 → 真机扫码

这次**不需要**开调试模式也能跑：域名在合法名单里、HTTPS 齐全——和正式发布环境一模一样。

## 已知限制

- 容器文件系统不持久：重新部署新版本后，MySQL/Redis 数据会重置（体验期可接受；正式化时迁移 VPS 或挂云存储卷）
- 免费环境在小程序**发布上线后**开始计费，月内体验不受影响
- 默认域名有 QPS/带宽限制，仅测试用途
