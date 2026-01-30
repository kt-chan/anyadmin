const backupService = require('../../../frontend/services/backup.service');
const apiClient = require('../../../frontend/utils/apiClient');
const logger = require('../../../frontend/utils/logger');

jest.mock('../../../frontend/utils/apiClient');
jest.mock('../../../frontend/utils/logger');

describe('Backup Service', () => {
  const mockToken = 'test-token';
  const config = { headers: { Authorization: `Bearer ${mockToken}` } };

  afterEach(() => {
    jest.clearAllMocks();
  });

  describe('getBackupData', () => {
    it('should fetch backup data successfully', async () => {
      const mockData = [{ id: 1, type: 'full' }];
      apiClient.get.mockResolvedValue({ data: mockData });

      const result = await backupService.getBackupData(mockToken);

      expect(apiClient.get).toHaveBeenCalledWith('/api/v1/backups', config);
      expect(result).toEqual({ backups: mockData });
    });

    it('should throw error when fetching fails', async () => {
      const error = new Error('Network Error');
      apiClient.get.mockRejectedValue(error);

      await expect(backupService.getBackupData(mockToken)).rejects.toThrow('Network Error');
      expect(logger.error).toHaveBeenCalledWith('Error fetching backup data', error);
    });
  });

  describe('createBackup', () => {
    it('should create a backup successfully', async () => {
      const type = 'full';
      const mockResponse = { id: 2, status: 'pending' };
      apiClient.post.mockResolvedValue({ data: mockResponse });

      const result = await backupService.createBackup(mockToken, type);

      expect(apiClient.post).toHaveBeenCalledWith('/api/v1/backups', { type }, config);
      expect(result).toEqual(mockResponse);
    });
  });

  describe('restoreFromBackup', () => {
    it('should start restore successfully', async () => {
      const backupId = 'backup-123';
      const mockResponse = { status: 'restoring' };
      apiClient.post.mockResolvedValue({ data: mockResponse });

      const result = await backupService.restoreFromBackup(mockToken, backupId);

      expect(apiClient.post).toHaveBeenCalledWith(`/api/v1/backups/restore/${backupId}`, {}, config);
      expect(result).toEqual(mockResponse);
    });
  });

  describe('deleteBackup', () => {
    it('should delete backup successfully', async () => {
      const backupId = 'backup-123';
      const result = await backupService.deleteBackup(mockToken, backupId);

      // Current implementation just returns true
      expect(result).toBe(true);
    });
  });
});
