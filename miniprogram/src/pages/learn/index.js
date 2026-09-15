const api = require('../../api/index')

Page({
  data: {
    loading: true,
    plan: null,      // 当前冲刺计划（null 表示未定制）
    planStats: null, // 计划进度统计

    // 学习路径：分阶段学习指南（覆盖面试五大题型）
    stages: [
      {
        id: 'basic',
        name: '基础阶段',
        icon: '🌱',
        duration: '1-2 周',
        desc: '掌握核心概念，建立知识框架，同时梳理个人经历亮点',
        topics: [
          { id: 'basic', icon: '📖', name: '基础知识', level: 'easy', difficulty: '入门', desc: '概念、原理、定义类问题' },
          { id: 'behavior', icon: '🤝', name: '行为面试', level: 'easy', difficulty: '入门', desc: '团队协作和职业素养' }
        ]
      },
      {
        id: 'intermediate',
        name: '进阶阶段',
        icon: '🚀',
        duration: '2-3 周',
        desc: '深挖项目技术细节，强化算法与手写代码能力',
        topics: [
          { id: 'project', icon: '💼', name: '项目经验', level: 'medium', difficulty: '中等', desc: '围绕简历项目深挖技术细节' },
          { id: 'algorithm', icon: '🧮', name: '算法与数据结构', level: 'medium', difficulty: '中等', desc: '逻辑思维和编码能力' }
        ]
      },
      {
        id: 'advanced',
        name: '高级阶段',
        icon: '🎓',
        duration: '2-3 周',
        desc: '培养架构设计思维，并通过综合模拟面试查漏补缺',
        topics: [
          { id: 'system_design', icon: '🏗️', name: '系统设计', level: 'hard', difficulty: '进阶', desc: '整体架构和扩展性思考' },
          { id: 'mock', icon: '🎯', name: '综合模拟面试', level: 'hard', difficulty: '进阶', desc: '综合所有题型完整演练，检验备考成果' }
        ]
      }
    ]
  },

  onShow() {
    this.loadPlan()
  },

  onPullDownRefresh() {
    this.loadPlan(() => wx.stopPullDownRefresh())
  },

  // 加载冲刺计划（用于入口横幅展示当前进度）
  loadPlan(done) {
    api.learn.getPlan().then(res => {
      const plan = res.data || null
      this.setData({
        plan: plan,
        planStats: plan ? this.calcStats(plan) : null,
        loading: false
      })
      if (done) done()
    }).catch(err => {
      // 加载失败保留已有数据；无计划时 res.data 为 null，走生成入口
      console.error('获取冲刺计划失败:', err)
      this.setData({ loading: false })
      if (done) done()
    })
  },

  // 计算计划进度统计（口径与冲刺计划页保持一致）
  calcStats(plan) {
    const config = plan.config || {}
    const totalDays = config.totalDays || 0

    // 冲刺进行到第几天（从 1 开始）
    let currentDay = 1
    if (plan.startDate) {
      const start = new Date(plan.startDate.replace(/-/g, '/')).getTime()
      currentDay = Math.max(1, Math.floor((Date.now() - start) / 86400000) + 1)
    }

    // 任务完成率
    let total = 0
    let done = 0
    const phases = plan.phases || []
    phases.forEach(phase => {
      const tasks = phase.tasks || []
      tasks.forEach(task => {
        total++
        if (task.done) done++
      })
    })

    return {
      currentDay: currentDay,
      totalDays: totalDays,
      remainingDays: Math.max(0, totalDays - currentDay + 1),
      taskProgress: total ? Math.round(done / total * 100) : 0
    }
  },

  // 跳转冲刺计划页
  goPlan() {
    wx.navigateTo({ url: '/pages/learn/plan' })
  }
})
