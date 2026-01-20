const { getSystemUsersData, getSystemAuditLogs } = require('../data/mockData');

const systemController = {
  // 显示系统管理页面
  showSystem: (req, res) => {
    res.render('system', {
      user: req.session.user,
      users: getSystemUsersData(),
      auditLogs: getSystemAuditLogs(),
      page: 'system'
    });
  }
};

module.exports = systemController;