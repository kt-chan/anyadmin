const logger = require('../utils/logger');

const deploymentService = {
  nextDeploymentStep: async (currentStep, data) => {
    logger.info(`Deployment step ${currentStep} completed`, data);
    return parseInt(currentStep) + 1;
  }
};

module.exports = deploymentService;
