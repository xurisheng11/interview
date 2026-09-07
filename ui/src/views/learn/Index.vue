<template>
  <div class="layout">
    <Sidebar :items="sidebarItems" />

    <div class="main-content">
      <!-- 冲刺计划入口横幅 -->
      <div class="sprint-banner" @click="$router.push('/learn/plan')">
        <div class="sprint-banner-text">
          <strong>🏃 定制专属冲刺计划</strong>
          <span>设定目标岗位和冲刺天数（30/60/90天），AI 教练为你生成分阶段、可打卡的备考计划</span>
        </div>
        <el-button size="small" class="sprint-banner-btn">去定制 →</el-button>
      </div>

      <!-- 顶部阶段导航 -->
      <div class="stage-nav">
        <div
          v-for="(stage, idx) in stages"
          :key="idx"
          class="stage-item"
          :class="{
            'stage-active': currentStage === idx,
            'stage-done': currentStage > idx,
            'stage-locked': currentStage < idx
          }"
          @click="currentStage = idx"
        >
          <div class="stage-num">{{ idx + 1 }}</div>
          <div class="stage-label">{{ stage.label }}</div>
        </div>
      </div>

      <!-- ====== 阶段1: 了解题型类目 ====== -->
      <div v-if="currentStage === 0" class="stage-content">
        <div class="stage-header">
          <h2>📋 面试题型全览</h2>
          <p class="stage-desc">
            面试考察通常分为 <strong>5 大类题型</strong>。先了解每类题型的考察重点，
            再逐类突破效率最高。下方为你推荐备考顺序。
          </p>
        </div>

        <!-- 题型分类卡片（按备考顺序） -->
        <div class="category-grid">
          <div
            v-for="(cat, idx) in categories"
            :key="cat.id"
            class="category-card"
            :class="{ 'cat-recommended': idx === 0 }"
            @click="startCategory(cat)"
          >
            <div class="cat-header">
              <span class="cat-icon">{{ cat.icon }}</span>
              <div>
                <div class="cat-name">{{ cat.name }}</div>
                <div class="cat-count">{{ cat.desc }}</div>
              </div>
              <div v-if="idx === 0" class="cat-badge">推荐先练</div>
            </div>
            <div class="cat-questions">
              <el-tag
                v-for="q in cat.sampleQuestions.slice(0, 3)"
                :key="q"
                size="mini"
                type="info"
                class="cat-q-tag"
              >{{ q }}</el-tag>
              <span v-if="cat.sampleQuestions.length > 3" class="cat-q-more">...</span>
            </div>
            <div class="cat-footer">
              <span class="cat-difficulty">{{ cat.difficulty }}</span>
              <el-button size="mini" type="primary" class="cat-btn">
                {{ getCatProgress(cat.id) > 0 ? '继续练习' : '开始练习' }} →
              </el-button>
            </div>
          </div>
        </div>

        <!-- 备考路线图 -->
        <el-card shadow="hover" class="roadmap-card">
          <div slot="header" class="card-title">🗺️ 推荐备考顺序</div>
          <div class="roadmap-flow">
            <div v-for="(cat, idx) in categories" :key="cat.id" class="roadmap-step">
              <div class="step-num">{{ idx + 1 }}</div>
              <div class="step-name">{{ cat.name }}</div>
              <div class="step-arrow" v-if="idx < categories.length - 1">→</div>
            </div>
          </div>
          <div class="roadmap-tip">
            💡 建议：先把每一类练到 60 分以上，再进入下一类；全部准备完后做一次完整模拟面试检验效果。
          </div>
        </el-card>
      </div>

      <!-- ====== 阶段2: 单点突破（分类练习）====== -->
      <div v-if="currentStage === 1" class="stage-content">
        <div class="stage-header">
          <h2>🎯 分类专项练习</h2>
          <p class="stage-desc">
            选择一个知识点类别，深入练习该类所有题型，直到掌握。
          </p>
        </div>

        <!-- 当前选中类别 -->
        <div v-if="selectedCategory" class="selected-cat-banner">
          <span>当前类别：</span>
          <strong>{{ selectedCategory.icon }} {{ selectedCategory.name }}</strong>
          <span class="cat-progress-label">
            进度：{{ getCatProgress(selectedCategory.id) }}%
            <el-progress :percentage="getCatProgress(selectedCategory.id)" :stroke-width="8" class="inline-progress" />
          </span>
          <el-button size="mini" @click="selectedCategory = null">切换类别</el-button>
        </div>

        <!-- 类别选择 -->
        <div v-if="!selectedCategory" class="category-select-grid">
          <div
            v-for="cat in categories"
            :key="cat.id"
            class="cat-select-card"
            :class="{ selected: selectedCategory?.id === cat.id }"
            @click="selectedCategory = cat"
          >
            <div class="cat-select-icon">{{ cat.icon }}</div>
            <div class="cat-select-name">{{ cat.name }}</div>
            <div class="cat-select-progress">
              <el-progress
                type="circle"
                :percentage="getCatProgress(cat.id)"
                :width="60"
                :stroke-width="6"
              />
            </div>
            <div class="cat-select-done">
              <span v-if="getCatProgress(cat.id) >= 60" class="done-badge">✅ 可解锁下一类</span>
              <span v-else class="pending-badge">进行中</span>
            </div>
          </div>
        </div>

        <!-- 该类别的题库列表 -->
        <div v-else class="cat-practice-area">
          <div class="practice-header">
            <el-tag type="primary" size="medium">{{ selectedCategory.name }} 专项</el-tag>
            <div class="practice-header-right">
              <el-tag :type="practiceDifficulty === 'easy' ? 'success' : 'info'" size="small">
                难度: {{ practiceDifficultyLabel }}
              </el-tag>
              <el-button size="mini" @click="practiceDifficulty = practiceDifficulty === 'easy' ? 'medium' : 'easy'">
                切换难度
              </el-button>
            </div>
          </div>

          <!-- 练习说明 -->
          <el-card shadow="never" class="practice-tip-card">
            <div class="practice-tip">
              <div class="tip-icon">💡</div>
              <div class="tip-text">
                <strong>{{ selectedCategory.name }}</strong> 通常考{{ selectedCategory.desc }}。<br />
                练习策略：先看参考思路，再自己组织语言回答，重点关注"有没有说到点上"。
              </div>
            </div>
          </el-card>

          <!-- 快速开始练习按钮 -->
          <div class="practice-start-zone">
            <el-button type="primary" size="large" class="practice-start-btn" :loading="generating" @click="startCategoryPractice">
              🚀 开始 {{ selectedCategory.name }} 练习
            </el-button>
            <span class="practice-start-hint">AI 将生成 5 道 {{ selectedCategory.name }} 相关的题目</span>
          </div>

          <!-- 关联题目示例 -->
          <div v-if="practiceQuestions.length" class="practice-questions-list">
            <div class="practice-questions-header">
              <span>本轮练习题预览</span>
              <span class="practice-q-count">{{ practiceQuestions.length }} 道</span>
            </div>
            <div
              v-for="(q, idx) in practiceQuestions"
              :key="idx"
              class="practice-q-item"
              @click="goPracticeQuestion(q)"
            >
              <span class="pq-num">Q{{ idx + 1 }}</span>
              <span class="pq-content">{{ q.content }}</span>
              <el-tag size="mini" :type="diffTagType(q.difficulty)">{{ diffLabel(q.difficulty) }}</el-tag>
            </div>
            <div class="practice-questions-footer">
              <el-button type="primary" @click="startPracticeSession">开始作答 →</el-button>
            </div>
          </div>
        </div>
      </div>

      <!-- ====== 阶段3: 串联练习 ====== -->
      <div v-if="currentStage === 2" class="stage-content">
        <div class="stage-header">
          <h2>🔗 串联强化练习</h2>
          <p class="stage-desc">
            把多个类别的题目混合在一起练习，模拟真实面试中题目随机出现的场景，
            训练快速切换思维的能力。
          </p>
        </div>

        <el-card shadow="hover" class="mix-config-card">
          <div slot="header" class="card-title">选择混合类别</div>
          <div class="mix-categories">
            <el-checkbox
              v-for="cat in categories"
              :key="cat.id"
              v-model="mixSelectedCats"
              :label="cat.id"
              class="mix-cat-check"
            >
              {{ cat.icon }} {{ cat.name }}
            </el-checkbox>
          </div>
          <div class="mix-config-row">
            <span class="config-label">每类题目数：</span>
            <el-input-number v-model="mixCountPerCat" :min="1" :max="5" size="small" />
            <span class="mix-total">共约 {{ mixSelectedCats.length * mixCountPerCat }} 道题</span>
          </div>
          <el-button
            type="primary"
            class="mix-start-btn"
            :disabled="mixSelectedCats.length === 0"
            :loading="mixGenerating"
            @click="startMixedPractice"
          >
            🔀 开始串联练习
          </el-button>
        </el-card>
      </div>

      <!-- ====== 阶段4: 综合模拟面试 ====== -->
      <div v-if="currentStage === 3" class="stage-content">
        <div class="stage-header">
          <h2>🎓 综合模拟面试</h2>
          <p class="stage-desc">
            恭喜你来到最后阶段！完整走一遍面试流程，查漏补缺。
          </p>
        </div>

        <!-- 学习成果总览 -->
        <el-card shadow="hover" class="summary-card">
          <div slot="header" class="card-title">📊 学习成果总览</div>
          <div class="summary-grid">
            <div class="summary-item">
              <div class="summary-num">{{ totalPracticeCount }}</div>
              <div class="summary-label">累计练习题数</div>
            </div>
            <div class="summary-item">
              <div class="summary-num">{{ avgScore }}分</div>
              <div class="summary-label">平均得分</div>
            </div>
            <div class="summary-item">
              <div class="summary-num">{{ masteredCategories }}</div>
              <div class="summary-label">已掌握类别</div>
            </div>
            <div class="summary-item">
              <div class="summary-num">{{ weakestCategory }}</div>
              <div class="summary-label">待加强类别</div>
            </div>
          </div>
        </el-card>

        <!-- 发起模拟面试入口 -->
        <el-card shadow="hover" class="interview-entry-card">
          <div class="interview-entry">
            <div class="entry-text">
              <h3>🎯 发起完整模拟面试</h3>
              <p>综合所有类别，随机出题，模拟真实面试环境。</p>
              <el-tag v-if="totalPracticeCount === 0" type="warning" size="small">
                建议先完成前3个阶段的练习再进行模拟面试
              </el-tag>
              <el-tag v-else-if="avgScore < 60" type="warning" size="small">
                平均分不足60分，建议先巩固薄弱环节
              </el-tag>
              <el-tag v-else type="success" size="small">
                你已准备好，开始模拟面试！
              </el-tag>
            </div>
            <el-button type="primary" size="large" class="entry-btn" @click="$router.push('/interview/config')">
              🚀 发起模拟面试
            </el-button>
          </div>
        </el-card>
      </div>
    </div>
  </div>
</template>

<script>
import Sidebar from '@/components/layout/Sidebar.vue'
import { getQuestions } from '@/api/question'
import { createInterview } from '@/api/interview'

export default {
  name: 'LearningCenter',
  components: { Sidebar },
  data() {
    return {
      currentStage: 0,
      // 五大题型分类（按备考顺序）
      categories: [
        {
          id: 'basic',
          name: '基础知识',
          icon: '📖',
          desc: '概念、原理、定义类问题',
          difficulty: '🟢 入门',
          sampleQuestions: [
            '解释一下什么是 RESTful API？',
            'HTTP 和 HTTPS 的区别是什么？',
            '什么是线程与进程的区别？',
            '数据库事务的四大特性是什么？'
          ]
        },
        {
          id: 'project',
          name: '项目经验',
          icon: '💼',
          desc: '围绕简历项目深挖技术细节',
          difficulty: '🟡 中等',
          sampleQuestions: [
            '这个项目最大的技术难点是什么？你怎么解决的？',
            '为什么选择这个技术栈？有没有考虑过其他方案？',
            '项目中遇到的最棘手的 bug 是？怎么排查的？'
          ]
        },
        {
          id: 'algorithm',
          name: '算法与数据结构',
          icon: '🧮',
          desc: '逻辑思维和编码能力',
          difficulty: '🟡 中等',
          sampleQuestions: [
            '如何反转一个单向链表？',
            '请用最优算法查找数组中和为目标值的两个数',
            '什么是哈希表？哈希冲突有哪些解决方案？'
          ]
        },
        {
          id: 'system_design',
          name: '系统设计',
          icon: '🏗️',
          desc: '整体架构和扩展性思考',
          difficulty: '🔴 进阶',
          sampleQuestions: [
            '如何设计一个短链接服务？',
            '如果用户量增长100倍，系统要如何改造？',
            '如何设计一个高可用的消息队列？'
          ]
        },
        {
          id: 'behavior',
          name: '行为面试',
          icon: '🤝',
          desc: '团队协作和职业素养',
          difficulty: '🟢 入门',
          sampleQuestions: [
            '最有成就感的一件事是什么？',
            '和同事产生分歧时你怎么处理？',
            '你的职业规划是什么？'
          ]
        }
      ],

      // 阶段2: 分类练习
      selectedCategory: null,
      practiceDifficulty: 'easy',
      practiceQuestions: [],
      generating: false,

      // 阶段3: 串联练习
      mixSelectedCats: ['basic', 'project'],
      mixCountPerCat: 2,
      mixGenerating: false,

      // 阶段4: 统计数据（从本地存储读取）
      practiceHistory: []
    }
  },

  computed: {
    sidebarItems() {
      return [
        {
          title: '学习中心',
          children: [
            { icon: '📚', label: '学习路径', path: '/learn' },
            { icon: '🏃', label: '冲刺计划', path: '/learn/plan' },
            { icon: '📝', label: '题库练习', path: '/questions' },
            { icon: '🚀', label: '发起面试', path: '/interview/config' }
          ]
        }
      ]
    },
    practiceDifficultyLabel() {
      const map = { easy: '🟢 简单', medium: '🟡 中等' }
      return map[this.practiceDifficulty] || '简单'
    },
    totalPracticeCount() {
      return this.practiceHistory.length
    },
    avgScore() {
      if (!this.practiceHistory.length) return 0
      const sum = this.practiceHistory.reduce((acc, h) => acc + (h.score || 0), 0)
      return Math.round(sum / this.practiceHistory.length)
    },
    masteredCategories() {
      // 已掌握：进度 >= 60% 的类别数
      return this.categories.filter(c => this.getCatProgress(c.id) >= 60).length
    },
    weakestCategory() {
      let min = null
      let minScore = Infinity
      this.categories.forEach(c => {
        const p = this.getCatProgress(c.id)
        if (p < minScore && this.practiceHistory.some(h => h.category === c.id)) {
          minScore = p
          min = c.name
        }
      })
      return min || '—'
    }
  },

  created() {
    this.loadPracticeHistory()
    // 根据历史数据自动推断当前阶段
    this.inferCurrentStage()
  },

  methods: {
    // 从本地存储加载练习历史
    loadPracticeHistory() {
      try {
        const raw = localStorage.getItem('practice_history')
        this.practiceHistory = raw ? JSON.parse(raw) : []
      } catch (e) {
        this.practiceHistory = []
      }
    },

    // 保存练习结果到历史
    savePracticeResult(score, category) {
      this.practiceHistory.push({ score, category, date: new Date().toISOString() })
      localStorage.setItem('practice_history', JSON.stringify(this.practiceHistory))
    },

    // 计算某类别进度（基于历史得分）
    getCatProgress(categoryId) {
      const catHistory = this.practiceHistory.filter(h => h.category === categoryId)
      if (!catHistory.length) return 0
      const avg = catHistory.reduce((s, h) => s + (h.score || 0), 0) / catHistory.length
      // 映射到 0-100
      return Math.min(100, Math.round(avg))
    },

    // 根据历史数据推断用户当前处于哪个阶段
    inferCurrentStage() {
      const done = this.practiceHistory.length
      if (done === 0) this.currentStage = 0
      else if (done < 3) this.currentStage = 1
      else if (done < 10) this.currentStage = 2
      else this.currentStage = 3
    },

    // 点击类别卡片，开始分类练习
    startCategory(cat) {
      this.selectedCategory = cat
      this.currentStage = 1
    },

    // 发起分类专项练习
    async startCategoryPractice() {
      if (!this.selectedCategory) return
      this.generating = true
      this.practiceQuestions = []
      try {
        // 调用后端生成该类别的练习题
        const res = await createInterview({
          jobTitle: '通用岗位',
          difficulty: this.practiceDifficulty,
          experience: 'fresh',
          round: 'round1',
          focusAreas: [this.selectedCategory.id],
          remark: `请只生成 ${this.selectedCategory.name} 相关的面试题，共5道`
        })
        const d = res.data || res
        const questions = d.questions || []
        if (questions.length) {
          this.practiceQuestions = questions
          // 存入本地，标记类别
          this.practiceQuestions.forEach(q => {
            q._category = this.selectedCategory.id
          })
        }
      } catch (e) {
        this.$message.error('题目生成失败，请重试')
      } finally {
        this.generating = false
      }
    },

    // 开始作答（跳转答题页）
    async startPracticeSession() {
      if (!this.practiceQuestions.length) return
      try {
        const res = await createInterview({
          jobTitle: '通用岗位',
          difficulty: this.practiceDifficulty,
          experience: 'fresh',
          round: 'round1',
          focusAreas: [this.selectedCategory.id],
          remark: `单题练习模式，只生成 ${this.selectedCategory.name} 相关题目`
        })
        const d = res.data || res
        const interviewId = d.id || d.interviewId || d.data?.id
        if (interviewId) {
          this.$store.commit('interview/SET_CURRENT_ID', interviewId)
          this.$store.commit('interview/SET_QUESTIONS', this.practiceQuestions)
          this.$store.commit('interview/SET_INTERVIEW_MODE', 'text')
          this.$router.push('/interview/loading')
        }
      } catch (e) {
        this.$message.error('发起练习失败')
      }
    },

    // 点击单题预览
    goPracticeQuestion(q) {
      this.$router.push(`/questions`)
    },

    // 串联练习
    async startMixedPractice() {
      this.mixGenerating = true
      try {
        const focusAreas = this.mixSelectedCats
        const res = await createInterview({
          jobTitle: '通用岗位',
          difficulty: 'medium',
          experience: 'fresh',
          round: 'round1',
          focusAreas,
          remark: `串联练习模式，混合 ${focusAreas.join('、')} 类别，各${this.mixCountPerCat}道题`
        })
        const d = res.data || res
        const interviewId = d.id || d.interviewId || d.data?.id
        if (interviewId) {
          this.$store.commit('interview/SET_CURRENT_ID', interviewId)
          this.$store.commit('interview/SET_INTERVIEW_MODE', 'text')
          this.$message.success('串联练习配置成功，开始生成...')
          this.$router.push('/interview/loading')
        }
      } catch (e) {
        this.$message.error('发起串联练习失败')
      } finally {
        this.mixGenerating = false
      }
    },

    diffTagType(d) {
      const map = { easy: 'success', medium: 'warning', hard: 'danger' }
      return map[d] || 'info'
    },
    diffLabel(d) {
      const map = { easy: '🟢 简单', medium: '🟡 中等', hard: '🔴 困难' }
      return map[d] || d || '—'
    }
  }
}
</script>

<style scoped>
.layout {
  display: flex;
  min-height: calc(100vh - 90px);
  width: 100%;
}

.main-content {
  flex: 1;
  padding: 20px;
  background: #fafafa;
  overflow: auto;
  min-width: 0;
}

/* ===== 冲刺计划横幅 ===== */
.sprint-banner {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  background: linear-gradient(135deg, #131921 0%, #232f3e 60%, #3a4a5c 100%);
  border-radius: 8px;
  padding: 14px 20px;
  margin-bottom: 16px;
  cursor: pointer;
  transition: box-shadow 0.2s;
}
.sprint-banner:hover { box-shadow: 0 4px 14px rgba(19, 25, 33, 0.35); }
.sprint-banner-text { display: flex; flex-direction: column; gap: 3px; min-width: 0; }
.sprint-banner-text strong { font-size: 15px; color: #ff9900; }
.sprint-banner-text span { font-size: 12px; color: #aab7c4; }
.sprint-banner-btn {
  background: #ff9900 !important;
  border-color: #ff9900 !important;
  color: #111 !important;
  font-weight: bold !important;
  flex-shrink: 0;
}

/* ===== 阶段导航 ===== */
.stage-nav {
  display: flex;
  gap: 8px;
  margin-bottom: 24px;
  background: #fff;
  padding: 12px 16px;
  border-radius: 8px;
  box-shadow: 0 1px 4px rgba(0,0,0,0.08);
}

.stage-item {
  flex: 1;
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 16px;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s;
  border: 2px solid transparent;
  background: #f5f5f5;
}

.stage-item:hover { background: #f0f0f0; }

.stage-item.stage-active {
  background: #fff7e6;
  border-color: #ff9900;
}

.stage-item.stage-done {
  background: #f6ffed;
  border-color: #52c41a;
}

.stage-item.stage-locked {
  opacity: 0.5;
  cursor: not-allowed;
}

.stage-num {
  width: 28px;
  height: 28px;
  border-radius: 50%;
  background: #ddd;
  color: #fff;
  font-weight: bold;
  font-size: 14px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.stage-active .stage-num { background: #ff9900; }
.stage-done .stage-num { background: #52c41a; }

.stage-label {
  font-size: 14px;
  font-weight: bold;
  color: #555;
}

.stage-active .stage-label { color: #a05c00; }

/* ===== 阶段内容通用 ===== */
.stage-content { }
.stage-header { margin-bottom: 20px; }
.stage-header h2 {
  font-size: 20px;
  color: #111;
  margin: 0 0 8px;
}
.stage-desc {
  font-size: 14px;
  color: #666;
  line-height: 1.7;
  margin: 0;
}

.card-title {
  font-size: 15px;
  font-weight: bold;
  color: #111;
  border-left: 4px solid #ff9900;
  padding-left: 10px;
}

/* ===== 阶段1: 题型概览 ===== */
.category-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 16px;
  margin-bottom: 20px;
}

.category-card {
  background: #fff;
  border: 1px solid #e8e8e8;
  border-radius: 10px;
  padding: 16px;
  cursor: pointer;
  transition: all 0.2s;
  position: relative;
}

.category-card:hover {
  border-color: #ff9900;
  box-shadow: 0 4px 12px rgba(255, 153, 0, 0.15);
  transform: translateY(-2px);
}

.cat-recommended {
  border-color: #ff9900;
  background: linear-gradient(135deg, #fff9f0, #fff);
}

.cat-header {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  margin-bottom: 12px;
}

.cat-icon { font-size: 32px; flex-shrink: 0; }

.cat-name { font-size: 16px; font-weight: bold; color: #222; }
.cat-count { font-size: 12px; color: #888; margin-top: 2px; }

.cat-badge {
  position: absolute;
  top: 10px;
  right: 10px;
  background: #ff9900;
  color: #111;
  font-size: 11px;
  font-weight: bold;
  padding: 2px 8px;
  border-radius: 10px;
}

.cat-questions {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
  margin-bottom: 12px;
}

.cat-q-tag { font-size: 11px; }
.cat-q-more { font-size: 12px; color: #999; align-self: center; }

.cat-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.cat-difficulty { font-size: 12px; color: #888; }
.cat-btn { background: #ff9900 !important; border-color: #ff9900 !important; color: #111 !important; }

/* 路线图 */
.roadmap-card { margin-top: 8px; }

.roadmap-flow {
  display: flex;
  align-items: center;
  gap: 0;
  flex-wrap: wrap;
  margin-bottom: 16px;
}

.roadmap-step {
  display: flex;
  align-items: center;
  gap: 6px;
}

.step-num {
  width: 24px;
  height: 24px;
  border-radius: 50%;
  background: #ff9900;
  color: #111;
  font-size: 12px;
  font-weight: bold;
  display: flex;
  align-items: center;
  justify-content: center;
}

.step-name { font-size: 13px; font-weight: bold; color: #333; white-space: nowrap; }
.step-arrow { font-size: 18px; color: #ccc; margin: 0 8px; }

.roadmap-tip {
  font-size: 13px;
  color: #666;
  background: #fff7e6;
  padding: 10px 14px;
  border-radius: 6px;
  border-left: 3px solid #ff9900;
  line-height: 1.7;
}

/* ===== 阶段2: 分类练习 ===== */
.selected-cat-banner {
  display: flex;
  align-items: center;
  gap: 16px;
  background: #fff;
  padding: 12px 16px;
  border-radius: 8px;
  margin-bottom: 16px;
  font-size: 14px;
  flex-wrap: wrap;
}

.cat-progress-label {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-left: auto;
  font-size: 13px;
  color: #555;
}

.inline-progress { width: 120px; }

.category-select-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
  gap: 14px;
}

.cat-select-card {
  background: #fff;
  border: 2px solid #e8e8e8;
  border-radius: 10px;
  padding: 16px;
  cursor: pointer;
  transition: all 0.2s;
  text-align: center;
}

.cat-select-card:hover { border-color: #ff9900; transform: translateY(-2px); }
.cat-select-card.selected { border-color: #ff9900; background: #fff9f0; }
.cat-select-icon { font-size: 36px; margin-bottom: 8px; }
.cat-select-name { font-size: 15px; font-weight: bold; color: #333; margin-bottom: 10px; }
.cat-select-done { margin-top: 8px; font-size: 12px; }
.done-badge { color: #52c41a; font-weight: bold; }
.pending-badge { color: #888; }

.practice-tip-card { margin-bottom: 16px; }
.practice-tip { display: flex; align-items: flex-start; gap: 10px; }
.tip-icon { font-size: 24px; flex-shrink: 0; }
.tip-text { font-size: 14px; color: #444; line-height: 1.7; }

.practice-start-zone {
  text-align: center;
  padding: 32px;
  background: #fff;
  border-radius: 10px;
  border: 2px dashed #ff9900;
  margin-bottom: 20px;
}

.practice-start-btn {
  background: #ff9900 !important;
  border-color: #ff9900 !important;
  color: #111 !important;
  font-size: 16px !important;
  font-weight: bold !important;
  padding: 12px 40px !important;
  height: auto !important;
}

.practice-start-hint {
  display: block;
  font-size: 13px;
  color: #999;
  margin-top: 8px;
}

.practice-questions-list {
  background: #fff;
  border-radius: 8px;
  overflow: hidden;
}

.practice-questions-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  background: #fafafa;
  border-bottom: 1px solid #f0f0f0;
  font-size: 14px;
  font-weight: bold;
  color: #333;
}

.practice-q-count { color: #ff9900; font-size: 13px; }

.practice-q-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 12px 16px;
  border-bottom: 1px solid #f5f5f5;
  cursor: pointer;
  transition: background 0.15s;
}

.practice-q-item:hover { background: #fff9f0; }
.practice-q-item:last-of-type { border-bottom: none; }

.pq-num { font-weight: bold; color: #ff9900; min-width: 28px; font-size: 13px; }
.pq-content { flex: 1; font-size: 14px; color: #333; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }

.practice-questions-footer {
  padding: 12px 16px;
  text-align: center;
  border-top: 1px solid #f0f0f0;
}

/* ===== 阶段3: 串联练习 ===== */
.mix-config-card { }

.mix-categories {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  margin-bottom: 16px;
}

.mix-cat-check { font-size: 14px; }

.mix-config-row {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 16px;
}

.config-label { font-size: 14px; color: #555; }
.mix-total { font-size: 13px; color: #ff9900; font-weight: bold; }

.mix-start-btn {
  background: #ff9900 !important;
  border-color: #ff9900 !important;
  color: #111 !important;
  font-weight: bold !important;
}

/* ===== 阶段4: 综合模拟 ===== */
.summary-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
  text-align: center;
}

.summary-item { padding: 16px; }
.summary-num { font-size: 32px; font-weight: bold; color: #ff9900; }
.summary-label { font-size: 13px; color: #888; margin-top: 4px; }

.interview-entry-card { margin-top: 16px; }
.interview-entry {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 24px;
  flex-wrap: wrap;
}

.entry-text h3 { margin: 0 0 6px; font-size: 18px; color: #111; }
.entry-text p { margin: 0 0 10px; font-size: 14px; color: #666; }

.entry-btn {
  background: #ff9900 !important;
  border-color: #ff9900 !important;
  color: #111 !important;
  font-weight: bold !important;
  font-size: 15px !important;
  padding: 12px 32px !important;
  height: auto !important;
  flex-shrink: 0;
}

@media (max-width: 768px) {
  .stage-nav { flex-direction: column; }
  .summary-grid { grid-template-columns: repeat(2, 1fr); }
  .category-grid { grid-template-columns: 1fr; }
  .category-select-grid { grid-template-columns: repeat(2, 1fr); }
  .interview-entry { flex-direction: column; }
  .entry-btn { width: 100%; }
}
</style>
