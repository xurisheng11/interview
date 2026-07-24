<template>
  <div class="doing-layout">
    <!-- 左侧题目导航 -->
    <div class="nav-panel">
      <div class="nav-title">题目导航</div>
      <div class="nav-progress-text">{{ answeredCount }}/{{ total }} 已完成</div>
      <div class="nav-list">
        <div
          v-for="(q, idx) in questions"
          :key="idx"
          class="nav-item"
          :class="navItemClass(idx)"
          @click="jumpTo(idx)"
        >
          <span class="nav-num">{{ idx + 1 }}</span>
          <span class="nav-status-icon">
            <span v-if="answers[idx] && answers[idx].skipped">⊘</span>
            <span v-else-if="answers[idx] && answers[idx].submitted">✓</span>
            <span v-else-if="idx === currentIdx">●</span>
            <span v-else>○</span>
          </span>
        </div>
      </div>
      <div class="nav-legend">
        <div class="legend-item"><span class="legend-dot answered"></span>已答</div>
        <div class="legend-item"><span class="legend-dot skipped"></span>跳过</div>
        <div class="legend-item"><span class="legend-dot current"></span>当前</div>
      </div>
      <!-- 暂停按钮 -->
      <el-button
        type="danger"
        size="small"
        class="pause-btn"
        :loading="pausing"
        @click="handlePause"
      >⏸ 暂停面试</el-button>
    </div>

    <!-- 右侧主内容 -->
    <div class="main-area">
      <!-- 顶部进度条 + 计时器 -->
      <div class="top-bar">
        <div class="progress-wrap">
          <span class="progress-label">答题进度</span>
          <el-progress
            :percentage="progressPercent"
            :stroke-width="10"
            :color="'#ff9900'"
            class="progress-bar"
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
            class="q-tag"
          >{{ tag }}</el-tag>
          <span class="q-time-hint" v-if="currentQuestion.estimatedMinutes">
            建议 {{ currentQuestion.estimatedMinutes }} 分钟
          </span>
        </div>
        <div class="question-content">{{ currentQuestion.content }}</div>
      </div>

      <!-- 答案输入区（未提交时显示） -->
      <div class="answer-area" v-if="!currentAnswerState.submitted && !currentAnswerState.skipped">
        <div class="answer-label">📝 你的答案 <span class="hint">（支持代码块，使用 ``` 包裹）</span></div>
        <textarea
          v-model="currentAnswer"
          class="answer-input"
          placeholder="请输入你的答案..."
          :disabled="submitting"
        ></textarea>
        <div class="action-row">
          <el-button
            type="primary"
            class="btn-submit"
            :loading="submitting"
            :disabled="!currentAnswer.trim()"
            @click="handleSubmit"
          >提交答案</el-button>
          <el-button
            class="btn-skip"
            :disabled="submitting"
            @click="handleSkip"
          >跳过此题</el-button>
        </div>
      </div>

      <!-- AI 点评区（提交后显示） -->
      <div class="review-area" v-if="currentAnswerState.submitted && !currentAnswerState.skipped">
        <div class="review-score-row">
          <span class="review-score-label">AI 评分</span>
          <el-progress
            :percentage="currentAnswerState.score || 0"
            :stroke-width="14"
            :color="scoreColor(currentAnswerState.score)"
            class="review-progress"
          ></el-progress>
          <span class="review-score-num" :style="{ color: scoreColor(currentAnswerState.score) }">
            {{ currentAnswerState.score }}分
          </span>
        </div>
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
        <el-collapse class="ref-collapse">
          <el-collapse-item title="📖 查看参考答案" name="ref">
            <div class="ref-answer">{{ currentAnswerState.referenceAnswer || '暂无参考答案' }}</div>
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
import { submitAnswer, pauseInterview, completeInterview } from '@/api/interview'

export default {
  name: 'InterviewDoing',

  data() {
    return {
      currentIdx: 0,
      currentAnswer: '',
      submitting: false,
      pausing: false,
      completing: false,
      // answers[idx] = { submitted, skipped, score, pros, cons, referenceAnswer, userAnswer }
      answers: [],
      timerSeconds: 0,
      timerHandle: null
    }
  },

  computed: {
    interviewId() {
      return this.$store.state.interview.currentInterviewId
    },
    questions() {
      return this.$store.state.interview.currentQuestions || []
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
      return this.timerSeconds <= 60 && this.timerSeconds > 0
    }
  },

  created() {
    if (!this.interviewId || !this.questions.length) {
      this.$message.warning('未找到面试信息，请重新配置')
      this.$router.replace('/interview/config')
      return
    }
    this.answers = Array(this.total).fill(null).map(() => ({}))
    this.startTimer()
  },

  beforeDestroy() {
    this.clearTimer()
  },

  methods: {
    navItemClass(idx) {
      const a = this.answers[idx] || {}
      if (a.skipped) return 'skipped'
      if (a.submitted) return 'answered'
      if (idx === this.currentIdx) return 'current'
      return ''
    },

    jumpTo(idx) {
      const a = this.answers[idx] || {}
      // 只允许跳到已答/已跳过的题，或当前题
      if (a.submitted || a.skipped || idx === this.currentIdx) {
        this.currentIdx = idx
        this.currentAnswer = a.userAnswer || ''
        this.resetTimer()
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
          // 超时自动跳过
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
      if (!this.currentAnswer.trim()) return
      this.submitting = true
      try {
        const res = await submitAnswer(this.interviewId, {
          questionIndex: this.currentIdx,
          answer: this.currentAnswer.trim()
        })
        const d = res?.data?.data || res?.data || res
        this.$set(this.answers, this.currentIdx, {
          submitted: true,
          skipped: false,
          userAnswer: this.currentAnswer.trim(),
          score: d.score || 0,
          pros: Array.isArray(d.pros) ? d.pros : (d.pros ? [d.pros] : []),
          cons: Array.isArray(d.cons) ? d.cons : (d.cons ? [d.cons] : []),
          referenceAnswer: d.referenceAnswer || d.reference_answer || ''
        })
        this.clearTimer()
      } catch (err) {
        const msg = err?.response?.data?.message || err?.message || '提交失败，请重试'
        this.$message.error(msg)
      } finally {
        this.submitting = false
      }
    },

    handleSkip() {
      this.doSkip()
    },

    doSkip() {
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
        this.currentAnswer = ''
        this.resetTimer()
      }
    },

    async handleComplete() {
      this.completing = true
      try {
        await completeInterview(this.interviewId)
        this.$message.success('面试完成！正在生成报告...')
        this.$router.push(`/report/${this.interviewId}`)
      } catch (err) {
        const msg = err?.response?.data?.message || err?.message || '完成面试失败，请重试'
        this.$message.error(msg)
      } finally {
        this.completing = false
      }
    },

    async handlePause() {
      this.pausing = true
      try {
        await pauseInterview(this.interviewId)
        this.$message.success('面试已暂停，下次登录可继续')
        this.$router.push('/dashboard')
      } catch (err) {
        const msg = err?.response?.data?.message || err?.message || '暂停失败，请重试'
        this.$message.error(msg)
      } finally {
        this.pausing = false
      }
    }
  }
}
</script>

<style scoped>
.doing-layout {
  display: flex;
  min-height: calc(100vh - 90px);
  background: #f3f3f3;
}

/* ===== 左侧导航面板 ===== */
.nav-panel {
  width: 180px;
  min-width: 180px;
  background: #fff;
  border-right: 1px solid #ddd;
  padding: 16px 12px;
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.nav-title {
  font-size: 13px;
  font-weight: bold;
  color: #333;
  border-left: 3px solid #ff9900;
  padding-left: 8px;
}

.nav-progress-text {
  font-size: 12px;
  color: #999;
}

.nav-list {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.nav-item {
  width: 36px;
  height: 36px;
  border-radius: 4px;
  border: 2px solid #ddd;
  background: #fafafa;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: all 0.2s;
  position: relative;
}

.nav-item:hover { border-color: #ff9900; }

.nav-item.answered  { background: #e6f9ed; border-color: #67c23a; }
.nav-item.skipped   { background: #f5f5f5; border-color: #bbb; }
.nav-item.current   { border-color: #ff9900; border-width: 2px; box-shadow: 0 0 0 2px rgba(255,153,0,0.2); }

.nav-num  { font-size: 11px; font-weight: bold; color: #555; line-height: 1; }
.nav-status-icon { font-size: 10px; color: #aaa; line-height: 1; }
.nav-item.answered  .nav-status-icon { color: #67c23a; }
.nav-item.skipped   .nav-status-icon { color: #bbb; }
.nav-item.current   .nav-status-icon { color: #ff9900; }

.nav-legend { display: flex; flex-direction: column; gap: 4px; margin-top: 4px; }
.legend-item { display: flex; align-items: center; gap: 6px; font-size: 11px; color: #666; }
.legend-dot { width: 10px; height: 10px; border-radius: 2px; }
.legend-dot.answered { background: #e6f9ed; border: 1px solid #67c23a; }
.legend-dot.skipped  { background: #f5f5f5; border: 1px solid #bbb; }
.legend-dot.current  { background: #fff; border: 2px solid #ff9900; }

.pause-btn { margin-top: auto; width: 100%; }

/* ===== 右侧主区 ===== */
.main-area {
  flex: 1;
  padding: 20px;
  overflow: auto;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

/* 顶部进度条+计时器 */
.top-bar {
  display: flex;
  align-items: center;
  gap: 16px;
  background: #fff;
  border: 1px solid #ddd;
  border-radius: 4px;
  padding: 12px 16px;
}

.progress-wrap {
  flex: 1;
  display: flex;
  align-items: center;
  gap: 10px;
}

.progress-label { font-size: 13px; color: #555; white-space: nowrap; }
.progress-bar   { flex: 1; }
.progress-text  { font-size: 13px; color: #555; white-space: nowrap; }

.timer {
  font-size: 18px;
  font-weight: bold;
  color: #333;
  background: #f8f8f8;
  border: 1px solid #ddd;
  border-radius: 4px;
  padding: 4px 14px;
  display: flex;
  align-items: center;
  gap: 6px;
  white-space: nowrap;
  transition: color 0.3s, border-color 0.3s;
}

.timer.warning { color: #f56c6c; border-color: #f56c6c; background: #fff6f6; }

/* 题目卡片 */
.question-card {
  background: #fff;
  border: 1px solid #ddd;
  border-radius: 4px;
  padding: 20px;
}

.question-header {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 8px;
  margin-bottom: 14px;
}

.q-index {
  font-size: 15px;
  font-weight: bold;
  color: #111;
  background: #ff9900;
  padding: 2px 10px;
  border-radius: 12px;
}

.q-difficulty {
  font-size: 12px;
  padding: 2px 10px;
  border-radius: 12px;
  font-weight: bold;
}
.diff-初级, .diff-easy   { background: #e6f9ed; color: #67c23a; }
.diff-中级, .diff-medium { background: #fff3e0; color: #ff9900; }
.diff-高级, .diff-hard   { background: #fee; color: #f56c6c; }

.q-tag { margin: 0 2px; }

.q-time-hint {
  font-size: 12px;
  color: #999;
  margin-left: auto;
}

.question-content {
  font-size: 15px;
  line-height: 1.8;
  color: #222;
  white-space: pre-wrap;
  word-break: break-word;
}

/* 答案输入区 */
.answer-area {
  background: #fff;
  border: 1px solid #ddd;
  border-radius: 4px;
  padding: 20px;
}

.answer-label {
  font-size: 13px;
  font-weight: bold;
  color: #333;
  margin-bottom: 10px;
}

.hint { font-size: 12px; color: #aaa; font-weight: normal; }

.answer-input {
  width: 100%;
  min-height: 180px;
  padding: 12px;
  border: 1px solid #ccc;
  border-radius: 4px;
  font-size: 14px;
  font-family: 'Courier New', monospace;
  resize: vertical;
  outline: none;
  transition: border 0.2s;
  box-sizing: border-box;
  line-height: 1.6;
}

.answer-input:focus { border-color: #ff9900; box-shadow: 0 0 0 2px rgba(255,153,0,0.12); }

.action-row {
  display: flex;
  gap: 12px;
  margin-top: 14px;
}

.btn-submit { background: #ff9900 !important; border-color: #ff9900 !important; color: #111 !important; font-weight: bold !important; }
.btn-submit:hover { background: #f3a847 !important; border-color: #f3a847 !important; }
.btn-submit.is-disabled, .btn-submit.is-disabled:hover { background: #e0e0e0 !important; border-color: #e0e0e0 !important; color: #aaa !important; }

/* AI 点评区 */
.review-area {
  background: #fff;
  border: 1px solid #ddd;
  border-radius: 4px;
  padding: 20px;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.review-score-row {
  display: flex;
  align-items: center;
  gap: 12px;
}

.review-score-label { font-size: 14px; font-weight: bold; color: #333; white-space: nowrap; }
.review-progress { flex: 1; }
.review-score-num { font-size: 22px; font-weight: bold; min-width: 52px; text-align: right; }

.review-section-title {
  font-size: 13px;
  font-weight: bold;
  color: #333;
  margin-bottom: 8px;
}

.review-list {
  margin: 0;
  padding-left: 20px;
}

.review-list li {
  font-size: 14px;
  line-height: 1.7;
  color: #444;
}

.review-list.pros li { color: #2d7a4a; }
.review-list.cons li { color: #a05c00; }

.ref-answer {
  font-size: 14px;
  line-height: 1.8;
  color: #333;
  white-space: pre-wrap;
  padding: 10px;
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

.skip-icon { font-size: 40px; color: #bbb; }
.skip-text { font-size: 16px; color: #999; margin: 8px 0 20px; }

.next-row { display: flex; justify-content: center; }
.btn-next { background: #ff9900 !important; border-color: #ff9900 !important; color: #111 !important; font-weight: bold !important; padding: 10px 32px !important; font-size: 15px !important; }
.btn-next:hover { background: #f3a847 !important; border-color: #f3a847 !important; }

/* el-progress 颜色覆盖 */
::v-deep .el-progress-bar__inner { transition: width 0.4s ease; }
::v-deep .el-collapse-item__header { font-weight: bold; color: #333; }
</style>
