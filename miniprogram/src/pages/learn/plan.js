const api = require('../../api/index')

// 任务关联功能的跳转配置（action → 小程序页面路径与文案）
const ACTION_MAP = {
  practice: { path: '/pages/questions/index', label: '去练题' },
  mock: { path: '/pages/interview/create', label: '去面试' },
  resume: { path: '/pages/resume/list', label: '去简历' },
  company: { path: '/pages/company/intel', label: '看情报' },
  community: { path: '/pages/community/index', label: '看面经' }
}

Page({
  data: {
    loading: true,
    generating: false,
    plan: null,   // 当前冲刺计划（null 表示未定制，展示定制表单）
    phases: [],   // 处理后的阶段列表（含进度、任务跳转文案）
    stats: null,  // 计划统计（天数、完成率、当前阶段等）

    // ===== 定制表单 =====
    jobTitle: '',
    targetCompany: '',
    selectedDays: 60,         // 冲刺天数（7-180）
    useInterviewDate: false,  // 是否按面试日期定制
    interviewDate: '',        // 面试日期 YYYY-MM-DD
    daysUntilInterview: 0,    // 距面试天数
    dailyMinutes: 60,
    experience: 'fresh',
    weakAreas: [],

    daysOptions: [
      { value: 30, name: '30天', desc: '短期突击' },
      { value: 60, name: '60天', desc: '稳步冲刺 · 推荐' },
      { value: 90, name: '90天', desc: '系统备战' }
    ],
    minutesOptions: [
      { value: 30, label: '30分钟' },
      { value: 60, label: '1小时' },
      { value: 120, label: '2小时' },
      { value: 180, label: '3小时以上' }
    ],
    experienceOptions: [
      { value: 'fresh', label: '应届/在校' },
      { value: '1-3', label: '1-3年' },
      { value: '3-5', label: '3-5年' },
      { value: '5+', label: '5年以上' }
    ],
    // 薄弱环节选项（与学习路径五大题型保持一致）
    categories: [
      { id: 'basic', name: '基础知识', icon: '📖' },
      { id: 'project', name: '项目经验', icon: '💼' },
      { id: 'algorithm', name: '算法与数据结构', icon: '🧮' },
      { id: 'system_design', name: '系统设计', icon: '🏗️' },
      { id: 'behavior', name: '行为面试', icon: '🤝' }
    ],

    minDate: '', // 面试日期可选范围：今天 ~ 180 天后
    maxDate: ''
  },

  onLoad() {
    const now = new Date()
    this.setData({
      minDate: this.formatDate(now),
      maxDate: this.formatDate(new Date(now.getTime() + 180 * 86400000))
    })
    this.loadPlan()
  },

  onPullDownRefresh() {
    // 生成中避免打断请求
    if (this.data.generating) {
      wx.stopPullDownRefresh()
      return
    }
    this.loadPlan(() => wx.stopPullDownRefresh())
  },

  // ===== 计划加载 =====

  loadPlan(done) {
    api.learn.getPlan().then(res => {
      const plan = res.data || null
      const view = plan ? this.buildPlanView(plan) : { phases: [], stats: null }
      this.setData(Object.assign({ plan: plan, loading: false }, view))
      if (done) done()
    }).catch(err => {
      // 加载失败保留已有数据；无计划时 res.data 为 null，展示定制表单
      console.error('获取冲刺计划失败:', err)
      this.setData({ loading: false })
      if (done) done()
    })
  },

  // 将原始计划处理为视图数据：阶段进度、任务跳转文案、统计信息
  buildPlanView(plan) {
    const config = plan.config || {}
    const totalDays = config.totalDays || 0

    // 冲刺进行到第几天（从 1 开始）
    let currentDay = 1
    if (plan.startDate) {
      const start = new Date(plan.startDate.replace(/-/g, '/')).getTime()
      currentDay = Math.max(1, Math.floor((Date.now() - start) / 86400000) + 1)
    }

    // 阶段列表（附带进度与任务跳转文案）
    const phases = (plan.phases || []).map((phase, pi) => {
      const tasks = (phase.tasks || []).map((task, ti) => ({
        id: pi + '-' + ti,
        content: task.content,
        action: task.action,
        actionLabel: (ACTION_MAP[task.action] || {}).label || '',
        done: !!task.done
      }))
      const doneCount = tasks.filter(t => t.done).length
      return {
        id: 'phase-' + pi,
        index: pi,
        name: phase.name,
        dayStart: phase.dayStart,
        dayEnd: phase.dayEnd,
        goal: phase.goal,
        tasks: tasks,
        progress: tasks.length ? Math.round(doneCount / tasks.length * 100) : 0,
        isCurrent: currentDay >= phase.dayStart && currentDay <= phase.dayEnd,
        isPast: currentDay > phase.dayEnd
      }
    })

    // 全局统计
    let total = 0
    let done = 0
    phases.forEach(p => {
      p.tasks.forEach(t => {
        total++
        if (t.done) done++
      })
    })
    const currentPhase = phases.find(p => p.isCurrent)
    const minutes = config.dailyMinutes || 60

    return {
      phases: phases,
      stats: {
        currentDay: currentDay,
        totalDays: totalDays,
        remainingDays: Math.max(0, totalDays - currentDay + 1),
        dayProgress: totalDays ? Math.min(100, Math.round(currentDay / totalDays * 100)) : 0,
        taskProgress: total ? Math.round(done / total * 100) : 0,
        currentPhaseName: currentPhase ? currentPhase.name : '',
        dailyLabel: minutes >= 60 ? (Math.round(minutes / 60 * 10) / 10) + '小时' : minutes + '分钟'
      }
    }
  },

  // ===== 表单交互 =====

  onJobTitleInput(e) {
    this.setData({ jobTitle: e.detail.value })
  },

  onCompanyInput(e) {
    this.setData({ targetCompany: e.detail.value })
  },

  // 选择冲刺天数快捷项（30/60/90 天）
  selectDays(e) {
    this.setData({
      selectedDays: e.currentTarget.dataset.value,
      useInterviewDate: false,
      interviewDate: '',
      daysUntilInterview: 0
    })
  },

  // 切换为按面试日期定制
  selectDateMode() {
    this.setData({ useInterviewDate: true })
  },

  // 选择面试日期，自动换算冲刺天数（限制 7-180 天）
  onDateChange(e) {
    const date = e.detail.value
    const target = new Date(date.replace(/-/g, '/')).getTime()
    const days = Math.floor((target - Date.now()) / 86400000) + 1
    this.setData({
      interviewDate: date,
      daysUntilInterview: days,
      selectedDays: Math.min(180, Math.max(7, days)),
      useInterviewDate: true
    })
  },

  selectMinutes(e) {
    this.setData({ dailyMinutes: e.currentTarget.dataset.value })
  },

  selectExperience(e) {
    this.setData({ experience: e.currentTarget.dataset.value })
  },

  // 切换薄弱环节（多选）
  toggleWeakArea(e) {
    const id = e.currentTarget.dataset.id
    const areas = this.data.weakAreas
    const index = areas.indexOf(id)
    if (index > -1) {
      areas.splice(index, 1)
    } else {
      areas.push(id)
    }
    this.setData({ weakAreas: [...areas] })
  },

  // ===== 生成 / 管理计划 =====

  // 生成冲刺计划（AI 生成约需 30 秒）
  generatePlan() {
    const data = this.data
    if (data.generating) return
    if (!data.jobTitle.trim()) {
      wx.showToast({ title: '请填写目标岗位', icon: 'none' })
      return
    }
    if (data.useInterviewDate && !data.interviewDate) {
      wx.showToast({ title: '请选择面试日期', icon: 'none' })
      return
    }

    this.setData({ generating: true })
    wx.showLoading({ title: 'AI 教练定制中…', mask: true })

    api.learn.generatePlan({
      jobTitle: data.jobTitle.trim(),
      targetCompany: data.targetCompany.trim(),
      totalDays: data.selectedDays,
      dailyMinutes: data.dailyMinutes,
      experience: data.experience,
      weakAreas: data.weakAreas
    }).then(res => {
      wx.hideLoading()
      const plan = res.data
      if (!plan) {
        this.setData({ generating: false })
        wx.showToast({ title: '生成失败，请重试', icon: 'none' })
        return
      }
      const view = this.buildPlanView(plan)
      this.setData({
        generating: false,
        plan: plan,
        phases: view.phases,
        stats: view.stats
      })
      wx.showToast({ title: '计划已生成，开始打卡吧！', icon: 'success' })
    }).catch(err => {
      wx.hideLoading()
      this.setData({ generating: false })
      console.error('生成冲刺计划失败:', err)
      wx.showToast({ title: '生成失败，请稍后重试', icon: 'none' })
    })
  },

  // 勾选/取消任务（乐观更新，接口失败回滚）
  toggleTask(e) {
    const pi = e.currentTarget.dataset.phase
    const ti = e.currentTarget.dataset.task
    const plan = this.data.plan
    const phases = (plan && plan.phases) || []
    if (!phases[pi] || !phases[pi].tasks || !phases[pi].tasks[ti]) return

    const task = phases[pi].tasks[ti]
    const prev = !!task.done
    const next = !prev

    // 先本地更新
    task.done = next
    const view = this.buildPlanView(plan)
    this.setData({ phases: view.phases, stats: view.stats })

    api.learn.updateTask({ phaseIndex: pi, taskIndex: ti, done: next }).catch(err => {
      // 更新失败回滚
      console.error('更新任务状态失败:', err)
      task.done = prev
      const rollback = this.buildPlanView(plan)
      this.setData({ phases: rollback.phases, stats: rollback.stats })
    })
  },

  // 跳转任务关联功能（去练题/去面试等）
  goAction(e) {
    const action = e.currentTarget.dataset.action
    const conf = ACTION_MAP[action]
    if (conf) {
      wx.navigateTo({ url: conf.path })
    }
  },

  // 重新定制（带入原配置方便微调）
  recustomize() {
    wx.showModal({
      title: '重新定制',
      content: '重新定制将覆盖当前计划及打卡进度，确定继续吗？',
      confirmColor: '#1677ff',
      success: res => {
        if (!res.confirm) return
        const config = (this.data.plan && this.data.plan.config) || {}
        this.setData({
          plan: null,
          phases: [],
          stats: null,
          jobTitle: config.jobTitle || '',
          targetCompany: config.targetCompany || '',
          selectedDays: config.totalDays || 60,
          dailyMinutes: config.dailyMinutes || 60,
          experience: config.experience || 'fresh',
          weakAreas: config.weakAreas ? [...config.weakAreas] : [],
          useInterviewDate: false,
          interviewDate: '',
          daysUntilInterview: 0
        })
        wx.pageScrollTo({ scrollTop: 0, duration: 300 })
      }
    })
  },

  // 删除计划
  removePlan() {
    wx.showModal({
      title: '删除计划',
      content: '删除后计划及打卡进度将无法恢复，确定删除吗？',
      confirmColor: '#ff4d4f',
      confirmText: '删除',
      success: res => {
        if (!res.confirm) return
        wx.showLoading({ title: '删除中…' })
        api.learn.deletePlan().then(() => {
          wx.hideLoading()
          this.setData({ plan: null, phases: [], stats: null })
          wx.showToast({ title: '计划已删除', icon: 'success' })
          wx.pageScrollTo({ scrollTop: 0, duration: 300 })
        }).catch(err => {
          wx.hideLoading()
          console.error('删除计划失败:', err)
          wx.showToast({ title: '删除失败，请重试', icon: 'none' })
        })
      }
    })
  },

  // 日期格式化为 YYYY-MM-DD
  formatDate(d) {
    const month = d.getMonth() + 1
    const day = d.getDate()
    return d.getFullYear() + '-' + (month < 10 ? '0' + month : month) + '-' + (day < 10 ? '0' + day : day)
  }
})
