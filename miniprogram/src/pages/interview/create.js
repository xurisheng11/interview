const api = require('../../api/index')

Page({
  data: {
    // 岗位选择（按类别分组，与 UI 前端一致）
    selectedJobTitle: '前端开发',
    selectedCategory: 'tech',
    isCustomJob: false,
    customJobInput: '',
    jobCategories: [
      {
        name: 'tech',
        label: '技术类',
        icon: '💻',
        jobs: [
          '后端开发', '前端开发', '全栈开发', '移动端开发(Android)',
          '移动端开发(iOS)', '大数据工程师', 'AI算法工程师', '测试工程师',
          '运维/DevOps', '网络安全', '嵌入式开发', '游戏开发',
          '游戏客户端开发', '游戏服务端开发', '数据分析', '数据工程',
          '机器学习工程师', '深度学习工程师', 'NLP工程师', '推荐算法工程师'
        ]
      },
      {
        name: 'product',
        label: '产品与设计',
        icon: '🎨',
        jobs: [
          '产品经理', '产品助理', '高级产品经理', '数据产品经理',
          'AI产品经理', 'C端产品经理', 'B端产品经理', '平台产品经理',
          'UI设计师', 'UX设计师', '视觉设计师', '交互设计师',
          '平面设计师', '品牌设计师', '视频设计师'
        ]
      },
      {
        name: 'operation',
        label: '运营与市场',
        icon: '📈',
        jobs: [
          '运营专员', '内容运营', '用户运营', '活动运营',
          '新媒体运营', '电商运营', '社群运营', '游戏运营',
          '市场策划', '市场营销', '商务拓展', '销售代表',
          '客户经理', '渠道运营', '增长运营'
        ]
      },
      {
        name: 'admin',
        label: '职能类',
        icon: '📋',
        jobs: [
          '会计/财务', '人力资源', '行政管理', '法务/合规',
          '采购/供应链', '质量管理', '项目协调', '总裁助理',
          '投资关系', '公关媒介', '党建专员', '行政前台'
        ]
      }
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
    // WXML 不支持 .indexOf() 方法调用，用映射表驱动选中态渲染
    selectedTypeMap: { structured: true },
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
    selectedFocusMap: {},
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

    // 面试模式（value 保持 video 兼容后端；小程序端实现为录音答题，摄像头视频为 Web 端能力）
    selectedMode: 'text',
    modes: [
      { value: 'text', name: '文字模式', desc: '打字输入回答' },
      { value: 'video', name: '语音模式', desc: '说话答题，录音停止后自动转文字' }
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
    currentStep: 1,
    // 开始按钮可用态（wxml 只能绑 data 字段，getter 在模板里取不到）
    canStart: true
  },

  onLoad() {
    this.updateCanStart()
  },

  // 计算开始按钮可用态（结果写入 data 供 wxml 绑定）
  updateCanStart() {
    const canStart = !!(this.data.selectedJobTitle && this.data.selectedRound && this.data.selectedDifficulty)
    if (canStart !== this.data.canStart) {
      this.setData({ canStart })
    }
  },

  // 步骤切换
  setStep(e) {
    const step = e.currentTarget.dataset.step
    this.setData({ currentStep: step })
  },

  // 选择岗位分类
  selectCategory(e) {
    this.setData({ 
      selectedCategory: e.currentTarget.dataset.name,
      customJobInput: '',
      isCustomJob: false
    })
  },

  // 选择岗位
  selectJobTitle(e) {
    this.setData({ 
      selectedJobTitle: e.currentTarget.dataset.name,
      isCustomJob: false,
      customJobInput: ''
    })
    this.updateCanStart()
  },

  // 自定义岗位输入
  onCustomJobInput(e) {
    this.setData({ customJobInput: e.detail.value })
  },

  // 应用自定义岗位
  applyCustomJob() {
    const val = this.data.customJobInput.trim()
    if (val) {
      this.setData({ 
        selectedJobTitle: val,
        isCustomJob: true 
      })
      this.updateCanStart()
    }
  },

  // 选择轮次
  selectRound(e) {
    this.setData({ selectedRound: e.currentTarget.dataset.value })
    this.updateCanStart()
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
    const map = {}
    types.forEach(t => { map[t] = true })
    this.setData({ selectedTypes: types, selectedTypeMap: map })
  },

  // 选择难度
  selectDifficulty(e) {
    this.setData({ selectedDifficulty: e.currentTarget.dataset.value })
    this.updateCanStart()
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
    const map = {}
    areas.forEach(a => { map[a] = true })
    this.setData({ selectedFocusAreas: areas, selectedFocusMap: map })
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
    return this.data.selectedJobTitle || '前端开发'
  },

  // 开始面试
  startInterview() {
    if (!this.data.canStart) {
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
