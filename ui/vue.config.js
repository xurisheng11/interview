module.exports = {
  devServer: {
    port: 3000,
    allowedHosts: 'all',
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
        proxyTimeout: 300000  // 5 分钟，和前端 axios 超时一致
      }
    }
  }
}
