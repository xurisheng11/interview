import Vue from 'vue'
import VueRouter from 'vue-router'

Vue.use(VueRouter)

// 懒加载视图组件
const Login = () => import('@/views/Login.vue')
const Dashboard = () => import('@/views/Dashboard.vue')
const InterviewConfig = () => import('@/views/interview/Config.vue')
const InterviewLoading = () => import('@/views/interview/Loading.vue')
const InterviewDoing = () => import('@/views/interview/Doing.vue')
const InterviewHistory = () => import('@/views/interview/History.vue')
const ReportDetail = () => import('@/views/report/Detail.vue')
const ReportShare = () => import('@/views/report/Share.vue')
const QuestionList = () => import('@/views/question/List.vue')
const QuestionPractice = () => import('@/views/question/Practice.vue')
const CommunityIndex = () => import('@/views/community/Index.vue')
const ArticleDetail = () => import('@/views/community/Article.vue')
const ProfileIndex = () => import('@/views/profile/Index.vue')
const CompanyIntel = () => import('@/views/company/Intel.vue')

// 管理后台
const AdminLogin = () => import('@/views/admin/Login.vue')
const AdminDashboard = () => import('@/views/admin/Dashboard.vue')

const routes = [
  {
    path: '/login',
    name: 'Login',
    component: Login,
    meta: { requiresAuth: false, title: '登录' }
  },
  {
    path: '/',
    redirect: '/dashboard'
  },
  {
    path: '/dashboard',
    name: 'Dashboard',
    component: Dashboard,
    meta: { requiresAuth: true, title: '首页仪表盘' }
  },
  {
    path: '/interview',
    component: { render: h => h('router-view') },
    meta: { requiresAuth: true },
    children: [
      { path: 'config', name: 'InterviewConfig', component: InterviewConfig, meta: { requiresAuth: true, title: '发起面试' } },
      { path: 'loading', name: 'InterviewLoading', component: InterviewLoading, meta: { requiresAuth: true, title: 'AI 生成中' } },
      { path: ':id/doing', name: 'InterviewDoing', component: InterviewDoing, meta: { requiresAuth: true, title: '面试中' } },
      { path: ':id/video-prep', name: 'VideoPreparation', component: () => import('@/views/interview/VideoPreparation.vue'), meta: { requiresAuth: true, title: '视频面试准备' } },
      { path: ':id/video-doing', name: 'VideoDoing', component: () => import('@/views/interview/VideoDoing.vue'), meta: { requiresAuth: true, title: '视频面试进行中' } },
      { path: 'history', name: 'InterviewHistory', component: InterviewHistory, meta: { requiresAuth: true, title: '面试历史' } }
    ]
  },
  {
    path: '/report/:interviewId',
    name: 'ReportDetail',
    component: ReportDetail,
    meta: { requiresAuth: true, title: '面试报告' }
  },
  {
    path: '/report/share/:token',
    name: 'ReportShare',
    component: ReportShare,
    meta: { requiresAuth: false, title: '分享报告' }
  },
  {
    path: '/questions',
    component: { render: h => h('router-view') },
    meta: { requiresAuth: true },
    children: [
      { path: '', name: 'QuestionList', component: QuestionList, meta: { requiresAuth: true, title: '题库练习' } },
      { path: ':id/practice', name: 'QuestionPractice', component: QuestionPractice, meta: { requiresAuth: true, title: '单题练习' } }
    ]
  },
  {
    path: '/community',
    component: { render: h => h('router-view') },
    meta: { requiresAuth: true },
    children: [
      { path: '', name: 'Community', component: CommunityIndex, meta: { requiresAuth: true, title: '知识社区' } },
      { path: ':id', name: 'ArticleDetail', component: ArticleDetail, meta: { requiresAuth: true, title: '文章详情' } }
    ]
  },
  {
    path: '/profile',
    name: 'Profile',
    component: ProfileIndex,
    meta: { requiresAuth: true, title: '个人中心' }
  },
  {
    path: '/company/intel',
    name: 'CompanyIntel',
    component: CompanyIntel,
    meta: { requiresAuth: true, title: '公司面试知识库' }
  },
  // 管理后台路由（独立，不受主应用 Navbar 影响）
  {
    path: '/admin/login',
    name: 'AdminLogin',
    component: AdminLogin,
    meta: { requiresAuth: false, isAdmin: true, title: '管理员登录' }
  },
  {
    path: '/admin/dashboard',
    name: 'AdminDashboard',
    component: AdminDashboard,
    meta: { requiresAdminAuth: true, isAdmin: true, title: '后台管理' }
  },
  // 404 兜底
  {
    path: '*',
    redirect: '/dashboard'
  }
]

const router = new VueRouter({
  mode: 'history',
  base: '/',
  routes
})

// 路由守卫
router.beforeEach((to, from, next) => {
  const token = localStorage.getItem('token')
  const adminToken = localStorage.getItem('admin_token')
  const adminUser = JSON.parse(localStorage.getItem('admin_user') || 'null')

  // 管理后台路由守卫
  if (to.meta.requiresAdminAuth) {
    if (!adminToken || !adminUser || adminUser.role !== 'admin') {
      next({ path: '/admin/login' })
      return
    }
    if (to.meta && to.meta.title) {
      document.title = `${to.meta.title} - MockInterview`
    }
    next()
    return
  }

  // 管理后台登录页：已登录则直接跳仪表盘
  if (to.path === '/admin/login' && adminToken && adminUser && adminUser.role === 'admin') {
    next({ path: '/admin/dashboard', replace: true })
    return
  }

  // 普通路由守卫
  if (to.meta.requiresAuth && !token) {
    next({ path: '/login', query: { redirect: to.fullPath } })
  } else if (to.path === '/login' && token) {
    next({ path: '/dashboard', replace: true })
  } else {
    next()
  }

  // 设置页面标题
  if (to.meta && to.meta.title) {
    document.title = `${to.meta.title} - MockInterview`
  }
})

// 全局处理 NavigationDuplicated / Redirected 错误，避免控制台噪音
const originalPush = VueRouter.prototype.push
VueRouter.prototype.push = function push(location) {
  return originalPush.call(this, location).catch(err => {
    if (err && err.name !== 'NavigationDuplicated' && !err.message.includes('Redirected')) {
      return Promise.reject(err)
    }
  })
}

const originalReplace = VueRouter.prototype.replace
VueRouter.prototype.replace = function replace(location) {
  return originalReplace.call(this, location).catch(err => {
    if (err && err.name !== 'NavigationDuplicated' && !err.message.includes('Redirected')) {
      return Promise.reject(err)
    }
  })
}

export default router
