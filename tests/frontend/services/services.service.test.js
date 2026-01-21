const servicesService = require('../../../frontend/services/services.service');
const apiClient = require('../../../frontend/utils/apiClient');
const logger = require('../../../frontend/utils/logger');

jest.mock('../../../frontend/utils/apiClient');
jest.mock('../../../frontend/utils/logger');

describe('Services Service', () => {
  afterEach(() => {
    jest.clearAllMocks();
  });

  describe('getServicesList', () => {
    it('should fetch services list successfully', async () => {
      const mockServices = [{ name: 'Service 1' }];
      apiClient.get.mockResolvedValue({ data: mockServices });

      const result = await servicesService.getServicesList();

      expect(apiClient.get).toHaveBeenCalledWith('/services');
      expect(result).toEqual({ services: mockServices });
    });

    it('should throw error on failure', async () => {
      apiClient.get.mockRejectedValue(new Error('API Error'));
      await expect(servicesService.getServicesList()).rejects.toThrow('API Error');
      expect(logger.error).toHaveBeenCalled();
    });
  });

  describe('getServicesStatus', () => {
    it('should fetch services status successfully', async () => {
      const mockStatus = [{ id: 1, status: 'running' }];
      apiClient.get.mockResolvedValue({ data: mockStatus });

      const result = await servicesService.getServicesStatus();

      expect(apiClient.get).toHaveBeenCalledWith('/dashboard/services');
      expect(result).toEqual(mockStatus);
    });
  });

  describe('restartService', () => {
    it('should restart service successfully', async () => {
      const result = await servicesService.restartService('svc_1');
      expect(logger.info).toHaveBeenCalledWith(expect.stringContaining('Restarting service'));
      expect(result).toBe(true);
    });
  });

  describe('stopService', () => {
    it('should stop service successfully', async () => {
      const result = await servicesService.stopService('svc_1');
      expect(logger.info).toHaveBeenCalledWith(expect.stringContaining('Stopping service'));
      expect(result).toBe(true);
    });
  });
});
