const systemService = require('../../../frontend/services/system.service');
const apiClient = require('../../../frontend/utils/apiClient');
const logger = require('../../../frontend/utils/logger');

jest.mock('../../../frontend/utils/apiClient');
jest.mock('../../../frontend/utils/logger');

describe('System Service', () => {
  const mockToken = 'test-token';
  const config = { headers: { Authorization: `Bearer ${mockToken}` } };

  afterEach(() => {
    jest.clearAllMocks();
  });

  describe('getSystemData', () => {
    it('should fetch system data successfully', async () => {
      // Mock two calls: /api/v1/users and /api/v1/dashboard/stats
      apiClient.get.mockImplementation((url) => {
        if (url === '/api/v1/users') return Promise.resolve({ data: [{ user: 1 }] });
        if (url === '/api/v1/dashboard/stats') return Promise.resolve({ data: { logs: [{ username: 'test', action: 'login', createdAt: '2023-01-01' }] } });
        return Promise.reject(new Error(`Unknown URL: ${url}`));
      });

      const result = await systemService.getSystemData(mockToken);

      expect(apiClient.get).toHaveBeenCalledWith('/api/v1/users', config);
      expect(apiClient.get).toHaveBeenCalledWith('/api/v1/dashboard/stats', config);
      expect(result.users).toHaveLength(1);
      expect(result.auditLogs).toHaveLength(1);
    });

    it('should throw error on failure', async () => {
      apiClient.get.mockRejectedValue(new Error('API Error'));
      await expect(systemService.getSystemData(mockToken)).rejects.toThrow('API Error');
    });
  });

  describe('createUser', () => {
    it('should create user successfully', async () => {
      const username = 'test';
      const role = 'admin';
      const password = 'pass';
      
      apiClient.post.mockResolvedValue({ data: { id: 1, username } });

      const result = await systemService.createUser(mockToken, username, role, password);

      expect(apiClient.post).toHaveBeenCalledWith('/api/v1/users', { username, role, password }, config);
      expect(result).toEqual({ id: 1, username });
    });
  });

  describe('appReflash', () => {
    it('should start app reflash successfully', async () => {
      const result = await systemService.appReflash(mockToken);
      expect(result).toEqual({ estimatedTime: '10分钟' });
    });
  });
});
