# 面试模拟系统

基于 AI（DeepSeek）的模拟面试练习平台，支持根据岗位、难度、经验自动生成题目，提供实时点评与面试报告。

---

## 技术栈

- **前端**：Vue2 + ElementUI + ECharts + Axios
- **后端**：Go + Gin + Redis
- **AI**：DeepSeek API
- **缓存**：Redis（用户信息永久存储，面试数据7天，题目缓存24h）

---

## 快速启动

### 前置要求

- Go 1.19+
- Node.js 16+
- Redis 服务（默认地址 `172.19.31.128:6379`，DB=1）

---

### 1. 启动后端

```bash
cd api
go run main.go
```

后端默认监听 `http://localhost:8080`

---

### 2. 启动前端

```bash
cd ui
npm install
npm run serve
```

前端默认访问地址：`http://localhost:3000`

---

## 配置说明

### Redis 连接配置

编辑 `api/.env`：

```env
REDIS_HOST=172.19.31.128
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=1
```

### DeepSeek API Key 配置

编辑 `api/.env`：

```env
DEEPSEEK_API_KEY=sk-xxxxxxxxxxxxxxxxxxxxxxxx
DEEPSEEK_BASE_URL=https://api.deepseek.com
```

### 完整环境变量示例（`api/.env`）

```env
# 服务
SERVER_PORT=8080
GIN_MODE=debug

# JWT
JWT_SECRET=interview-sim-secret-2024
JWT_EXPIRE_DAYS=7

# Redis
REDIS_HOST=172.19.31.128
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=1

# DeepSeek
DEEPSEEK_API_KEY=your_deepseek_api_key_here
DEEPSEEK_BASE_URL=https://api.deepseek.com

# 业务配置
AI_DAILY_LIMIT=10
INTERVIEW_QUESTION_COUNT=10
```

---

## 功能模块

| 模块 | 说明 |
|------|------|
| 用户注册/登录 | JWT 鉴权，支持邮箱/手机号登录 |
| 面试配置 | 选择岗位、难度、经验、轮次，AI 生成题目 |
| 模拟答题 | 逐题作答，AI 实时点评，支持暂停续题 |
| 视频面试模式 | 摄像头预览 + 语音转文字答题 + 非语言行为分析，推荐 Chrome 88+ |
| 面试报告 | 综合评分、雷达图、逐题明细、AI建议、分享链接；视频模式含表达能力分析 |
| 题库练习 | 多维度筛选，单题 AI 点评，收藏功能 |
| 知识社区 | AI 生成知识文章，支持点赞/收藏/评论 |
| 个人中心 | 个人信息修改、面试历史、成长轨迹、我的收藏 |

---

## 浏览器兼容性

| 功能 | Chrome 88+ | Edge 88+ | Firefox | Safari |
|------|-----------|----------|---------|--------|
| 摄像头/麦克风 | ✅ | ✅ | ✅ | ✅ |
| 语音识别（STT） | ✅ | ✅ | ❌ | ⚠️ 部分 |
| 视频录制 | ✅ | ✅ | ✅ | ⚠️ 14.1+ |

**推荐使用 Chrome 88+ 以获得完整视频面试体验。**  
Firefox 用户可进入视频模式但语音识别不可用，需手动输入答案。

---

## 目录结构

```
interview/
├── api/          # Go 后端
│   ├── main.go
│   ├── .env
│   ├── handler/
│   ├── service/
│   ├── repository/
│   ├── model/
│   └── ...
└── ui/           # Vue2 前端
    ├── src/
    │   ├── views/
    │   ├── components/
    │   ├── api/
    │   └── ...
    └── package.json
```
