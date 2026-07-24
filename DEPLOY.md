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
