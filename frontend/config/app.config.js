const appConfig = {
  port: process.env.PORT || 3000,
  env: process.env.NODE_ENV || 'development',
  appName: '知识库管理平台',
  version: '1.0.0',
  backendApiUrl: process.env.BACKEND_API_URL || 'http://127.0.0.1:8080'
};

module.exports = appConfig;
