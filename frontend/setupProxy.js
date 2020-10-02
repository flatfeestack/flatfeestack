// https://create-react-app.dev/docs/proxying-api-requests-in-development/
const { createProxyMiddleware } = require("http-proxy-middleware");

// we use this proxy to allow "../api" requests which work for production build
// also that way, cors enabling is avoided
module.exports = function (app) {
  app.use(
    "/api*",
    createProxyMiddleware({
      target: "http://localhost:8000",
      pathRewrite: { "^/auth": "" },
      changeOrigin: false,
    })
  );
};
