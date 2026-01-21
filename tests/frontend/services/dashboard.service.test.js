const dashboardService = require('../../../frontend/services/dashboard.service');
const apiClient = require('../../../frontend/utils/apiClient');
const logger = require('../../../frontend/utils/logger');

jest.mock('../../../frontend/utils/apiClient');
jest.mock('../../../frontend/utils/logger');

describe('Dashboard Service', () => {
  afterEach(() => {
    jest.clearAllMocks();
  });

  describe('getOverviewData', () => {
    it('should fetch all dashboard data successfully', async () => {
      apiClient.get.mockImplementation((url) => {
        switch (url) {
          case '/dashboard/metrics': return Promise.resolve({ data: { metrics: true } });
          case '/dashboard/services': return Promise.resolve({ data: { services: true } });
          case '/backup/info': return Promise.resolve({ data: { backup: true } });
          case '/dashboard/config': return Promise.resolve({ data: { config: true } });
          case '/dashboard/audit-logs': return Promise.resolve({ data: { logs: true } });
          default: return Promise.reject(new Error(`Unknown URL: ${url}`));
        }
      });

      const result = await dashboardService.getOverviewData();

      expect(apiClient.get).toHaveBeenCalledTimes(5);
      expect(result).toEqual({
        metrics: { metrics: true },
        services: { services: true },
        backupInfo: { backup: true },
        config: { config: true },
        auditLogs: { logs: true }
      });
    });

    it('should throw error when any request fails', async () => {
      apiClient.get.mockRejectedValue(new Error('API Error'));

      await expect(dashboardService.getOverviewData()).rejects.toThrow('API Error');
      expect(logger.error).toHaveBeenCalled();
    });
  });

  describe('getMetrics', () => {
    it('should fetch metrics successfully', async () => {
        apiClient.get.mockResolvedValue({ data: { cpu: '10%' } });
        const result = await dashboardService.getMetrics();
        expect(apiClient.get).toHaveBeenCalledWith('/dashboard/metrics');
        expect(result).toEqual({ cpu: '10%' });
    });
  });

  describe('saveConfig', () => {
    it('should save config successfully', async () => {
      const result = await dashboardService.saveConfig({ key: 'val' });
      expect(logger.info).toHaveBeenCalledWith('Saving config', { key: 'val' });
      expect(result).toBe(true);
    });
  });

  describe('updateConcurrency', () => {
    it('should update concurrency successfully', async () => {
      const result = await dashboardService.updateConcurrency(10);
      expect(logger.info).toHaveBeenCalledWith('Updating concurrency to 10');
      expect(result).toBe(10);
    });
  });

  describe('updateTokenLimit', () => {
    it('should update token limit successfully', async () => {
      const result = await dashboardService.updateTokenLimit(100);
      expect(logger.info).toHaveBeenCalledWith('Updating token limit to 100');
      expect(result).toBe(100);
    });
  });
});
