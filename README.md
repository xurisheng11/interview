# 面试模拟系统

基于 AI（DeepSeek）的模拟面试练习平台：**Vue2 Web 端 + 原生微信小程序**双前端，共享同一套 Go 后端。根据岗位、难度、经验自动生成题目，AI 实时点评，生成面试报告与成长轨迹。

---

## 功能特性

| 模块 | 说明 |
|------|------|
| 双端入口 | Web（Vue2 + ElementUI）与微信小程序（微信一键登录 / 账号密码注册登录） |
| 面试配置 | 选择岗位、难度、经验、轮次，AI 生成题目 |
| 模拟答题 | 逐题作答，AI 实时点评，支持暂停续题 |
| 视频面试模式 | 摄像头预览 + 语音转文字答题 + 非语言行为分析（Web 端，推荐 Chrome 88+） |
| 面试报告 | 综合评分、雷达图、逐题明细、AI 建议、分享链接 |
| 题库练习 | 多维度筛选，单题 AI 点评，收藏功能 |
| 学习中心 | 定制冲刺计划、模拟套题（Web 端） |
| 知识社区 | AI 生成知识文章，支持点赞 / 收藏 / 评论 |
| 简历中心 | 简历上传与 AI 解析（小程序端） |
| 个人中心 | 资料编辑、面试历史、成长轨迹、我的收藏 |

---

## 技术栈

- **后端**：Go + Gin + JWT；Redis 为主存储，MySQL 为持久层（启动自动建库、自动迁移）
- **AI**：DeepSeek API（出题、点评、报告、知识文章）
- **Web 前端**：Vue2 + ElementUI + ECharts + Axios
- **小程序**：原生微信小程序（tabBar：首页 / 面试 / 我的）
- **部署**：云托管 all-in-one 单容器 / docker-compose（VPS）/ Railway

---

## 仓库结构

```
interview/
├── api/                  # Go 后端
│   ├── main.go
│   ├── handler/          # Gin 控制器
│   ├── service/          # 业务层（含 DeepSeek 调用）
│   ├── repository/       # Redis / MySQL 数据访问
│   ├── model/            # 数据模型
│   ├── db/migrations/    # MySQL 迁移脚本
│   └── Dockerfile        # 后端镜像
├── ui/                   # Web 前端（Vue2）
│   └── src/views/        # 页面
├── miniprogram/          # 微信小程序（原生）
│   └── src/pages/        # 页面
├── deploy/               # 部署资产
│   ├── Dockerfile.cloudrun / cloudrun-start.sh / pack-cloudrun.ps1   # 云托管 all-in-one
│   └── docker-compose.yml / nginx.conf / .env.example                # VPS 一体化
├── DEPLOY.md             # 部署指南（三条路线）
└── README.md
```

---

## 本地开发

### 前置要求

- Go 1.21+（推荐 1.24）
- Node.js 16+（Web 前端）
- Redis 服务（必需，后端启动依赖）
- MySQL 8（可选，连接失败自动降级为纯 Redis 模式）
- 微信开发者工具（小程序）

### 1. 启动后端

```bash
cd api
# 先编辑 .env 配置 Redis / DeepSeek / 微信密钥（见下文"配置说明"）
go run main.go
```

后端监听 `http://localhost:8080`。

### 2. 启动 Web 前端

```bash
cd ui
npm install
npm run serve
```

访问 `http://localhost:3000`。

### 3. 微信小程序

1. 微信开发者工具导入 `miniprogram/` 目录
2. 后端地址在 `miniprogram/src/app.js` 的 `globalData.apiBaseUrl` 配置（本地调试填 `http://localhost:8080/api/v1`）
3. 开发期勾选"不校验合法域名"（设置 → 项目设置 → 本地设置）

---

## 配置说明（api/.env）

```env
# 服务
SERVER_PORT=8080
GIN_MODE=debug

# JWT（生产环境务必换成长随机串；更换后所有旧登录态失效）
JWT_SECRET=please-change-me
JWT_EXPIRE_DAYS=7

# Redis（必需）
REDIS_HOST=127.0.0.1
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0

# MySQL（可选，失败自动降级纯 Redis 模式）
MYSQL_HOST=127.0.0.1
MYSQL_PORT=3306
MYSQL_USER=root
MYSQL_PASSWORD=
MYSQL_DB=interview_sim

# DeepSeek
DEEPSEEK_API_KEY=sk-xxxxxxxx
DEEPSEEK_BASE_URL=https://api.deepseek.com

# 微信小程序（一键登录必需）
WX_APP_ID=wx你的AppID
WX_SECRET=你的AppSecret

# 业务配置
AI_DAILY_LIMIT=10
INTERVIEW_QUESTION_COUNT=10
ADMIN_USERNAME=admin
ADMIN_PASSWORD=请设置强密码
```

---

## 部署上线

| 路线 | 成本 | 适用场景 |
|------|------|----------|
| **微信云托管（免费环境）** | 0 元（小程序未发布阶段免费，发布后 19.9 元/月起） | 零成本跑通「部署 → HTTPS → 真机体验」全流程 |
| **VPS Docker Compose** | 约 38 元/年（秒杀档轻量服务器） | 正式发布：自有备案域名 + Nginx HTTPS |
| **Railway** | 免费额度 $5/月 | Web 端快速上线 |

> 详细步骤、环境变量清单、域名白名单注意事项见 **[DEPLOY.md](DEPLOY.md)**。

部署资产已就绪：

- **云托管 all-in-one**：`deploy/Dockerfile.cloudrun` + `cloudrun-start.sh` + `pack-cloudrun.ps1`（Ubuntu + MySQL 8 + Redis 7 + Go 后端单容器，打包脚本自动排除 `.env` 防密钥泄漏）
- **VPS 一体化**：`deploy/docker-compose.yml` + `nginx.conf` + `.env.example`（Go + Redis + MySQL + Nginx 反代 + Certbot 自动续期）

---

## 浏览器兼容性（视频面试）

| 功能 | Chrome 88+ | Edge 88+ | Firefox | Safari |
|------|-----------|----------|---------|--------|
| 摄像头/麦克风 | ✅ | ✅ | ✅ | ✅ |
| 语音识别（STT） | ✅ | ✅ | ❌ | ⚠️ 部分 |
| 视频录制 | ✅ | ✅ | ✅ | ⚠️ 14.1+ |

推荐使用 Chrome 88+ 获得完整视频面试体验；Firefox 语音识别不可用，需手动输入答案。
