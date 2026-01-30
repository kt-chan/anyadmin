const deploymentService = require('../../../frontend/services/deployment.service');
const apiClient = require('../../../frontend/utils/apiClient');
const logger = require('../../../frontend/utils/logger');

jest.mock('../../../frontend/utils/apiClient');
jest.mock('../../../frontend/utils/logger');

describe('Deployment Service', () => {
  const mockToken = 'test-token';
  const config = { headers: { Authorization: `Bearer ${mockToken}` } };

  afterEach(() => {
    jest.clearAllMocks();
  });

  describe('generateDeployment', () => {
    it('should call generate endpoint', async () => {
      const deployConfig = { mode: 'new' };
      apiClient.post.mockResolvedValue({ data: { success: true } });

      const result = await deploymentService.generateDeployment(mockToken, deployConfig);

      expect(apiClient.post).toHaveBeenCalledWith('/api/v1/deploy/generate', deployConfig, config);
      expect(result).toEqual({ success: true });
    });
  });

  describe('controlAgent', () => {
    it('should call agent control endpoint', async () => {
      const ip = '1.1.1.1';
      const action = 'restart';
      apiClient.post.mockResolvedValue({ data: { success: true } });

      const result = await deploymentService.controlAgent(mockToken, { ip, action });

      expect(apiClient.post).toHaveBeenCalledWith('/api/v1/deploy/agent/control', { ip, action }, config);
      expect(result).toEqual({ success: true });
    });
  });

  describe('removeNode', () => {
    it('should call delete endpoint for node', async () => {
      const ip = '1.1.1.1';
      apiClient.delete.mockResolvedValue({ data: { success: true } });

      const result = await deploymentService.removeNode(mockToken, ip);

      expect(apiClient.delete).toHaveBeenCalledWith('/api/v1/deploy/nodes', {
        headers: { Authorization: `Bearer ${mockToken}` },
        params: { ip }
      });
      expect(result).toEqual({ success: true });
    });
  });
});