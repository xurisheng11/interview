module.exports = {
  devServer: {
    port: 3000,
    allowedHosts: 'all',
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true
      }
    }
  }
}
