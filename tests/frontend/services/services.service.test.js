const servicesService = require('../../../frontend/services/services.service');
const apiClient = require('../../../frontend/utils/apiClient');
const logger = require('../../../frontend/utils/logger');

jest.mock('../../../frontend/utils/apiClient');
jest.mock('../../../frontend/utils/logger');

describe('Services Service', () => {
  const token = 'test-token';
  const mockServices = [
    { name: 'vllm', status: 'Running', type: 'Container' }
  ];

  afterEach(() => {
    jest.clearAllMocks();
  });

  describe('getServicesList', () => {
    it('should fetch services list successfully', async () => {
      apiClient.get.mockResolvedValue({ data: { services: mockServices } });

      const result = await servicesService.getServicesList(token);

      expect(apiClient.get).toHaveBeenCalledWith('/api/v1/dashboard/stats', {
        headers: { Authorization: `Bearer ${token}` }
      });
      expect(result).toEqual({ services: mockServices });
    });
  });

  describe('getServicesStatus', () => {
    it('should fetch services status successfully', async () => {
      apiClient.get.mockResolvedValue({ data: { services: mockServices } });

      const result = await servicesService.getServicesStatus(token);

      expect(apiClient.get).toHaveBeenCalledWith('/api/v1/dashboard/stats', {
        headers: { Authorization: `Bearer ${token}` }
      });
      expect(result).toEqual(mockServices);
    });
  });

  describe('restartService', () => {
    it('should restart container service successfully', async () => {
      apiClient.post.mockResolvedValue({ data: { message: 'ok' } });

      const result = await servicesService.restartService('vllm', '172.20.0.10', token, 'Container');

      expect(apiClient.post).toHaveBeenCalledWith('/api/v1/container/control', 
        { name: 'vllm', action: 'restart', node_ip: '172.20.0.10' },
        { headers: { Authorization: `Bearer ${token}` } }
      );
      expect(result).toBe(true);
    });

    it('should restart agent successfully', async () => {
      apiClient.post.mockResolvedValue({ data: { message: 'ok' } });

      const result = await servicesService.restartService('Agent (node1)', '172.20.0.10', token, 'Agent');

      expect(apiClient.post).toHaveBeenCalledWith('/api/v1/deploy/agent/control', 
        { ip: '172.20.0.10', action: 'restart' },
        { headers: { Authorization: `Bearer ${token}` } }
      );
      expect(result).toBe(true);
    });
  });

  describe('stopService', () => {
    it('should stop container service successfully', async () => {
      apiClient.post.mockResolvedValue({ data: { message: 'ok' } });

      const result = await servicesService.stopService('vllm', '172.20.0.10', token, 'Container');

      expect(apiClient.post).toHaveBeenCalledWith('/api/v1/container/control', 
        { name: 'vllm', action: 'stop', node_ip: '172.20.0.10' },
        { headers: { Authorization: `Bearer ${token}` } }
      );
      expect(result).toBe(true);
    });

    it('should stop agent successfully', async () => {
      apiClient.post.mockResolvedValue({ data: { message: 'ok' } });

      const result = await servicesService.stopService('Agent (node1)', '172.20.0.10', token, 'Agent');

      expect(apiClient.post).toHaveBeenCalledWith('/api/v1/deploy/agent/control', 
        { ip: '172.20.0.10', action: 'stop' },
        { headers: { Authorization: `Bearer ${token}` } }
      );
      expect(result).toBe(true);
    });
  });
});