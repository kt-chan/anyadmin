const deploymentService = require('../../../frontend/services/deployment.service');
const fs = require('fs').promises;
const apiClient = require('../../../frontend/utils/apiClient');

// Mock fs and apiClient
jest.mock('fs', () => ({
  promises: {
    readdir: jest.fn()
  }
}));

jest.mock('../../../frontend/utils/apiClient', () => ({
  post: jest.fn(),
  get: jest.fn()
}));

// Mock logger to avoid console noise
jest.mock('../../../frontend/utils/logger', () => ({
  error: jest.fn(),
  info: jest.fn()
}));

describe('Deployment Service - Local Model Discovery', () => {
  const mockToken = 'test-token';

  afterEach(() => {
    jest.clearAllMocks();
  });

  test('should list local models when mode is "local"', async () => {
    // Mock fs.readdir to return some directories
    const mockFiles = [
      { name: 'Model-A', isDirectory: () => true },
      { name: 'Model-B', isDirectory: () => true },
      { name: 'README.md', isDirectory: () => false }
    ];
    fs.readdir.mockResolvedValue(mockFiles);

    const result = await deploymentService.discoverModels(mockToken, { mode: 'local' });

    expect(fs.readdir).toHaveBeenCalled();
    
    expect(result).toEqual({
      data: [
        { id: 'Model-A' },
        { id: 'Model-B' }
      ]
    });
    
    // Ensure apiClient was NOT called
    expect(apiClient.post).not.toHaveBeenCalled();
  });

  test('should handle "new_deployment" mode alias for local', async () => {
    const mockFiles = [{ name: 'Model-C', isDirectory: () => true }];
    fs.readdir.mockResolvedValue(mockFiles);

    const result = await deploymentService.discoverModels(mockToken, { mode: 'new_deployment' });

    expect(fs.readdir).toHaveBeenCalled();
    expect(result).toEqual({ data: [{ id: 'Model-C' }] });
  });

  test('should return empty list if local scan fails', async () => {
    fs.readdir.mockRejectedValue(new Error('ENOENT'));

    const result = await deploymentService.discoverModels(mockToken, { mode: 'local' });

    expect(result).toEqual({ data: [] });
  });

  test('should use apiClient when mode is not local', async () => {
    const mockResponse = { data: { data: [{ id: 'Remote-Model' }] } };
    apiClient.post.mockResolvedValue(mockResponse);

    const result = await deploymentService.discoverModels(mockToken, { host: '1.2.3.4', port: '8000', mode: 'remote' });

    expect(apiClient.post).toHaveBeenCalledWith(
      '/api/v1/deploy/vllm-models',
      { host: '1.2.3.4', port: '8000' },
      expect.any(Object)
    );
    expect(result).toEqual({ data: [{ id: 'Remote-Model' }] });
    expect(fs.readdir).not.toHaveBeenCalled();
  });
});
