// API 请求模块

const request = (options) => {
  return new Promise((resolve, reject) => {
    const app = getApp()
    const token = wx.getStorageSync('token')
    
    wx.request({
      url: app.globalData.apiBaseUrl + options.url,
      method: options.method || 'GET',
      data: options.data || {},
      header: {
        'Content-Type': 'application/json',
        'Authorization': token ? `Bearer ${token}` : ''
      },
      timeout: options.timeout || 30000,
      success: (res) => {
        if (res.statusCode === 200) {
          if (res.data.code === 200) {
            resolve(res.data)
          } else if (res.data.code === 401) {
            // Token 过期，跳转登录
            wx.removeStorageSync('token')
            wx.removeStorageSync('userInfo')
            wx.redirectTo({ url: '/pages/login/login' })
            reject(new Error(res.data.message || '登录已过期'))
          } else {
            wx.showToast({
              title: res.data.message || '请求失败',
              icon: 'none'
            })
            reject(new Error(res.data.message))
          }
        } else {
          wx.showToast({
            title: '服务器错误',
            icon: 'none'
          })
          reject(new Error('服务器错误'))
        }
      },
      fail: (err) => {
        wx.showToast({
          title: '网络连接失败',
          icon: 'none'
        })
        reject(err)
      }
    })
  })
}

// 认证相关 API
export const auth = {
  // 微信登录
  wxLogin: (code) => request({
    url: '/auth/wxlogin',
    method: 'POST',
    data: { code }
  }),
  
  // 用户名密码登录
  login: (account, password) => request({
    url: '/auth/login',
    method: 'POST',
    data: { account, password }
  }),
  
  // 注册
  register: (data) => request({
    url: '/auth/register',
    method: 'POST',
    data
  }),
  
  // 获取当前用户信息
  getMe: () => request({
    url: '/auth/me'
  })
}

// 面试相关 API
export const interview = {
  // 创建面试
  create: (data) => request({
    url: '/interviews',
    method: 'POST',
    data
  }),
  
  // 获取面试列表
  list: () => request({
    url: '/interviews'
  }),
  
  // 获取面试详情
  get: (id) => request({
    url: `/interviews/${id}`
  }),
  
  // 提交回答
  submitAnswer: (id, data) => request({
    url: `/interviews/${id}/answers`,
    method: 'POST',
    data
  }),
  
  // 完成面试
  complete: (id) => request({
    url: `/interviews/${id}/complete`,
    method: 'PUT'
  })
}

// 题库相关 API
export const question = {
  // 获取题库列表
  list: (params) => request({
    url: '/questions',
    data: params || {}
  }),
  
  // 获取题目详情
  get: (id) => request({
    url: `/questions/${id}`
  }),

  // 单题练习（AI 点评）
  practice: (id, data) => request({
    url: `/questions/${id}/practice`,
    method: 'POST',
    data
  }),

  // 收藏/取消收藏题目
  collect: (id) => request({
    url: `/questions/${id}/collect`,
    method: 'POST'
  })
}

// 报告相关 API
export const report = {
  // 获取面试报告
  get: (interviewId) => request({
    url: `/reports/${interviewId}`
  }),
  
  // 获取分享报告
  getShare: (token) => request({
    url: `/reports/share/${token}`
  })
}

// 用户资料相关 API
export const profile = {
  get: () => request({
    url: '/profile'
  }),
  
  update: (data) => request({
    url: '/profile',
    method: 'PUT',
    data
  }),
  
  getStats: () => request({
    url: '/profile/stats'
  }),
  
  getTrend: () => request({
    url: '/profile/trend'
  }),

  // 我的收藏
  getCollections: () => request({
    url: '/profile/collections'
  }),

  // 修改密码
  changePassword: (data) => request({
    url: '/profile/password',
    method: 'PUT',
    data
  })
}

// 简历管理 API
export const resume = {
  // 获取简历列表
  list: () => request({
    url: '/resumes'
  }),

  // 获取简历详情
  get: (id) => request({
    url: `/resumes/${id}`
  }),

  // 删除简历
  delete: (id) => request({
    url: `/resumes/${id}`,
    method: 'DELETE'
  }),

  // 重新分析简历
  retryAnalyze: (id) => request({
    url: `/resumes/${id}/retry`,
    method: 'POST'
  }),

  // 基于简历创建面试
  createInterview: (id) => request({
    url: `/resumes/${id}/interview`,
    method: 'POST'
  }),

  // 上传简历（multipart/form-data）
  upload: (filePath) => {
    return new Promise((resolve, reject) => {
      const app = getApp()
      const token = wx.getStorageSync('token')
      wx.uploadFile({
        url: app.globalData.apiBaseUrl + '/resumes',
        filePath: filePath,
        name: 'file',
        header: {
          'Authorization': token ? `Bearer ${token}` : ''
        },
        timeout: 300000,
        success: (res) => {
          try {
            const data = JSON.parse(res.data)
            if (data.code === 200) {
              resolve(data)
            } else {
              wx.showToast({ title: data.message || '上传失败', icon: 'none' })
              reject(new Error(data.message))
            }
          } catch (e) {
            reject(new Error('解析上传结果失败'))
          }
        },
        fail: (err) => {
          wx.showToast({ title: '上传失败', icon: 'none' })
          reject(err)
        }
      })
    })
  }
}

// 知识社区 API
export const community = {
  // 获取文章列表
  getArticles: (params) => request({
    url: '/community/articles',
    data: params || {}
  }),

  // AI 生成文章
  generateArticle: (data) => request({
    url: '/community/articles/ai',
    method: 'POST',
    data
  }),

  // 获取文章详情
  getArticle: (id) => request({
    url: `/community/articles/${id}`
  }),

  // 点赞文章
  like: (id) => request({
    url: `/community/articles/${id}/like`,
    method: 'POST'
  }),

  // 收藏文章
  collect: (id) => request({
    url: `/community/articles/${id}/collect`,
    method: 'POST'
  }),

  // 获取评论列表
  getComments: (id) => request({
    url: `/community/articles/${id}/comments`
  }),

  // 发表评论
  addComment: (id, data) => request({
    url: `/community/articles/${id}/comments`,
    method: 'POST',
    data
  })
}

// 学习中心 API
export const learn = {
  // 生成冲刺计划
  generatePlan: (data) => request({
    url: '/learn/plan',
    method: 'POST',
    data,
    timeout: 120000
  }),

  // 获取当前冲刺计划
  getPlan: () => request({
    url: '/learn/plan'
  }),

  // 更新任务状态
  updateTask: (data) => request({
    url: '/learn/plan/task',
    method: 'PUT',
    data
  }),

  // 删除冲刺计划
  deletePlan: () => request({
    url: '/learn/plan',
    method: 'DELETE'
  })
}

// 贡献管理 API
export const contribution = {
  // 提交面试题
  submit: (data) => request({
    url: '/contributions/questions',
    method: 'POST',
    data
  }),

  // 获取我的贡献
  getMine: () => request({
    url: '/contributions/me'
  }),

  // 获取积分
  getCredits: () => request({
    url: '/contributions/credits'
  }),

  // 扣除积分（查看已验证题目时）
  deductCredits: (amount) => request({
    url: '/contributions/deduct',
    method: 'POST',
    data: { amount: amount || 1 }
  }),

  // 获取公司贡献题目
  getCompanyQuestions: (company, jobTitle) => request({
    url: `/contributions/company/${encodeURIComponent(company)}/questions`,
    data: jobTitle ? { jobTitle } : {}
  }),

  // 获取单道题目详情
  getQuestion: (id) => request({
    url: `/contributions/questions/${id}`
  }),

  // 投票
  vote: (id, vote) => request({
    url: `/contributions/questions/${id}/vote`,
    method: 'POST',
    data: { vote }
  }),

  // 举报
  report: (id, reason) => request({
    url: `/contributions/questions/${id}/report`,
    method: 'POST',
    data: { reason }
  }),

  // 搜索降级方案
  search: (company, jobTitle) => request({
    url: '/contributions/search',
    data: { company, jobTitle }
  })
}

// 公司面试知识库 API
export const company = {
  // 查询公司面试情报
  getIntel: (companyName, jobTitle) => request({
    url: '/company/intel',
    data: { company: companyName, jobTitle }
  }),

  // 获取公司题目参考答案
  getQuestionAnswer: (companyName, jobTitle, content) => request({
    url: '/company/question-answer',
    data: { company: companyName, jobTitle, content }
  })
}

// 面试 API 补充
export const interviewExtra = {
  // 暂停面试
  pause: (id) => request({
    url: `/interviews/${id}/pause`,
    method: 'PUT'
  }),

  // 搜索公司
  searchCompanies: (query) => request({
    url: '/companies/search',
    data: { q: query }
  })
}

export default {
  wxLogin: auth.wxLogin,
  login: auth.login,
  register: auth.register,
  getMe: auth.getMe,
  interview,
  interviewExtra,
  question,
  report,
  profile,
  resume,
  community,
  learn,
  contribution,
  company
}
