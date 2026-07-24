import Vue from 'vue'
import Vuex from 'vuex'
import user from './modules/user'
import interview from './modules/interview'

Vue.use(Vuex)

export default new Vuex.Store({
  modules: { user, interview }
})
