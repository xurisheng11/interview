/**
 * 浏览器媒体能力兼容性检测工具
 */

/**
 * 检测浏览器媒体 API 支持情况
 * @returns {{ getUserMedia: boolean, speechRecognition: boolean, mediaRecorder: boolean, isMobile: boolean }}
 */
export function checkMediaSupport() {
  // 检查 getUserMedia（包含旧版 API fallback）
  const hasGetUserMedia = !!(
    (navigator.mediaDevices && navigator.mediaDevices.getUserMedia) ||
    navigator.getUserMedia ||
    navigator.webkitGetUserMedia ||
    navigator.mozGetUserMedia ||
    navigator.msGetUserMedia
  )

  return {
    getUserMedia: hasGetUserMedia,
    speechRecognition: !!(window.SpeechRecognition || window.webkitSpeechRecognition),
    mediaRecorder: typeof window.MediaRecorder !== 'undefined',
    isMobile: /Android|webOS|iPhone|iPad|iPod|BlackBerry|IEMobile|Opera Mini/i.test(
      navigator.userAgent
    )
  }
}

/**
 * 按优先级返回当前浏览器支持的视频录制 MIME 类型
 * @returns {string} 支持的 mimeType，若均不支持返回空字符串
 */
export function getRecommendedMimeType() {
  if (typeof window.MediaRecorder === 'undefined') return ''
  const types = [
    'video/webm;codecs=vp9',
    'video/webm;codecs=vp8',
    'video/webm',
    'video/mp4'
  ]
  for (const type of types) {
    if (MediaRecorder.isTypeSupported(type)) return type
  }
  return ''
}
