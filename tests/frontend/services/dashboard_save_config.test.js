const dashboardService = require('../../../frontend/services/dashboard.service');
const apiClient = require('../../../frontend/utils/apiClient');

jest.mock('../../../frontend/utils/apiClient');

describe('Dashboard Service - saveConfig', () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  it('should call backend API to save config and return data', async () => {
    const mockToken = 'test-token';
    const mockConfigData = {
        name: 'test-model',
        mode: 'balanced',
        max_model_len: 4096,
        max_num_seqs: 256,
        max_num_batched_tokens: 2048,
        gpu_memory_utilization: 0.85
    };
    const mockResponse = { data: { ...mockConfigData, status: 'saved' } };

    apiClient.post.mockResolvedValue(mockResponse);

    const result = await dashboardService.saveConfig(mockToken, mockConfigData);

    expect(result).toEqual(mockResponse.data);
    expect(apiClient.post).toHaveBeenCalledTimes(1);
    expect(apiClient.post).toHaveBeenCalledWith(
      '/api/v1/configs/inference',
      mockConfigData,
      expect.objectContaining({
        headers: { Authorization: `Bearer ${mockToken}` }
      })
    );
  });

  it('should throw error when backend API fails', async () => {
    const mockToken = 'test-token';
    const mockConfigData = { name: 'test-model', mode: 'balanced' };
    const mockError = new Error('API Error');

    apiClient.post.mockRejectedValue(mockError);

    await expect(dashboardService.saveConfig(mockToken, mockConfigData)).rejects.toThrow('API Error');
  });
});
