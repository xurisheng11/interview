# Requirements

## Introduction

本需求文档定义了**视频面试模式**功能的详细需求。该功能是在现有「面试模拟系统」基础上新增的模拟面试方式，允许用户通过摄像头和麦克风进行真实的视频面试模拟，系统将实时或录制用户的音视频表现，并结合AI技术对答案内容、语言表达、非语言行为等多维度进行评估。

### Background

现有系统提供文字问答式面试模拟（用户输入文字答案 → AI点评内容）。但真实面试场景中，除了答案内容，面试官还会关注应聘者的语言表达能力、情绪管理、眼神交流、肢体语言等非语言要素。视频面试模式旨在模拟真实面试场景，帮助用户全方位提升面试表现。

### Goals

1. 提供基于摄像头+麦克风的视频面试模拟体验
2. 实时语音转文字，自动记录用户回答内容
3. （可选）录制面试视频，供用户回放和AI分析
4. AI多维度评估：内容质量、语速、语调、情绪、眼神接触、肢体语言
5. 生成包含视频表现维度的面试报告

### Glossary

| 术语 | 定义 |
|------|------|
| 视频面试模式 | 通过摄像头+麦克克风进行的模拟面试方式，区别于现有的文字输入模式 |
| 语音转文字（STT） | Speech-to-Text，将用户语音实时转换为文字答案 |
| 非语言要素 | 包括语速、语调、情绪、眼神接触、肢体语言等非内容性的表现指标 |
| 视频流 | 实时采集的摄像头画面数据流 |
| 音频流 | 实时采集的麦克风音频数据流 |


## Requirements

### Requirement 1: 视频面试模式入口

**User Story:** 作为用户，我希望在面试配置页能选择「视频面试模式」，以便通过摄像头和麦克风进行更真实的面试模拟。

#### Acceptance Criteria

1. WHERE 用户访问面试配置页 (`/interview/config`)，THE 页面 SHALL 在「答题方式」区域新增「视频面试模式」选项（Radio 单选）
2. WHEN 用户选择「视频面试模式」，THE 系统 SHALL 显示设备权限提示（需要摄像头和麦克风权限）
3. WHEN 用户点击「发起面试」且选择了视频模式，THE 系统 SHALL 先检测浏览器是否支持 `navigator.mediaDevices.getUserMedia` API
4. IF 浏览器不支持或用户拒绝权限，THE 系统 SHALL 显示错误提示并禁止进入视频面试
5. IF 权限检测通过，THE 系统 SHALL 创建视频面试会话并跳转到视频答题页


### Requirement 2: 设备检测与权限授权

**User Story:** 作为用户，我希望系统在开始视频面试前帮我检测摄像头和麦克风是否正常工作，以便在正式开始前排除设备问题。

#### Acceptance Criteria

1. WHEN 进入视频面试准备页，THE 系统 SHALL 自动请求摄像头和麦克风权限
2. WHILE 等待用户授权，THE 页面 SHALL 显示「正在请求设备权限...」提示
3. IF 用户授予权限，THE 系统 SHALL 展示摄像头实时预览画面（镜像显示）
4. IF 用户拒绝摄像头权限，THE 系统 SHALL 提示「摄像头权限被拒绝，视频面试需要摄像头权限，请在浏览器设置中允许」
5. IF 用户拒绝麦克风权限，THE 系统 SHALL 提示「麦克风权限被拒绝，请允许麦克风权限以便录制语音答案」
6. THE 系统 SHALL 提供「音频测试」功能，用户说话时显示实时音量波形指示器
7. WHEN 设备检测完成且无异常，THE 页面 SHALL 显示「开始面试」按钮


### Requirement 3: 视频面试答题界面

**User Story:** 作为用户，我希望视频面试答题页能展示我的摄像头画面和题目内容，同时支持语音作答，以便在接近真实面试的环境中练习。

#### Acceptance Criteria

1. THE 视频答题页 SHALL 分为左侧摄像头预览区和右侧题目+答案区两栏布局
2. WHILE 面试进行中，THE 左侧 SHALL 持续显示摄像头实时画面（镜像模式，类似视频通话体验）
3. THE 左侧 SHALL 显示实时录制状态指示器（红点闪烁 + 计时器）
4. THE 右侧 SHALL 显示当前题目信息（题号/难度/标签/题目内容），与现有文字模式保持一致
5. THE 右侧 SHALL 提供「语音作答」区域，包含实时语音转文字结果（Textarea 可编辑）
6. WHEN 用户开始说话，THE 系统 SHALL 通过 Web Speech API 实时转写语音为文字并填入 Textarea
7. THE 系统 SHALL 保留「手动编辑」能力，用户可对自动转写的文字进行修改
8. THE 页面 SHALL 显示每题倒计时（与现有模式相同规则：estimatedMinutes × 2 分钟）
9. THE 页面 SHALL 提供「提交答案」「跳过此题」「暂停面试」操作按钮，行为与现有模式一致


### Requirement 4: 语音转文字（STT）

**User Story:** 作为用户，我希望系统能实时将我的语音转换为文字，自动填充到答案框中，这样我就不需要手动打字，可以专注于口头表达。

#### Acceptance Criteria

1. THE 系统 SHALL 使用浏览器内置 Web Speech API (`SpeechRecognition`) 实现实时语音转文字
2. WHEN 开始作答一道题，THE 系统 SHALL 自动启动语音识别（语言设置为中文 `zh-CN`，可切换英文 `en-US`）
3. WHILE 语音识别进行中，THE 系统 SHALL 以「临时结果」（interim results）形式实时显示正在识别的内容（灰色斜体）
4. WHEN 一句话识别完成，THE 系统 SHALL 将「最终结果」追加到 Textarea 中（正常字体）
5. IF Web Speech API 不被浏览器支持（如 Firefox），THE 系统 SHALL 提示「当前浏览器不支持语音识别，请使用 Chrome 或 Edge，或手动输入答案」并降级为纯文字模式
6. THE 系统 SHALL 提供「麦克风开关」按钮，允许用户暂停/恢复语音识别
7. WHEN 倒计时结束或用户点击「提交」，THE 系统 SHALL 停止语音识别并以 Textarea 当前内容作为最终答案


### Requirement 5: 视频录制（可选）

**User Story:** 作为用户，我希望可以选择是否录制我的视频面试，以便事后回看自己的表现并持续改进。

#### Acceptance Criteria

1. WHEN 用户在准备页开启「录制面试视频」开关（默认关闭），THE 系统 SHALL 在面试开始时通过 MediaRecorder API 录制摄像头+麦克风合流
2. WHILE 录制进行中，THE 系统 SHALL 将视频分块（每30秒一个 Blob chunk）存储在浏览器内存中
3. WHEN 面试完成，IF 用户开启了录制，THE 系统 SHALL 将所有 chunks 合并为完整的 WebM/MP4 视频文件
4. THE 系统 SHALL 提供「下载视频」按钮，允许用户将录制视频下载到本地
5. THE 系统 SHALL 明确提示：「视频仅存储在本地，不会上传至服务器，关闭页面后视频将丢失」
6. IF 录制过程中发生错误（如内存不足），THE 系统 SHALL 提示「录制中断：存储空间不足」并停止录制，但面试可继续进行
7. THE 录制功能 SHALL 为可选项，不影响面试的核心流程


### Requirement 6: AI答案点评（视频模式增强）

**User Story:** 作为用户，我希望提交答案后，AI不仅点评答案内容，还能结合语音数据分析我的口头表达能力，给出更全面的反馈。

#### Acceptance Criteria

1. WHEN 用户在视频模式下提交答案，THE 系统 SHALL 将 Textarea 中的文字答案发送至后端（与现有模式相同接口）
2. THE 后端 SHALL 在调用 DeepSeek 点评时，附加「视频面试模式」的 prompt 上下文，要求 AI 评估「口语化表达质量」
3. THE AI点评结果 SHALL 在现有维度（score/pros/cons/referenceAnswer）基础上，新增：
   - `expressionScore`（表达得分，0-100）：评估答案的口语化流畅程度
   - `expressionFeedback`（表达反馈）：针对口语表达的具体建议
4. THE 前端 SHALL 在 AI 点评区域新增「表达能力」反馈卡片，展示 expressionScore 和 expressionFeedback
5. IF 答案文字过短（少于20字），THE 系统 SHALL 提示「答案较短，建议语音作答时尽量展开说明」
6. THE 内容评分与表达评分 SHALL 分别显示，内容得分用于计算面试总分，表达得分仅作参考


### Requirement 7: 非语言行为分析（MVP阶段简化实现）

**User Story:** 作为用户，我希望系统能分析我的非语言行为（如语速、停顿、眼神），并给出改进建议，帮助我提升整体面试表现。

#### Acceptance Criteria（MVP阶段）

1. THE 系统 SHALL 在前端通过 Web Speech API 的 `SpeechRecognitionResult` 获取语音片段时间戳，计算：
   - `speechRate`（语速）：每分钟字数（WPM，Words Per Minute）
   - `pauseCount`（停顿次数）：大于2秒的静默次数
2. WHEN 每题作答完成，THE 前端 SHALL 将 `{ speechRate, pauseCount, duration }` 作为 metadata 与答案一起提交后端
3. THE 后端 SHALL 在 AnswerRecord 中新增 `nonVerbalMetrics` 字段（JSON 格式）存储上述指标
4. THE 后端 SHALL 在 AI 点评时将这些指标以 prompt 形式告知 DeepSeek，要求 AI 评估：
   - 语速是否合适（建议范围 120-150 WPM）
   - 是否有过多停顿或语气词
5. THE AI 返回的 expressionFeedback SHALL 包含对语速和停顿的具体反馈
6. THE MVP阶段 SHALL NOT 实现视频流分析（情绪识别、眼神追踪），这些功能作为未来扩展
7. THE 系统 SHALL 在面试报告中展示每题的语速统计和平均停顿时长


### Requirement 8: 视频面试报告

**User Story:** 作为用户，我希望视频面试完成后能生成包含表达能力维度的面试报告，以便全面了解自己的表现。

#### Acceptance Criteria

1. WHEN 用户完成视频面试，THE 系统 SHALL 生成面试报告，包含现有所有维度（内容得分/模块分/AI综合评价）
2. THE 报告 SHALL 新增「表达能力」模块，展示：
   - 平均表达得分（所有题目 expressionScore 的平均值）
   - 平均语速（WPM）及与推荐范围的对比
   - 总停顿次数及平均每题停顿次数
   - AI 对整体口头表达的综合评价
3. THE 报告 SHALL 在「逐题明细」中为每道题展示：
   - 内容得分
   - 表达得分
   - 语速和停顿数据
   - AI 表达反馈
4. THE 报告雷达图 SHALL 新增「表达能力」维度（与「算法能力」「系统设计」等并列）
5. THE 报告 SHALL 标注「视频面试模式」标识，与普通文字模式报告区分
6. IF 用户录制了视频，THE 报告 SHALL 提示「视频已保存在本地，可通过浏览器下载记录查看」
7. THE 分享报告功能 SHALL 支持视频模式报告，但不包含录制视频（仅分享数据和评价）


### Requirement 9: 面试暂停与恢复（视频模式）

**User Story:** 作为用户，我希望视频面试可以暂停和恢复，这样在网络不稳定或需要临时中断时不会丢失进度。

#### Acceptance Criteria

1. WHEN 用户在视频面试中点击「暂停面试」，THE 系统 SHALL 停止摄像头预览和语音识别，关闭媒体流
2. WHEN 暂停时，THE 系统 SHALL 通过现有 `PUT /api/v1/interviews/:id/pause` 接口保存进度到 Redis（7天TTL）
3. WHEN 用户从仪表盘「继续面试」进入一个暂停的视频面试，THE 系统 SHALL 重新请求摄像头/麦克风权限并恢复到暂停时的题目
4. IF 用户在视频面试中途刷新页面，THE 系统 SHALL 检测到未完成的视频面试会话，提示「检测到未完成的视频面试，是否继续？」
5. WHEN 继续面试时，THE 系统 SHALL 从上次作答的题目开始（已提交的答案不重复作答）


### Requirement 10: 隐私与数据安全

**User Story:** 作为用户，我希望系统明确告知我的视频和音频数据如何处理，确保隐私不被泄露。

#### Acceptance Criteria

1. THE 系统 SHALL 在首次启动视频面试前展示隐私政策弹窗，说明：
   - 摄像头和麦克风仅用于本地模拟面试，不会上传至服务器
   - 语音转文字通过浏览器内置 API 完成，在本地处理
   - 视频录制（如开启）仅存储在浏览器内存，不上传服务器
   - 后端仅接收文字答案和统计数据（语速/停顿），不接收音视频原始数据
2. THE 用户 SHALL 必须点击「同意并开始」才能进入视频面试，否则返回配置页
3. THE 系统 SHALL 在视频录制开始前再次确认「是否录制视频？录制将消耗存储空间」
4. THE 系统 SHALL 在页面关闭或刷新前检测到正在进行的视频流，弹出确认「确定离开？媒体流将被关闭」
5. WHEN 面试完成或暂停，THE 系统 SHALL 立即释放摄像头和麦克风资源，浏览器指示灯熄灭
6. THE 后端 SHALL NOT 存储任何音视频二进制数据，仅存储文字答案和统计指标（JSON 格式）


### Requirement 11: 后端数据模型扩展

**User Story:** 作为开发者，我需要扩展现有面试数据模型以支持视频面试模式的额外数据字段。

#### Acceptance Criteria

1. THE `InterviewSession` 模型 SHALL 新增 `mode` 字段（值：`"text"` 或 `"video"`，默认 `"text"`）
2. THE `AnswerRecord` 模型 SHALL 新增以下可选字段：
   - `expressionScore int`（表达得分，0-100，视频模式专有）
   - `expressionFeedback string`（表达反馈，视频模式专有）
   - `nonVerbalMetrics struct`：包含 `speechRate float64`、`pauseCount int`、`duration int`（秒）
3. THE `InterviewReport` 模型 SHALL 新增：
   - `expressionSummary string`（整体表达能力 AI 评价）
   - `avgExpressionScore int`（平均表达得分）
   - `avgSpeechRate float64`（平均语速 WPM）
4. THE DeepSeek Prompt SHALL 在视频模式下附加非语言指标上下文，格式为：`"用户语速为X字/分钟，停顿N次，请结合以下回答内容给出表达能力评估"`
5. THE 现有文字面试模式数据 SHALL NOT 受到影响，所有新字段对文字模式为 omitempty（JSON序列化时忽略空值）


### Requirement 12: API接口扩展

**User Story:** 作为开发者，我需要扩展现有 API 接口以支持视频面试模式的附加数据传递。

#### Acceptance Criteria

1. `POST /api/v1/interviews` 接口 SHALL 在请求体中新增 `mode` 字段（`"text"` / `"video"`），后端据此设置会话模式
2. `POST /api/v1/interviews/:id/answers` 接口 SHALL 在请求体中新增可选字段：
   - `nonVerbalMetrics: { speechRate, pauseCount, duration }`
3. `GET /api/v1/reports/:interviewId` 接口响应 SHALL 在视频模式报告中包含 `expressionSummary`、`avgExpressionScore`、`avgSpeechRate`
4. 所有扩展字段 SHALL 向后兼容：文字模式下这些字段为 null 或不存在，不影响现有功能
5. THE 现有 `PUT /api/v1/interviews/:id/pause` 和 `PUT /api/v1/interviews/:id/complete` 接口 SHALL 无需修改，行为一致


### Requirement 13: 兼容性与降级处理

**User Story:** 作为用户，我希望即使我的浏览器不完全支持某些视频功能，系统也能给出清晰提示并提供可用的替代方案。

#### Acceptance Criteria

1. IF 浏览器不支持 `getUserMedia`，THE 系统 SHALL 禁止进入视频模式并提示「当前浏览器不支持视频面试，请使用 Chrome 88+ 或 Edge 88+」
2. IF 浏览器支持 `getUserMedia` 但不支持 `SpeechRecognition`（如 Firefox），THE 系统 SHALL 允许进入视频模式但禁用语音转文字，提示「当前浏览器不支持语音识别，请手动输入答案」
3. IF `MediaRecorder` 不支持 WebM 格式，THE 系统 SHALL 尝试 MP4 格式，若都不支持则禁用录制功能但不阻塞面试
4. THE 系统 SHALL 在视频面试准备页展示「浏览器兼容性检查」结果，用绿色/橙色/红色图标分别标注三项功能（摄像头、语音识别、视频录制）的支持状态
5. THE 系统 SHALL 推荐用户使用 Chrome 最新版以获得最佳体验


### Requirement 14: 性能与用户体验

**User Story:** 作为用户，我希望视频面试过程流畅不卡顿，并且操作直观易懂。

#### Acceptance Criteria

1. THE 摄像头预览 SHALL 使用 CSS `transform: scaleX(-1)` 实现镜像效果，避免 JS 实时处理视频流
2. THE 语音转文字 SHALL 使用 `continuous: true` 和 `interimResults: true` 保证实时流畅的转写体验
3. THE 视频录制 SHALL 设置合理的 bitrate（视频 2.5Mbps，音频 128kbps）避免文件过大
4. THE 系统 SHALL 在倒计时最后10秒显示醒目提示（进度条变红色 + 闪烁），提醒用户即将超时
5. THE 系统 SHALL 在每题提交后自动停止语音识别，进入下一题时重新启动，避免误识别
6. THE 页面 SHALL 在视频流初始化成功后移除 loading 遮罩，避免用户等待焦虑
7. THE 系统 SHALL 在移动设备上禁用视频面试模式（小屏幕体验差），提示「请在电脑端使用视频面试功能」

