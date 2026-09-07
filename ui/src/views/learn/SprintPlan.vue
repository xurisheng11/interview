<template>
  <div class="layout">
    <Sidebar :items="sidebarItems" />

    <div class="main-content" v-loading="loading">
      <!-- ====== 未定制：定制表单 ====== -->
      <div v-if="!loading && !plan" class="customize-area">
        <div class="customize-header">
          <h2>🏃 定制我的冲刺计划</h2>
          <p class="customize-desc">
            告诉我们你的目标和时间安排，AI 教练将为你量身定制一份分阶段、可打卡的面试冲刺计划。
          </p>
        </div>

        <el-card shadow="never" class="form-card">
          <el-form label-width="110px" label-position="left">
            <el-form-item label="🎯 目标岗位" required>
              <el-input v-model="form.jobTitle" placeholder="如：Go 后端开发工程师" clearable class="form-input" />
            </el-form-item>

            <el-form-item label="🏢 目标公司">
              <el-input v-model="form.targetCompany" placeholder="选填，如：字节跳动" clearable class="form-input" />
            </el-form-item>

            <el-form-item label="📅 冲刺天数" required>
              <div class="days-options">
                <div
                  v-for="d in daysOptions"
                  :key="d.value"
                  class="days-item"
                  :class="{ selected: form.totalDays === d.value && !customDays }"
                  @click="selectDays(d.value)"
                >
                  <div class="days-num">{{ d.value }}天</div>
                  <div class="days-desc">{{ d.desc }}</div>
                </div>
                <div class="days-item days-custom" :class="{ selected: customDays }" @click="customDays = true">
                  <div class="days-num">自定义</div>
                  <el-input-number
                    v-if="customDays"
                    v-model="form.totalDays"
                    :min="7"
                    :max="180"
                    size="mini"
                    @click.native.stop
                  />
                  <div v-else class="days-desc">7-180天</div>
                </div>
              </div>
            </el-form-item>

            <el-form-item label="⏰ 每日投入" required>
              <el-radio-group v-model="form.dailyMinutes">
                <el-radio-button :label="30">30分钟</el-radio-button>
                <el-radio-button :label="60">1小时</el-radio-button>
                <el-radio-button :label="120">2小时</el-radio-button>
                <el-radio-button :label="180">3小时以上</el-radio-button>
              </el-radio-group>
            </el-form-item>

            <el-form-item label="💼 工作经验">
              <el-radio-group v-model="form.experience">
                <el-radio-button label="fresh">应届/在校</el-radio-button>
                <el-radio-button label="1-3">1-3年</el-radio-button>
                <el-radio-button label="3-5">3-5年</el-radio-button>
                <el-radio-button label="5+">5年以上</el-radio-button>
              </el-radio-group>
            </el-form-item>

            <el-form-item label="📉 薄弱环节">
              <el-checkbox-group v-model="form.weakAreas">
                <el-checkbox v-for="c in categories" :key="c.id" :label="c.id">
                  {{ c.icon }} {{ c.name }}
                </el-checkbox>
              </el-checkbox-group>
              <div class="form-hint">选择你自认为薄弱的题型，AI 会为其安排更多练习任务</div>
            </el-form-item>

            <el-form-item>
              <el-button
                type="primary"
                size="large"
                class="generate-btn"
                :loading="generating"
                :disabled="!form.jobTitle"
                @click="generatePlan"
              >
                {{ generating ? 'AI 教练正在为你定制计划，约需30秒…' : '✨ 生成我的冲刺计划' }}
              </el-button>
            </el-form-item>
          </el-form>
        </el-card>
      </div>

      <!-- ====== 已定制：计划展示 ====== -->
      <div v-if="!loading && plan" class="plan-area">
        <!-- 计划总览 -->
        <div class="plan-header">
          <div class="plan-header-main">
            <h2 class="plan-title">🏃 {{ plan.title }}</h2>
            <p class="plan-summary">{{ plan.summary }}</p>
            <div class="plan-meta">
              <el-tag size="small" type="warning">🎯 {{ plan.config.jobTitle }}</el-tag>
              <el-tag v-if="plan.config.targetCompany" size="small" type="info">🏢 {{ plan.config.targetCompany }}</el-tag>
              <el-tag size="small" type="info">⏰ 每日 {{ dailyLabel }}</el-tag>
              <el-tag size="small" type="info">📅 {{ plan.startDate }} 开始</el-tag>
            </div>
          </div>
          <div class="plan-header-stats">
            <div class="stat-block">
              <div class="stat-num">{{ currentDay }}<span class="stat-total">/{{ plan.config.totalDays }}</span></div>
              <div class="stat-label">冲刺第几天</div>
            </div>
            <div class="stat-block">
              <div class="stat-num">{{ taskProgress }}<span class="stat-total">%</span></div>
              <div class="stat-label">任务完成率</div>
            </div>
            <div class="stat-block">
              <div class="stat-num">{{ remainingDays }}<span class="stat-total">天</span></div>
              <div class="stat-label">剩余时间</div>
            </div>
          </div>
        </div>

        <el-progress
          :percentage="dayProgress"
          :stroke-width="10"
          :show-text="false"
          class="day-progress"
        />

        <!-- 操作栏 -->
        <div class="plan-actions">
          <span class="today-phase" v-if="currentPhaseIndex >= 0">
            📍 当前阶段：<strong>{{ plan.phases[currentPhaseIndex].name }}</strong>
          </span>
          <span class="today-phase" v-else-if="currentDay > plan.config.totalDays">
            🎉 冲刺已结束，去发起一次模拟面试检验成果吧！
          </span>
          <div class="action-btns">
            <el-button size="small" @click="recustomize">🔄 重新定制</el-button>
            <el-button size="small" type="danger" plain @click="removePlan">删除计划</el-button>
          </div>
        </div>

        <!-- 阶段列表 -->
        <div
          v-for="(phase, pi) in plan.phases"
          :key="pi"
          class="phase-card"
          :class="{
            'phase-current': pi === currentPhaseIndex,
            'phase-past': isPhasePast(phase)
          }"
        >
          <div class="phase-head">
            <div class="phase-badge">阶段 {{ pi + 1 }}</div>
            <div class="phase-title-wrap">
              <div class="phase-name">
                {{ phase.name }}
                <el-tag v-if="pi === currentPhaseIndex" size="mini" type="warning">进行中</el-tag>
                <el-tag v-else-if="isPhasePast(phase)" size="mini" type="success">已过期</el-tag>
              </div>
              <div class="phase-range">第 {{ phase.dayStart }} - {{ phase.dayEnd }} 天 · {{ phase.goal }}</div>
            </div>
            <div class="phase-progress">
              <el-progress
                type="circle"
                :percentage="phaseProgress(phase)"
                :width="52"
                :stroke-width="5"
              />
            </div>
          </div>
          <div class="phase-tasks">
            <div
              v-for="(task, ti) in phase.tasks"
              :key="ti"
              class="task-item"
              :class="{ 'task-done': task.done }"
            >
              <el-checkbox
                :value="task.done"
                @change="val => toggleTask(pi, ti, val)"
              />
              <span class="task-content" @click="toggleTask(pi, ti, !task.done)">{{ task.content }}</span>
              <el-button
                v-if="actionRoute(task.action)"
                type="text"
                size="mini"
                class="task-action"
                @click="$router.push(actionRoute(task.action))"
              >{{ actionLabel(task.action) }} →</el-button>
            </div>
          </div>
        </div>

        <!-- 里程碑 + 每日例行 -->
        <el-row :gutter="16">
          <el-col :xs="24" :sm="12">
            <el-card shadow="never" class="side-card">
              <div slot="header" class="card-title">🚩 里程碑检验点</div>
              <div
                v-for="(m, i) in plan.milestones"
                :key="i"
                class="milestone-item"
                :class="{ 'milestone-past': currentDay > m.day }"
              >
                <div class="milestone-day">第{{ m.day }}天</div>
                <div class="milestone-content">{{ m.content }}</div>
              </div>
              <div v-if="!plan.milestones || !plan.milestones.length" class="empty-tip">暂无里程碑</div>
            </el-card>
          </el-col>
          <el-col :xs="24" :sm="12">
            <el-card shadow="never" class="side-card">
              <div slot="header" class="card-title">☀️ 每日例行任务</div>
              <ul class="routine-list">
                <li v-for="(r, i) in plan.dailyRoutine" :key="i">{{ r }}</li>
              </ul>
              <div v-if="plan.tips" class="plan-tips">💪 {{ plan.tips }}</div>
            </el-card>
          </el-col>
        </el-row>
      </div>
    </div>
  </div>
</template>

<script>
import Sidebar from '@/components/layout/Sidebar.vue'
import { generateSprintPlan, getSprintPlan, updateSprintTask, deleteSprintPlan } from '@/api/learn'

export default {
  name: 'SprintPlan',
  components: { Sidebar },
  data() {
    return {
      loading: true,
      generating: false,
      plan: null,
      customDays: false,
      form: {
        jobTitle: '',
        targetCompany: '',
        totalDays: 60,
        dailyMinutes: 60,
        experience: 'fresh',
        weakAreas: []
      },
      daysOptions: [
        { value: 30, desc: '短期突击' },
        { value: 60, desc: '稳步冲刺 · 推荐' },
        { value: 90, desc: '系统备战' }
      ],
      // 与学习中心五大题型分类保持一致
      categories: [
        { id: 'basic', name: '基础知识', icon: '📖' },
        { id: 'project', name: '项目经验', icon: '💼' },
        { id: 'algorithm', name: '算法与数据结构', icon: '🧮' },
        { id: 'system_design', name: '系统设计', icon: '🏗️' },
        { id: 'behavior', name: '行为面试', icon: '🤝' }
      ]
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
    dailyLabel() {
      const m = this.plan?.config?.dailyMinutes || 60
      return m >= 60 ? `${Math.round(m / 60 * 10) / 10}小时` : `${m}分钟`
    },
    // 冲刺进行到第几天（从1开始，未开始不小于1）
    currentDay() {
      if (!this.plan) return 1
      const start = new Date(this.plan.startDate + 'T00:00:00')
      const diff = Math.floor((Date.now() - start.getTime()) / 86400000) + 1
      return Math.max(1, diff)
    },
    remainingDays() {
      if (!this.plan) return 0
      return Math.max(0, this.plan.config.totalDays - this.currentDay + 1)
    },
    dayProgress() {
      if (!this.plan) return 0
      return Math.min(100, Math.round(this.currentDay / this.plan.config.totalDays * 100))
    },
    taskProgress() {
      if (!this.plan) return 0
      let total = 0
      let done = 0
      this.plan.phases.forEach(p => {
        (p.tasks || []).forEach(t => {
          total++
          if (t.done) done++
        })
      })
      return total ? Math.round(done / total * 100) : 0
    },
    currentPhaseIndex() {
      if (!this.plan) return -1
      return this.plan.phases.findIndex(
        p => this.currentDay >= p.dayStart && this.currentDay <= p.dayEnd
      )
    }
  },

  created() {
    this.loadPlan()
    // 从个人资料带出工作经验，减少重复填写
    const exp = this.$store.state.user?.userInfo?.experience
    if (exp) this.form.experience = exp
  },

  methods: {
    async loadPlan() {
      this.loading = true
      try {
        const res = await getSprintPlan()
        this.plan = res.data || null
      } catch (e) {
        this.plan = null
      } finally {
        this.loading = false
      }
    },

    selectDays(v) {
      this.customDays = false
      this.form.totalDays = v
    },

    async generatePlan() {
      if (!this.form.jobTitle) {
        this.$message.warning('请填写目标岗位')
        return
      }
      this.generating = true
      try {
        const res = await generateSprintPlan(this.form)
        this.plan = res.data || null
        this.$message.success('冲刺计划已生成，开始打卡吧！')
      } catch (e) {
        // 错误提示已由拦截器统一处理
      } finally {
        this.generating = false
      }
    },

    async toggleTask(phaseIndex, taskIndex, done) {
      // 先本地更新，接口失败再回滚
      const task = this.plan.phases[phaseIndex].tasks[taskIndex]
      const prev = task.done
      task.done = done
      try {
        await updateSprintTask({ phaseIndex, taskIndex, done })
      } catch (e) {
        task.done = prev
      }
    },

    recustomize() {
      this.$confirm('重新定制将覆盖当前计划及打卡进度，确定继续吗？', '提示', { type: 'warning' })
        .then(() => {
          // 带入原配置方便微调
          this.form = { ...this.form, ...this.plan.config }
          this.customDays = ![30, 60, 90].includes(this.form.totalDays)
          this.plan = null
        })
        .catch(() => {})
    },

    removePlan() {
      this.$confirm('删除后计划及打卡进度将无法恢复，确定删除吗？', '提示', { type: 'warning' })
        .then(async () => {
          await deleteSprintPlan()
          this.plan = null
          this.$message.success('计划已删除')
        })
        .catch(() => {})
    },

    isPhasePast(phase) {
      return this.currentDay > phase.dayEnd
    },

    phaseProgress(phase) {
      const tasks = phase.tasks || []
      if (!tasks.length) return 0
      const done = tasks.filter(t => t.done).length
      return Math.round(done / tasks.length * 100)
    },

    actionRoute(action) {
      const map = {
        practice: '/questions',
        mock: '/interview/config',
        resume: '/resume',
        company: '/company/intel',
        community: '/community'
      }
      return map[action] || ''
    },
    actionLabel(action) {
      const map = {
        practice: '去练题',
        mock: '去面试',
        resume: '去简历',
        company: '看情报',
        community: '看面经'
      }
      return map[action] || ''
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

/* ===== 定制表单 ===== */
.customize-area {
  max-width: 760px;
  margin: 0 auto;
}

.customize-header {
  text-align: center;
  margin: 20px 0 24px;
}
.customize-header h2 { font-size: 24px; color: #111; margin: 0 0 8px; }
.customize-desc { font-size: 14px; color: #666; margin: 0; line-height: 1.7; }

.form-card { border-radius: 10px; }
.form-input { max-width: 320px; }
.form-hint { font-size: 12px; color: #999; line-height: 1.6; }

.days-options {
  display: flex;
  gap: 12px;
  flex-wrap: wrap;
}
.days-item {
  border: 2px solid #e8e8e8;
  border-radius: 8px;
  padding: 10px 18px;
  cursor: pointer;
  text-align: center;
  transition: all 0.2s;
  min-width: 96px;
}
.days-item:hover { border-color: #ff9900; }
.days-item.selected { border-color: #ff9900; background: #fff9f0; }
.days-num { font-size: 16px; font-weight: bold; color: #333; line-height: 1.4; }
.days-desc { font-size: 12px; color: #999; }

.generate-btn {
  background: #ff9900 !important;
  border-color: #ff9900 !important;
  color: #111 !important;
  font-weight: bold !important;
  font-size: 15px !important;
  padding: 12px 36px !important;
  height: auto !important;
}

/* ===== 计划展示 ===== */
.plan-header {
  display: flex;
  justify-content: space-between;
  gap: 24px;
  background: linear-gradient(135deg, #131921 0%, #232f3e 60%, #3a4a5c 100%);
  border-radius: 12px;
  padding: 24px 28px;
  flex-wrap: wrap;
}

.plan-title { font-size: 22px; color: #fff; margin: 0 0 8px; }
.plan-summary { font-size: 13px; color: #aab7c4; margin: 0 0 12px; line-height: 1.7; max-width: 520px; }
.plan-meta { display: flex; gap: 8px; flex-wrap: wrap; }

.plan-header-stats { display: flex; gap: 28px; align-items: center; }
.stat-block { text-align: center; }
.stat-num { font-size: 30px; font-weight: bold; color: #ff9900; line-height: 1.2; }
.stat-total { font-size: 14px; color: #aab7c4; font-weight: normal; }
.stat-label { font-size: 12px; color: #aab7c4; margin-top: 2px; }

.day-progress { margin: 14px 0 10px; }

.plan-actions {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
  flex-wrap: wrap;
  gap: 10px;
}
.today-phase { font-size: 14px; color: #555; }
.today-phase strong { color: #a05c00; }

/* 阶段卡片 */
.phase-card {
  background: #fff;
  border: 1px solid #e8e8e8;
  border-radius: 10px;
  padding: 16px 20px;
  margin-bottom: 14px;
  transition: all 0.2s;
}
.phase-card.phase-current { border-color: #ff9900; box-shadow: 0 2px 10px rgba(255, 153, 0, 0.12); }
.phase-card.phase-past { opacity: 0.75; }

.phase-head {
  display: flex;
  align-items: center;
  gap: 14px;
  margin-bottom: 12px;
}
.phase-badge {
  background: #ff9900;
  color: #111;
  font-size: 12px;
  font-weight: bold;
  padding: 4px 10px;
  border-radius: 12px;
  flex-shrink: 0;
}
.phase-title-wrap { flex: 1; min-width: 0; }
.phase-name { font-size: 16px; font-weight: bold; color: #222; display: flex; align-items: center; gap: 8px; }
.phase-range { font-size: 12px; color: #888; margin-top: 2px; }
.phase-progress { flex-shrink: 0; }

.phase-tasks { border-top: 1px dashed #eee; padding-top: 10px; }
.task-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 7px 4px;
  border-radius: 6px;
  transition: background 0.15s;
}
.task-item:hover { background: #fff9f0; }
.task-content { flex: 1; font-size: 14px; color: #333; cursor: pointer; line-height: 1.6; }
.task-done .task-content { color: #aaa; text-decoration: line-through; }
.task-action { color: #ff9900 !important; flex-shrink: 0; }

/* 里程碑 & 例行任务 */
.side-card { margin-bottom: 16px; }
.card-title {
  font-size: 15px;
  font-weight: bold;
  color: #111;
  border-left: 4px solid #ff9900;
  padding-left: 10px;
}

.milestone-item {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  padding: 8px 0;
  border-bottom: 1px dashed #f0f0f0;
}
.milestone-item:last-child { border-bottom: none; }
.milestone-day {
  background: #fff7e6;
  color: #a05c00;
  font-size: 12px;
  font-weight: bold;
  padding: 2px 8px;
  border-radius: 10px;
  flex-shrink: 0;
}
.milestone-content { font-size: 13px; color: #444; line-height: 1.6; }
.milestone-past { opacity: 0.6; }
.milestone-past .milestone-day { background: #f6ffed; color: #52c41a; }

.routine-list { margin: 0; padding-left: 20px; }
.routine-list li { font-size: 13px; color: #444; line-height: 2; }

.plan-tips {
  margin-top: 12px;
  font-size: 13px;
  color: #666;
  background: #fff7e6;
  padding: 10px 14px;
  border-radius: 6px;
  border-left: 3px solid #ff9900;
  line-height: 1.7;
}

.empty-tip { font-size: 13px; color: #999; text-align: center; padding: 12px 0; }

@media (max-width: 768px) {
  .plan-header { flex-direction: column; }
  .plan-header-stats { gap: 20px; }
  .plan-actions { flex-direction: column; align-items: flex-start; }
}
</style>
