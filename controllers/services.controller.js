const { getServicesData } = require('../data/mockData');

const servicesController = {
  // 显示服务管理页面
  showServices: (req, res) => {
    res.render('services', {
      user: req.session.user,
      services: getServicesData(),
      page: 'services'
    });
  }
};

module.exports = servicesController;