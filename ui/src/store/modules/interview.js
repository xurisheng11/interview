const namespaced = true

const state = {
  currentInterview: null,
  currentInterviewId: null,
  currentQuestions: [],
  currentIndex: 0,
  // 视频面试模式
  mediaStream: null,
  interviewMode: 'text',
  enableRecording: false,
  recordedVideo: null
}

const mutations = {
  SET_INTERVIEW(state, interview) { state.currentInterview = interview },
  SET_CURRENT_ID(state, id) { state.currentInterviewId = id },
  SET_QUESTIONS(state, questions) { state.currentQuestions = questions },
  SET_CURRENT_INDEX(state, index) { state.currentIndex = index },
  CLEAR_INTERVIEW(state) {
    state.currentInterview = null
    state.currentInterviewId = null
    state.currentQuestions = []
    state.currentIndex = 0
  },
  SET_MEDIA_STREAM(state, stream) {
    state.mediaStream = stream
  },
  SET_INTERVIEW_MODE(state, mode) {
    state.interviewMode = mode
  },
  SET_ENABLE_RECORDING(state, bool) {
    state.enableRecording = bool
  },
  SET_RECORDED_VIDEO(state, { url, blob, size, duration }) {
    state.recordedVideo = { url, blob, size, duration }
  },
  RELEASE_MEDIA_STREAM(state) {
    if (state.mediaStream) {
      state.mediaStream.getTracks().forEach(t => t.stop())
      state.mediaStream = null
    }
  }
}

const actions = {
  setInterview({ commit }, { interview, questions }) {
    commit('SET_INTERVIEW', interview)
    commit('SET_QUESTIONS', questions)
  }
}

const getters = {
  currentInterview: state => state.currentInterview,
  currentQuestions: state => state.currentQuestions,
  currentIndex: state => state.currentIndex
}

export default { namespaced, state, mutations, actions, getters }
