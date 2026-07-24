<template>
  <div class="loading-page">
    <!-- 正常加载态 -->
    <div v-if="!hasError" class="loading-container">
      <!-- 动画区域 -->
      <div class="spinner-wrapper">
        <div class="spinner-ring">
          <div class="spinner-inner">
            <span class="spinner-icon">🤖</span>
          </div>
        </div>
        <div class="pulse-ring pulse-ring-1"></div>
        <div class="pulse-ring pulse-ring-2"></div>
      </div>

      <!-- 标题文字 -->
      <div class="loading-title">AI 正在生成面试题目</div>
      <div class="loading-subtitle">{{ statusText }}</div>

      <!-- 进度点动画 -->
      <div class="dots">
        <span class="dot" :class="{ active: dotIdx >= 1 }"></span>
        <span class="dot" :class="{ active: dotIdx >= 2 }"></span>
        <span class="dot" :class="{ active: dotIdx >= 3 }"></span>
      </div>

      <!-- 配置摘要卡片 -->
      <div class="config-summary" v-if="interviewInfo">
        <div class="summary-title">本次面试配置</div>
        <div class="summary-grid">
          <div class="summary-item" v-if="interviewInfo.jobTitle">
            <span class="summary-label">目标岗位</span>
            <span class="summary-value highlight">{{ interviewInfo.jobTitle }}</span>
          </div>
          <div class="summary-item" v-if="interviewInfo.difficulty">
            <span class="summary-label">面试难度</span>
            <span class="summary-value">{{ interviewInfo.difficulty }}</span>
          </div>
          <div class="summary-item" v-if="interviewInfo.experience">
            <span class="summary-label">工作经验</span>
            <span class="summary-value">{{ interviewInfo.experience }}</span>
          </div>
          <div class="summary-item" v-if="interviewInfo.round">
            <span class="summary-label">面试轮次</span>
            <span class="summary-value">{{ roundLabel }}</span>
          </div>
          <div class="summary-item full-width" v-if="interviewInfo.focusAreas && interviewInfo.focusAreas.length">
            <span class="summary-label">重点方向</span>
            <span class="summary-value">
              <span
                v-for="area in interviewInfo.focusAreas"
                :key="area"
                class="focus-tag"
              >{{ area }}</span>
            </span>
          </div>
        </div>
      </div>

      <!-- 超时提示（非错误，仅提示） -->
      <div class="timeout-hint" v-if="isTimeout && !hasError">
        <i class="el-icon-time"></i>
        AI 生成耗时较长，请耐心等待...
      </div>
    </div>

    <!-- 错误态 -->
    <div v-else class="error-container">
      <div class="error-icon">⚠️</div>
      <div class="error-title">题目生成失败</div>
      <div class="error-msg">{{ errorMsg }}</div>
      <div class="error-actions">
        <el-button type="primary" class="action-btn" @click="handleRetry">
          🔄 重新尝试
        </el-button>
        <el-button class="action-btn-ghost" @click="handleBackConfig">
          ← 返回配置
        </el-button>
      </div>
    </div>
  </div>
</template>

<script>
import { getInterview } from '@/api/interview'

const POLL_INTERVAL = 2000   // 2 秒轮询一次
const TIMEOUT_MS   = 30000  // 30 秒超时

const ROUND_LABELS = {
  round1: '一面（基础）',
  round2: '二面（技术深度）',
  round3: '三面（综合/HR）'
}

const STATUS_TEXTS = [
  '正在分析岗位要求...',
  '正在匹配知识点范围...',
  '正在生成面试题目...',
  '正在优化题目质量...',
  '即将完成，请稍候...'
]

export default {
  name: 'InterviewLoading',

  data() {
    return {
      // 轮询相关
      pollTimer: null,
      timeoutTimer: null,
      startTime: null,

      // UI 状态
      hasError: false,
      isTimeout: false,
      errorMsg: '',
      dotIdx: 0,
      dotTimer: null,
      statusIdx: 0,
      statusTimer: null
    }
  },

  computed: {
    interviewId() {
      return this.$store.state.interview.currentInterviewId
    },

    interviewInfo() {
      // 优先取 store 里已有的 interview 对象
      const stored = this.$store.state.interview.currentInterview
      if (stored) return stored
      // 没有则返回 null，摘要区域不显示
      return null
    },

    roundLabel() {
      const r = this.interviewInfo && this.interviewInfo.round
      return ROUND_LABELS[r] || r || ''
    },

    statusText() {
      return STATUS_TEXTS[this.statusIdx % STATUS_TEXTS.length]
    }
  },

  created() {
    // 若没有 interviewId，直接跳回配置页
    if (!this.interviewId) {
      this.$message.warning('未找到面试信息，请重新配置')
      this.$router.replace('/interview/config')
      return
    }
    this.startPolling()
  },

  beforeDestroy() {
    this.clearTimers()
  },

  methods: {
    // ---- 启动轮询 ----
    startPolling() {
      this.hasError = false
      this.isTimeout = false
      this.errorMsg = ''
      this.startTime = Date.now()

      // 打点动画
      this.dotTimer = setInterval(() => {
        this.dotIdx = (this.dotIdx % 3) + 1
      }, 600)

      // 状态文字轮换
      this.statusTimer = setInterval(() => {
        this.statusIdx += 1
      }, 3000)

      // 超时定时器
      this.timeoutTimer = setTimeout(() => {
        this.isTimeout = true
        // 超时后再给 10s 最终机会，若仍未成功则报错
        this.finalTimeoutTimer = setTimeout(() => {
          this.clearTimers()
          this.showError('AI 生成超时，请稍后重试')
        }, 10000)
      }, TIMEOUT_MS)

      // 立即执行一次，再开始定时
      this.poll()
    },

    // ---- 单次轮询 ----
    async poll() {
      try {
        const res = await getInterview(this.interviewId)

        // 兼容不同响应结构
        const data = res?.data?.data || res?.data || res

        // 判断 questions 是否已生成
        const questions = data?.questions || data?.data?.questions
        const hasQuestions = Array.isArray(questions) && questions.length > 0

        if (hasQuestions) {
          // 题目已就绪
          this.clearTimers()
          // 将 interview 和 questions 存入 store
          this.$store.commit('interview/SET_INTERVIEW', data)
          this.$store.commit('interview/SET_QUESTIONS', questions)
          this.$message.success('题目生成完成，即将进入面试...')
          this.$router.replace(`/interview/${this.interviewId}/doing`)
        } else {
          // 继续轮询
          this.pollTimer = setTimeout(() => this.poll(), POLL_INTERVAL)
        }
      } catch (err) {
        // 网络/接口错误：先尝试重试，超过重试次数再报错
        this.retryCount = (this.retryCount || 0) + 1
        if (this.retryCount >= 3) {
          this.clearTimers()
          const msg =
            err?.response?.data?.message ||
            err?.response?.data?.msg ||
            err?.message ||
            '获取面试数据失败，请重试'
          this.showError(msg)
        } else {
          // 等待后继续轮询
          this.pollTimer = setTimeout(() => this.poll(), POLL_INTERVAL * 2)
        }
      }
    },

    // ---- 清除所有定时器 ----
    clearTimers() {
      clearTimeout(this.pollTimer)
      clearTimeout(this.timeoutTimer)
      clearTimeout(this.finalTimeoutTimer)
      clearInterval(this.dotTimer)
      clearInterval(this.statusTimer)
    },

    // ---- 显示错误 ----
    showError(msg) {
      this.hasError = true
      this.errorMsg = msg
    },

    // ---- 重试 ----
    handleRetry() {
      this.retryCount = 0
      this.dotIdx = 0
      this.statusIdx = 0
      this.startPolling()
    },

    // ---- 返回配置 ----
    handleBackConfig() {
      this.clearTimers()
      this.$router.push('/interview/config')
    }
  }
}
</script>

<style scoped>
/* 整页深色背景 */
.loading-page {
  min-height: calc(100vh - 90px);
  background: #232f3e;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 40px 20px;
}

/* ===== 加载容器 ===== */
.loading-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 24px;
  width: 100%;
  max-width: 520px;
}

/* ===== Spinner ===== */
.spinner-wrapper {
  position: relative;
  width: 100px;
  height: 100px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.spinner-ring {
  width: 80px;
  height: 80px;
  border-radius: 50%;
  border: 3px solid rgba(255, 153, 0, 0.2);
  border-top-color: #ff9900;
  animation: spin 1s linear infinite;
  position: relative;
  display: flex;
  align-items: center;
  justify-content: center;
}

.spinner-inner {
  width: 56px;
  height: 56px;
  background: rgba(255, 153, 0, 0.12);
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
}

.spinner-icon {
  font-size: 24px;
  animation: counter-spin 1s linear infinite;
}

/* 脉冲光环 */
.pulse-ring {
  position: absolute;
  border-radius: 50%;
  border: 2px solid rgba(255, 153, 0, 0.3);
  animation: pulse 2s ease-out infinite;
}

.pulse-ring-1 {
  width: 90px;
  height: 90px;
  animation-delay: 0s;
}

.pulse-ring-2 {
  width: 100px;
  height: 100px;
  animation-delay: 0.7s;
}

/* ===== 文字 ===== */
.loading-title {
  font-size: 22px;
  font-weight: bold;
  color: #fff;
  letter-spacing: 1px;
}

.loading-subtitle {
  font-size: 14px;
  color: #b8c0cc;
  min-height: 20px;
  transition: opacity 0.3s;
}

/* ===== 进度点 ===== */
.dots {
  display: flex;
  gap: 8px;
  align-items: center;
}

.dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: rgba(255, 153, 0, 0.25);
  transition: background 0.3s, transform 0.3s;
}

.dot.active {
  background: #ff9900;
  transform: scale(1.3);
}

/* ===== 配置摘要卡片 ===== */
.config-summary {
  background: rgba(255, 255, 255, 0.06);
  border: 1px solid rgba(255, 153, 0, 0.25);
  border-radius: 8px;
  padding: 20px 24px;
  width: 100%;
}

.summary-title {
  font-size: 13px;
  color: #ff9900;
  font-weight: bold;
  margin-bottom: 14px;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.summary-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px 16px;
}

.summary-item {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.summary-item.full-width {
  grid-column: 1 / -1;
}

.summary-label {
  font-size: 11px;
  color: #7a8a99;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.summary-value {
  font-size: 14px;
  color: #e8edf2;
  font-weight: 500;
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.summary-value.highlight {
  color: #ff9900;
  font-size: 15px;
}

.focus-tag {
  display: inline-block;
  background: rgba(255, 153, 0, 0.15);
  border: 1px solid rgba(255, 153, 0, 0.3);
  color: #ffb84d;
  font-size: 12px;
  padding: 2px 10px;
  border-radius: 12px;
}

/* ===== 超时提示 ===== */
.timeout-hint {
  font-size: 13px;
  color: #febd69;
  display: flex;
  align-items: center;
  gap: 6px;
  background: rgba(255, 189, 105, 0.08);
  border: 1px solid rgba(255, 189, 105, 0.2);
  padding: 8px 16px;
  border-radius: 20px;
}

/* ===== 错误容器 ===== */
.error-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 16px;
  max-width: 440px;
  width: 100%;
  text-align: center;
}

.error-icon {
  font-size: 52px;
}

.error-title {
  font-size: 20px;
  font-weight: bold;
  color: #fff;
}

.error-msg {
  font-size: 14px;
  color: #b8c0cc;
  line-height: 1.6;
}

.error-actions {
  display: flex;
  gap: 12px;
  margin-top: 8px;
  flex-wrap: wrap;
  justify-content: center;
}

/* ===== 按钮样式 ===== */
.action-btn {
  padding: 10px 28px !important;
  font-size: 14px !important;
  font-weight: bold !important;
  height: auto !important;
}

.action-btn-ghost {
  padding: 10px 28px;
  font-size: 14px;
  font-weight: bold;
  border: 1px solid rgba(255, 255, 255, 0.25);
  background: transparent;
  color: #b8c0cc;
  border-radius: 4px;
  cursor: pointer;
  transition: all 0.2s;
}

.action-btn-ghost:hover {
  border-color: #ff9900;
  color: #ff9900;
  background: rgba(255, 153, 0, 0.08);
}

/* ===== Element UI 覆盖 ===== */
::v-deep .el-button--primary {
  background: #ff9900;
  border-color: #ff9900;
  color: #111;
}

::v-deep .el-button--primary:hover,
::v-deep .el-button--primary:focus {
  background: #f3a847;
  border-color: #f3a847;
  color: #111;
}

/* ===== 关键帧 ===== */
@keyframes spin {
  to { transform: rotate(360deg); }
}

@keyframes counter-spin {
  to { transform: rotate(-360deg); }
}

@keyframes pulse {
  0% {
    transform: scale(1);
    opacity: 0.6;
  }
  100% {
    transform: scale(1.6);
    opacity: 0;
  }
}
</style>
