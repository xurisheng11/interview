Page({
  data: {
    feedbackContent: '',
    submitting: false,
    faqList: [
      {
        id: 1,
        question: '如何开始一次面试练习？',
        answer: '1. 登录账号后，点击首页的"开始面试"按钮\n2. 填写公司名称和岗位信息\n3. 选择面试类型和难度\n4. 设置题目数量和答题时间\n5. 点击"开始面试"即可开始',
        expanded: false
      },
      {
        id: 2,
        question: '面试过程中可以暂停吗？',
        answer: '可以。您可以在面试过程中随时点击"跳过"按钮跳过当前题目，或者返回列表页。面试记录会自动保存，您可以在之后继续完成。',
        expanded: false
      },
      {
        id: 3,
        question: '如何查看面试报告？',
        answer: '完成面试后，系统会自动生成评估报告。您可以在"面试记录"中找到该面试，点击查看详细的评分、点评和建议。',
        expanded: false
      },
      {
        id: 4,
        question: '录音功能如何使用？',
        answer: '1. 在答题界面，点击下方的"录音"按钮\n2. 允许微信录音权限\n3. 开始说话录制语音回答\n4. 再次点击停止录音\n5. 可以播放试听或重新录制',
        expanded: false
      },
      {
        id: 5,
        question: '面试报告可以分享吗？',
        answer: '可以。在面试详情页面，点击"分享报告"按钮，可以将您的面试报告分享给好友或保存到相册。',
        expanded: false
      },
      {
        id: 6,
        question: '忘记了密码怎么办？',
        answer: '目前支持微信一键登录，推荐使用微信登录。如果使用账号密码登录，请在登录页面点击"忘记密码"或联系客服重置密码。',
        expanded: false
      }
    ]
  },

  // 展开/收起 FAQ
  toggleFaq(e) {
    const id = e.currentTarget.dataset.id
    const faqList = this.data.faqList.map(item => {
      if (item.id === id) {
        return { ...item, expanded: !item.expanded }
      }
      return item
    })
    this.setData({ faqList })
  },

  // 反馈内容输入
  onFeedbackInput(e) {
    this.setData({ feedbackContent: e.detail.value })
  },

  // 提交反馈
  submitFeedback() {
    if (!this.data.feedbackContent.trim()) {
      wx.showToast({ title: '请输入反馈内容', icon: 'none' })
      return
    }

    this.setData({ submitting: true })

    // 模拟提交反馈
    setTimeout(() => {
      this.setData({ 
        submitting: false,
        feedbackContent: ''
      })
      
      wx.showModal({
        title: '提交成功',
        content: '感谢您的反馈，我们会尽快处理！',
        showCancel: false
      })
    }, 1500)
  },

  // 查看教程
  viewTutorial(e) {
    const type = e.currentTarget.dataset.type
    const tutorials = {
      new: '面试练习教程\n\n1. 登录账号\n2. 点击"开始面试"\n3. 填写面试信息\n4. 开始答题\n5. 完成面试查看报告',
      record: '录音功能教程\n\n1. 进入答题页面\n2. 点击"录音"按钮\n3. 对着手机说话\n4. 再次点击停止\n5. 可试听或重新录制',
      report: '报告解读教程\n\n1. 进入面试详情\n2. 查看综合评分\n3. 查看分项得分\n4. 阅读AI点评\n5. 参考改进建议'
    }
    
    wx.showModal({
      title: '使用教程',
      content: tutorials[type] || '教程内容',
      showCancel: false
    })
  },

  // 联系微信（复制微信号）
  contactWechat() {
    wx.setClipboardData({
      data: 'xurixuxuxu',
      success: () => {
        wx.showToast({ 
          title: '微信号已复制', 
          icon: 'success' 
        })
      }
    })
  },

  // 联系邮箱（复制邮箱）
  contactEmail() {
    wx.setClipboardData({
      data: 'xurisheng1133@163.com',
      success: () => {
        wx.showToast({ 
          title: '邮箱已复制', 
          icon: 'success' 
        })
      }
    })
  }
})
