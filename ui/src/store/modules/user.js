import { login as loginApi, getMe } from '@/api/auth'

const state = {
  token: localStorage.getItem('token') || '',
  userInfo: JSON.parse(localStorage.getItem('userInfo') || 'null')
}

const mutations = {
  SET_TOKEN(state, token) {
    state.token = token
    if (token) {
      localStorage.setItem('token', token)
    } else {
      localStorage.removeItem('token')
    }
  },
  SET_USER_INFO(state, userInfo) {
    state.userInfo = userInfo
    if (userInfo) {
      localStorage.setItem('userInfo', JSON.stringify(userInfo))
    } else {
      localStorage.removeItem('userInfo')
    }
  },
  CLEAR_AUTH(state) {
    state.token = ''
    state.userInfo = null
    localStorage.removeItem('token')
    localStorage.removeItem('userInfo')
  }
}

const actions = {
  async login({ commit }, { account, password }) {
    const res = await loginApi({ account, password })
    commit('SET_TOKEN', res.data.token)
    commit('SET_USER_INFO', res.data.user)
    return res.data
  },
  logout({ commit }) {
    commit('CLEAR_AUTH')
  },
  async fetchMe({ commit }) {
    const res = await getMe()
    commit('SET_USER_INFO', res.data)
    return res.data
  }
}

const getters = {
  isLoggedIn: state => !!state.token,
  userInfo: state => state.userInfo,
  userId: state => state.userInfo ? state.userInfo.userId : null
}

export default { namespaced: true, state, mutations, actions, getters }
