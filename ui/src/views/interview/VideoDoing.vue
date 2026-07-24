<template>
  <div class="video-doing-layout">
    <!-- 左侧摄像头区 -->
    <div class="camera-panel">
      <div class="camera-preview-wrapper">
        <VideoPreview :stream="mediaStream" :mirror="true" />
      </div>
      
      <!-- 录制状态指示 -->
      <div v-if="isRecording" class="recording-status">
        <span class="rec-dot"></span>
        <span class="rec-text">录制中 {{ formatRecordingDuration(recordingDuration) }}</span>
      </div>

      <!-- 语音识别状态 -->
      <div class="speech-status-row">
        <SpeechIndicator 
          :stream="mediaStream" 
          :active="isSpeechActive" 
          :supported="isSpeechSupported" 
        />
      </div>
    </div>

    <!-- 右侧答题区 -->
    <div class="main-panel">
      <!-- 顶部进度+计时 -->
      <div class="top-bar">
        <div class="progress-wrap">
          <span class="progress-label">答题进度</span>
          <el-progress
            :percentage="progressPercent"
            :stroke-width="10"
            :color="'#ff9900'"
          ></el-progress>
          <span class="progress-text">{{ answeredCount }}/{{ total }}</span>
        </div>
        <div class="timer" :class="{ warning: timerWarning }">
          <i class="el-icon-time"></i>
          {{ timerText }}
        </div>
      </div>

      <!-- 题目卡片 -->
      <div class="question-card" v-if="currentQuestion">
        <div class="question-header">
          <span class="q-index">第 {{ currentIdx + 1 }} 题</span>
          <span class="q-difficulty" :class="'diff-' + currentQuestion.difficulty">
            {{ currentQuestion.difficulty || '中等' }}
          </span>
          <el-tag
            v-for="tag in (currentQuestion.tags || [])"
            :key="tag"
            size="mini"
          >{{ tag }}</el-tag>
          <span class="q-time-hint" v-if="currentQuestion.estimatedMinutes">
            建议 {{ currentQuestion.estimatedMinutes }} 分钟
          </span>
        </div>
        <div class="question-content">{{ currentQuestion.content }}</div>
      </div>

      <!-- 答案输入区（未提交） -->
      <div class="answer-area" v-if="!currentAnswerState.submitted && !currentAnswerState.skipped">
        <div class="answer-label">
          📝 你的答案
          <span class="hint">（语音自动填入，也可手动编辑）</span>
        </div>
        <textarea
          v-model="userAnswer"
          class="answer-input"
          placeholder="开启麦克风后，语音将自动转为文字..."
          :disabled="submitting"
        ></textarea>
        
        <!-- 临时识别结果 -->
        <div v-if="interimTranscript" class="interim-text">
          <i class="el-icon-loading"></i> {{ interimTranscript }}
        </div>

        <!-- 控制按钮 -->
        <div class="control-row">
          <el-switch
            v-model="isSpeechActive"
            :disabled="!isSpeechSupported"
            @change="toggleSpeech"
            active-text="麦克风"
            inactive-text="麦克风"
          ></el-switch>
          <span v-if="!isSpeechSupported" class="warn-tip">语音识别不可用，请手动输入</span>
        </div>

        <div class="action-row">
          <el-button
            type="primary"
            class="btn-submit"
            :loading="submitting"
            :disabled="!userAnswer.trim()"
            @click="handleSubmit"
          >提交答案</el-button>
          <el-button @click="handleSkip" :disabled="submitting">跳过此题</el-button>
          <el-button type="danger" @click="handlePause" :loading="pausing">暂停面试</el-button>
        </div>
      </div>

      <!-- AI 点评区（已提交） -->
      <div class="review-area" v-if="currentAnswerState.submitted && !currentAnswerState.skipped">
        <div class="review-score-row">
          <span class="review-label">内容得分</span>
          <el-progress
            :percentage="currentAnswerState.score || 0"
            :stroke-width="14"
            :color="scoreColor(currentAnswerState.score)"
          ></el-progress>
          <span class="review-score-num" :style="{ color: scoreColor(currentAnswerState.score) }">
            {{ currentAnswerState.score }}分
          </span>
        </div>

        <!-- 表达得分（视频模式） -->
        <div class="review-score-row" v-if="currentAnswerState.expressionScore">
          <span class="review-label">表达得分</span>
          <el-progress
            :percentage="currentAnswerState.expressionScore || 0"
            :stroke-width="14"
            color="#409eff"
          ></el-progress>
          <span class="review-score-num" style="color:#409eff">
            {{ currentAnswerState.expressionScore }}分
          </span>
        </div>

        <!-- 非语言指标 -->
        <div v-if="currentAnswerState.nonVerbalMetrics" class="metrics-row">
          <span class="metric-item">
            🗣️ 语速：{{ currentAnswerState.nonVerbalMetrics.speechRate }} 字/分钟
          </span>
          <span class="metric-item">
            ⏸️ 停顿：{{ currentAnswerState.nonVerbalMetrics.pauseCount }} 次
          </span>
          <span class="metric-item">
            ⏱️ 用时：{{ currentAnswerState.nonVerbalMetrics.duration }} 秒
          </span>
        </div>

        <!-- AI 反馈 -->
        <div class="review-section" v-if="currentAnswerState.pros && currentAnswerState.pros.length">
          <div class="review-section-title">✅ 优点</div>
          <ul class="review-list pros">
            <li v-for="(p, i) in currentAnswerState.pros" :key="i">{{ p }}</li>
          </ul>
        </div>
        <div class="review-section" v-if="currentAnswerState.cons && currentAnswerState.cons.length">
          <div class="review-section-title">⚠️ 不足</div>
          <ul class="review-list cons">
            <li v-for="(c, i) in currentAnswerState.cons" :key="i">{{ c }}</li>
          </ul>
        </div>

        <!-- 表达反馈 -->
        <div class="review-section" v-if="currentAnswerState.expressionFeedback">
          <div class="review-section-title">🎤 表达反馈</div>
          <p class="expression-feedback">{{ currentAnswerState.expressionFeedback }}</p>
        </div>

        <el-collapse class="ref-collapse">
          <el-collapse-item title="📖 查看参考答案" name="ref">
            <div class="ref-answer">{{ currentAnswerState.referenceAnswer || '暂无' }}</div>
          </el-collapse-item>
        </el-collapse>

        <div class="next-row">
          <el-button type="primary" class="btn-next" @click="handleNext">
            {{ isLastQuestion ? '🎉 完成面试' : '下一题 →' }}
          </el-button>
        </div>
      </div>

      <!-- 跳过提示 -->
      <div class="skip-tip" v-if="currentAnswerState.skipped">
        <div class="skip-icon">⊘</div>
        <div class="skip-text">已跳过此题</div>
        <div class="next-row">
          <el-button type="primary" class="btn-next" @click="handleNext">
            {{ isLastQuestion ? '🎉 完成面试' : '下一题 →' }}
          </el-button>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { speechMixin } from '@/mixins/speechMixin'
import { recordingMixin } from '@/mixins/recordingMixin'
import VideoPreview from '@/components/interview/VideoPreview.vue'
import SpeechIndicator from '@/components/interview/SpeechIndicator.vue'
import { submitAnswer, completeInterview, pauseInterview } from '@/api/interview'

export default {
  name: 'VideoDoing',
  components: { VideoPreview, SpeechIndicator },
  mixins: [speechMixin, recordingMixin],

  data() {
    return {
      currentIdx: 0,
      userAnswer: '',
      submitting: false,
      pausing: false,
      completing: false,
      answers: [],
      timerSeconds: 0,
      timerHandle: null,
      speechLang: 'zh-CN'
    }
  },

  computed: {
    interviewId() {
      return this.$store.state.interview.currentInterviewId
    },
    questions() {
      return this.$store.state.interview.currentQuestions || []
    },
    mediaStream() {
      return this.$store.state.interview.mediaStream
    },
    enableRecording() {
      return this.$store.state.interview.enableRecording
    },
    total() {
      return this.questions.length
    },
    currentQuestion() {
      return this.questions[this.currentIdx] || null
    },
    currentAnswerState() {
      return this.answers[this.currentIdx] || {}
    },
    answeredCount() {
      return this.answers.filter(a => a && (a.submitted || a.skipped)).length
    },
    progressPercent() {
      if (!this.total) return 0
      return Math.round((this.answeredCount / this.total) * 100)
    },
    isLastQuestion() {
      return this.currentIdx >= this.total - 1
    },
    timerText() {
      const m = Math.floor(this.timerSeconds / 60)
      const s = this.timerSeconds % 60
      return `${String(m).padStart(2, '0')}:${String(s).padStart(2, '0')}`
    },
    timerWarning() {
      return this.timerSeconds <= 10 && this.timerSeconds > 0
    }
  },

  created() {
    if (!this.interviewId || !this.questions.length) {
      this.$message.warning('未找到面试信息，请重新配置')
      this.$router.replace('/interview/config')
      return
    }
    // 读取语言设置
    this.speechLang = sessionStorage.getItem('speechLang') || 'zh-CN'
    this.answers = Array(this.total).fill(null).map(() => ({}))
    this.startTimer()
  },

  mounted() {
    // 初始化语音识别
    if (this.initSpeechRecognition(this.speechLang)) {
      this.startSpeech()
    }
    // 启动录制
    if (this.enableRecording && this.mediaStream) {
      this.startRecording(this.mediaStream)
    }
    // 绑定页面关闭提示
    window.addEventListener('beforeunload', this.handleBeforeUnload)
  },

  beforeDestroy() {
    this.clearTimer()
    window.removeEventListener('beforeunload', this.handleBeforeUnload)
  },

  beforeRouteLeave(to, from, next) {
    // 面试进行中但未手动完成时弹出确认
    const allDone = this.answers.every(a => a && (a.submitted || a.skipped))
    if (!allDone && !this.completing) {
      this.$confirm('面试正在进行中，确定离开？媒体流将被关闭。', '提示', {
        confirmButtonText: '离开',
        cancelButtonText: '继续面试',
        type: 'warning'
      }).then(() => {
        this._cleanup()
        next()
      }).catch(() => next(false))
    } else {
      next()
    }
  },

  methods: {
    handleBeforeUnload(e) {
      const allDone = this.answers.every(a => a && (a.submitted || a.skipped))
      if (!allDone) {
        e.preventDefault()
        e.returnValue = '面试正在进行中，确定离开？媒体流将被关闭。'
      }
    },

    startTimer() {
      const q = this.currentQuestion
      const mins = (q && q.estimatedMinutes) ? q.estimatedMinutes * 2 : 10
      this.timerSeconds = mins * 60
      this.clearTimer()
      this.timerHandle = setInterval(() => {
        if (this.timerSeconds > 0) {
          this.timerSeconds--
        } else {
          this.clearTimer()
          if (!this.currentAnswerState.submitted && !this.currentAnswerState.skipped) {
            this.$message.warning('作答时间到，自动跳过此题')
            this.doSkip()
          }
        }
      }, 1000)
    },

    resetTimer() {
      this.clearTimer()
      this.startTimer()
    },

    clearTimer() {
      if (this.timerHandle) {
        clearInterval(this.timerHandle)
        this.timerHandle = null
      }
    },

    scoreColor(score) {
      if (!score && score !== 0) return '#909399'
      if (score >= 80) return '#67c23a'
      if (score >= 60) return '#ff9900'
      return '#f56c6c'
    },

    async handleSubmit() {
      if (!this.userAnswer.trim()) return
      this.submitting = true

      // 停止语音获取非语言指标
      this.stopSpeech()
      const metrics = this.getNonVerbalMetrics()

      try {
        const res = await submitAnswer(this.interviewId, {
          questionIndex: this.currentIdx,
          answer: this.userAnswer.trim(),
          nonVerbalMetrics: metrics
        })
        const d = res?.data?.data || res?.data || res
        this.$set(this.answers, this.currentIdx, {
          submitted: true,
          skipped: false,
          userAnswer: this.userAnswer.trim(),
          score: d.score || 0,
          pros: Array.isArray(d.pros) ? d.pros : (d.pros ? [d.pros] : []),
          cons: Array.isArray(d.cons) ? d.cons : (d.cons ? [d.cons] : []),
          referenceAnswer: d.referenceAnswer || '',
          expressionScore: d.expressionScore || 0,
          expressionFeedback: d.expressionFeedback || '',
          nonVerbalMetrics: metrics
        })
        this.clearTimer()
      } catch (err) {
        const msg = err?.response?.data?.message || err?.message || '提交失败，请重试'
        this.$message.error(msg)
        // 恢复语音识别
        this.startSpeech()
      } finally {
        this.submitting = false
      }
    },

    handleSkip() {
      this.doSkip()
    },

    doSkip() {
      this.stopSpeech()
      this.$set(this.answers, this.currentIdx, {
        submitted: false,
        skipped: true,
        userAnswer: ''
      })
      this.clearTimer()
    },

    handleNext() {
      if (this.isLastQuestion) {
        this.handleComplete()
      } else {
        this.currentIdx++
        this.userAnswer = ''
        this.resetTimer()
        // 开始下一题语音识别
        if (this.isSpeechSupported) {
          this.startSpeech()
        }
      }
    },

    async handleComplete() {
      this.completing = true
      try {
        // 停止录制
        if (this.enableRecording) {
          this.stopRecording()
        }
        await completeInterview(this.interviewId)
        this._cleanup()
        this.$message.success('面试完成！正在生成报告...')
        this.$router.push(`/report/${this.interviewId}`)
      } catch (err) {
        const msg = err?.response?.data?.message || err?.message || '完成面试失败，请重试'
        this.$message.error(msg)
        this.completing = false
      }
    },

    async handlePause() {
      this.pausing = true
      try {
        this.stopSpeech()
        this.stopRecording()
        await pauseInterview(this.interviewId)
        this._cleanup()
        this.$message.success('面试已暂停')
        this.$router.push('/dashboard')
      } catch (err) {
        const msg = err?.response?.data?.message || err?.message || '暂停失败，请重试'
        this.$message.error(msg)
      } finally {
        this.pausing = false
      }
    },

    _cleanup() {
      this.stopSpeech()
      this.clearTimer()
      this.$store.commit('interview/RELEASE_MEDIA_STREAM')
    }
  }
}
</script>

<style scoped>
/* Task 16.2 样式 */
.video-doing-layout {
  display: flex;
  min-height: calc(100vh - 90px);
  background: #f3f3f3;
}

/* 左侧摄像头区 40% */
.camera-panel {
  width: 40%;
  min-width: 280px;
  background: #111;
  padding: 12px;
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.camera-preview-wrapper {
  width: 100%;
  padding-top: 75%; /* 4:3 */
  position: relative;
  border-radius: 8px;
  overflow: hidden;
}
.camera-preview-wrapper > * {
  position: absolute;
  inset: 0;
}

/* 录制状态红点闪烁 */
.recording-status {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 10px;
  background: rgba(255,255,255,0.1);
  border-radius: 4px;
}
.rec-dot {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  background: #f56c6c;
  animation: blink 1s infinite;
}
.rec-text { color: #fff; font-size: 13px; }

@keyframes blink {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.2; }
}

.speech-status-row {
  padding: 6px 10px;
  background: rgba(255,255,255,0.08);
  border-radius: 4px;
}

/* 右侧答题区 60% */
.main-panel {
  flex: 1;
  padding: 16px;
  overflow: auto;
  display: flex;
  flex-direction: column;
  gap: 14px;
}

/* 进度条 */
.top-bar {
  display: flex;
  align-items: center;
  gap: 12px;
  background: #fff;
  border: 1px solid #ddd;
  border-radius: 4px;
  padding: 10px 14px;
}
.progress-wrap {
  flex: 1;
  display: flex;
  align-items: center;
  gap: 10px;
}
.progress-label, .progress-text { font-size: 13px; color: #555; white-space: nowrap; }

.timer {
  font-size: 18px;
  font-weight: bold;
  color: #333;
  background: #f8f8f8;
  border: 1px solid #ddd;
  border-radius: 4px;
  padding: 4px 12px;
  display: flex;
  align-items: center;
  gap: 6px;
  white-space: nowrap;
  transition: color 0.3s, border-color 0.3s;
}
/* 倒计时最后10秒变红 */
.timer.warning { color: #f56c6c; border-color: #f56c6c; background: #fff6f6; }

/* 题目卡片 */
.question-card {
  background: #fff;
  border: 1px solid #ddd;
  border-radius: 4px;
  padding: 16px;
}
.question-header {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 8px;
  margin-bottom: 12px;
}
.q-index {
  font-size: 14px;
  font-weight: bold;
  background: #ff9900;
  color: #111;
  padding: 2px 10px;
  border-radius: 12px;
}
.q-difficulty {
  font-size: 12px;
  padding: 2px 8px;
  border-radius: 12px;
  font-weight: bold;
}
.diff-初级, .diff-easy   { background: #e6f9ed; color: #67c23a; }
.diff-中级, .diff-medium { background: #fff3e0; color: #ff9900; }
.diff-高级, .diff-hard   { background: #fee; color: #f56c6c; }
.q-time-hint { font-size: 12px; color: #999; margin-left: auto; }
.question-content { font-size: 15px; line-height: 1.8; color: #222; white-space: pre-wrap; }

/* 答案区 */
.answer-area {
  background: #fff;
  border: 1px solid #ddd;
  border-radius: 4px;
  padding: 16px;
  display: flex;
  flex-direction: column;
  gap: 10px;
}
.answer-label { font-size: 13px; font-weight: bold; color: #333; }
.hint { font-size: 12px; color: #aaa; font-weight: normal; }

.answer-input {
  width: 100%;
  min-height: 140px;
  padding: 10px;
  border: 1px solid #ccc;
  border-radius: 4px;
  font-size: 14px;
  font-family: inherit;
  resize: vertical;
  outline: none;
  transition: border 0.2s;
  box-sizing: border-box;
  line-height: 1.6;
}
.answer-input:focus { border-color: #ff9900; }

/* 临时识别结果（灰色斜体） */
.interim-text {
  font-size: 13px;
  color: #999;
  font-style: italic;
  padding: 4px 8px;
  background: #f5f5f5;
  border-radius: 4px;
  min-height: 28px;
}

.control-row {
  display: flex;
  align-items: center;
  gap: 12px;
}
.warn-tip { font-size: 12px; color: #e6a817; }

.action-row {
  display: flex;
  gap: 10px;
  flex-wrap: wrap;
}

.btn-submit {
  background: #ff9900 !important;
  border-color: #ff9900 !important;
  color: #111 !important;
  font-weight: bold !important;
}

/* AI点评区 */
.review-area {
  background: #fff;
  border: 1px solid #ddd;
  border-radius: 4px;
  padding: 16px;
  display: flex;
  flex-direction: column;
  gap: 12px;
}
.review-score-row {
  display: flex;
  align-items: center;
  gap: 10px;
}
.review-label { font-size: 13px; font-weight: bold; color: #333; white-space: nowrap; width: 60px; }
.review-score-num { font-size: 20px; font-weight: bold; min-width: 46px; text-align: right; }

.metrics-row {
  display: flex;
  gap: 16px;
  flex-wrap: wrap;
  background: #f8f8f8;
  border-radius: 4px;
  padding: 8px 12px;
}
.metric-item { font-size: 12px; color: #555; }

.review-section-title { font-size: 13px; font-weight: bold; color: #333; margin-bottom: 6px; }
.review-list { margin: 0; padding-left: 18px; }
.review-list li { font-size: 13px; line-height: 1.7; }
.review-list.pros li { color: #2d7a4a; }
.review-list.cons li { color: #a05c00; }

.expression-feedback { font-size: 13px; color: #444; line-height: 1.7; margin: 0; }

.ref-answer {
  font-size: 13px;
  color: #444;
  line-height: 1.8;
  white-space: pre-wrap;
  padding: 8px 10px;
  background: #f8f8f8;
  border-radius: 4px;
}

/* 跳过提示 */
.skip-tip {
  background: #fff;
  border: 1px dashed #ccc;
  border-radius: 4px;
  padding: 40px 20px;
  text-align: center;
}
.skip-icon { font-size: 36px; color: #bbb; }
.skip-text { font-size: 15px; color: #999; margin: 8px 0 20px; }

.next-row { display: flex; justify-content: center; }
.btn-next {
  background: #ff9900 !important;
  border-color: #ff9900 !important;
  color: #111 !important;
  font-weight: bold !important;
  padding: 10px 28px !important;
}

@media (max-width: 768px) {
  .video-doing-layout { flex-direction: column; }
  .camera-panel { width: 100%; }
}
</style>
