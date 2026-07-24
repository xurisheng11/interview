# Design Document

## 面试模拟系统 - 技术设计文档

---

## 1. 系统架构概览

```
┌─────────────────────────────────────────────────────┐
│                    浏览器客户端                        │
│              Vue2 + ElementUI (ui/)                  │
│         Vue Router / Vuex / Axios                    │
└──────────────────────┬──────────────────────────────┘
                       │ HTTP / RESTful API
                       │ Authorization: Bearer <JWT>
┌──────────────────────▼──────────────────────────────┐
│                  Gin 后端服务 (api/)                   │
│                                                      │
│  ┌─────────────┐  ┌──────────────┐  ┌────────────┐  │
│  │  路由层      │  │  中间件层     │  │  Handler层 │  │
│  │  router/    │  │  JWT Auth    │  │  handler/  │  │
│  │             │  │  CORS        │  │            │  │
│  └─────────────┘  └──────────────┘  └────────────┘  │
│                                                      │
│  ┌─────────────┐  ┌──────────────┐  ┌────────────┐  │
│  │  Service层  │  │  Repository层 │  │  Model层   │  │
│  │  service/   │  │  repository/ │  │  model/    │  │
│  └─────────────┘  └──────────────┘  └────────────┘  │
└────────────────┬──────────────────┬─────────────────┘
                 │                  │
    ┌────────────▼───┐    ┌─────────▼──────────┐
    │   Redis        │    │   DeepSeek API      │
    │ 172.19.31.128  │    │ api.deepseek.com    │
    │ DB=1           │    │                     │
    │ 用户信息(永久)  │    │ 题目生成/答案点评/  │
    │ 面试数据(7天)   │    │ 报告生成/知识文章   │
    │ 题目缓存(24h)   │    │                     │
    └────────────────┘    └─────────────────────┘
```

---

## 2. 目录结构

### 2.1 后端目录 `api/`

```
api/
├── main.go                    # 程序入口
├── go.mod
├── go.sum
├── .env                       # 环境变量
├── config/
│   └── config.go              # 配置加载（viper）
├── middleware/
│   ├── auth.go                # JWT 鉴权中间件
│   ├── cors.go                # 跨域中间件
│   └── ratelimit.go           # AI 调用限流中间件
├── router/
│   └── router.go              # 路由注册
├── handler/
│   ├── auth.go                # 注册/登录
│   ├── interview.go           # 面试配置、题目生成、答题
│   ├── report.go              # 面试报告
│   ├── question.go            # 题库管理
│   ├── community.go           # 知识社区
│   └── profile.go             # 个人中心
├── service/
│   ├── auth_service.go
│   ├── interview_service.go
│   ├── report_service.go
│   ├── question_service.go
│   ├── community_service.go
│   ├── profile_service.go
│   └── deepseek_service.go    # DeepSeek API 封装
├── repository/
│   ├── redis.go               # Redis 操作封装
│   └── user_repo.go           # 用户数据操作
├── model/
│   ├── user.go
│   ├── interview.go
│   ├── question.go
│   ├── report.go
│   └── community.go
├── pkg/
│   ├── jwt/
│   │   └── jwt.go             # JWT 工具
│   ├── response/
│   │   └── response.go        # 统一响应格式
│   └── validator/
│       └── validator.go       # 参数校验
└── pkg/utils/
    └── utils.go               # 通用工具函数
```

### 2.2 前端目录 `ui/`

```
ui/
├── package.json
├── vue.config.js              # 代理配置
├── public/
│   └── index.html
└── src/
    ├── main.js                # 入口，注册 Element UI
    ├── App.vue
    ├── router/
    │   └── index.js           # Vue Router 路由表
    ├── store/
    │   ├── index.js           # Vuex store
    │   └── modules/
    │       ├── user.js        # 用户状态
    │       └── interview.js   # 面试状态
    ├── api/
    │   ├── request.js         # Axios 封装（含拦截器）
    │   ├── auth.js
    │   ├── interview.js
    │   ├── report.js
    │   ├── question.js
    │   ├── community.js
    │   └── profile.js
    ├── views/
    │   ├── Login.vue          # 登录/注册
    │   ├── Dashboard.vue      # 首页仪表盘
    │   ├── interview/
    │   │   ├── Config.vue     # 面试配置
    │   │   ├── Loading.vue    # AI 生成中
    │   │   ├── Doing.vue      # 答题页
    │   │   └── History.vue    # 面试历史
    │   ├── report/
    │   │   └── Detail.vue     # 报告详情
    │   ├── question/
    │   │   ├── List.vue       # 题库列表
    │   │   └── Practice.vue   # 单题练习
    │   ├── community/
    │   │   ├── Index.vue      # 社区首页
    │   │   └── Article.vue    # 文章详情
    │   └── profile/
    │       └── Index.vue      # 个人中心
    ├── components/
    │   ├── layout/
    │   │   ├── Navbar.vue     # 顶部导航
    │   │   ├── Subnav.vue     # 二级导航
    │   │   └── Sidebar.vue    # 侧边栏
    │   ├── common/
    │   │   ├── ScoreCircle.vue    # 圆形得分仪表
    │   │   ├── ScoreBar.vue       # 评分进度条
    │   │   ├── RadarChart.vue     # 雷达图（ECharts）
    │   │   └── MarkdownRender.vue # Markdown 渲染
    │   └── interview/
    │       ├── QuestionCard.vue   # 题目卡片
    │       └── AiComment.vue      # AI 点评组件
    └── utils/
        ├── auth.js            # token 存取工具
        └── format.js          # 格式化工具
```

---

## 3. 数据模型设计

### 3.1 Redis Key 设计

| Key 格式 | 类型 | TTL | 说明 |
|---------|------|-----|------|
| `user:{userId}` | Hash | 永久 | 用户基本信息 |
| `user:account:{account}` | String | 永久 | account→userId 映射（account=手机号/邮箱/用户名） |
| `interview:session:{interviewId}` | Hash | 7天 | 进行中的面试会话（题目列表+作答进度） |
| `interview:questions:{cacheKey}` | String(JSON) | 24h | 题目缓存，cacheKey=md5(岗位+难度+经验+轮次) |
| `report:{userId}:{interviewId}` | String(JSON) | 永久 | 面试报告 |
| `report:share:{shareToken}` | String | 7天 | 分享链接 token→reportId |
| `ai:limit:{userId}:{date}` | Integer | 当天到期 | 用户当天 AI 文章生成次数 |
| `question:cache:{questionId}` | String(JSON) | 24h | 单题缓存 |
| `community:article:{articleId}` | String(JSON) | 1h | 文章缓存 |

### 3.2 用户 Redis Hash 字段（`user:{userId}`）

```
userId        string   用户唯一ID（UUID）
username      string   用户名
phone         string   手机号（可选）
email         string   邮箱（可选）
passwordHash  string   bcrypt 加密密码
avatar        string   头像URL
nickname      string   昵称
bio           string   个人简介
createdAt     string   注册时间（RFC3339）
role          string   角色：user / admin
```

### 3.3 面试会话 Redis Hash 字段（`interview:session:{interviewId}`）

```
interviewId   string   面试ID
userId        string   用户ID
jobTitle      string   目标岗位
difficulty    string   难度：junior/middle/senior
experience    string   经验：fresh/1-3/3-5/5+
round         string   轮次：round1/round2/round3
focusAreas    string   重点方向（JSON数组）
questions     string   题目列表（JSON数组）
currentIndex  int      当前题目索引
answers       string   作答记录（JSON对象，key=题目索引）
status        string   状态：ongoing/paused/completed
startTime     string   开始时间
pauseTime     string   暂停时间
```

### 3.4 面试报告结构（JSON，存 Redis）

```json
{
  "interviewId": "uuid",
  "userId": "uuid",
  "jobTitle": "后端开发",
  "difficulty": "middle",
  "round": "round1",
  "experience": "1-3",
  "totalScore": 85,
  "grade": "良好",
  "passed": true,
  "passReason": "",
  "startTime": "2024-01-01T10:00:00Z",
  "endTime": "2024-01-01T10:42:00Z",
  "totalSeconds": 2520,
  "answeredCount": 9,
  "skippedCount": 1,
  "totalCount": 10,
  "questions": [
    {
      "index": 0,
      "content": "请做一下自我介绍",
      "tags": ["综合"],
      "difficulty": "easy",
      "estimatedMinutes": 2,
      "userAnswer": "...",
      "score": 90,
      "pros": ["表达清晰"],
      "cons": ["未突出技术亮点"],
      "referenceAnswer": "...",
      "skipped": false
    }
  ],
  "moduleScores": [
    {"module": "算法", "count": 2, "avgScore": 82, "level": "良好"}
  ],
  "aiSummary": {
    "strengths": ["...", "...", "..."],
    "weaknesses": ["...", "...", "..."],
    "roadmap": "..."
  },
  "createdAt": "2024-01-01T10:42:00Z"
}
```

### 3.5 知识社区文章结构（存 Redis + 内存列表）

```json
{
  "articleId": "uuid",
  "userId": "uuid",
  "title": "Redis 持久化详解",
  "content": "## Markdown正文...",
  "tags": ["Redis", "数据库"],
  "jobCategory": "backend",
  "isAiGenerated": true,
  "likeCount": 12,
  "collectCount": 5,
  "commentCount": 3,
  "createdAt": "2024-01-01T10:00:00Z"
}
```

---

## 4. API 接口设计

### 统一响应格式

```json
{
  "code": 200,
  "message": "success",
  "data": {}
}
```

错误码约定：`200` 成功，`400` 参数错误，`401` 未授权，`403` 禁止，`404` 不存在，`429` 限流，`500` 服务器错误。

---

### 4.1 认证模块 `/api/v1/auth`

| 方法 | 路径 | 说明 | 鉴权 |
|------|------|------|------|
| POST | `/api/v1/auth/register` | 注册 | 否 |
| POST | `/api/v1/auth/login` | 登录 | 否 |
| GET  | `/api/v1/auth/me` | 获取当前用户信息 | 是 |

**POST /api/v1/auth/register**
```json
// Request
{ "username": "zhang", "account": "zhang@163.com", "password": "Abc12345", "confirmPassword": "Abc12345" }
// Response
{ "code": 200, "message": "注册成功", "data": { "userId": "uuid", "token": "jwt..." } }
```

**POST /api/v1/auth/login**
```json
// Request
{ "account": "zhang@163.com", "password": "Abc12345" }
// Response
{ "code": 200, "data": { "token": "jwt...", "expiresAt": "2024-01-08T...", "user": { "userId":"...", "username":"...", "nickname":"...", "avatar":"..." } } }
```

---

### 4.2 面试模块 `/api/v1/interviews`

| 方法 | 路径 | 说明 | 鉴权 |
|------|------|------|------|
| POST | `/api/v1/interviews` | 创建面试（配置+生成题目） | 是 |
| GET  | `/api/v1/interviews` | 获取面试历史列表 | 是 |
| GET  | `/api/v1/interviews/:id` | 获取面试会话详情 | 是 |
| POST | `/api/v1/interviews/:id/answers` | 提交单题答案（AI点评） | 是 |
| PUT  | `/api/v1/interviews/:id/pause` | 暂停面试 | 是 |
| PUT  | `/api/v1/interviews/:id/complete` | 完成面试（触发报告生成） | 是 |

**POST /api/v1/interviews**
```json
// Request
{
  "jobTitle": "后端开发",
  "difficulty": "middle",
  "experience": "1-3",
  "round": "round1",
  "focusAreas": ["算法", "数据库"],
  "remark": "熟悉Go和Redis"
}
// Response
{
  "code": 200,
  "data": {
    "interviewId": "uuid",
    "questions": [
      {
        "index": 0,
        "content": "请做一下自我介绍",
        "tags": ["综合"],
        "difficulty": "easy",
        "estimatedMinutes": 2
      }
    ]
  }
}
```

**POST /api/v1/interviews/:id/answers**
```json
// Request
{ "questionIndex": 2, "answer": "RDB是快照持久化..." }
// Response
{
  "code": 200,
  "data": {
    "score": 82,
    "pros": ["基本概念描述准确"],
    "cons": ["未提及AOF重写机制"],
    "referenceAnswer": "..."
  }
}
```

---

### 4.3 报告模块 `/api/v1/reports`

| 方法 | 路径 | 说明 | 鉴权 |
|------|------|------|------|
| GET  | `/api/v1/reports/:interviewId` | 获取面试报告 | 是 |
| POST | `/api/v1/reports/:interviewId/share` | 生成分享链接 | 是 |
| GET  | `/api/v1/reports/share/:token` | 访问分享报告 | 否 |

---

### 4.4 题库模块 `/api/v1/questions`

| 方法 | 路径 | 说明 | 鉴权 |
|------|------|------|------|
| GET  | `/api/v1/questions` | 题库列表（支持筛选分页） | 是 |
| GET  | `/api/v1/questions/:id` | 获取题目详情 | 是 |
| POST | `/api/v1/questions/:id/practice` | 单题练习（AI点评） | 是 |
| POST | `/api/v1/questions/:id/collect` | 收藏/取消收藏 | 是 |
| POST | `/api/v1/questions` | 创建题目（管理员） | 是(admin) |
| PUT  | `/api/v1/questions/:id` | 编辑题目（管理员） | 是(admin) |
| DELETE | `/api/v1/questions/:id` | 删除题目（管理员） | 是(admin) |

**GET /api/v1/questions 查询参数**
```
page=1&pageSize=20&jobTitle=后端开发&difficulty=middle&tag=算法&type=basic&keyword=Redis
```

---

### 4.5 社区模块 `/api/v1/community`

| 方法 | 路径 | 说明 | 鉴权 |
|------|------|------|------|
| GET  | `/api/v1/community/articles` | 文章列表 | 是 |
| POST | `/api/v1/community/articles/ai` | AI 生成文章 | 是 |
| GET  | `/api/v1/community/articles/:id` | 文章详情 | 是 |
| POST | `/api/v1/community/articles/:id/like` | 点赞 | 是 |
| POST | `/api/v1/community/articles/:id/collect` | 收藏 | 是 |
| GET  | `/api/v1/community/articles/:id/comments` | 评论列表 | 是 |
| POST | `/api/v1/community/articles/:id/comments` | 发表评论 | 是 |

---

### 4.6 个人中心模块 `/api/v1/profile`

| 方法 | 路径 | 说明 | 鉴权 |
|------|------|------|------|
| GET  | `/api/v1/profile` | 获取个人信息 | 是 |
| PUT  | `/api/v1/profile` | 更新个人信息 | 是 |
| PUT  | `/api/v1/profile/password` | 修改密码 | 是 |
| GET  | `/api/v1/profile/stats` | 个人统计数据 | 是 |
| GET  | `/api/v1/profile/trend` | 得分趋势（近30次） | 是 |
| GET  | `/api/v1/profile/collections` | 我的收藏 | 是 |

---

## 5. 前端路由设计

```javascript
// router/index.js
const routes = [
  { path: '/login',    name: 'Login',     component: Login,    meta: { requiresAuth: false } },
  { path: '/',         redirect: '/dashboard' },
  {
    path: '/dashboard', name: 'Dashboard', component: Dashboard, meta: { requiresAuth: true }
  },
  {
    path: '/interview',
    meta: { requiresAuth: true },
    children: [
      { path: 'config',    name: 'InterviewConfig',  component: InterviewConfig },
      { path: 'loading',   name: 'InterviewLoading', component: InterviewLoading },
      { path: ':id/doing', name: 'InterviewDoing',   component: InterviewDoing },
      { path: 'history',   name: 'InterviewHistory', component: InterviewHistory },
    ]
  },
  {
    path: '/report/:interviewId', name: 'ReportDetail',
    component: ReportDetail, meta: { requiresAuth: true }
  },
  {
    path: '/report/share/:token', name: 'ReportShare',
    component: ReportShare, meta: { requiresAuth: false }
  },
  {
    path: '/questions',
    meta: { requiresAuth: true },
    children: [
      { path: '',       name: 'QuestionList',    component: QuestionList },
      { path: ':id/practice', name: 'QuestionPractice', component: QuestionPractice },
    ]
  },
  {
    path: '/community',
    meta: { requiresAuth: true },
    children: [
      { path: '',    name: 'Community',      component: CommunityIndex },
      { path: ':id', name: 'ArticleDetail',  component: ArticleDetail },
    ]
  },
  { path: '/profile', name: 'Profile', component: ProfileIndex, meta: { requiresAuth: true } },
]

// 路由守卫：未登录跳转 /login，已登录访问 /login 跳转 /dashboard
router.beforeEach((to, from, next) => {
  const token = localStorage.getItem('token')
  if (to.meta.requiresAuth && !token) {
    next({ path: '/login', query: { redirect: to.fullPath } })
  } else if (!to.meta.requiresAuth && token && to.path === '/login') {
    next('/dashboard')
  } else {
    next()
  }
})
```

---

## 6. JWT 鉴权流程

```
注册/登录
  │
  ▼
后端生成 JWT Token
  payload: { userId, role, exp: now+7天 }
  secret: 环境变量 JWT_SECRET
  │
  ▼
前端存入 localStorage('token')
  │
  ▼
每次请求 Axios 拦截器自动携带
  Header: Authorization: Bearer <token>
  │
  ▼
Gin JWT 中间件验证
  ├── 无 token → 401
  ├── token 过期 → 401 + message: "token expired"
  ├── token 非法 → 401 + message: "invalid token"
  └── 验证通过 → ctx.Set("userId", claims.UserId)
  │
  ▼
Handler 从 ctx 取 userId 使用

前端响应拦截器：
  401 → 清除 token → 跳转 /login?redirect=当前路径
```

---

## 7. DeepSeek API 调用设计

### 7.1 接口封装

```go
// service/deepseek_service.go
type DeepSeekService struct {
    apiKey  string
    baseURL string
    client  *http.Client
}

// 请求结构（OpenAI 兼容格式）
POST https://api.deepseek.com/v1/chat/completions
Header: Authorization: Bearer {DEEPSEEK_API_KEY}
Body: {
  "model": "deepseek-chat",
  "messages": [{"role": "user", "content": "<prompt>"}],
  "temperature": 0.7,
  "max_tokens": 4096
}
```

### 7.2 题目生成 Prompt 模板

```
你是一位资深技术面试官，请为以下面试场景生成 {count} 道面试题目。

面试信息：
- 目标岗位：{jobTitle}
- 面试难度：{difficulty}（初级/中级/高级）
- 工作经验：{experience}
- 面试轮次：{round}
- 重点方向：{focusAreas}
- 补充说明：{remark}

题目构成要求（{round}）：
{roundRequirement}

请严格按照以下 JSON 格式返回，不要包含任何其他文字：
[
  {
    "index": 0,
    "content": "题目内容",
    "tags": ["知识点标签"],
    "difficulty": "easy|medium|hard",
    "estimatedMinutes": 3,
    "type": "basic|algorithm|design|hr"
  }
]
```

### 7.3 答案点评 Prompt 模板

```
你是一位专业技术面试官，请对以下面试答案进行评分和点评。

题目：{questionContent}
知识点：{tags}
岗位：{jobTitle}，难度：{difficulty}
候选人答案：{userAnswer}

请严格按照以下 JSON 格式返回：
{
  "score": 85,
  "pros": ["优点1", "优点2"],
  "cons": ["不足1", "不足2"],
  "referenceAnswer": "参考答案正文（支持Markdown）"
}

评分标准：完整准确100分，基本正确60-80分，有误20-50分，完全错误0-10分。
```

### 7.4 报告综合评价 Prompt 模板

```
你是一位资深面试顾问，请根据以下面试数据生成综合评价。

岗位：{jobTitle}，轮次：{round}，综合得分：{totalScore}
各题得分摘要：{scoresSummary}

请严格按照以下 JSON 格式返回：
{
  "strengths": ["优势1", "优势2", "优势3"],
  "weaknesses": [
    {"point": "不足1", "suggestion": "改进建议", "resource": "推荐资源链接"},
    {"point": "不足2", "suggestion": "改进建议", "resource": "推荐资源链接"},
    {"point": "不足3", "suggestion": "改进建议", "resource": "推荐资源链接"}
  ],
  "roadmap": "备考路线图文字描述"
}
```

### 7.5 缓存策略

```
题目生成缓存：
  cacheKey = md5(jobTitle + difficulty + experience + round + focusAreas)
  Redis Key: interview:questions:{cacheKey}
  TTL: 24小时
  流程: 先查 Redis → 命中直接返回 → 未命中调 DeepSeek → 写入 Redis

AI 文章生成限流：
  Redis Key: ai:limit:{userId}:{date}  (date=YYYYMMDD)
  TTL: 当天 23:59:59 到期
  限制: 每天最多 10 次（可配置 AI_DAILY_LIMIT 环境变量）
  超限返回 429
```

---

## 8. 关键业务流程

### 8.1 完整面试流程

```
用户填写面试配置
        │
        ▼
POST /api/v1/interviews
  1. 生成 interviewId（UUID）
  2. 构造 cacheKey，查 Redis 题目缓存
     ├── 命中 → 直接使用缓存题目
     └── 未命中 → 调 DeepSeek 生成 → 写缓存
  3. 将面试会话写入 Redis（status=ongoing）
  4. 返回 interviewId + questions
        │
        ▼
前端跳转 /interview/{id}/doing
每题作答：
  POST /api/v1/interviews/:id/answers
    1. 调 DeepSeek 点评（传题目+答案）
    2. 将答案+点评写入 Redis session
    3. 返回点评结果
        │
  用户完成所有题目 or 主动结束
        │
        ▼
PUT /api/v1/interviews/:id/complete
  1. 计算综合得分（加权平均，跳过=0）
  2. 统计各知识点模块分
  3. 调 DeepSeek 生成综合评价
  4. 构建完整报告 JSON
  5. 写入 Redis：report:{userId}:{interviewId}（永久）
  6. 返回 interviewId
        │
        ▼
前端跳转 /report/{interviewId}
```

### 8.2 面试通过判定逻辑

```go
func DeterminePassStatus(report *Report) string {
    // 跳过超过50%直接未通过
    if float64(report.SkippedCount)/float64(report.TotalCount) > 0.5 {
        return "fail" // 跳过题目过多
    }
    score := report.TotalScore
    if score >= 75 { return "pass" }
    if score >= 60 { return "pending" }
    return "fail"
}
```

### 8.3 暂停/续题流程

```
用户点击「暂停面试」
  PUT /api/v1/interviews/:id/pause
    → Redis session status = "paused"，记录 pauseTime
    → TTL 重置为 7 天

下次登录，仪表盘显示「继续面试」入口
  GET /api/v1/interviews（status=ongoing/paused）
    → 前端检测到有未完成面试
    → 跳转 /interview/{id}/doing，恢复到 currentIndex
```

---

## 9. 配置与环境变量

### 9.1 后端 `.env`

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

### 9.2 前端 `vue.config.js` 代理

```javascript
module.exports = {
  devServer: {
    port: 3000,
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true
      }
    }
  }
}
```

---

## 10. 非功能性设计

### 10.1 错误处理

- 所有 DeepSeek 调用包裹 retry 逻辑（最多重试 2 次，间隔 1s）
- DeepSeek 调用超时设置 30s
- Redis 操作失败时记录日志，不影响主流程（降级处理）

### 10.2 安全

- 密码使用 bcrypt 加密存储（cost=10）
- JWT Secret 通过环境变量注入，不硬编码
- 所有需鉴权接口统一走 JWT 中间件
- 社区评论做基础敏感词过滤（内置词表）

### 10.3 CORS 配置

```go
// 开发环境允许 localhost:3000
// 生产环境配置具体域名
AllowOrigins: []string{"http://localhost:3000"}
AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
AllowHeaders: []string{"Authorization", "Content-Type"}
```
