# Tasks

## Task Dependency Graph

```
T1(项目初始化)
  ├── T2(后端基础框架)
  │     ├── T3(用户注册登录)
  │     │     ├── T5(面试配置与题目生成)
  │     │     │     ├── T6(答题与AI点评)
  │     │     │     │     └── T7(面试报告)
  │     │     ├── T8(题库管理)
  │     │     ├── T9(知识社区)
  │     │     └── T10(个人中心)
  └── T4(前端基础框架)
        ├── T11(登录注册页)
        │     ├── T12(仪表盘)
        │     ├── T13(面试配置与答题页)
        │     │     └── T14(面试报告页)
        │     ├── T15(题库页)
        │     ├── T16(社区页)
        │     └── T17(个人中心页)
```

---

## Task 1: 项目脚手架初始化

- [x] 1.1 创建后端项目结构
  - 在 `api/` 目录下执行 `go mod init interview-sim`
  - 创建完整目录树：`config/ middleware/ router/ handler/ service/ repository/ model/ pkg/`
  - 安装依赖：`github.com/gin-gonic/gin`, `github.com/go-redis/redis/v8`, `github.com/golang-jwt/jwt/v5`, `golang.org/x/crypto`, `github.com/joho/godotenv`, `github.com/google/uuid`

- [x] 1.2 创建前端项目结构
  - 在 `ui/` 目录下创建 Vue2 项目结构
  - 安装依赖：`vue@2`, `vue-router@3`, `vuex@3`, `element-ui@2`, `axios`, `echarts`, `marked`
  - 配置 `vue.config.js` 代理到 `http://localhost:8080`

- [x] 1.3 创建环境配置文件
  - 创建 `api/.env`，写入所有环境变量（参考 design.md 第9节）
  - 创建 `api/config/config.go`，使用 `godotenv` 加载 `.env`

---

## Task 2: 后端基础框架

- [x] 2.1 统一响应格式
  - 创建 `api/pkg/response/response.go`
  - 实现 `Success(c, data)` / `Fail(c, code, msg)` 辅助函数

- [x] 2.2 Redis 连接与封装
  - 创建 `api/repository/redis.go`
  - 初始化 Redis 客户端（读取 config），导出全局 `RDB`
  - 封装常用操作：`Set`, `Get`, `HSet`, `HGetAll`, `Del`, `Expire`, `IncrBy`

- [x] 2.3 JWT 工具
  - 创建 `api/pkg/jwt/jwt.go`
  - 实现 `GenerateToken(userId, role string) (string, error)`
  - 实现 `ParseToken(tokenStr string) (*Claims, error)`
  - Token 有效期从 config 读取（默认7天）

- [x] 2.4 JWT 鉴权中间件
  - 创建 `api/middleware/auth.go`
  - 从 `Authorization: Bearer <token>` 提取并验证 token
  - 验证通过将 `userId`、`role` 写入 gin.Context
  - 验证失败返回 401

- [x] 2.5 CORS 中间件
  - 创建 `api/middleware/cors.go`
  - 允许来自 `localhost:3000` 的请求，支持 OPTIONS 预检

- [x] 2.6 路由注册与 main.go
  - 创建 `api/router/router.go`，注册所有路由分组
  - 创建 `api/main.go`，初始化 config/redis/router，启动 Gin 服务

---

## Task 3: 用户注册与登录（后端）

- [x] 3.1 用户数据模型
  - 创建 `api/model/user.go`
  - 定义 `User` struct（userId, username, phone, email, passwordHash, avatar, nickname, bio, createdAt, role）
  - 实现 `ToRedisHash()` 和 `FromRedisHash()` 序列化方法

- [x] 3.2 用户 Repository
  - 创建 `api/repository/user_repo.go`
  - 实现：`SaveUser(user)` — 写 `user:{userId}` Hash（永久，无TTL）
  - 实现：`SaveAccountIndex(account, userId)` — 写 `user:account:{account}`（永久）
  - 实现：`GetUserById(userId)` — 读 `user:{userId}`
  - 实现：`GetUserByAccount(account)` — 读 `user:account:{account}` → 再读 user Hash
  - 实现：`AccountExists(account) bool`

- [x] 3.3 认证 Service
  - 创建 `api/service/auth_service.go`
  - 实现 `Register(username, account, password, confirmPwd)` — 校验格式、检查唯一性、bcrypt加密、存Redis
  - 实现 `Login(account, password)` — 查用户、验密码、生成JWT

- [x] 3.4 认证 Handler
  - 创建 `api/handler/auth.go`
  - 实现 `Register` Handler：绑定参数 → 调 service → 返回 token
  - 实现 `Login` Handler：绑定参数 → 调 service → 返回 token + 用户信息
  - 实现 `GetMe` Handler：从 ctx 取 userId → 查用户 → 返回用户信息

- [x] 3.5 注册路由
  - 在 `router.go` 中注册：`POST /api/v1/auth/register`、`POST /api/v1/auth/login`、`GET /api/v1/auth/me`（需鉴权）

---

## Task 4: 前端基础框架

- [x] 4.1 入口与插件注册
  - 创建 `ui/src/main.js`：引入 Vue2、ElementUI、Router、Vuex，全局注册
  - 创建 `ui/src/App.vue`：包含 `<router-view>`，根据登录状态显示/隐藏 Navbar 和 Subnav

- [x] 4.2 Axios 请求封装
  - 创建 `ui/src/api/request.js`
  - 请求拦截器：从 localStorage 取 token，自动加 `Authorization: Bearer <token>`
  - 响应拦截器：401 自动清除 token 并跳转 `/login?redirect=当前路径`；统一处理错误提示（Element UI Message）

- [x] 4.3 Vuex Store
  - 创建 `ui/src/store/modules/user.js`
  - state: `{ token, userInfo }`
  - mutations: `SET_TOKEN`, `SET_USER_INFO`, `CLEAR_AUTH`
  - actions: `login(account, password)`, `logout()`, `fetchMe()`
  - 创建 `ui/src/store/index.js` 组合模块

- [x] 4.4 Vue Router 配置
  - 创建 `ui/src/router/index.js`，按 design.md 第5节注册所有路由
  - 实现路由守卫：未登录访问需鉴权页 → 跳 `/login`；已登录访问 `/login` → 跳 `/dashboard`

- [x] 4.5 公共布局组件
  - 创建 `ui/src/components/layout/Navbar.vue`：顶部导航，根据 store.user.token 决定显示「登录/注册」还是用户名+退出
  - 创建 `ui/src/components/layout/Subnav.vue`：二级导航，仅登录后显示
  - 创建 `ui/src/components/layout/Sidebar.vue`：侧边栏，接收 `items` prop

---

## Task 5: DeepSeek 服务与面试题目生成（后端）

- [x] 5.1 DeepSeek Service 封装
  - 创建 `api/service/deepseek_service.go`
  - 实现 `Chat(prompt string) (string, error)`：POST 到 DeepSeek API，超时30s，失败重试2次
  - 实现 `GenerateQuestions(config InterviewConfig) ([]Question, error)`：调 Chat，解析 JSON 返回题目列表
  - 实现 `ReviewAnswer(question, answer, config) (*ReviewResult, error)`：调 Chat，解析JSON返回点评
  - 实现 `GenerateReport(summary) (*AISummary, error)`：调 Chat，解析JSON返回综合评价

- [x] 5.2 面试数据模型
  - 创建 `api/model/interview.go`
  - 定义 `Question` struct（index, content, tags, difficulty, estimatedMinutes, type）
  - 定义 `InterviewSession` struct（interviewId, userId, jobTitle, difficulty, experience, round, focusAreas, questions, currentIndex, answers, status, startTime）
  - 定义 `AnswerRecord` struct（userAnswer, score, pros, cons, referenceAnswer, skipped, submittedAt）
  - 实现序列化/反序列化 JSON 方法

- [x] 5.3 面试 Repository
  - 创建面试相关 Redis 操作：`SaveSession`, `GetSession`, `UpdateSessionAnswer`, `UpdateSessionStatus`
  - 题目缓存操作：`GetQuestionsCache(cacheKey)`, `SetQuestionsCache(cacheKey, questions)`

- [x] 5.4 面试 Service
  - 创建 `api/service/interview_service.go`
  - 实现 `CreateInterview(userId, config)` — 生成ID、查题目缓存或调DeepSeek生成、存Redis会话、返回题目
  - 实现 `SubmitAnswer(userId, interviewId, questionIndex, answer)` — 调DeepSeek点评、更新Redis会话
  - 实现 `PauseInterview(userId, interviewId)` — 更新状态为paused，重置TTL为7天
  - 实现 `GetInterviewList(userId)` — 从Redis获取用户所有面试会话摘要
  - 实现 `GetInterviewSession(userId, interviewId)` — 获取完整会话

- [x] 5.5 面试 Handler
  - 创建 `api/handler/interview.go`
  - 实现 `CreateInterview`、`GetInterviewList`、`GetInterview`、`SubmitAnswer`、`PauseInterview`、`CompleteInterview` Handler

- [x] 5.6 注册路由
  - `POST /api/v1/interviews`、`GET /api/v1/interviews`、`GET /api/v1/interviews/:id`
  - `POST /api/v1/interviews/:id/answers`、`PUT /api/v1/interviews/:id/pause`、`PUT /api/v1/interviews/:id/complete`
  - 全部需要 JWT 鉴权

---

## Task 6: 面试报告（后端）

- [x] 6.1 报告数据模型
  - 创建 `api/model/report.go`
  - 定义完整报告结构（参考 design.md 3.4节）
  - 实现得分计算逻辑：加权平均、模块分统计、通过判定

- [x] 6.2 报告 Service
  - 创建 `api/service/report_service.go`
  - 实现 `GenerateReport(userId, interviewId)` — 从Redis取会话、计算分数、调DeepSeek生成综合评价、存报告到Redis（永久）
  - 实现 `GetReport(userId, interviewId)` — 从Redis读取报告
  - 实现 `CreateShareLink(userId, interviewId)` — 生成token，写 `report:share:{token}`（7天TTL）
  - 实现 `GetSharedReport(token)` — 从Redis取token对应报告（脱敏，隐藏userAnswer）

- [x] 6.3 报告 Handler
  - 创建 `api/handler/report.go`
  - 实现 `GetReport`、`CreateShare`、`GetSharedReport` Handler

- [x] 6.4 注册路由
  - `GET /api/v1/reports/:interviewId`（需鉴权）
  - `POST /api/v1/reports/:interviewId/share`（需鉴权）
  - `GET /api/v1/reports/share/:token`（不需鉴权）

---

## Task 7: 题库管理（后端）

- [x] 7.1 题目数据模型
  - 创建 `api/model/question.go`
  - 定义 `QuestionItem` struct（questionId, content, jobTitle, difficulty, tags, type, answerCount, avgScore, createdBy, createdAt）

- [x] 7.2 题库 Service
  - 创建 `api/service/question_service.go`
  - 实现 `ListQuestions(filters, page, pageSize)` — Redis缓存题目列表，未命中通过DeepSeek动态生成预置题目
  - 实现 `GetQuestion(questionId)` — 优先读Redis缓存
  - 实现 `PracticeQuestion(userId, questionId, answer)` — 调DeepSeek点评单题
  - 实现 `CollectQuestion(userId, questionId)` — 写收藏记录到 `user:collect:question:{userId}`（Set类型）
  - 实现 `CreateQuestion`, `UpdateQuestion`, `DeleteQuestion`（管理员操作）

- [x] 7.3 题库 Handler 与路由
  - 创建 `api/handler/question.go`，实现所有 Handler
  - 注册路由（`GET /api/v1/questions`、`POST /api/v1/questions/:id/practice` 等）

---

## Task 8: 知识社区（后端）

- [x] 8.1 社区数据模型
  - 创建 `api/model/community.go`
  - 定义 `Article` struct、`Comment` struct

- [x] 8.2 AI 生成文章 Prompt
  - 在 `deepseek_service.go` 中添加 `GenerateArticle(topic, jobCategory string) (*Article, error)`
  - Prompt 要求返回 JSON：`{ "title": "...", "content": "Markdown正文", "tags": [] }`

- [x] 8.3 社区 Service
  - 创建 `api/service/community_service.go`
  - 实现 `GenerateAIArticle(userId, topic, jobCategory)` — 检查每日限额（Redis incr）、调DeepSeek、存文章
  - 实现 `ListArticles(jobCategory, sortBy, page)` — 文章列表（存Redis Hash + ZSet排序）
  - 实现 `GetArticle(articleId)` — 读文章详情（Redis缓存1h）
  - 实现 `LikeArticle`, `CollectArticle` — 写入对应 Set/计数
  - 实现 `AddComment(userId, articleId, content)` — 敏感词过滤 → 存评论
  - 实现 `ListComments(articleId)`

- [x] 8.4 社区 Handler 与路由
  - 创建 `api/handler/community.go`，实现所有 Handler
  - 注册路由（全部需鉴权）

---

## Task 9: 个人中心（后端）

- [x] 9.1 个人中心 Service
  - 创建 `api/service/profile_service.go`
  - 实现 `GetProfile(userId)` — 从Redis读取用户信息
  - 实现 `UpdateProfile(userId, nickname, avatar, bio)` — HSet 更新对应字段
  - 实现 `ChangePassword(userId, oldPwd, newPwd)` — 验旧密码 → bcrypt新密码 → 更新Redis
  - 实现 `GetStats(userId)` — 统计累计次数/平均分/最高分/常练岗位（遍历用户面试历史）
  - 实现 `GetScoreTrend(userId)` — 返回近30次面试得分数组
  - 实现 `GetCollections(userId)` — 返回收藏题目+收藏文章列表

- [x] 9.2 个人中心 Handler 与路由
  - 创建 `api/handler/profile.go`
  - 注册路由：`GET/PUT /api/v1/profile`、`PUT /api/v1/profile/password`、`GET /api/v1/profile/stats`、`GET /api/v1/profile/trend`、`GET /api/v1/profile/collections`

---

## Task 10: 前端登录注册页

- [x] 10.1 创建 API 模块
  - 创建 `ui/src/api/auth.js`：封装 `register(data)`, `login(data)`, `getMe()` 请求

- [x] 10.2 登录注册页面
  - 创建 `ui/src/views/Login.vue`
  - Tab 切换登录/注册，参考 prototype.html 样式（亚马逊风格深色导航+橙色主题）
  - 登录表单：账号+密码，点击登录 → dispatch store/login → 跳转 `/dashboard`（或 redirect 参数目标）
  - 注册表单：用户名+账号+密码+确认密码，前端格式校验（Element Form rules），提交 → 注册成功 → 自动登录
  - 错误提示使用 `this.$message.error()`

---

## Task 11: 前端仪表盘

- [x] 11.1 仪表盘页面
  - 创建 `ui/src/views/Dashboard.vue`
  - 布局：侧边栏 + 主内容区（参考 prototype.html）
  - 顶部 4 格统计卡片：累计面试次数/平均得分/最高得分/连续练习天数（调 `GET /api/v1/profile/stats`）
  - 近期面试记录表格（调 `GET /api/v1/interviews?page=1&pageSize=5`）
  - 今日推荐练习（随机3题，调 `GET /api/v1/questions?page=1&pageSize=3`）
  - 得分趋势折线图（ECharts，调 `GET /api/v1/profile/trend`）

- [x] 11.2 ScoreCircle 公共组件
  - 创建 `ui/src/components/common/ScoreCircle.vue`
  - 接收 `score` prop，渲染圆形得分仪表（CSS 渐变圆形）

---

## Task 12: 前端面试配置与答题页

- [x] 12.1 API 模块
  - 创建 `ui/src/api/interview.js`：封装所有面试相关请求

- [x] 12.2 面试配置页
  - 创建 `ui/src/views/interview/Config.vue`
  - Tag 单选/多选组件（岗位、难度、经验、轮次、重点方向）
  - 必填项未选时「发起面试」按钮禁用（:disabled）
  - 点击发起 → 调 `POST /api/v1/interviews` → 存 interviewId 到 store → 跳转 `/interview/loading`

- [x] 12.3 AI 加载页
  - 创建 `ui/src/views/interview/Loading.vue`
  - 显示 loading 动画 + 配置摘要文字
  - 轮询 `GET /api/v1/interviews/:id` 直到 questions 返回 → 自动跳转 `/interview/:id/doing`

- [x] 12.4 答题页
  - 创建 `ui/src/views/interview/Doing.vue`
  - 进度条（已完成/总题数）+ 计时器（倒计时，每题 estimatedMinutes * 2 分钟）
  - 左侧题目导航（题号，已答绿色/跳过灰色/当前橙色）
  - 题目卡片：题号/标签/难度/题目内容/作答时间提示
  - Textarea 输入框（支持代码块提示）
  - 「提交答案」→ 调 `POST /api/v1/interviews/:id/answers` → 显示 AI 点评区域（score/pros/cons/参考答案折叠）
  - 「跳过此题」→ 直接进下一题（本地标记 skipped，不调后端）
  - 「暂停面试」→ 调 `PUT /api/v1/interviews/:id/pause` → 跳转仪表盘
  - 所有题目完成 → 调 `PUT /api/v1/interviews/:id/complete` → 跳转 `/report/:id`

- [x] 12.5 面试历史页
  - 创建 `ui/src/views/interview/History.vue`
  - 列表展示历史面试（时间/岗位/得分/轮次/状态）
  - 支持查看报告、继续未完成面试

---

## Task 13: 前端面试报告页

- [x] 13.1 API 模块
  - 创建 `ui/src/api/report.js`

- [x] 13.2 RadarChart 公共组件
  - 创建 `ui/src/components/common/RadarChart.vue`
  - 使用 ECharts 渲染雷达图，接收 `modules` prop（模块名+得分数组）

- [x] 13.3 ScoreBar 公共组件
  - 创建 `ui/src/components/common/ScoreBar.vue`
  - 接收 `score` prop，按 80+/60-79/0-59 显示绿/橙/红色进度条

- [x] 13.4 报告详情页
  - 创建 `ui/src/views/report/Detail.vue`
  - 顶部通过状态横幅（pass/pending/fail 三种颜色样式，参考 prototype.html）
  - 圆形综合得分 + 基本信息（岗位/轮次/难度/用时/作答题数）
  - 知识点雷达图（RadarChart 组件）
  - 逐题得分明细（可折叠，ScoreBar组件）：题目/用户答案/AI得分/AI点评/参考答案折叠
  - 题目列表筛选 Tab（全部/已作答/已跳过/高分/低分）
  - 知识点模块得分卡片（3列网格）
  - AI 综合评价与建议（优势/不足/备考路线图）
  - 生成分享链接按钮 → 调 `POST /api/v1/reports/:id/share` → 显示链接+复制

---

## Task 14: 前端题库页

- [x] 14.1 API 模块
  - 创建 `ui/src/api/question.js`

- [x] 14.2 MarkdownRender 公共组件
  - 创建 `ui/src/components/common/MarkdownRender.vue`
  - 使用 `marked` 库渲染 Markdown 内容

- [x] 14.3 题库列表页
  - 创建 `ui/src/views/question/List.vue`
  - 顶部筛选栏：岗位/难度/标签/类型（Element Select + Tag）+ 关键词搜索
  - 题目列表（每项：标题/岗位标签/难度徽章/知识点/作答人数/平均分/收藏按钮）
  - 分页（Element Pagination，每页20条）
  - 点击题目 → 跳转 `/questions/:id/practice`

- [x] 14.4 单题练习页
  - 创建 `ui/src/views/question/Practice.vue`
  - 显示完整题目 + Textarea 输入答案
  - 提交后显示 AI 点评（score/pros/cons/参考答案折叠）
  - 收藏/取消收藏按钮

---

## Task 15: 前端知识社区页

- [x] 15.1 API 模块
  - 创建 `ui/src/api/community.js`

- [x] 15.2 社区首页
  - 创建 `ui/src/views/community/Index.vue`
  - 顶部岗位分类 Tab（后端/前端/大数据/AI/会计/通用）
  - 排序切换（热度/最新）
  - 文章列表（标题/标签/点赞数/收藏数/评论数/时间）
  - 「AI 知识获取」按钮 → 弹出 Dialog，输入知识点 → 调 `POST /api/v1/community/articles/ai` → 刷新列表
  - 429 错误时提示「今日 AI 生成次数已用完」

- [x] 15.3 文章详情页
  - 创建 `ui/src/views/community/Article.vue`
  - 标题/标签/时间 + Markdown 渲染正文（MarkdownRender 组件）
  - 点赞/收藏操作按钮
  - 评论列表 + 发表评论输入框

---

## Task 16: 前端个人中心页

- [x] 16.1 API 模块
  - 创建 `ui/src/api/profile.js`

- [x] 16.2 个人中心页
  - 创建 `ui/src/views/profile/Index.vue`
  - Tab 切换：个人信息/面试历史/成长轨迹/我的收藏
  - 个人信息 Tab：头像/昵称/简介修改（Element Form）+ 修改密码（验证旧密码）
  - 面试历史 Tab：表格列表（时间/岗位/得分/轮次），点击查看报告
  - 成长轨迹 Tab：ECharts 折线图（近30次得分趋势）
  - 我的收藏 Tab：分「题目」「文章」两个 Tag 面板

---

## Task 17: 集成测试与联调

- [~] 17.1 前后端联调
  - 验证注册/登录流程（JWT 存储、路由守卫）
  - 验证完整面试流程（配置 → 出题 → 答题 → 报告）
  - 验证 Redis 数据写入（用户信息永久、会话7天、题目缓存24h）

- [~] 17.2 边界情况处理
  - DeepSeek 调用失败时的 fallback 提示和重试按钮
  - 面试中途刷新页面能从 Redis 恢复进度
  - 未登录访问任意页面自动跳转 `/login`

- [x] 17.3 启动文档
  - 创建 `README.md`，说明后端（`cd api && go run main.go`）和前端（`cd ui && npm run serve`）启动方式
  - 说明 Redis 连接配置和 DeepSeek API Key 配置方式
