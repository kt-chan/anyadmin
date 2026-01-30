const dashboardService = require('../../../frontend/services/dashboard.service');
const apiClient = require('../../../frontend/utils/apiClient');
const logger = require('../../../frontend/utils/logger');

jest.mock('../../../frontend/utils/apiClient');
jest.mock('../../../frontend/utils/logger');

describe('Dashboard Service', () => {
  const token = 'test-token';
  const mockStatsResponse = {
    data: {
      system: { cpuUsage: 10, memoryUsed: 1024, memoryTotal: 4096 },
      services: [{ name: 'vllm', status: 'Running' }],
      logs: [{ username: 'admin', action: 'login', createdAt: new Date().toISOString() }]
    }
  };

  afterEach(() => {
    jest.clearAllMocks();
  });

  describe('getOverviewData', () => {
    it('should fetch all dashboard data successfully', async () => {
      apiClient.get.mockResolvedValue(mockStatsResponse);

      const result = await dashboardService.getOverviewData(token);

      expect(apiClient.get).toHaveBeenCalledWith('/api/v1/dashboard/stats', {
        headers: { Authorization: `Bearer ${token}` }
      });
      
      expect(result).toHaveProperty('metrics');
      expect(result).toHaveProperty('services');
      expect(result).toHaveProperty('backupInfo');
      expect(result).toHaveProperty('config');
      expect(result.config).toHaveProperty('mode');
      expect(result.config.mode).toBe('balanced');
      expect(result).toHaveProperty('auditLogs');
    });

    it('should throw error when request fails', async () => {
      apiClient.get.mockRejectedValue(new Error('API Error'));

      await expect(dashboardService.getOverviewData(token)).rejects.toThrow('API Error');
      expect(logger.error).toHaveBeenCalled();
    });
  });

  describe('getMetrics', () => {
    it('should fetch metrics successfully', async () => {
        apiClient.get.mockResolvedValue({ data: { cpu: '10%' } });
        const result = await dashboardService.getMetrics(token);
        expect(apiClient.get).toHaveBeenCalledWith('/api/v1/system/stats', {
            headers: { Authorization: `Bearer ${token}` }
        });
        expect(result).toEqual({ cpu: '10%' });
    });
  });

  describe('saveConfig', () => {
    it('should save config successfully', async () => {
      const configData = { key: 'val' };
      const result = await dashboardService.saveConfig(token, configData);
      expect(logger.info).toHaveBeenCalledWith('Saving config', configData);
      expect(result).toBe(true);
    });
  });

  describe('calculateVllmConfig', () => {
    it('should call backend api correctly', async () => {
      const mockData = { model_name: 'test', node_ip: '1.2.3.4' };
      const mockResponse = { data: { vllm_config: {} } };
      apiClient.post.mockResolvedValue(mockResponse);

      const result = await dashboardService.calculateVllmConfig(token, mockData);

      expect(apiClient.post).toHaveBeenCalledWith('/api/v1/configs/vllm-calculate', mockData, {
        headers: { Authorization: `Bearer ${token}` }
      });
      expect(result).toEqual(mockResponse.data);
    });
  });
});