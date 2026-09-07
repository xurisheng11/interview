# 面试轮次选择与思考时间改造 - 需求变更文档

## 简介

本文档是对 `requirements.md` 中"需求二：面试配置"和"需求九：视频面试增强"的增量变更。核心改造三点：
1. **面试轮次选择**：新增"综合面试"选项，增强引导文案（覆盖范围 + 风险提示）
2. **思考时间**：从"用户手动设定固定秒数"改为"系统自动捕捉真实思考时长"
3. **引导文案**：每个轮次选项必须明确标注覆盖范围和风险

## 术语表（增量）

| 术语 | 说明 |
|------|------|
| 综合面试 | 新增轮次选项（comprehensive），题目均衡覆盖基础、技术深度、综合/HR，适合只有一轮面试的单位 |
| 自动思考捕捉 | 系统自动记录"题目展示→用户首次开口/输入"的时间差，作为该题真实思考时长 |
| 最长思考上限 | 用户可选配置，超过该时间仍未作答时系统给出提醒（默认 120 秒） |

---

## 需求背景

### 用户反馈来源

求职者对现有"面试轮次"和"思考时间"设计提出三点核心问题：

1. **轮次不适配单轮面试场景**：很多单位（尤其当前趋势）只有一轮面试，甚至医院等单位分"科面"和"院面"，现有 一面/二面/三面 的分类无法覆盖。用户不知道"只有一轮面试时该选哪个"。
2. **思考时间不应手动设定**：思考时间取决于题目难度和掌握程度，用户无法提前预估。应该由系统自动捕捉真实思考时长。
3. **引导文案缺失风险提示**：现有"如果只有一轮面试，建议选三面"的引导没有告知用户"三面不含基础题"这一关键风险。

### 改造目标

- 让用户在面对任何面试场景时，都能快速选择到最适配的练习模式
- 让思考时间数据真实反映用户状态，为报告提供更精准的分析维度
- 让每个选项的覆盖范围和风险一目了然，用户不会"踩坑"

---

## 需求一：面试轮次选择改造

**用户故事：** 作为面试者，我希望在选择面试轮次时能看到清晰的覆盖范围和适用场景说明，以便选到最匹配我目标面试的练习模式。

### 功能设计

#### 1.1 新增"综合面试"选项（P0）

在现有 round1/round2/round3 基础上，新增第四个选项 `comprehensive`：

| 值 | 显示名 | 题目构成 |
|----|--------|---------|
| round1 | 一面 | 自我介绍(1题) + 基础知识(60%) + 简单算法/逻辑(20%) + 项目简述(20%) |
| round2 | 二面 | 技术深度(50%) + 系统设计(30%) + 项目追问(20%) |
| round3 | 三面 | 综合能力(40%) + 职业规划(30%) + HR(30%) |
| **comprehensive** | **综合面试** | **自我介绍(1题) + 基础知识(25%) + 技术深度(25%) + 项目经验(25%) + 综合/HR(25%)** |

**后端 prompt 改造**：`deepseek_service.go` 的 `getRoundRequirement` 和 `getRoundName` 新增 `comprehensive` 分支。

#### 1.2 增强引导文案（P0）

每个轮次选项的展示信息从"名称 + 6字描述"扩展为"名称 + 覆盖范围 + 适合人群 + 风险提示"：

| 选项 | 覆盖范围 | 适合人群 | 风险提示 |
|------|---------|---------|---------|
| 一面 | 基础概念、算法入门、项目概述 | 正在准备第一轮面试，或想夯实基础 | 不含深度技术题和系统设计 |
| 二面 | 技术深挖、系统设计、项目追问 | 已通过一面，准备技术深面 | 不含基础题和HR类问题 |
| 三面 | 综合素质、职业规划、HR问答 | 准备终面/HR面 | ⚠️ 不含基础和技术深度题 |
| 综合面试 | 均衡覆盖基础→深度→综合全链路 | 目标单位只有一轮面试，或想全面练习 | 题量有限，各方向覆盖不如单轮深入 |

**UI 交互**：选中某轮次后，下方显示 `el-alert` 提示框，展示完整的覆盖范围和风险提示。

#### 1.3 存量数据兼容（P0）

- 已有面试记录中 `round` 字段值（round1/round2/round3）不变，前端展示和历史报告均正常
- `InterviewListItem.Round` 和 `InterviewReport.Round` 需能正确展示 `comprehensive` 为"综合面试"
- 题目缓存 key（`buildQuestionsCacheKey`）已包含 `cfg.Round`，新增 `comprehensive` 值自然兼容

#### 1.4 验收标准（GWT 格式）

1. WHEN 用户进入面试配置页，THEN 轮次选择区展示 4 个选项：一面、二面、三面、综合面试。
2. WHEN 用户点击"综合面试"，THEN 该选项高亮选中，下方显示 alert 提示"均衡覆盖基础→深度→综合全链路，适合只有一轮面试的场景。注意：各方向题量有限，不如单轮深入。"
3. WHEN 用户点击"三面"，THEN 下方显示 alert 提示"⚠️ 三面仅包含综合能力、职业规划和HR类问题，不含基础知识和技术深度题。如果你的目标面试只有一轮，建议选择「综合面试」以获得更全面的覆盖。"
4. WHEN 用户选择"综合面试"并提交，THEN 后端 `getRoundRequirement` 返回综合面试的题目构成要求，`getRoundName` 返回"综合面试"。
5. WHEN 用户查看历史面试列表，THEN round 为 `comprehensive` 的记录正确显示"综合面试"。
6. WHEN 已有 round1/round2/round3 的历史记录存在时，THEN 这些记录的展示和功能不受影响。

---

## 需求二：思考时间自动捕捉

**用户故事：** 作为面试者，我希望系统能自动记录我每道题的真实思考时间（从看到题目到开始作答），以便在报告中看到自己的反应速度和准备程度。

### 功能设计

#### 2.1 视频模式（VideoDoing.vue）自动捕捉（P0）

**核心逻辑**：
- 记录 `questionDisplayedAt`：每道题目渲染到页面的时间戳（`Date.now()`）
- 记录 `firstSpeechAt`：语音识别 `onresult` 回调中首次收到**非空** `final` 或 `interim` 结果的时间戳
- `thinkDuration = Math.round((firstSpeechAt - questionDisplayedAt) / 1000)`（秒）
- 若用户手动输入（未使用语音），则 `firstSpeechAt` 取 textarea 首次 `input` 事件时间

**实现要点**：
- `speechMixin.js` 的 `onresult` 回调中新增 `firstSpeechTime` 字段（仅首次非空结果时写入）
- `VideoDoing.vue` 的 `handleNext()` 中记录 `questionDisplayedAt = Date.now()`，同时重置 `firstSpeechTime`
- `getNonVerbalMetrics()` 返回值新增 `thinkDuration` 字段
- 提交答案时将 `thinkDuration` 通过 `SubmitAnswerReq.ThinkDuration` 发送到后端

**作答时长修正**：
- 现有 `duration` 计算从 `speechMetrics.startTime`（即 `startSpeech()` 调用时）到 `stopSpeech()`
- 改造后 `duration` 仍从 `startSpeech()` 开始计（即语音识别启动时），但 `thinkDuration` 独立计算
- 实际作答时长 = `duration - thinkDuration`（在报告中展示）

#### 2.2 文字模式（Doing.vue）自动捕捉（P1）

**核心逻辑**：
- 记录 `questionDisplayedAt`：每道题目渲染到页面的时间戳
- 记录 `firstInputAt`：textarea 首次 `input` 事件（或 `focus` 后首次按键）的时间戳
- `thinkDuration = Math.round((firstInputAt - questionDisplayedAt) / 1000)`（秒）
- 提交答案时将 `thinkDuration` 通过 `SubmitAnswerReq.ThinkDuration` 发送到后端

#### 2.3 配置页改造（P0）

**移除**：现有的思考时间滑块（`el-slider` 0-60秒）

**替换为**：最长思考上限配置（`el-select` 下拉选择）

| 选项值 | 显示文案 |
|--------|---------|
| 60 | 60秒（快节奏练习） |
| 120 | 2分钟（推荐，接近真实面试） |
| 180 | 3分钟（充分思考） |
| 0 | 不限制 |

- 默认值：120秒
- 当思考时间超过上限时，前端显示**温和提醒**（如"思考时间已较长，建议开始作答"），但不强制中断
- 此配置仅影响提醒行为，不影响数据记录

**文案更新**：
- 旧："显示题目后，系统会等待 X 秒再开始录制你的回答"
- 新："系统会自动记录你每道题的思考时间。超过设定时长会温和提醒你开始作答。"

#### 2.4 后端改造（P0）

- `CreateInterviewReq.ThinkTime` 语义变更：从"固定等待秒数"改为"最长思考上限秒数"（字段名不变，保持兼容）
- `interview_service.go` 默认值从 15 改为 120
- `SubmitAnswerReq.ThinkDuration`：前端现在会正确填充此字段（之前始终为 0）
- `deepseek_service.go` ReviewAnswer prompt 中 `思考时长：%d秒` 现在能接收到真实数据
- 无需改动 `report_service.go`：ThinkDuration 已在 NonVerbalMetrics 中持久化，报告页面可直接展示

#### 2.5 报告展示增强（P1）

- **单题报告**：在现有"用时：X秒"旁新增"思考：Y秒"展示
- **综合报告**：新增"平均思考时间"统计（`CalcAvgThinkDuration` 函数，类似 `CalcAvgSpeechRate`）
- **AI 评价**：`GenerateVideoExpressionSummary` prompt 中注入平均思考时间数据

#### 2.6 验收标准（GWT 格式）

1. WHEN 用户在视频模式下看到新题目，THEN 系统记录 `questionDisplayedAt` 时间戳。
2. WHEN 用户在视频模式下首次说话（语音识别收到非空结果），THEN 系统记录 `firstSpeechAt` 并计算 `thinkDuration`。
3. WHEN 用户提交视频模式答案，THEN `nonVerbalMetrics.thinkDuration` 包含真实思考秒数（>0）。
4. WHEN 用户在文字模式下看到新题目，THEN 系统记录 `questionDisplayedAt` 时间戳。
5. WHEN 用户在文字模式下首次在 textarea 中输入内容，THEN 系统记录 `firstInputAt` 并计算 `thinkDuration`。
6. WHEN 用户提交文字模式答案，THEN `thinkDuration` 字段包含真实思考秒数。
7. WHEN 面试配置页选择视频模式，THEN 显示"最长思考上限"下拉选择（默认2分钟），不再显示 0-60秒滑块。
8. WHEN 用户思考时间超过设定的上限，THEN 页面显示温和提醒"思考时间已较长，建议开始作答"。
9. WHEN 查看单题报告时，THEN 展示"思考：X秒"（位于"用时"指标旁）。
10. WHEN 查看综合报告时，THEN 视频模式报告中展示"平均思考时间"统计。
11. WHEN 用户切到下一题（handleNext），THEN `questionDisplayedAt` 和 `firstSpeechTime` 被正确重置。
12. WHEN 用户跳过某题，THEN 该题不记录 thinkDuration（跳过的题 thinkDuration = 0）。

---

## 数据结构变更

### 前端 Config.vue

```javascript
// roundOptions 新增 comprehensive
roundOptions: [
  {
    label: '一面',
    value: 'round1',
    desc: '基础能力考察',
    coverage: '自我介绍 + 基础知识 + 简单算法 + 项目概述',
    audience: '正在准备第一轮面试，或想夯实基础',
    risk: '不含深度技术题和系统设计'
  },
  {
    label: '二面',
    value: 'round2',
    desc: '技术深度考察',
    coverage: '技术深挖 + 系统设计 + 项目追问',
    audience: '已通过一面，准备技术深面',
    risk: '不含基础题和HR类问题'
  },
  {
    label: '三面',
    value: 'round3',
    desc: '综合能力/HR面',
    coverage: '综合素质 + 职业规划 + HR问答',
    audience: '准备终面/HR面',
    risk: '⚠️ 不含基础和技术深度题。如果目标面试只有一轮，建议选择「综合面试」'
  },
  {
    label: '综合面试',
    value: 'comprehensive',
    desc: '全链路覆盖',
    coverage: '自我介绍 + 基础 + 技术深度 + 项目经验 + 综合/HR 均衡分布',
    audience: '目标单位只有一轮面试，或想全面练习',
    risk: '题量有限，各方向覆盖不如单轮深入'
  }
]

// 思考时间配置变更
config: {
  // 旧字段（保留兼容，语义变更）
  thinkTime: 120,  // 从 15 改为 120，语义从"固定等待"变为"最长上限"
}
```

### 前端 speechMixin.js

```javascript
// data 新增
speechMetrics: {
  // ...现有字段...
  firstSpeechTime: null  // 首次检测到语音的时间戳
}

// onresult 中新增
if (!this.speechMetrics.firstSpeechTime && (final || interim)) {
  this.speechMetrics.firstSpeechTime = now
}

// startSpeech 中重置
this.speechMetrics.firstSpeechTime = null

// getNonVerbalMetrics 返回新增 thinkDuration
getNonVerbalMetrics() {
  // ...现有逻辑...
  const thinkDuration = (this.speechMetrics.firstSpeechTime && this.questionDisplayedAt)
    ? Math.round((this.speechMetrics.firstSpeechTime - this.questionDisplayedAt) / 1000)
    : 0
  return {
    speechRate: this.calcSpeechRate(),
    pauseCount: this.speechMetrics.pauseCount,
    duration,
    thinkDuration,  // 新增
    verbalTics
  }
}
```

### 前端 VideoDoing.vue

```javascript
// data 新增
data() {
  return {
    // ...现有字段...
    questionDisplayedAt: null  // 当前题目展示时间戳
  }
}

// mounted 中记录
mounted() {
  this.questionDisplayedAt = Date.now()
  // ...现有逻辑...
}

// handleNext 中重置
handleNext() {
  if (this.isLastQuestion) {
    this.handleComplete()
  } else {
    this.currentIdx++
    this.userAnswer = ''
    this.questionDisplayedAt = Date.now()  // 新增
    this.resetTimer()
    this.resetTicState()
    if (this.isSpeechSupported) {
      this.startSpeech()
    }
  }
}

// handleSubmit 中传递 thinkDuration
async handleSubmit() {
  this.stopSpeech()
  const metrics = this.getNonVerbalMetrics()  // 已包含 thinkDuration
  // submitAnswer 时 metrics 已包含 thinkDuration，无需额外处理
}
```

### 前端 Doing.vue

```javascript
// data 新增
data() {
  return {
    // ...现有字段...
    questionDisplayedAt: null,  // 当前题目展示时间戳
    firstInputAt: null          // 首次输入时间戳
  }
}

// 初始化时记录
init() {
  // ...现有逻辑...
  this.questionDisplayedAt = Date.now()
}

// handleNext 中重置
handleNext() {
  if (this.isLastQuestion) {
    this.handleComplete()
  } else {
    this.currentIdx++
    this.currentAnswer = ''
    this.questionDisplayedAt = Date.now()  // 新增
    this.firstInputAt = null               // 新增
    this.resetTimer()
  }
}

// handleSubmit 中计算并传递 thinkDuration
async handleSubmit() {
  if (!this.currentAnswer.trim()) return
  this.submitting = true
  const thinkDuration = (this.firstInputAt && this.questionDisplayedAt)
    ? Math.round((this.firstInputAt - this.questionDisplayedAt) / 1000)
    : 0
  try {
    const res = await submitAnswer(this.interviewId, {
      questionIndex: this.currentIdx,
      answer: this.currentAnswer.trim(),
      thinkDuration  // 新增
    })
    // ...
  }
}
```

### 后端变更

#### `api/service/deepseek_service.go`

```go
// getRoundName 新增 comprehensive 分支
func getRoundName(round string) string {
    switch round {
    case "round1":
        return "一面（基础）"
    case "round2":
        return "二面（技术深度）"
    case "round3":
        return "三面（综合/HR）"
    case "comprehensive":
        return "综合面试"
    default:
        return round
    }
}

// getRoundRequirement 新增 comprehensive 分支
func getRoundRequirement(round string) string {
    switch round {
    case "round1":
        return "一面：必含自我介绍(1题) + 基础知识题(60%) + 简单算法/逻辑题(20%) + 项目经历简述(20%)"
    case "round2":
        return "二面：技术深度题(50%) + 系统设计题(30%) + 项目经验追问(20%)"
    case "round3":
        return "三面：综合能力题(40%) + 职业规划(30%) + HR类问题(30%)"
    case "comprehensive":
        return "综合面试：必含自我介绍(1题) + 基础知识题(25%) + 技术深度题(25%) + 项目经验题(25%) + 综合/HR题(25%)"
    default:
        return "均衡分配各类题目"
    }
}
```

#### `api/service/interview_service.go`

```go
// 默认值变更
if req.ThinkTime <= 0 {
    req.ThinkTime = 120 // 默认120秒（最长思考上限，而非固定等待时间）
}
```

#### `api/model/report.go`

```go
// 新增函数
func CalcAvgThinkDuration(questions []ReportQuestion) int {
    total, count := 0, 0
    for _, q := range questions {
        if !q.Skipped && q.NonVerbalMetrics != nil && q.NonVerbalMetrics.ThinkDuration > 0 {
            total += q.NonVerbalMetrics.ThinkDuration
            count++
        }
    }
    if count == 0 {
        return 0
    }
    return total / count
}
```

#### `api/model/interview.go`

无需改动。`InterviewConfig.Round` 为 string 类型，新增值 `comprehensive` 自然兼容。`NonVerbalMetrics.ThinkDuration` 已存在。

---

## API 变更

### 创建面试 POST /interviews

请求体新增/变更：
- `round`：新增值 `"comprehensive"`（与现有 round1/round2/round3 并列）
- `thinkTime`：语义从"固定等待秒数"改为"最长思考上限秒数"，默认值从 15 改为 120

响应体无变更。

### 提交答案 POST /interviews/:id/answers

请求体变更：
- `thinkDuration`：之前存在但前端从未发送（始终为 0），改造后前端会发送真实值

响应体无变更。

---

## 任务拆解

| 编号 | 任务 | 涉及文件 | 复杂度 | 依赖 |
|------|------|---------|--------|------|
| T1 | 后端 `getRoundName` 和 `getRoundRequirement` 新增 `comprehensive` 分支 | `api/service/deepseek_service.go` | 低 | 无 |
| T2 | 后端 `interview_service.go` ThinkTime 默认值从 15 改为 120 | `api/service/interview_service.go` | 低 | 无 |
| T3 | 后端新增 `CalcAvgThinkDuration` 函数 | `api/model/report.go` | 低 | 无 |
| T4 | 前端 Config.vue `roundOptions` 新增 `comprehensive` 选项及增强描述字段 | `ui/src/views/interview/Config.vue` | 低 | T1 |
| T5 | 前端 Config.vue 轮次选择 UI 增加选中后的 alert 提示（覆盖范围+风险） | `ui/src/views/interview/Config.vue` | 中 | T4 |
| T6 | 前端 Config.vue 思考时间配置改为"最长思考上限"下拉选择 | `ui/src/views/interview/Config.vue` | 中 | T2 |
| T7 | 前端 speechMixin.js 新增 `firstSpeechTime` 追踪，改造 `getNonVerbalMetrics` | `ui/src/mixins/speechMixin.js` | 中 | 无 |
| T8 | 前端 VideoDoing.vue 新增 `questionDisplayedAt` 追踪，handleNext 中重置 | `ui/src/views/interview/VideoDoing.vue` | 中 | T7 |
| T9 | 前端 Doing.vue 新增 `questionDisplayedAt` 和 `firstInputAt` 追踪，handleSubmit 中传递 thinkDuration | `ui/src/views/interview/Doing.vue` | 中 | 无 |
| T10 | 前端报告页面展示思考时间数据 | 报告详情页（待确认文件） | 低 | T3, T7, T8, T9 |
| T11 | 后端 `GenerateVideoExpressionSummary` prompt 中注入平均思考时间 | `api/service/deepseek_service.go` | 低 | T3 |
| T12 | 面试列表/历史中 `comprehensive` 的正确显示 | `api/service/interview_service.go` | 低 | T1 |

---

## 正确性属性（增量）

1. **轮次互斥性**：每次面试只能选择一个轮次值（round1/round2/round3/comprehensive 四选一）。
2. **思考时间非负性**：`thinkDuration` 始终 ≥ 0，且 ≤ `questionDisplayedAt` 到 `submittedAt` 的总时间差。
3. **思考时间独立性**：`thinkDuration` 不计入作答时长（`duration`），两者在报告中分别展示。
4. **缓存键兼容性**：`buildQuestionsCacheKey` 中 `cfg.Round` 为 `comprehensive` 时生成的缓存键与其他轮次不冲突。
5. **存量数据不变性**：已有 round1/round2/round3 的面试记录和报告，在升级后展示和功能完全不变。

---

## 风险与待确认项

| 编号 | 问题 | 建议 |
|------|------|------|
| R1 | 科面/院面等细分场景是否需要支持？ | 建议放到 v3，当前通过"补充说明"字段 + AI 自适应出题部分覆盖 |
| R2 | 渐进式练习路径（从一面练到三面）是否需要？ | 建议放到 v3，当前用户可手动分三次面试完成 |
| R3 | 文字模式的 thinkDuration 精度问题（用户可能先看了很久再开始打字） | 可接受，这正是真实思考时间的体现 |
| R4 | 语音识别启动延迟（浏览器 SpeechRecognition 有 1-2秒启动时间） | 可接受，误差在合理范围内；`firstSpeechTime` 取 onresult 首次非空结果，比 start() 更准确 |
| R5 | ThinkTime 语义变更是否影响存量数据？ | 不影响。已有记录中 ThinkTime 仅作为配置存档，不参与运行时逻辑 |
