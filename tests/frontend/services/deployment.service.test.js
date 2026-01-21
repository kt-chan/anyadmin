const deploymentService = require('../../../frontend/services/deployment.service');
const logger = require('../../../frontend/utils/logger');

jest.mock('../../../frontend/utils/logger');

describe('Deployment Service', () => {
  afterEach(() => {
    jest.clearAllMocks();
  });

  describe('nextDeploymentStep', () => {
    it('should increment step successfully', async () => {
      const result = await deploymentService.nextDeploymentStep('1', { status: 'ok' });
      
      expect(logger.info).toHaveBeenCalledWith('Deployment step 1 completed', { status: 'ok' });
      expect(result).toBe(2);
    });
  });
});
