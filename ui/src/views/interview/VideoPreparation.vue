<template>
  <div class="prep-layout">
    <!-- 移动端警告 -->
    <div v-if="compat.isMobile" class="full-warning">
      <div class="warning-icon">📱</div>
      <div class="warning-title">请在电脑端使用视频面试功能</div>
      <div class="warning-desc">视频面试需要较大屏幕和稳定的设备性能，请在 PC 端使用 Chrome、Edge 等现代浏览器访问。</div>
      <el-button type="primary" class="btn-orange" @click="$router.push('/interview/config')">返回配置</el-button>
    </div>

    <template v-else>
      <div class="prep-container">
        <!-- 左侧：摄像头预览 -->
        <div class="camera-side">
          <div class="camera-wrapper">
            <VideoPreview :stream="cameraStream" :mirror="true" />
          </div>

          <!-- 音频波形指示 -->
          <div class="audio-row">
            <SpeechIndicator :stream="cameraStream" :active="!!cameraStream" :supported="compat.speechRecognition" />
            <span class="audio-label">麦克风检测</span>
          </div>

          <!-- 无摄像头设备，但有麦克风，允许继续 -->
          <el-alert
            v-if="permError.noCamera"
            type="warning"
            :closable="false"
            show-icon
            title="未检测到摄像头"
            description="摄像头不可用，将以仅音频模式进行面试（无摄像头预览）。语音识别仍可正常使用。"
            style="margin-top:12px"
          ></el-alert>
          <!-- 权限错误 -->
          <el-alert
            v-if="permError.security"
            type="error"
            :closable="false"
            show-icon
            title="无法访问摄像头：需要 HTTPS 环境"
            description="浏览器安全策略要求摄像头/麦克风只能在 HTTPS 或 localhost 下使用。请通过 https://localhost:3000 访问，或让管理员配置 HTTPS。"
            style="margin-top:12px"
          ></el-alert>
          <div v-else-if="permError.occupied" style="margin-top:12px">
            <el-alert
              type="error"
              :closable="false"
              show-icon
              title="摄像头被其他程序占用"
              description="请关闭正在使用摄像头的其他应用（如微信、钉钉视频等），然后点击重试。"
            ></el-alert>
            <el-button size="small" style="margin-top:8px" @click="requestPermissions">🔄 重试</el-button>
          </div>
          <div v-else-if="permError.camera && permError.microphone && !permError.occupied" style="margin-top:12px">
            <el-alert
              type="error"
              :closable="false"
              show-icon
              title="未检测到摄像头和麦克风"
              description="您的设备没有摄像头和麦克风，无法使用视频面试功能。请连接设备后重试，或返回配置页选择文字模式。"
            ></el-alert>
            <el-button size="small" style="margin-top:8px" @click="requestPermissions">🔄 重新检测</el-button>
            <el-button size="small" style="margin-top:8px" @click="$router.push('/interview/config')">返回配置</el-button>
          </div>
          <div v-else-if="permError.camera" style="margin-top:12px">
            <el-alert
              type="error"
              :closable="false"
              show-icon
              title="摄像头/麦克风权限被拒绝"
              description="请点击浏览器地址栏左侧的🔒图标，将摄像头和麦克风权限改为「允许」，然后点击重试。"
            ></el-alert>
            <el-button size="small" style="margin-top:8px" @click="requestPermissions">🔄 重新请求权限</el-button>
          </div>
          <el-alert
            v-if="permLoading"
            type="info"
            :closable="false"
            show-icon
            title="正在请求设备权限，请在浏览器弹窗中点击「允许」..."
            style="margin-top:12px"
          ></el-alert>
        </div>

        <!-- 右侧：设置区 -->
        <div class="settings-side">
          <div class="card">
            <div class="card-title">🎥 视频面试准备</div>

            <!-- 兼容性检测 -->
            <div class="section">
              <div class="section-title">浏览器兼容性检查</div>
              <div class="compat-list">
                <div class="compat-item">
                  <span :class="['compat-icon', compat.getUserMedia ? 'ok' : 'fail']">
                    {{ compat.getUserMedia ? '✓' : '✗' }}
                  </span>
                  <span>摄像头 / 麦克风</span>
                  <span v-if="!compat.getUserMedia" class="compat-tip">请更新浏览器</span>
                </div>
                <div class="compat-item">
                  <span :class="['compat-icon', compat.speechRecognition ? 'ok' : 'warn']">
                    {{ compat.speechRecognition ? '✓' : '⚠' }}
                  </span>
                  <span>语音识别</span>
                  <span v-if="!compat.speechRecognition" class="compat-tip warn-text">将使用手动输入模式</span>
                </div>
                <div class="compat-item">
                  <span :class="['compat-icon', compat.mediaRecorder ? 'ok' : 'warn']">
                    {{ compat.mediaRecorder ? '✓' : '⚠' }}
                  </span>
                  <span>视频录制</span>
                  <span v-if="!compat.mediaRecorder" class="compat-tip warn-text">录制功能不可用</span>
                </div>
              </div>

              <!-- getUserMedia 不支持时的硬性拦截 -->
              <el-alert
                v-if="!compat.getUserMedia"
                type="error"
                :closable="false"
                show-icon
                title="浏览器不支持视频面试"
                description="当前浏览器不支持摄像头/麦克风访问，推荐使用 Chrome、Edge、Firefox 或 Safari 最新版本。"
                style="margin-top:12px"
              ></el-alert>
              <!-- 语音识别不支持时的橙色警告 -->
              <el-alert
                v-else-if="!compat.speechRecognition"
                type="warning"
                :closable="false"
                show-icon
                title="语音识别不可用，将使用手动输入模式"
                description="当前浏览器不支持 Web Speech API，您可以手动在文本框中输入答案。"
                style="margin-top:12px"
              ></el-alert>
            </div>

            <!-- 语言选择 -->
            <div class="section">
              <div class="section-title">识别语言</div>
              <el-select v-model="speechLang" size="small" style="width:180px">
                <el-option label="普通话（中文）" value="zh-CN"></el-option>
                <el-option label="English" value="en-US"></el-option>
              </el-select>
            </div>

            <!-- 录制开关 -->
            <div class="section">
              <div class="section-title">视频录制</div>
              <div class="switch-row">
                <el-switch
                  v-model="enableRecording"
                  :disabled="!compat.mediaRecorder"
                  active-text="开启"
                  inactive-text="关闭"
                ></el-switch>
                <span class="switch-hint">录制视频仅存储在本地，不会上传服务器</span>
              </div>
            </div>

            <!-- 开始按钮 -->
            <div class="action-row">
              <el-button
                type="primary"
                class="btn-orange btn-start"
                :disabled="!canStart"
                :loading="permLoading"
                @click="showPrivacyDialog = true"
              >
                同意并开始面试
              </el-button>
              <el-button class="btn-back" @click="$router.push('/interview/config')">返回配置</el-button>
            </div>
          </div>
        </div>
      </div>

      <!-- 隐私声明弹窗 -->
      <el-dialog
        title="隐私声明"
        :visible.sync="showPrivacyDialog"
        :close-on-click-modal="false"
        width="480px"
      >
        <ul class="privacy-list">
          <li>摄像头和麦克风画面<strong>不会上传至服务器</strong></li>
          <li>语音识别在<strong>浏览器本地</strong>完成</li>
          <li>后端仅接收文字答案和语速/停顿统计数据</li>
          <li v-if="enableRecording">视频录制仅存储于本地内存，<strong>关闭页面后消失</strong></li>
        </ul>
        <div slot="footer">
          <el-button @click="showPrivacyDialog = false">取消</el-button>
          <el-button type="primary" class="btn-orange" @click="startInterview">同意并开始</el-button>
        </div>
      </el-dialog>
    </template>
  </div>
</template>

<script>
import { checkMediaSupport } from '@/utils/mediaCompatibility'
import VideoPreview from '@/components/interview/VideoPreview.vue'
import SpeechIndicator from '@/components/interview/SpeechIndicator.vue'

export default {
  name: 'VideoPreparation',
  components: { VideoPreview, SpeechIndicator },

  data() {
    return {
      compat: {
        getUserMedia: true,
        speechRecognition: true,
        mediaRecorder: true,
        isMobile: false
      },
      cameraStream: null,
      permLoading: false,
      permError: {
        camera: false,
        microphone: false,
        occupied: false,
        security: false,
        noCamera: false
      },
      speechLang: 'zh-CN',
      enableRecording: false,
      showPrivacyDialog: false,
      started: false
    }
  },

  computed: {
    canStart() {
      // 有流（摄像头+麦克风 或 仅麦克风）且没有阻塞性错误即可开始
      return (
        !!this.cameraStream &&
        !this.permError.security &&
        !(this.permError.camera && this.permError.microphone)
      )
    },
    interviewId() {
      return this.$route.params.id
    }
  },

  mounted() {
    this.compat = checkMediaSupport()
    // 不管兼容性检测结果，直接尝试请求权限
    // 让浏览器自己弹出授权框，失败了再提示
    this.requestPermissions()
  },

  beforeDestroy() {
    // 若用户未点开始，释放媒体流
    if (!this.started && this.cameraStream) {
      this.cameraStream.getTracks().forEach(t => t.stop())
      this.$store.commit('interview/RELEASE_MEDIA_STREAM')
    }
  },

  methods: {
    async requestPermissions() {
      this.permLoading = true
      this.permError.camera = false
      this.permError.microphone = false
      this.permError.occupied = false
      this.permError.security = false
      this.permError.noCamera = false

      // 检查 mediaDevices 是否存在（非 HTTPS/localhost 下不存在）
      // 尝试 polyfill 旧浏览器的 getUserMedia
      if (!navigator.mediaDevices) {
        navigator.mediaDevices = {}
      }
      if (!navigator.mediaDevices.getUserMedia) {
        const legacyGetUserMedia = (
          navigator.getUserMedia ||
          navigator.webkitGetUserMedia ||
          navigator.mozGetUserMedia ||
          navigator.msGetUserMedia
        )
        if (legacyGetUserMedia) {
          // 把旧 API 包装成 Promise
          navigator.mediaDevices.getUserMedia = (constraints) => {
            return new Promise((resolve, reject) => {
              legacyGetUserMedia.call(navigator, constraints, resolve, reject)
            })
          }
        } else {
          this.permError.security = true
          this.permLoading = false
          return
        }
      }

      try {
        let stream
        try {
          // 优先请求摄像头+麦克风
          stream = await navigator.mediaDevices.getUserMedia({ video: true, audio: true })
        } catch (err) {
          if (err.name === 'NotFoundError' || err.name === 'DevicesNotFoundError') {
            // 没有摄像头，尝试只请求麦克风
            try {
              stream = await navigator.mediaDevices.getUserMedia({ video: false, audio: true })
              this.permError.noCamera = true // 标记无摄像头，仅音频模式
            } catch (audioErr) {
              throw audioErr // 麦克风也没有，抛出
            }
          } else {
            throw err
          }
        }
        this.cameraStream = stream
        this.$store.commit('interview/SET_MEDIA_STREAM', stream)
        // 成功后修正兼容性标记
        this.compat.getUserMedia = true
      } catch (err) {
        console.warn('getUserMedia error:', err.name, err.message)
        if (err.name === 'NotAllowedError' || err.name === 'PermissionDeniedError') {
          this.permError.camera = true
          this.permError.microphone = true
        } else if (err.name === 'NotFoundError' || err.name === 'DevicesNotFoundError') {
          // 设备都不存在
          this.permError.camera = true
          this.permError.microphone = true
        } else if (err.name === 'NotReadableError' || err.name === 'TrackStartError') {
          this.permError.camera = true
          this.permError.occupied = true
        } else if (err.name === 'OverconstrainedError') {
          this.permError.camera = true
        } else if (err.name === 'SecurityError') {
          this.permError.security = true
        } else {
          this.permError.camera = true
        }
      } finally {
        this.permLoading = false
      }
    },

    startInterview() {
      this.showPrivacyDialog = false
      this.started = true
      this.$store.commit('interview/SET_INTERVIEW_MODE', 'video')
      this.$store.commit('interview/SET_ENABLE_RECORDING', this.enableRecording)
      // 将语言存入 sessionStorage 供答题页使用
      sessionStorage.setItem('speechLang', this.speechLang)
      this.$router.push(`/interview/${this.interviewId}/video-doing`)
    }
  }
}
</script>

<style scoped>
.prep-layout {
  min-height: calc(100vh - 90px);
  background: #f3f3f3;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 20px;
}

.full-warning {
  text-align: center;
  padding: 60px 20px;
}
.warning-icon { font-size: 64px; }
.warning-title { font-size: 22px; font-weight: bold; margin: 16px 0 8px; }
.warning-desc { font-size: 14px; color: #666; margin-bottom: 24px; max-width: 400px; margin-left: auto; margin-right: auto; }

.prep-container {
  display: flex;
  gap: 24px;
  width: 100%;
  max-width: 1000px;
}

/* 左侧摄像头区 */
.camera-side {
  flex: 0 0 420px;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.camera-wrapper {
  width: 100%;
  /* 16:9 比例 */
  padding-top: 56.25%;
  position: relative;
  border-radius: 8px;
  overflow: hidden;
  background: #000;
  border: 2px solid #ddd;
}

.camera-wrapper > * {
  position: absolute;
  inset: 0;
}

.audio-row {
  display: flex;
  align-items: center;
  gap: 10px;
  background: #fff;
  border: 1px solid #e0e0e0;
  border-radius: 6px;
  padding: 8px 12px;
}
.audio-label { font-size: 12px; color: #555; }

/* 右侧设置区 */
.settings-side {
  flex: 1;
}

.card {
  background: #fff;
  border: 1px solid #ddd;
  border-radius: 6px;
  padding: 20px;
}

.card-title {
  font-size: 18px;
  font-weight: bold;
  margin-bottom: 20px;
  color: #111;
  border-left: 4px solid #ff9900;
  padding-left: 10px;
}

.section {
  margin-bottom: 20px;
}

.section-title {
  font-size: 13px;
  font-weight: bold;
  color: #333;
  margin-bottom: 10px;
}

/* 兼容性列表 */
.compat-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.compat-item {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
  color: #444;
}

.compat-icon {
  width: 20px;
  height: 20px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 12px;
  font-weight: bold;
  flex-shrink: 0;
}
.compat-icon.ok   { background: #e6f9ed; color: #67c23a; }
.compat-icon.warn { background: #fff3e0; color: #e6a817; }
.compat-icon.fail { background: #fee; color: #f56c6c; }

.compat-tip { font-size: 11px; color: #999; margin-left: 4px; }
.warn-text  { color: #e6a817; }

/* 录制开关 */
.switch-row {
  display: flex;
  align-items: center;
  gap: 12px;
}
.switch-hint { font-size: 12px; color: #999; }

/* 操作按钮 */
.action-row {
  display: flex;
  gap: 12px;
  margin-top: 24px;
}

.btn-start {
  flex: 1;
  font-weight: bold !important;
}

.btn-back { }

.btn-orange {
  background: #ff9900 !important;
  border-color: #ff9900 !important;
  color: #111 !important;
}
.btn-orange:hover {
  background: #f3a847 !important;
  border-color: #f3a847 !important;
}
::v-deep .el-button--primary.is-disabled {
  background: #d3d3d3 !important;
  border-color: #d3d3d3 !important;
  color: #888 !important;
}

/* 隐私声明 */
.privacy-list {
  padding-left: 20px;
  line-height: 2;
  font-size: 14px;
  color: #444;
}

@media (max-width: 768px) {
  .prep-container { flex-direction: column; }
  .camera-side { flex: none; }
}
</style>
