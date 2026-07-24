<template>
  <div class="practice-page" v-loading="pageLoading">
    <div v-if="question">
      <!-- 题目卡片 -->
      <el-card shadow="hover" class="question-card">
        <div slot="header" class="card-header">
          <div class="q-tags">
            <el-tag type="primary" effect="plain">{{ question.jobTitle }}</el-tag>
            <el-tag :type="diffTagType(question.difficulty)" class="ml-6">{{ diffLabel(question.difficulty) }}</el-tag>
            <el-tag v-for="tag in (question.tags || [])" :key="tag" type="info" class="ml-6">{{ tag }}</el-tag>
          </div>
          <el-button
            :type="collected ? 'warning' : 'default'"
            size="small"
            :icon="collected ? 'el-icon-star-on' : 'el-icon-star-off'"
            @click="toggleCollect"
          >{{ collected ? '已收藏' : '收藏题目' }}</el-button>
        </div>
        <div class="q-content">{{ question.content }}</div>
      </el-card>

      <!-- 答题区 -->
      <el-card shadow="never" class="answer-card" v-if="!reviewResult">
        <div slot="header" class="card-header">
          <span class="card-title">✏️ 我的作答</span>
        </div>
        <el-input
          v-model="answer"
          type="textarea"
          :rows="10"
          placeholder="请在这里输入你的答案，支持代码块格式..."
          resize="vertical"
        />
        <div class="answer-actions">
          <el-button
            type="primary"
            :loading="submitting"
            :disabled="!answer.trim()"
            @click="submitAnswer"
          >提交答案获取 AI 点评</el-button>
          <el-button @click="$router.push('/questions')">返回题库</el-button>
        </div>
      </el-card>

      <!-- AI 点评结果 -->
      <div v-if="reviewResult">
        <el-card shadow="hover" class="review-card">
          <div slot="header" class="card-header">
            <span class="card-title">🤖 AI 点评结果</span>
            <el-button type="text" @click="resetAnswer">重新作答</el-button>
          </div>

          <!-- 得分 -->
          <div class="score-row">
            <span class="score-label">得分：</span>
            <div class="score-bar-wrap">
              <score-bar :score="reviewResult.score || 0" />
            </div>
            <span class="score-num" :style="{ color: scoreColor(reviewResult.score) }">
              {{ reviewResult.score }}
            </span>
          </div>

          <!-- 优点 -->
          <div class="review-block" v-if="reviewResult.pros && reviewResult.pros.length">
            <div class="review-title pros-title">✅ 优点</div>
            <ul class="review-list">
              <li v-for="(p, i) in reviewResult.pros" :key="i">{{ p }}</li>
            </ul>
          </div>

          <!-- 不足 -->
          <div class="review-block" v-if="reviewResult.cons && reviewResult.cons.length">
            <div class="review-title cons-title">⚠️ 不足</div>
            <ul class="review-list">
              <li v-for="(c, i) in reviewResult.cons" :key="i">{{ c }}</li>
            </ul>
          </div>

          <!-- 我的作答 -->
          <div class="review-block">
            <div class="review-title">📝 我的作答</div>
            <div class="my-answer">{{ answer }}</div>
          </div>

          <!-- 参考答案（折叠） -->
          <el-collapse class="ref-collapse">
            <el-collapse-item title="📖 查看参考答案" name="ref">
              <markdown-render :content="reviewResult.referenceAnswer" />
            </el-collapse-item>
          </el-collapse>

          <div class="bottom-actions">
            <el-button type="primary" @click="resetAnswer">再练一次</el-button>
            <el-button @click="$router.push('/questions')">返回题库</el-button>
          </div>
        </el-card>
      </div>
    </div>

    <el-empty v-if="!pageLoading && !question" description="题目加载失败">
      <el-button @click="$router.push('/questions')">返回题库</el-button>
    </el-empty>
  </div>
</template>

<script>
import { getQuestion, practiceQuestion, collectQuestion } from '@/api/question'
import ScoreBar from '@/components/common/ScoreBar.vue'
import MarkdownRender from '@/components/common/MarkdownRender.vue'

export default {
  name: 'QuestionPractice',
  components: { ScoreBar, MarkdownRender },
  data() {
    return {
      pageLoading: true,
      question: null,
      answer: '',
      submitting: false,
      reviewResult: null,
      collected: false
    }
  },
  mounted() {
    this.loadQuestion()
  },
  methods: {
    async loadQuestion() {
      const id = this.$route.params.id
      try {
        const res = await getQuestion(id)
        this.question = res.data || res
        this.collected = !!(this.question.collected)
      } catch (e) {
        this.$message.error('题目加载失败')
      } finally {
        this.pageLoading = false
      }
    },
    async submitAnswer() {
      if (!this.answer.trim()) return
      this.submitting = true
      const id = this.$route.params.id
      try {
        const res = await practiceQuestion(id, { answer: this.answer })
        this.reviewResult = res.data || res
      } catch (e) {
        this.$message.error('提交失败，请重试')
      } finally {
        this.submitting = false
      }
    },
    async toggleCollect() {
      const id = this.$route.params.id
      try {
        await collectQuestion(id)
        this.collected = !this.collected
        this.$message.success(this.collected ? '收藏成功' : '已取消收藏')
      } catch (e) {
        this.$message.error('操作失败')
      }
    },
    resetAnswer() {
      this.answer = ''
      this.reviewResult = null
    },
    diffTagType(d) {
      const map = { junior: 'success', middle: 'warning', senior: 'danger' }
      return map[d] || 'info'
    },
    diffLabel(d) {
      const map = { junior: '初级', middle: '中级', senior: '高级' }
      return map[d] || d || '—'
    },
    scoreColor(score) {
      if (score >= 80) return '#67c23a'
      if (score >= 60) return '#ff9900'
      return '#f56c6c'
    }
  }
}
</script>

<style scoped>
.practice-page { max-width: 800px; margin: 0 auto; padding: 20px 16px; }
.question-card { margin-bottom: 16px; }
.card-header { display: flex; align-items: center; justify-content: space-between; flex-wrap: wrap; gap: 8px; }
.card-title { font-size: 15px; font-weight: bold; color: #111; border-left: 4px solid #ff9900; padding-left: 10px; }
.q-tags { display: flex; flex-wrap: wrap; gap: 6px; }
.ml-6 { margin-left: 6px; }
.q-content { font-size: 16px; color: #222; line-height: 1.8; padding: 8px 0; }
.answer-card { margin-bottom: 16px; }
.answer-actions { display: flex; gap: 12px; margin-top: 16px; }
.review-card { margin-bottom: 16px; }
.score-row { display: flex; align-items: center; gap: 12px; margin-bottom: 20px; }
.score-label { font-size: 15px; font-weight: bold; color: #333; }
.score-bar-wrap { flex: 1; }
.score-num { font-size: 22px; font-weight: bold; min-width: 48px; text-align: right; }
.review-block { margin-bottom: 16px; }
.review-title { font-size: 14px; font-weight: bold; margin-bottom: 8px; }
.pros-title { color: #52c41a; }
.cons-title { color: #ff9900; }
.review-list { padding-left: 20px; margin: 0; }
.review-list li { font-size: 14px; color: #444; margin-bottom: 4px; line-height: 1.6; }
.my-answer { font-size: 14px; color: #444; background: #fafafa; padding: 10px 12px; border-radius: 6px; white-space: pre-wrap; line-height: 1.6; }
.ref-collapse { margin-top: 8px; }
.bottom-actions { display: flex; gap: 12px; margin-top: 20px; padding-top: 16px; border-top: 1px solid #f0f0f0; }
</style>
