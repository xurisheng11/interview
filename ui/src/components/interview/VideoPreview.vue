<template>
  <div class="video-preview-wrapper">
    <video
      ref="videoEl"
      autoplay
      muted
      playsinline
      :class="{ mirrored: mirror }"
      class="preview-video"
    ></video>
    <div class="no-camera" v-if="!stream">
      <i class="el-icon-video-camera" style="font-size:32px;color:#666"></i>
      <p>摄像头未启动</p>
    </div>
  </div>
</template>

<script>
export default {
  name: 'VideoPreview',
  props: {
    stream: {
      default: null
    },
    mirror: {
      type: Boolean,
      default: true
    }
  },
  watch: {
    stream(newStream) {
      this._bindStream(newStream)
    }
  },
  mounted() {
    if (this.stream) {
      this._bindStream(this.stream)
    }
  },
  methods: {
    _bindStream(stream) {
      const video = this.$refs.videoEl
      if (!video) return
      video.srcObject = stream
      if (stream) {
        video.play().catch(() => {/* autoplay policy, ignore */})
      }
    }
  }
}
</script>

<style scoped>
.video-preview-wrapper {
  position: relative;
  width: 100%;
  height: 100%;
  background: #000;
  border-radius: 8px;
  overflow: hidden;
}

.preview-video {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
}

.mirrored {
  transform: scaleX(-1);
}

.no-camera {
  position: absolute;
  inset: 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  color: #888;
  background: #111;
}

.no-camera p {
  margin-top: 8px;
  font-size: 13px;
}
</style>
