const api = require('../../api/index')

Page({
  data: {
    // 岗位选择
    selectedJobTitle: 'frontend',
    jobTitles: [
      { value: 'frontend', name: '前端开发', desc: 'Vue/React/Angular' },
      { value: 'backend', name: '后端开发', desc: 'Java/Go/Python' },
      { value: 'fullstack', name: '全栈开发', desc: '前后端都要' },
      { value: 'mobile', name: '移动端开发', desc: 'iOS/Android' },
      { value: 'algorithm', name: '算法工程师', desc: '数据结构/机器学习' },
      { value: 'devops', name: 'DevOps工程师', desc: '运维/云原生' },
      { value: 'data', name: '数据工程师', desc: '大数据/数据分析' },
      { value: 'qa', name: '测试开发', desc: '自动化测试' }
    ],

    // 面试轮次
    selectedRound: 'comprehensive',
    rounds: [
      { value: 'round1', name: '初试', desc: '基础能力考察' },
      { value: 'round2', name: '复试', desc: '技术深度考察' },
      { value: 'round3', name: '终面', desc: '综合能力考察' },
      { value: 'comprehensive', name: '综合面试', desc: '全面模拟练习' }
    ],

    // 面试形式
    selectedTypes: ['structured'],
    interviewTypeOptions: [
      { value: 'structured', name: '结构化面试', desc: '标准化题目' },
      { value: 'semi-structured', name: '半结构化', desc: '灵活调整' },
      { value: 'random', name: '随机问答', desc: '临场发挥' }
    ],

    // 难度
    selectedDifficulty: 'medium',
    difficulties: [
      { value: 'easy', name: '简单', desc: '巩固基础知识' },
      { value: 'medium', name: '中等', desc: '常规面试难度' },
      { value: 'hard', name: '困难', desc: '挑战高薪岗位' }
    ],

    // 重点领域（可多选）
    selectedFocusAreas: [],
    focusAreaOptions: [
      { value: '算法', name: '算法与数据结构' },
      { value: '框架', name: '主流框架原理' },
      { value: '项目', name: '项目经验与架构' },
      { value: '性能', name: '性能优化' },
      { value: '安全', name: '安全与权限' },
      { value: '工程化', name: '工程化实践' },
      { value: '数据库', name: '数据库设计' },
      { value: '网络', name: '计算机网络' },
      { value: '设计模式', name: '设计模式' },
      { value: '系统设计', name: '系统设计' },
      { value: '软技能', name: '沟通与协作' },
      { value: '职业规划', name: '职业发展' }
    ],

    // 公司名称（可选）
    companyName: '',

    // 面试模式
    selectedMode: 'text',
    modes: [
      { value: 'text', name: '文字模式', desc: '打字输入回答' },
      { value: 'video', name: '视频模式', desc: '语音/视频回答' }
    ],

    // 答题时间（每题）
    selectedTime: 180,
    timeOptions: [
      { value: 60, label: '1分钟' },
      { value: 120, label: '2分钟' },
      { value: 180, label: '3分钟' },
      { value: 300, label: '5分钟' }
    ],

    // 思考时间（每题）
    selectedThinkTime: 30,
    thinkTimeOptions: [
      { value: 0, label: '不限' },
      { value: 15, label: '15秒' },
      { value: 30, label: '30秒' },
      { value: 60, label: '1分钟' },
      { value: 120, label: '2分钟' }
    ],

    // UI状态
    loading: false,
    currentStep: 1
  },

  // 计算属性
  get canStart() {
    return this.data.selectedJobTitle && this.data.selectedRound && this.data.selectedDifficulty
  },

  // 步骤切换
  setStep(e) {
    const step = e.currentTarget.dataset.step
    this.setData({ currentStep: step })
  },

  // 选择岗位
  selectJobTitle(e) {
    this.setData({ selectedJobTitle: e.currentTarget.dataset.value })
  },

  // 选择轮次
  selectRound(e) {
    this.setData({ selectedRound: e.currentTarget.dataset.value })
  },

  // 选择面试形式（可多选）
  toggleInterviewType(e) {
    const value = e.currentTarget.dataset.value
    const types = this.data.selectedTypes
    const index = types.indexOf(value)
    if (index > -1) {
      types.splice(index, 1)
    } else {
      if (types.length < 2) {
        types.push(value)
      } else {
        wx.showToast({ title: '最多选择2种形式', icon: 'none' })
        return
      }
    }
    this.setData({ selectedTypes: [...types] })
  },

  // 选择难度
  selectDifficulty(e) {
    this.setData({ selectedDifficulty: e.currentTarget.dataset.value })
  },

  // 选择重点领域（可多选）
  toggleFocusArea(e) {
    const value = e.currentTarget.dataset.value
    const areas = this.data.selectedFocusAreas
    const index = areas.indexOf(value)
    if (index > -1) {
      areas.splice(index, 1)
    } else {
      if (areas.length < 5) {
        areas.push(value)
      } else {
        wx.showToast({ title: '最多选择5个领域', icon: 'none' })
        return
      }
    }
    this.setData({ selectedFocusAreas: [...areas] })
  },

  // 公司名称输入
  onCompanyInput(e) {
    this.setData({ companyName: e.detail.value })
  },

  // 选择面试模式
  selectMode(e) {
    this.setData({ selectedMode: e.currentTarget.dataset.value })
  },

  // 选择答题时间
  selectTime(e) {
    this.setData({ selectedTime: e.currentTarget.dataset.value })
  },

  // 选择思考时间
  selectThinkTime(e) {
    this.setData({ selectedThinkTime: e.currentTarget.dataset.value })
  },

  // 获取岗位名称
  getJobTitleName() {
    const job = this.data.jobTitles.find(j => j.value === this.data.selectedJobTitle)
    return job ? job.name : '前端开发'
  },

  // 开始面试
  startInterview() {
    if (!this.canStart) {
      wx.showToast({ title: '请完善面试信息', icon: 'none' })
      return
    }

    this.setData({ loading: true })

    const { 
      selectedJobTitle, 
      selectedRound, 
      selectedDifficulty, 
      selectedTypes, 
      selectedFocusAreas,
      companyName,
      selectedMode,
      selectedTime,
      selectedThinkTime
    } = this.data

    const job = this.getJobTitleName()

    api.interview.create({
      jobTitle: job,
      difficulty: selectedDifficulty,
      round: selectedRound,
      companyName: companyName || undefined,
      interviewTypes: selectedTypes.length > 0 ? selectedTypes : ['structured'],
      focusAreas: selectedFocusAreas.length > 0 ? selectedFocusAreas : undefined,
      mode: selectedMode,
      thinkTime: selectedThinkTime
    }).then(res => {
      this.setData({ loading: false })
      
      wx.redirectTo({
        url: `/pages/interview/session?id=${res.data.interviewId}`
      })
    }).catch(err => {
      this.setData({ loading: false })
      console.error('创建面试失败', err)
      wx.showToast({ title: '创建面试失败', icon: 'none' })
    })
  }
})
