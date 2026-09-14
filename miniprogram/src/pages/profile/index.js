const api = require('../../api/index')

Page({
  data: {
    formData: {
      avatar: '',
      nickname: '',
      username: '',
      phone: '',
      email: '',
      bio: '',
      targetPosition: '',
      experience: ''
    },
    loading: false
  },

  onLoad() {
    this.loadProfile()
  },

  onShow() {
    // 每次显示时刷新数据
    const updatedData = this.data.formData
    this.setData({ formData: updatedData })
  },

  loadProfile() {
    // 从缓存获取用户信息
    const userInfo = wx.getStorageSync('userInfo') || {}
    
    this.setData({
      formData: {
        avatar: userInfo.avatar || '',
        nickname: userInfo.nickname || '',
        username: userInfo.username || '',
        phone: userInfo.phone || '',
        email: userInfo.email || '',
        bio: userInfo.bio || '',
        targetPosition: userInfo.targetPosition || '',
        experience: userInfo.experience || ''
      }
    })
  },

  // 选择头像
  chooseAvatar() {
    wx.chooseMedia({
      count: 1,
      mediaType: ['image'],
      sourceType: ['album', 'camera'],
      success: (res) => {
        const tempFilePath = res.tempFiles[0].tempFilePath
        this.setData({
          'formData.avatar': tempFilePath
        })
        
        // TODO: 上传头像到服务器
        wx.showToast({
          title: '头像已选择',
          icon: 'success'
        })
      }
    })
  },

  // 编辑字段
  editField(e) {
    const field = e.currentTarget.dataset.field
    const fieldConfig = {
      nickname: { title: '修改昵称', placeholder: '请输入昵称', maxlength: 20 },
      username: { title: '用户名', placeholder: '请输入用户名', maxlength: 20 },
      phone: { title: '手机号', placeholder: '请输入手机号', maxlength: 11 },
      email: { title: '邮箱', placeholder: '请输入邮箱', maxlength: 50 },
      bio: { title: '个人简介', placeholder: '介绍一下自己...', maxlength: 200 },
      targetPosition: { title: '目标岗位', placeholder: '例如：前端开发工程师', maxlength: 30 },
      experience: { title: '工作年限', placeholder: '例如：3年', maxlength: 10 }
    }

    const config = fieldConfig[field]
    if (!config) return

    wx.showModal({
      title: config.title,
      editable: true,
      placeholderText: config.placeholder,
      content: this.data.formData[field] || '',
      success: (res) => {
        if (res.confirm && res.content !== this.data.formData[field]) {
          const newValue = res.content.trim()
          
          // 验证
          if (field === 'phone' && newValue && !/^1[3-9]\d{9}$/.test(newValue)) {
            wx.showToast({ title: '手机号格式不正确', icon: 'none' })
            return
          }
          if (field === 'email' && newValue && !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(newValue)) {
            wx.showToast({ title: '邮箱格式不正确', icon: 'none' })
            return
          }

          this.setData({
            [`formData.${field}`]: newValue
          })
        }
      }
    })
  },

  // 保存资料
  saveProfile() {
    this.setData({ loading: true })

    api.profile.update(this.data.formData).then(res => {
      this.setData({ loading: false })
      
      // 更新本地缓存
      const userInfo = { ...wx.getStorageSync('userInfo'), ...this.data.formData }
      wx.setStorageSync('userInfo', userInfo)
      
      // 更新全局数据
      const app = getApp()
      app.globalData.userInfo = userInfo
      
      wx.showToast({
        title: '保存成功',
        icon: 'success'
      })
      
      setTimeout(() => {
        wx.navigateBack()
      }, 1500)
    }).catch(err => {
      this.setData({ loading: false })
      
      // 即使API失败也保存到本地（离线模式）
      const userInfo = { ...wx.getStorageSync('userInfo'), ...this.data.formData }
      wx.setStorageSync('userInfo', userInfo)
      
      wx.showToast({
        title: '已保存到本地',
        icon: 'success'
      })
    })
  }
})
