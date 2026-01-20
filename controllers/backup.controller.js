const { getBackupsData } = require('../data/mockData');

const backupController = {
  // 显示备份恢复页面
  showBackup: (req, res) => {
    res.render('backup', {
      user: req.session.user,
      backups: getBackupsData(),
      page: 'backup'
    });
  }
};

module.exports = backupController;