
module.exports = {
  productionSourceMap: false,
  devServer: {
    port: 8081,
    disableHostCheck: true,
    proxy: {
      "/train": {
        target: "http://127.0.0.1:8080"
      }
    }
  }
}
