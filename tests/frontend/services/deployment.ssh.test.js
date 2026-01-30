const deploymentService = require('../../../frontend/services/deployment.service');
const apiClient = require('../../../frontend/utils/apiClient');

jest.mock('../../../frontend/utils/apiClient');
jest.mock('../../../frontend/utils/logger');

describe('Deployment Service SSH Features', () => {
  afterEach(() => {
    jest.clearAllMocks();
  });

  describe('getSSHKey', () => {
    it('should call api to get key', async () => {
      apiClient.get.mockResolvedValue({ data: 'ssh-rsa MOCK_KEY' });
      
      const key = await deploymentService.getSSHKey('mock-token');
      
      expect(apiClient.get).toHaveBeenCalledWith('/api/v1/deploy/ssh-key', expect.objectContaining({
        headers: { Authorization: 'Bearer mock-token' },
        responseType: 'text'
      }));
      expect(key).toBe('ssh-rsa MOCK_KEY');
    });
  });

  describe('testConnection', () => {
    it('should call verify-ssh endpoint for ssh type', async () => {
      const payload = { host: '127.0.0.1:22', type: 'ssh' };
      apiClient.post.mockResolvedValue({ data: { status: 'success' } });

      const result = await deploymentService.testConnection('mock-token', payload);
      
      expect(apiClient.post).toHaveBeenCalledWith('/api/v1/deploy/test-connection', payload, expect.objectContaining({
        headers: { Authorization: 'Bearer mock-token' }
      }));
      expect(result).toEqual({ status: 'success' });
    });

    it('should call test-connection for other types', async () => {
       const payload = { type: 'inference' };
       apiClient.post.mockResolvedValue({ data: { success: true } });

       const result = await deploymentService.testConnection('mock-token', payload);
       
       expect(apiClient.post).toHaveBeenCalledWith('/api/v1/deploy/test-connection', payload, expect.objectContaining({
         headers: { Authorization: 'Bearer mock-token' }
       }));
       expect(result).toEqual({ success: true });
    });
  });
});
