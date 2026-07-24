# Design

## Overview

视频面试模式在现有「面试模拟系统」基础上扩展，复用所有现有接口和数据层，通过以下三个核心技术实现：

1. **浏览器 WebRTC API**（`getUserMedia`）：采集摄像头+麦克风
2. **Web Speech API**（`SpeechRecognition`）：实时语音转文字，纯前端，无需后端
3. **MediaRecorder API**：可选本地视频录制，不上传服务器

后端改动最小化：仅扩展数据模型和 AI Prompt，不新增独立服务。

---

## Architecture

```
┌─────────────────────────────────────────────────────┐
│                    前端（Vue2）                       │
│                                                     │
│  ┌──────────────────┐   ┌────────────────────────┐  │
│  │  VideoInterview  │   │  现有 interview/ 页面   │  │
│  │  Doing.vue (新)  │   │  Config.vue (扩展)     │  │
│  └────────┬─────────┘   └────────────────────────┘  │
│           │                                         │
│  ┌────────▼─────────────────────────────────────┐   │
│  │         VideoInterviewMixin.js               │   │
│  │  getUserMedia │ SpeechRecognition │ MediaRec  │   │
│  └────────┬─────────────────────────────────────┘   │
│           │ HTTP (复用现有 /api/v1/interviews)        │
└───────────┼─────────────────────────────────────────┘
            │
┌───────────▼─────────────────────────────────────────┐
│                   后端（Go + Gin）                    │
│                                                     │
│  interview_service.go  ←── 扩展 mode 字段            │
│  deepseek_service.go   ←── 扩展视频模式 Prompt        │
│  model/interview.go    ←── 扩展数据结构               │
│  model/report.go       ←── 扩展报告结构               │
└─────────────────────────────────────────────────────┘
```


## Frontend Components

### 1. 新增页面组件

#### `ui/src/views/interview/VideoPreparation.vue`
视频面试准备页（设备检测与授权）

**核心功能：**
- 请求摄像头+麦克风权限（`navigator.mediaDevices.getUserMedia`）
- 展示摄像头实时预览（`<video>` 元素，`srcObject` 绑定 `MediaStream`）
- 音频测试：实时显示音量波形（通过 `AudioContext.createAnalyser()`）
- 浏览器兼容性检查（检测 `getUserMedia`、`SpeechRecognition`、`MediaRecorder` 支持情况）
- 隐私政策弹窗确认

**状态管理：**
```javascript
data() {
  return {
    stream: null,          // MediaStream 对象
    videoReady: false,     // 摄像头是否就绪
    audioReady: false,     // 麦克风是否就绪
    permissions: {
      camera: 'pending',   // 'pending' | 'granted' | 'denied'
      microphone: 'pending'
    },
    compatibility: {
      getUserMedia: false,
      speechRecognition: false,
      mediaRecorder: false
    }
  }
}
```

**关键方法：**
- `checkCompatibility()`: 检测浏览器 API 支持
- `requestPermissions()`: 调用 `getUserMedia({ video: true, audio: true })`
- `startPreview()`: 将 stream 绑定到 `<video>` 元素
- `testAudio()`: 使用 `AudioContext` 分析音频振幅
- `startInterview()`: 跳转到 `/interview/:id/video-doing`，将 stream 传递至 Vuex

---

#### `ui/src/views/interview/VideoDoing.vue`
视频面试答题页

**布局结构：**
```vue
<template>
  <div class="video-interview">
    <!-- 左侧摄像头区 -->
    <div class="camera-panel">
      <video ref="cameraPreview" autoplay muted></video>
      <div class="recording-indicator" v-if="isRecording">
        <span class="dot"></span> 录制中 {{ recordingDuration }}
      </div>
    </div>

    <!-- 右侧答题区 -->
    <div class="question-panel">
      <div class="progress">{{ currentIndex + 1 }} / {{ questions.length }}</div>
      <div class="timer">{{ formatTime(timeLeft) }}</div>
      
      <div class="question-card">
        <h3>{{ currentQuestion.content }}</h3>
        <div class="tags">{{ currentQuestion.tags }}</div>
      </div>

      <div class="answer-area">
        <textarea v-model="userAnswer" placeholder="语音识别内容将自动填入..."></textarea>
        <div class="speech-status">
          <el-switch v-model="isSpeechActive" @change="toggleSpeech">麦克风</el-switch>
          <span class="interim-text">{{ interimTranscript }}</span>
        </div>
      </div>

      <div class="actions">
        <el-button @click="submitAnswer">提交答案</el-button>
        <el-button @click="skipQuestion">跳过此题</el-button>
        <el-button @click="pauseInterview">暂停面试</el-button>
      </div>

      <!-- AI 点评区（提交后显示）-->
      <div class="review-result" v-if="currentReview">
        <el-card>
          <div class="score">内容得分：{{ currentReview.score }}</div>
          <div class="score">表达得分：{{ currentReview.expressionScore }}</div>
          <p>{{ currentReview.expressionFeedback }}</p>
        </el-card>
      </div>
    </div>
  </div>
</template>
```

**核心功能：**
- 持续显示摄像头预览（从 Vuex 获取 stream）
- 语音识别实时转文字（`SpeechRecognition`，`continuous: true`）
- 非语言指标计算（语速、停顿次数）
- 可选视频录制（`MediaRecorder`）
- 倒计时与进度条
- 提交答案调用现有 `POST /api/v1/interviews/:id/answers`，附加 `nonVerbalMetrics`

**状态管理：**
```javascript
data() {
  return {
    stream: null,                // 从 Vuex 获取
    recognition: null,           // SpeechRecognition 实例
    recorder: null,              // MediaRecorder 实例
    chunks: [],                  // 录制的视频片段
    isSpeechActive: true,        // 语音识别开关
    interimTranscript: '',       // 临时识别结果
    userAnswer: '',              // 最终答案文字
    speechMetrics: {
      startTime: 0,
      wordCount: 0,
      pauseCount: 0,
      lastSpeechTime: 0
    }
  }
}
```

**关键方法：**
- `initSpeechRecognition()`: 初始化 `SpeechRecognition`，监听 `onresult` 事件
- `calculateMetrics()`: 根据识别时间戳计算语速（WPM）和停顿次数
- `startRecording()`: 启动 `MediaRecorder`，设置 `ondataavailable`
- `stopRecording()`: 停止录制，合并 chunks 为 Blob，提供下载
- `submitAnswer()`: 发送答案+非语言指标至后端
- `cleanup()`: 释放媒体流资源（`stream.getTracks().forEach(t => t.stop())`）


### 2. 扩展现有组件

#### `ui/src/views/interview/Config.vue` — 扩展

新增「答题方式」区域：

```vue
<el-form-item label="答题方式">
  <el-radio-group v-model="config.mode">
    <el-radio label="text">文字输入</el-radio>
    <el-radio label="video">视频面试</el-radio>
  </el-radio-group>
  <div v-if="config.mode === 'video'" class="video-hint">
    <i class="el-icon-video-camera"></i>
    需要摄像头和麦克风权限，建议使用 Chrome 浏览器
  </div>
</el-form-item>
```

逻辑变更：
- `handleStart()` 方法：若 `config.mode === 'video'`，创建面试会话后跳转到 `/interview/:id/video-prep`（视频准备页）而非 `/interview/loading`

---

### 3. 新增可复用组件

#### `ui/src/components/interview/VideoPreview.vue`
独立的摄像头预览组件，供准备页和答题页复用：

```vue
<template>
  <div class="video-preview-wrapper">
    <video ref="videoEl" autoplay muted playsinline
           :class="{ mirrored: mirror }"></video>
    <div class="no-camera" v-if="!active">摄像头未启动</div>
  </div>
</template>

<script>
export default {
  props: {
    stream: MediaStream,   // 媒体流
    mirror: { type: Boolean, default: true }  // 是否镜像
  }
}
</script>

<style scoped>
.mirrored { transform: scaleX(-1); }
</style>
```

#### `ui/src/components/interview/SpeechIndicator.vue`
语音识别状态指示组件（音量波形 + 状态文字）：

```vue
<template>
  <div class="speech-indicator">
    <canvas ref="waveCanvas" width="120" height="40"></canvas>
    <span class="status-text">{{ statusText }}</span>
  </div>
</template>
```

使用 `AudioContext.createAnalyser()` 绘制实时音量波形。

---

### 4. Vuex Store 扩展

在 `ui/src/store/modules/interview.js` 新增状态：

```javascript
state: {
  // ... 现有状态
  mediaStream: null,       // 全局媒体流，避免重复请求权限
  interviewMode: 'text',   // 'text' | 'video'
  enableRecording: false   // 是否开启录制
},
mutations: {
  SET_MEDIA_STREAM(state, stream) { state.mediaStream = stream },
  SET_INTERVIEW_MODE(state, mode) { state.interviewMode = mode },
  RELEASE_MEDIA_STREAM(state) {
    if (state.mediaStream) {
      state.mediaStream.getTracks().forEach(t => t.stop())
      state.mediaStream = null
    }
  }
}
```


## Speech Recognition Design

### Web Speech API 集成方案

```javascript
// VideoInterviewMixin.js - 可复用的语音识别逻辑
export const speechMixin = {
  data() {
    return {
      recognition: null,
      isSpeechActive: false,
      finalTranscript: '',
      interimTranscript: '',
      speechMetrics: {
        startTime: null,
        totalWords: 0,
        pauseThreshold: 2000,   // 2秒静默视为停顿
        pauseCount: 0,
        lastResultTime: null
      }
    }
  },
  methods: {
    initSpeechRecognition(lang = 'zh-CN') {
      const SpeechRecognition = window.SpeechRecognition 
        || window.webkitSpeechRecognition
      if (!SpeechRecognition) return false
      
      this.recognition = new SpeechRecognition()
      this.recognition.lang = lang
      this.recognition.continuous = true
      this.recognition.interimResults = true
      this.recognition.maxAlternatives = 1

      this.recognition.onresult = (event) => {
        const now = Date.now()
        // 检测停顿
        if (this.speechMetrics.lastResultTime) {
          const gap = now - this.speechMetrics.lastResultTime
          if (gap > this.speechMetrics.pauseThreshold) {
            this.speechMetrics.pauseCount++
          }
        }
        this.speechMetrics.lastResultTime = now

        let interim = ''
        let final = ''
        for (let i = event.resultIndex; i < event.results.length; i++) {
          const text = event.results[i][0].transcript
          if (event.results[i].isFinal) {
            final += text
            this.speechMetrics.totalWords += text.length
          } else {
            interim += text
          }
        }
        if (final) this.finalTranscript += final
        this.interimTranscript = interim
        this.userAnswer = this.finalTranscript
      }

      this.recognition.onerror = (event) => {
        if (event.error !== 'no-speech') {
          this.$message.warning(`语音识别错误：${event.error}`)
        }
      }
      return true
    },

    startSpeech() {
      if (!this.recognition) return
      this.finalTranscript = ''
      this.interimTranscript = ''
      this.speechMetrics.startTime = Date.now()
      this.speechMetrics.totalWords = 0
      this.speechMetrics.pauseCount = 0
      this.recognition.start()
      this.isSpeechActive = true
    },

    stopSpeech() {
      if (this.recognition) this.recognition.stop()
      this.isSpeechActive = false
    },

    // 计算语速（每分钟字数）
    calcSpeechRate() {
      const duration = (Date.now() - this.speechMetrics.startTime) / 60000
      return duration > 0 
        ? Math.round(this.speechMetrics.totalWords / duration) 
        : 0
    },

    getNonVerbalMetrics() {
      const duration = Math.round(
        (Date.now() - this.speechMetrics.startTime) / 1000
      )
      return {
        speechRate: this.calcSpeechRate(),
        pauseCount: this.speechMetrics.pauseCount,
        duration
      }
    }
  }
}
```

### 语言切换

准备页提供语言选择：
```vue
<el-select v-model="speechLang">
  <el-option label="普通话（中文）" value="zh-CN" />
  <el-option label="English" value="en-US" />
</el-select>
```


## Video Recording Design

### MediaRecorder 本地录制方案

```javascript
// VideoRecordingMixin.js
export const recordingMixin = {
  data() {
    return {
      recorder: null,
      recordedChunks: [],
      isRecording: false,
      recordingStartTime: null
    }
  },
  methods: {
    startRecording(stream) {
      if (!window.MediaRecorder) {
        this.$message.warning('当前浏览器不支持视频录制')
        return false
      }

      const options = { mimeType: 'video/webm;codecs=vp9' }
      // 降级尝试
      if (!MediaRecorder.isTypeSupported(options.mimeType)) {
        options.mimeType = 'video/webm'
        if (!MediaRecorder.isTypeSupported(options.mimeType)) {
          options.mimeType = 'video/mp4'
        }
      }

      try {
        this.recorder = new MediaRecorder(stream, options)
        this.recordedChunks = []
        
        this.recorder.ondataavailable = (event) => {
          if (event.data.size > 0) {
            this.recordedChunks.push(event.data)
          }
        }

        this.recorder.onstop = () => {
          this.generateVideoFile()
        }

        this.recorder.start(30000)  // 每30秒一个chunk
        this.isRecording = true
        this.recordingStartTime = Date.now()
        return true
      } catch (err) {
        this.$message.error('录制启动失败：' + err.message)
        return false
      }
    },

    stopRecording() {
      if (this.recorder && this.isRecording) {
        this.recorder.stop()
        this.isRecording = false
      }
    },

    generateVideoFile() {
      const blob = new Blob(this.recordedChunks, { type: 'video/webm' })
      const url = URL.createObjectURL(blob)
      const duration = Math.round((Date.now() - this.recordingStartTime) / 1000)
      
      // 存储到 Vuex，供报告页下载
      this.$store.commit('interview/SET_RECORDED_VIDEO', {
        url,
        blob,
        size: (blob.size / 1024 / 1024).toFixed(2) + ' MB',
        duration
      })
      
      this.$message.success(`视频录制完成，时长 ${duration}s，已可下载`)
    },

    downloadVideo() {
      const video = this.$store.state.interview.recordedVideo
      if (!video) return

      const a = document.createElement('a')
      a.href = video.url
      a.download = `interview-${Date.now()}.webm`
      a.click()
      URL.revokeObjectURL(video.url)
    }
  }
}
```

### 存储策略

- 录制文件**仅存储在浏览器内存**（Blob URL）
- 页面刷新或关闭后视频丢失（通过 `beforeunload` 提示用户）
- **不上传至服务器**（避免存储成本和隐私问题）
- 提供下载按钮，用户可保存至本地


## Backend Data Model Extensions

### `api/model/interview.go` 扩展

```go
// InterviewSession 新增字段
type InterviewSession struct {
    // ... 现有字段 ...
    Mode string `json:"mode"` // "text" | "video"，默认 "text"
}

// AnswerRecord 新增字段
type AnswerRecord struct {
    // ... 现有字段 ...
    
    // 视频面试专有字段（omitempty，文字模式不写入）
    ExpressionScore    int              `json:"expressionScore,omitempty"`
    ExpressionFeedback string           `json:"expressionFeedback,omitempty"`
    NonVerbalMetrics   *NonVerbalMetrics `json:"nonVerbalMetrics,omitempty"`
}

// NonVerbalMetrics 非语言行为指标
type NonVerbalMetrics struct {
    SpeechRate float64 `json:"speechRate"` // 每分钟字数（WPM）
    PauseCount int     `json:"pauseCount"` // 停顿次数（>2秒算一次）
    Duration   int     `json:"duration"`   // 作答时长（秒）
}
```

### `api/model/report.go` 扩展

```go
type InterviewReport struct {
    // ... 现有字段 ...
    
    // 视频面试专有字段
    Mode               string  `json:"mode"`
    ExpressionSummary  string  `json:"expressionSummary,omitempty"`
    AvgExpressionScore int     `json:"avgExpressionScore,omitempty"`
    AvgSpeechRate      float64 `json:"avgSpeechRate,omitempty"`
}
```


## Backend Service Changes

### `api/service/interview_service.go` 扩展

`CreateInterview` 方法接受 `mode` 参数并写入 session：
```go
func CreateInterview(userId string, config InterviewConfig) (*InterviewSession, error) {
    // 现有逻辑不变...
    session := &InterviewSession{
        // ... 现有字段 ...
        Mode: config.Mode, // 新增
    }
    // ...
}
```

`SubmitAnswer` 方法接受 `NonVerbalMetrics` 参数：
```go
type SubmitAnswerRequest struct {
    QuestionIndex    int               `json:"questionIndex"`
    Answer           string            `json:"answer"`
    NonVerbalMetrics *NonVerbalMetrics `json:"nonVerbalMetrics,omitempty"`
}
```

---

### `api/service/deepseek_service.go` 扩展

`ReviewAnswer` 函数增加视频模式分支：

```go
func ReviewAnswer(question, answer string, config InterviewConfig, metrics *NonVerbalMetrics) (*ReviewResult, error) {
    basePrompt := buildReviewPrompt(question, answer, config)
    
    if metrics != nil {
        // 视频模式：附加非语言指标上下文
        videoContext := fmt.Sprintf(
            "\n\n[语音表达数据]\n语速：%.0f字/分钟（推荐范围120-150），停顿次数：%d次，作答时长：%d秒。\n请额外评估该应聘者的口头表达能力，在JSON中新增 expressionScore（0-100）和 expressionFeedback 字段。",
            metrics.SpeechRate, metrics.PauseCount, metrics.Duration,
        )
        basePrompt += videoContext
    }
    
    resp, err := Chat(basePrompt)
    // ...
    
    result := &ReviewResult{}
    // ... 解析现有字段 ...
    if metrics != nil {
        result.ExpressionScore = parsed.ExpressionScore
        result.ExpressionFeedback = parsed.ExpressionFeedback
    }
    return result, nil
}
```

`GenerateReport` 函数在视频模式下追加表达能力汇总：

```go
func GenerateVideoExpressionSummary(answers []AnswerRecord) (*ExpressionSummary, error) {
    avgRate := calcAvgSpeechRate(answers)
    avgScore := calcAvgExpressionScore(answers)
    prompt := fmt.Sprintf(
        "该应聘者完成了%d道视频面试题，平均语速%.0f字/分钟，平均表达得分%d分。请给出整体口头表达能力评价和改进建议（JSON格式，字段：summary, suggestions数组）",
        len(answers), avgRate, avgScore,
    )
    // ...
}
```


## API Changes

### 修改接口

所有修改均向后兼容，现有文字模式调用无需改动。

#### `POST /api/v1/interviews` — 请求体扩展

```json
{
  "jobTitle": "后端开发",
  "difficulty": "medium",
  "experience": "3年",
  "round": "技术一面",
  "focusAreas": ["数据库", "算法"],
  "mode": "video"
}
```

`mode` 字段可选，默认 `"text"`。

---

#### `POST /api/v1/interviews/:id/answers` — 请求体扩展

```json
{
  "questionIndex": 0,
  "answer": "语音识别转写的文字答案...",
  "nonVerbalMetrics": {
    "speechRate": 135.5,
    "pauseCount": 2,
    "duration": 87
  }
}
```

`nonVerbalMetrics` 字段可选，文字模式不传。

---

#### `GET /api/v1/reports/:interviewId` — 响应体扩展（视频模式）

```json
{
  "interviewId": "xxx",
  "mode": "video",
  "totalScore": 78,
  "expressionSummary": "整体口头表达流畅，语速适中...",
  "avgExpressionScore": 82,
  "avgSpeechRate": 138.5,
  "questions": [
    {
      "score": 85,
      "expressionScore": 80,
      "expressionFeedback": "语速适中，但停顿较多...",
      "nonVerbalMetrics": { "speechRate": 130, "pauseCount": 3, "duration": 92 }
    }
  ]
}
```


## Frontend Routing

### 新增路由（`ui/src/router/index.js`）

```javascript
{
  path: '/interview/:id/video-prep',
  name: 'VideoPreparation',
  component: () => import('@/views/interview/VideoPreparation.vue'),
  meta: { requiresAuth: true, title: '视频面试准备' }
},
{
  path: '/interview/:id/video-doing',
  name: 'VideoDoing',
  component: () => import('@/views/interview/VideoDoing.vue'),
  meta: { requiresAuth: true, title: '视频面试进行中' }
}
```

### 路由流程对比

```
文字模式：
Config → POST /interviews → Loading (轮询) → Doing → Report

视频模式：
Config → POST /interviews → VideoPreparation (权限/预览) → VideoDoing → Report
```

VideoDoing 完成后跳转到现有 `/report/:id`，报告页自动检测 `mode === 'video'` 并展示额外维度。

---

## Report Page Extensions

### `ui/src/views/report/Detail.vue` 视频模式扩展

当 `report.mode === 'video'` 时，报告页新增以下内容：

**1. 模式标签**
```vue
<el-tag type="warning" v-if="report.mode === 'video'">
  <i class="el-icon-video-camera"></i> 视频面试
</el-tag>
```

**2. 表达能力概览卡片**
```vue
<el-card v-if="report.mode === 'video'" class="expression-card">
  <div slot="header">表达能力分析</div>
  <el-row>
    <el-col :span="8">
      <ScoreCircle :score="report.avgExpressionScore" label="表达得分" />
    </el-col>
    <el-col :span="8">
      <div class="metric">
        <div class="value">{{ report.avgSpeechRate }} <small>字/分钟</small></div>
        <div class="label">平均语速</div>
        <div class="hint" :class="speechRateClass">{{ speechRateHint }}</div>
      </div>
    </el-col>
    <el-col :span="8">
      <div class="metric">
        <div class="value">{{ totalPauses }</div>
        <div class="label">总停顿次数</div>
      </div>
    </el-col>
  </el-row>
  <p class="summary">{{ report.expressionSummary }}</p>
</el-card>
```

**3. 雷达图新维度**

`RadarChart.vue` 中，视频模式的 `modules` prop 额外包含「表达能力」维度（来自 `avgExpressionScore`）。

**4. 逐题明细新增列**

```vue
<div v-if="report.mode === 'video'" class="expression-detail">
  <ScoreBar :score="question.expressionScore" label="表达" />
  <p>{{ question.expressionFeedback }}</p>
  <small>语速：{{ question.nonVerbalMetrics.speechRate }}字/分钟，
         停顿：{{ question.nonVerbalMetrics.pauseCount }}次</small>
</div>
```


## Privacy & Resource Management

### 媒体流生命周期管理

```javascript
// 准备页 → 答题页：通过 Vuex 传递 stream（不重新请求权限）
// 答题页销毁（beforeDestroy / beforeRouteLeave）时释放资源

beforeDestroy() {
  this.stopSpeech()
  this.stopRecording()
  this.$store.commit('interview/RELEASE_MEDIA_STREAM')
},

// 监听页面关闭
mounted() {
  window.addEventListener('beforeunload', this.handleBeforeUnload)
},
methods: {
  handleBeforeUnload(e) {
    if (this.isInterviewActive) {
      e.preventDefault()
      e.returnValue = '面试正在进行中，确定离开？媒体流将被关闭。'
    }
  }
}
```

### 隐私弹窗实现

```vue
<el-dialog title="隐私声明" :visible.sync="showPrivacyDialog" :close-on-click-modal="false">
  <ul>
    <li>摄像头和麦克风画面<strong>不会上传至服务器</strong></li>
    <li>语音识别在<strong>浏览器本地</strong>完成</li>
    <li>后端仅接收文字答案和语速/停顿统计数据</li>
    <li>视频录制（如开启）仅存储于本地内存，关闭页面后消失</li>
  </ul>
  <div slot="footer">
    <el-button @click="cancelInterview">取消</el-button>
    <el-button type="primary" @click="agreeAndContinue">同意并开始</el-button>
  </div>
</el-dialog>
```

---

## Browser Compatibility

| 功能 | Chrome | Edge | Firefox | Safari |
|------|--------|------|---------|--------|
| `getUserMedia` | ✅ 47+ | ✅ 12+ | ✅ 36+ | ✅ 11+ |
| `SpeechRecognition` | ✅ 33+ | ✅ 79+ | ❌ | ⚠️ 部分支持 |
| `MediaRecorder` | ✅ 47+ | ✅ 79+ | ✅ 25+ | ⚠️ 14.1+ |

**推荐浏览器：Chrome 88+**

降级策略：
- 无 `SpeechRecognition` → 允许进入，但语音转文字功能不可用，显示「请手动输入答案」提示
- 无 `MediaRecorder` → 允许进入，禁用录制开关
- 无 `getUserMedia` → 完全禁止进入视频模式

---

## File Structure

新增/修改的文件清单：

```
ui/src/
├── views/interview/
│   ├── VideoPreparation.vue       【新增】设备检测与授权准备页
│   └── VideoDoing.vue             【新增】视频面试答题页
├── components/interview/
│   ├── VideoPreview.vue           【新增】摄像头预览组件
│   └── SpeechIndicator.vue        【新增】语音状态指示组件
├── mixins/
│   ├── speechMixin.js             【新增】语音识别可复用逻辑
│   └── recordingMixin.js          【新增】视频录制可复用逻辑
├── store/modules/interview.js     【修改】新增 mediaStream 状态
├── router/index.js                【修改】新增两条路由
├── api/interview.js               【修改】提交答案接口支持 nonVerbalMetrics
└── views/interview/Config.vue     【修改】新增答题方式选择
    views/report/Detail.vue        【修改】视频模式报告展示

api/
├── model/interview.go             【修改】新增 mode、NonVerbalMetrics 字段
├── model/report.go                【修改】新增表达能力汇总字段
├── service/interview_service.go   【修改】CreateInterview/SubmitAnswer 支持 mode
├── service/deepseek_service.go    【修改】ReviewAnswer/GenerateReport 视频模式 Prompt
└── handler/interview.go           【修改】绑定新字段
```


## Implementation Notes

### 关键技术要点

1. **Web Speech API 局限性**
   - `SpeechRecognition` 在中文环境下识别率约 85-90%，可能出现错别字
   - 需允许用户手动编辑识别结果
   - 静默检测不完美，停顿判定依赖结果时间戳差值

2. **MediaRecorder 内存管理**
   - 长面试（>30分钟）可能产生 500MB+ 视频文件
   - 建议每 30 秒生成一个 chunk，避免内存溢出
   - 提示用户关闭其他标签页以释放内存

3. **AI Prompt 优化**
   - DeepSeek 对口语化表达的评估需要明确 prompt 引导
   - 建议 prompt 示例：「请从流畅度、逻辑性、语言规范性三个维度评估口头表达，给出0-100分」

4. **跨浏览器兼容**
   - Safari 的 `SpeechRecognition` 需加 `webkit` 前缀且仅部分支持
   - Firefox 不支持 `SpeechRecognition`，必须降级为纯手动输入

5. **移动端策略**
   - 视频面试对设备性能和网络要求较高
   - 建议在移动端（检测 `navigator.userAgent`）完全禁用视频模式

---

## Testing Strategy

### 前端测试重点

- **设备权限测试**：模拟用户拒绝权限、浏览器不支持、无设备等异常场景
- **语音识别测试**：测试中文/英文切换、长时间静默、识别中断
- **视频录制测试**：长时间录制（>30分钟）、内存占用、录制中断恢复
- **媒体流释放测试**：验证路由跳转、页面刷新、主动退出时 stream 是否正确释放

### 后端测试重点

- **数据兼容性测试**：文字模式旧数据不受视频字段影响
- **AI Prompt 测试**：验证 DeepSeek 返回的 JSON 包含 `expressionScore` 字段
- **报告生成测试**：视频模式报告计算平均语速、表达得分

### 集成测试场景

1. 完整视频面试流程：配置 → 准备 → 答题（3题）→ 报告查看
2. 暂停恢复流程：答题中途暂停 → 关闭页面 → 再次进入恢复
3. 降级流程：Firefox 浏览器下进入视频模式但语音识别不可用
4. 录制下载流程：开启录制 → 完成面试 → 下载视频文件

---

## Performance Considerations

- **摄像头预览帧率**：建议限制在 15fps（通过 `getUserMedia` constraints）
- **语音识别频率**：`continuous: true` 模式下会持续监听，CPU 占用约 5-10%
- **视频编码效率**：WebM VP9 编码比 H.264 文件更小，优先使用
- **Redis 数据大小**：非语言指标 JSON 约 100 bytes，对 Redis 性能影响极小

---

## Security & Compliance

- ✅ 音视频数据不离开用户设备，符合 GDPR 隐私要求
- ✅ 后端仅接收文本和统计数据，无敏感二进制数据
- ✅ 用户可随时停止摄像头/麦克风访问
- ⚠️ 建议在隐私政策中明确说明「视频面试模拟仅用于个人练习，不做身份验证或人脸识别」

