# Tasks

## Implementation Plan

本任务清单基于需求文档（requirements.md）和设计文档（design.md），采用增量开发策略，优先实现核心功能，再扩展高级特性。

---

## Overview

视频面试模式功能的实施将分为 **5 个阶段**，共 **18 个任务**：

1. **阶段 1 - 前端基础**（T1-T4）：设备检测、权限授权、摄像头预览组件
2. **阶段 2 - 语音识别**（T5-T7）：Web Speech API 集成、实时转文字、非语言指标采集
3. **阶段 3 - 后端扩展**（T8-T11）：数据模型扩展、API 改造、AI Prompt 优化
4. **阶段 4 - 视频录制**（T12-T13）：MediaRecorder 本地录制、下载功能
5. **阶段 5 - 报告与优化**（T14-T18）：视频模式报告页、兼容性处理、集成测试

---

## Tasks


### Task 1: 后端数据模型扩展

**关联需求：** Requirement 11

- [ ] 1.1 扩展 `api/model/interview.go`
  - 在 `InterviewSession` struct 中新增 `Mode string` 字段（json tag: `"mode"`，默认值 `"text"`）
  - 新增 `NonVerbalMetrics` struct：`SpeechRate float64`、`PauseCount int`、`Duration int`
  - 在 `AnswerRecord` struct 中新增 `ExpressionScore int`、`ExpressionFeedback string`、`NonVerbalMetrics *NonVerbalMetrics`（均加 `omitempty`）
  - 更新 `ToRedisHash()` 和 `FromRedisHash()` 方法，确保新字段正确序列化/反序列化（JSON 存储）

- [ ] 1.2 扩展 `api/model/report.go`
  - 在 `InterviewReport` struct 中新增 `Mode string`、`ExpressionSummary string`、`AvgExpressionScore int`、`AvgSpeechRate float64`（均加 `omitempty`）
  - 新增 `calcAvgExpressionScore(answers []AnswerRecord) int` 辅助函数（过滤 ExpressionScore 为 0 的项）
  - 新增 `calcAvgSpeechRate(answers []AnswerRecord) float64` 辅助函数


### Task 2: 后端 API 接口扩展

**关联需求：** Requirement 12

- [ ] 2.1 修改 `api/service/interview_service.go`
  - `CreateInterview` 函数参数中新增 `mode string`，存入 `InterviewSession.Mode` 字段
  - `SubmitAnswer` 函数参数新增 `nonVerbalMetrics *NonVerbalMetrics`，存入 `AnswerRecord`
  - 确保向后兼容：文字模式调用时 `mode` 默认 `"text"`，`nonVerbalMetrics` 为 nil

- [ ] 2.2 修改 `api/handler/interview.go`
  - 在 `CreateInterview` Handler 的请求体绑定 struct 中新增 `Mode string`（json tag）
  - 在 `SubmitAnswer` Handler 的请求体绑定 struct 中新增 `NonVerbalMetrics *NonVerbalMetrics`（json tag）
  - 调用 service 时传递新字段

- [ ] 2.3 测试 API 向后兼容性
  - 使用 Postman/curl 测试文字模式创建面试（不传 `mode` 字段），验证默认为 `"text"`
  - 测试文字模式提交答案（不传 `nonVerbalMetrics`），验证不报错


### Task 3: DeepSeek Prompt 扩展（视频模式点评）

**关联需求：** Requirement 6, Requirement 7

- [ ] 3.1 修改 `api/service/deepseek_service.go` 的 `ReviewAnswer` 函数
  - 函数签名新增 `metrics *NonVerbalMetrics` 参数
  - 当 `metrics != nil` 时，在基础 prompt 末尾追加非语言指标上下文：
    ```
    [语音表达数据]
    语速：{speechRate}字/分钟（推荐120-150），停顿次数：{pauseCount}次，作答时长：{duration}秒。
    请额外在JSON中新增字段：expressionScore（0-100，评估口头表达质量）和 expressionFeedback（具体改进建议）。
    ```
  - 解析响应时提取 `expressionScore` 和 `expressionFeedback` 字段，写入 `ReviewResult`

- [ ] 3.2 新增 `GenerateVideoExpressionSummary` 函数
  - 函数接收 `[]AnswerRecord`，计算平均语速和平均表达得分
  - 构造 prompt 请求 DeepSeek 生成整体表达能力评价（返回 `summary string`）
  - 在 `report_service.go` 的 `GenerateReport` 函数中，若 `session.Mode == "video"` 则调用此函数，写入报告


### Task 4: 前端配置页扩展（答题方式选择）

**关联需求：** Requirement 1

- [ ] 4.1 修改 `ui/src/views/interview/Config.vue`
  - 在表单中新增「答题方式」区域（`<el-form-item label="答题方式">`）
  - 使用 `<el-radio-group v-model="config.mode">` 提供「文字输入」和「视频面试」两个选项
  - 当选择「视频面试」时，显示提示文字（需要摄像头和麦克风权限，推荐 Chrome）
  - 修改 `handleStart()` 方法：若 `config.mode === 'video'`，跳转到 `/interview/:id/video-prep` 而非 `/interview/loading`

- [ ] 4.2 更新 `ui/src/api/interview.js`
  - 在 `createInterview(data)` 请求中传递 `mode` 字段


### Task 5: 前端公共组件 - VideoPreview

**关联需求：** Requirement 2, Requirement 3

- [ ] 5.1 创建 `ui/src/components/interview/VideoPreview.vue`
  - 接收 prop：`stream`（MediaStream 对象）、`mirror`（Boolean，默认 true）
  - 使用 `<video ref="videoEl" autoplay muted playsinline>` 渲染画面
  - `watch stream`：当 stream 变化时，将 `videoEl.srcObject = stream`
  - CSS：镜像效果 `.mirrored { transform: scaleX(-1); }`，宽高自适应

- [ ] 5.2 创建 `ui/src/components/interview/SpeechIndicator.vue`
  - 接收 prop：`stream`（MediaStream）、`active`（Boolean，是否正在识别）
  - 使用 `AudioContext` + `createAnalyser()` 绘制实时音量波形（Canvas 120x40）
  - 显示状态文字：「正在识别...」/ 「已暂停」/ 「不支持」
  - `beforeDestroy` 时停止 AudioContext 防止内存泄漏


### Task 6: 前端 Mixin - 语音识别

**关联需求：** Requirement 4

- [ ] 6.1 创建 `ui/src/mixins/speechMixin.js`
  - 实现 `initSpeechRecognition(lang)` 方法，初始化 `SpeechRecognition`（兼容 `window.webkitSpeechRecognition`）
  - 配置参数：`continuous: true`、`interimResults: true`、`maxAlternatives: 1`
  - `onresult` 回调：区分 interim（临时）和 final（最终）结果，累计 `finalTranscript`，更新 `interimTranscript`
  - `onresult` 回调：记录时间戳，若两次结果间隔超过 2000ms，`pauseCount++`
  - 实现 `startSpeech()`、`stopSpeech()` 控制方法
  - 实现 `calcSpeechRate()` 计算语速（字数 / 时长（分钟））
  - 实现 `getNonVerbalMetrics()` 返回 `{ speechRate, pauseCount, duration }`

- [ ] 6.2 兼容性检测方法
  - 在 mixin 中提供 `checkSpeechSupport()` 方法，返回 Boolean
  - 若不支持，`isSpeechSupported = false`，页面显示降级提示


### Task 7: 前端 Mixin - 视频录制

**关联需求：** Requirement 5

- [ ] 7.1 创建 `ui/src/mixins/recordingMixin.js`
  - 实现 `startRecording(stream)` 方法：
    - 检测 `MediaRecorder` 支持情况，按优先级尝试 mimeType：`video/webm;codecs=vp9` → `video/webm` → `video/mp4`
    - `recorder.start(30000)`（每 30 秒一个 chunk）
  - 实现 `stopRecording()` 方法：停止录制，触发 `recorder.onstop`
  - `ondataavailable` 回调：将 chunk 推入 `recordedChunks[]`
  - `onstop` 回调：合并 chunks 为 Blob，生成 objectURL，通过 `$emit('recording-ready', { url, size, duration })` 通知父组件
  - 实现 `downloadVideo(url, filename)` 方法：创建临时 `<a>` 标签触发下载

- [ ] 7.2 错误处理
  - `try/catch` 包裹 `MediaRecorder` 初始化，捕获内存不足等异常
  - 录制出错时 `$emit('recording-error', errorMessage)`，面试可继续进行


### Task 8: Vuex Store 扩展

**关联需求：** Requirement 9

- [ ] 8.1 修改 `ui/src/store/modules/interview.js`
  - 新增 state 字段：`mediaStream: null`、`interviewMode: 'text'`、`enableRecording: false`、`recordedVideo: null`
  - 新增 mutations：
    - `SET_MEDIA_STREAM(state, stream)`
    - `SET_INTERVIEW_MODE(state, mode)`
    - `SET_ENABLE_RECORDING(state, bool)`
    - `SET_RECORDED_VIDEO(state, { url, blob, size, duration })`
    - `RELEASE_MEDIA_STREAM(state)`：调用 `stream.getTracks().forEach(t => t.stop())` 后置 null


### Task 9: 前端路由扩展

**关联需求：** Requirement 1, Requirement 2

- [ ] 9.1 修改 `ui/src/router/index.js`
  - 新增路由：`{ path: '/interview/:id/video-prep', name: 'VideoPreparation', component: () => import('@/views/interview/VideoPreparation.vue'), meta: { requiresAuth: true } }`
  - 新增路由：`{ path: '/interview/:id/video-doing', name: 'VideoDoing', component: () => import('@/views/interview/VideoDoing.vue'), meta: { requiresAuth: true } }`
  - 确保两条路由受现有路由守卫保护（requiresAuth）


### Task 10: 视频面试准备页

**关联需求：** Requirement 2, Requirement 10

- [ ] 10.1 创建 `ui/src/views/interview/VideoPreparation.vue`
  - 调用 `navigator.mediaDevices.getUserMedia({ video: true, audio: true })` 请求权限
  - 获取成功后，提交 stream 到 Vuex（`SET_MEDIA_STREAM`），同时显示摄像头预览（使用 `VideoPreview` 组件）
  - 权限被拒绝时，分别显示摄像头/麦克风权限拒绝提示（`<el-alert>`）
  - 兼容性检查区域：用绿色/橙色/红色图标展示三项功能支持状态（摄像头、语音识别、视频录制）
  - 提供语言选择（`<el-select>`：中文/英文），存入组件 data 备用
  - 提供「是否录制视频」开关（`<el-switch>`，默认关），提交到 Vuex（`SET_ENABLE_RECORDING`）
  - 隐私声明弹窗（`<el-dialog>`），用户点击「同意并开始」后调用 `startInterview()`

- [ ] 10.2 实现 `startInterview()` 方法
  - 跳转到 `/interview/:id/video-doing`（interviewId 从路由参数获取）
  - `beforeDestroy` 时：若用户未点击「开始」则调用 `RELEASE_MEDIA_STREAM` 释放资源


### Task 11: 视频面试答题页

**关联需求：** Requirement 3, Requirement 4, Requirement 5

- [ ] 11.1 创建 `ui/src/views/interview/VideoDoing.vue`（基础结构）
  - 引入 `speechMixin` 和 `recordingMixin`
  - 从 Vuex 获取 `mediaStream`、`questions`、`enableRecording`
  - 左右双栏布局：左侧 `VideoPreview` 组件 + 录制状态指示，右侧题目卡片 + 答题区
  - 挂载时：初始化语音识别（`initSpeechRecognition(lang)`），若 `enableRecording` 则 `startRecording(stream)`
  - 进度条、题目导航（题号颜色状态）、倒计时 - 复用现有 Doing.vue 的逻辑

- [ ] 11.2 语音转文字答题交互
  - Textarea 绑定 `userAnswer`（双向编辑）
  - 在 Textarea 下方显示临时识别结果（灰色斜体，`interimTranscript`）
  - 提供麦克风开关按钮（`<el-switch>`），调用 `toggleSpeech()`
  - 语言切换按钮（从准备页传入语言设置）

- [ ] 11.3 提交答案
  - 点击「提交答案」：调用 `stopSpeech()` 获取 `getNonVerbalMetrics()`
  - 调用 `POST /api/v1/interviews/:id/answers`，请求体中附加 `nonVerbalMetrics`
  - 收到响应后，显示 AI 点评区域，新增「表达得分」和「表达反馈」展示
  - 调用 `startSpeech()` 开始下一题语音识别

- [ ] 11.4 面试结束与资源释放
  - 最后一题提交后：调用 `stopRecording()`（若开启录制）
  - 调用 `PUT /api/v1/interviews/:id/complete`
  - 调用 `RELEASE_MEDIA_STREAM` 释放摄像头/麦克风
  - 跳转到 `/report/:id`
  - `beforeRouteLeave` 守卫：检测面试进行中时弹出确认框


### Task 12: 面试报告页扩展（视频模式）

**关联需求：** Requirement 8

- [ ] 12.1 修改 `ui/src/views/report/Detail.vue`
  - 检测 `report.mode === 'video'` 条件，以下功能仅视频模式显示
  - 报告标题区新增「视频面试」`<el-tag>` 标签
  - 新增「表达能力分析」卡片：三列展示 avgExpressionScore（`ScoreCircle`）、avgSpeechRate（带颜色提示）、总停顿次数
  - 卡片底部展示 `report.expressionSummary` AI 综合评价文字
  - 修改逐题明细展示：每题新增表达得分 `ScoreBar`、expressionFeedback 文字、语速/停顿数据

- [ ] 12.2 修改 `ui/src/components/common/RadarChart.vue`
  - 当传入的 `modules` 数组中包含「表达能力」维度时，正常渲染（无需特殊处理，数据由父组件控制）
  - 报告页在视频模式下，向 RadarChart 传入包含「表达能力」维度的 modules 数组

- [ ] 12.3 视频下载入口（若有录制）
  - 若 Vuex 中存在 `recordedVideo`，在报告页显示「下载面试录制」按钮
  - 点击按钮调用 `downloadVideo()` 方法（来自 recordingMixin），并显示提示「视频仅存储于本地，刷新后将丢失」


### Task 13: 兼容性检测工具函数

**关联需求：** Requirement 13

- [ ] 13.1 创建 `ui/src/utils/mediaCompatibility.js`
  - 导出 `checkMediaSupport()` 函数，返回：
    ```js
    {
      getUserMedia: Boolean,      // navigator.mediaDevices.getUserMedia 是否可用
      speechRecognition: Boolean, // SpeechRecognition / webkitSpeechRecognition 是否可用
      mediaRecorder: Boolean,     // MediaRecorder 是否可用
      isMobile: Boolean           // 是否移动端（通过 userAgent 判断）
    }
    ```
  - 导出 `getRecommendedMimeType()` 函数，按优先级返回支持的录制格式

- [ ] 13.2 在 `VideoPreparation.vue` 中使用兼容性检测
  - mounted 时调用 `checkMediaSupport()`，将结果存入组件 data
  - 若 `isMobile === true`，直接显示「请在电脑端使用视频面试功能」提示并禁用「开始面试」按钮
  - 若 `getUserMedia === false`，显示「浏览器不支持，请使用 Chrome 88+」并禁用「开始面试」按钮
  - 若 `speechRecognition === false`，以橙色警告提示「语音识别不可用，将使用手动输入模式」


### Task 14: 面试暂停与恢复（视频模式适配）

**关联需求：** Requirement 9

- [ ] 14.1 修改 `VideoDoing.vue` 的暂停逻辑
  - 点击「暂停面试」时：先调用 `stopSpeech()` 和 `stopRecording()`，再调用 `RELEASE_MEDIA_STREAM`
  - 再调用现有 `PUT /api/v1/interviews/:id/pause` 接口保存进度
  - 跳转到仪表盘

- [ ] 14.2 适配历史页继续面试
  - 修改 `ui/src/views/interview/History.vue` 的「继续面试」逻辑
  - 获取面试会话的 `mode` 字段，若为 `'video'`，跳转到 `/interview/:id/video-prep`（重新检测设备），而非 `/interview/:id/doing`

- [ ] 14.3 页面刷新恢复提示
  - 在 `VideoDoing.vue` 的 `mounted` 钩子中，检测是否为「已暂停/进行中」的视频面试
  - 若是，显示提示「检测到未完成的视频面试，已从上次题目继续」


### Task 15: 前端 API 模块更新

**关联需求：** Requirement 12

- [ ] 15.1 修改 `ui/src/api/interview.js`
  - `createInterview(data)` 确保 `data.mode` 字段被传递
  - 修改 `submitAnswer(id, data)` 支持传递 `nonVerbalMetrics` 字段

### Task 16: 样式与 UI 细节

**关联需求：** Requirement 14

- [ ] 16.1 `VideoPreparation.vue` 样式
  - 摄像头预览区：圆角边框，宽高 16:9 比例，背景 `#000`
  - 兼容性状态：绿色 ✓ / 橙色 ⚠ / 红色 ✗ 图标配色

- [ ] 16.2 `VideoDoing.vue` 样式
  - 左栏占 40%，右栏占 60%，左侧摄像头画面镜像（`transform: scaleX(-1)`）
  - 录制状态指示：红色圆点闪烁动画（`@keyframes blink`）
  - 倒计时最后 10 秒：进度条变为红色（`:class="{ danger: timeLeft <= 10 }"`）
  - 临时识别文字：灰色斜体（`color: #999; font-style: italic`）


### Task 17: 集成测试与联调

**关联需求：** 全部

- [ ] 17.1 完整流程测试（Chrome）
  - 配置页选择视频模式 → 准备页授权 → 答题页语音转文字 → 提交答案（含 AI 表达点评）→ 报告查看
  - 验证 `nonVerbalMetrics` 数据正确传递至后端并存储到 Redis
  - 验证报告页「表达能力」模块正确展示

- [ ] 17.2 降级场景测试（Firefox）
  - 进入视频模式，确认语音识别不可用时正确显示降级提示
  - 手动输入答案后提交，后端不传 `nonVerbalMetrics`，验证无报错

- [ ] 17.3 权限拒绝测试
  - 拒绝摄像头/麦克风权限，确认正确显示错误提示且不能进入面试
  - 权限拒绝后返回配置页，重新选择文字模式，确认正常流程不受影响

- [ ] 17.4 资源释放验证
  - 完成面试后，确认浏览器摄像头指示灯熄灭（媒体流已释放）
  - 暂停面试后，确认媒体流已释放
  - 主动关闭标签页时，确认 `beforeunload` 提示正确触发

### Task 18: 文档更新

**关联需求：** -

- [ ] 18.1 更新 `README.md`
  - 在功能模块表格中新增「视频面试模式」条目，说明功能简介和浏览器要求
  - 新增「浏览器兼容性」章节，说明 Chrome 88+ 为推荐浏览器

---

## Notes

- **MVP 范围**：本版本实现摄像头预览 + 语音转文字 + 语速/停顿分析，**不包含**情绪识别和眼神追踪（作为未来扩展）
- **数据策略**：音视频数据不上传服务器，后端仅存储文字和统计数据，对现有 Redis 存储无显著压力
- **向后兼容**：所有后端改动均向后兼容，现有文字模式功能不受影响
- **浏览器依赖**：Web Speech API 仅 Chromium 系浏览器完整支持，Firefox 用户需手动输入

