# 需求文档

## 简介

本功能为面试模拟系统新增**视频面试模式**。现有系统已支持用户通过文字输入回答面试题目，本功能在此基础上增加摄像头视频捕获与语音识别能力，使用户能够通过口头回答面试题，更真实地模拟线下面试场景。

系统通过浏览器原生 API 调用摄像头与麦克风，实时录制用户的视频与音频，并将语音转换为文字后作为答案提交给 AI 进行评分与点评，整个流程与现有文字输入流程保持一致。

---

## 词汇表

- **视频面试模式（Video_Interview_Mode）**：用户通过摄像头和麦克风进行面试的交互模式，与文字输入模式并列。
- **摄像头模块（Camera_Module）**：负责通过浏览器 MediaDevices API 获取用户摄像头视频流的前端模块。
- **语音识别模块（Speech_Recognition_Module）**：负责将用户语音实时转换为文字的前端模块，基于浏览器 Web Speech API 实现。
- **转写文本（Transcription_Text）**：语音识别模块将用户语音内容实时转换后得到的文字结果。
- **面试作答区（Answer_Panel）**：面试进行页面中供用户输入或录制答案的区域。
- **答案提交（Answer_Submission）**：将用户答案（文字或转写文本）发送至后端进行 AI 评分的操作。
- **浏览器兼容性（Browser_Compatibility）**：系统运行所需的浏览器环境要求，当前要求支持 MediaDevices API 和 Web Speech API 的现代浏览器（Chrome 25+、Edge 79+、Firefox 44+）。

---

## 需求列表

### 需求 1：面试模式选择

**用户故事：** 作为一名求职用户，我希望在发起面试时能够选择答题方式，以便根据自己的当前环境决定使用视频语音模式还是文字输入模式。

#### 验收标准

1. THE 面试配置页（Config_Page）SHALL 在现有配置项下方展示答题方式选择区域，包含"文字输入"和"视频语音"两个选项。
2. THE 面试配置页（Config_Page）SHALL 默认选中"文字输入"模式。
3. WHEN 用户选中"视频语音"模式，THE 面试配置页（Config_Page）SHALL 展示摄像头与麦克风权限说明提示。
4. WHEN 用户点击"发起面试"按钮，THE 面试配置页（Config_Page）SHALL 将用户选择的答题方式（`text` 或 `video`）随面试配置一同传递至面试进行页。

---

### 需求 2：摄像头与麦克风权限申请

**用户故事：** 作为一名求职用户，我希望系统在进入视频面试前明确告知我需要开启摄像头和麦克风，以便我能提前了解并授权。

#### 验收标准

1. WHEN 用户以"视频语音"模式进入面试进行页，THE Camera_Module SHALL 通过浏览器 `getUserMedia` API 申请摄像头和麦克风访问权限。
2. WHEN 用户授权成功，THE Camera_Module SHALL 将摄像头画面实时显示在面试作答区的视频预览框中。
3. IF 用户拒绝摄像头或麦克风权限，THEN THE Camera_Module SHALL 展示权限被拒绝的错误提示，并提供引导说明告知用户如何在浏览器中重新开启权限。
4. IF 用户设备不存在可用摄像头或麦克风，THEN THE Camera_Module SHALL 展示设备不可用的错误提示，并引导用户检查设备连接。
5. IF 当前浏览器不支持 MediaDevices API，THEN THE Camera_Module SHALL 展示浏览器不兼容提示，并建议用户切换至支持的浏览器（Chrome 25+、Edge 79+）。

---

### 需求 3：视频预览展示

**用户故事：** 作为一名求职用户，我希望在回答问题时能看到自己的摄像头画面，以便调整仪态、模拟真实面试状态。

#### 验收标准

1. WHILE 视频面试进行中，THE Camera_Module SHALL 在面试作答区持续展示用户摄像头的实时视频画面（镜像翻转显示）。
2. THE Camera_Module SHALL 展示视频预览框的宽高比为 16:9，最小尺寸不低于 240px × 135px。
3. WHEN 用户切换到下一道题，THE Camera_Module SHALL 保持视频预览持续显示，不重置或中断视频流。
4. WHEN 面试结束或用户离开面试页面，THE Camera_Module SHALL 停止摄像头视频流并释放设备资源。

---

### 需求 4：语音录制与实时转写

**用户故事：** 作为一名求职用户，我希望通过说话来回答面试题，系统能实时将我说的话转换为文字，以便我确认识别结果是否准确。

#### 验收标准

1. WHEN 用户点击"开始录音"按钮，THE Speech_Recognition_Module SHALL 启动语音识别，并实时将识别到的语音内容显示在转写文本区域。
2. WHILE 语音识别进行中，THE Speech_Recognition_Module SHALL 每隔不超过 500ms 更新一次转写文本区域中的内容，反映最新识别结果。
3. WHEN 用户点击"停止录音"按钮，THE Speech_Recognition_Module SHALL 停止语音识别并将最终转写结果定格在转写文本区域。
4. WHILE 语音识别进行中，THE 面试作答区（Answer_Panel）SHALL 展示明显的录音状态指示（如动态波形图标或红色录音标志）。
5. IF 语音识别过程中发生识别服务错误，THEN THE Speech_Recognition_Module SHALL 展示错误提示，并保留当前已转写的文本内容不丢失。
6. IF 当前浏览器不支持 Web Speech API，THEN THE Speech_Recognition_Module SHALL 展示不兼容提示，并建议用户使用 Chrome 或 Edge 浏览器。

---

### 需求 5：转写文本编辑

**用户故事：** 作为一名求职用户，我希望在语音识别结束后能够对转写的文字进行手动修改，以便纠正识别错误或补充内容。

#### 验收标准

1. WHEN 语音识别停止后，THE 面试作答区（Answer_Panel）SHALL 将转写文本区域变为可编辑状态，允许用户手动修改文字内容。
2. THE 面试作答区（Answer_Panel）SHALL 支持用户在已有转写文本基础上追加录音：WHEN 用户再次点击"开始录音"，THE Speech_Recognition_Module SHALL 将新的转写内容追加至现有文本末尾，而非覆盖。
3. THE 面试作答区（Answer_Panel）SHALL 提供"清空"按钮，WHEN 用户点击"清空"按钮，THE 面试作答区（Answer_Panel）SHALL 清除全部转写文本内容。

---

### 需求 6：以转写文本作为答案提交

**用户故事：** 作为一名求职用户，我希望确认转写文本无误后，能够将其作为答案提交给 AI 评分，流程与文字输入模式保持一致。

#### 验收标准

1. WHEN 用户点击"提交答案"按钮，THE Answer_Submission SHALL 使用转写文本区域中的当前内容作为答案，调用现有的答案提交接口（`POST /api/interviews/:id/answer`）。
2. IF 转写文本为空，THEN THE Answer_Submission SHALL 禁用"提交答案"按钮，并展示提示文字"请先录音或输入答案"。
3. WHEN 答案提交成功，THE 面试作答区（Answer_Panel）SHALL 展示与文字输入模式相同的 AI 评分与点评结果。
4. THE Answer_Submission SHALL 保持现有答案提交接口的请求参数结构（`questionIndex` 和 `answer` 字段）不变，将转写文本填入 `answer` 字段。

---

### 需求 7：视频面试模式与文字模式的视觉区分

**用户故事：** 作为一名求职用户，我希望在视频面试模式下，界面能清晰地与文字输入模式区分，让我明确知道当前处于哪种模式。

#### 验收标准

1. WHILE 视频面试模式进行中，THE 面试进行页（Doing_Page）SHALL 在页面顶部展示"视频面试模式"标识徽章。
2. WHILE 视频面试模式进行中，THE 面试进行页（Doing_Page）SHALL 隐藏文字输入文本框，仅展示视频预览区、录音控制区和转写文本区。
3. WHILE 文字输入模式进行中，THE 面试进行页（Doing_Page）SHALL 保持现有文字输入界面不变，不展示视频相关组件。

---

### 需求 8：面试会话中记录答题模式

**用户故事：** 作为系统管理员，我希望面试会话数据中能记录用户本次面试采用的答题模式，便于后续数据分析与功能迭代。

#### 验收标准

1. WHEN 创建面试会话时，THE 面试服务（Interview_Service）SHALL 在面试配置信息中记录 `answerMode` 字段，取值为 `text`（文字输入）或 `video`（视频语音）。
2. THE 面试服务（Interview_Service）SHALL 保持现有创建面试接口（`POST /api/interviews`）的兼容性：IF 请求中不包含 `answerMode` 字段，THEN THE 面试服务（Interview_Service）SHALL 默认将 `answerMode` 设置为 `text`。
