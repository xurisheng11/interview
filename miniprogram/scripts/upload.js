// 小程序代码上传脚本（基于官方 miniprogram-ci，无需打开开发者工具）
// 用法：npm run upload -- <版本号> "<版本描述>"
// 前置：mp 后台生成上传密钥，私钥保存为 miniprogram/private.wx437f73b0e2347eb7.key（已 gitignore，绝不入库）
const path = require('path')
const ci = require('miniprogram-ci')

const version = process.argv[2] || '1.0.0'
const desc = process.argv[3] || '体验版自动上传'
const keyPath = path.resolve(__dirname, '..', 'private.wx437f73b0e2347eb7.key')

const project = new ci.Project({
  appid: 'wx437f73b0e2347eb7',
  type: 'miniProgram',
  projectPath: path.resolve(__dirname, '..'),
  privateKeyPath: keyPath,
  ignores: ['node_modules/**/*']
})

ci.upload({
  project,
  version,
  desc,
  setting: {
    es6: true,
    minify: true,
    autoPrefixWXSS: true
  }
}).then(() => {
  console.log('[upload] 成功：版本 ' + version + ' - ' + desc)
  console.log('[upload] 下一步：mp 后台 → 版本管理 → 开发版本 → 选为体验版')
}).catch((err) => {
  console.error('[upload] 失败：', err && err.message ? err.message : err)
  process.exit(1)
})
