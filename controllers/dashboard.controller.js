const {
  getDashboardMetrics,
  getDashboardServices,
  getBackupInfo,
  getDashboardConfig,
  getDashboardAuditLogs
} = require('../data/mockData');

const dashboardController = {
  // 显示仪表板
  showDashboard: (req, res) => {
    res.render('dashboard', {
      user: req.session.user,
      page: 'dashboard',
      metrics: getDashboardMetrics(),
      services: getDashboardServices(),
      backupInfo: getBackupInfo(),
      config: getDashboardConfig(),
      auditLogs: getDashboardAuditLogs()
    });
  }
};

module.exports = dashboardController;