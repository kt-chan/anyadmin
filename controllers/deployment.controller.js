const deploymentController = {
  // 显示部署配置页面
  showDeployment: (req, res) => {
    res.render('deployment', {
      user: req.session.user,
      page: 'deployment'
    });
  }
};

module.exports = deploymentController;