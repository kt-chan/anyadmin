const systemService = require('../../../frontend/services/system.service');
const apiClient = require('../../../frontend/utils/apiClient');
const logger = require('../../../frontend/utils/logger');

jest.mock('../../../frontend/utils/apiClient');
jest.mock('../../../frontend/utils/logger');

describe('System Service', () => {
  afterEach(() => {
    jest.clearAllMocks();
  });

  describe('getSystemData', () => {
    it('should fetch system data successfully', async () => {
      apiClient.get.mockImplementation((url) => {
        if (url === '/system/users') return Promise.resolve({ data: [{ user: 1 }] });
        if (url === '/system/audit-logs') return Promise.resolve({ data: [{ log: 1 }] });
        return Promise.reject(new Error(`Unknown URL: ${url}`));
      });

      const result = await systemService.getSystemData();

      expect(apiClient.get).toHaveBeenCalledTimes(2);
      expect(result).toEqual({
        users: [{ user: 1 }],
        auditLogs: [{ log: 1 }]
      });
    });

    it('should throw error on failure', async () => {
      apiClient.get.mockRejectedValue(new Error('API Error'));
      await expect(systemService.getSystemData()).rejects.toThrow('API Error');
      expect(logger.error).toHaveBeenCalled();
    });
  });

  describe('createUser', () => {
    it('should create user successfully', async () => {
      const result = await systemService.createUser('newuser', 'admin');
      expect(logger.info).toHaveBeenCalledWith(expect.stringContaining('Creating user'));
      expect(result).toHaveProperty('userId');
    });
  });

  describe('appReflash', () => {
    it('should start app reflash successfully', async () => {
      const result = await systemService.appReflash();
      expect(logger.info).toHaveBeenCalledWith('Starting app reflash');
      expect(result).toHaveProperty('estimatedTime');
    });
  });
});
