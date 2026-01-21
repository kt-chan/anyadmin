const backupService = require('../../../frontend/services/backup.service');
const apiClient = require('../../../frontend/utils/apiClient');
const logger = require('../../../frontend/utils/logger');

jest.mock('../../../frontend/utils/apiClient');
jest.mock('../../../frontend/utils/logger');

describe('Backup Service', () => {
  afterEach(() => {
    jest.clearAllMocks();
  });

  describe('getBackupData', () => {
    it('should fetch backup data successfully', async () => {
      const mockData = [{ id: 'bk_1', time: '2024-01-01' }];
      apiClient.get.mockResolvedValue({ data: mockData });

      const result = await backupService.getBackupData();

      expect(apiClient.get).toHaveBeenCalledWith('/backups');
      expect(result).toEqual({ backups: mockData });
    });

    it('should throw error when fetching fails', async () => {
      const error = new Error('Network Error');
      apiClient.get.mockRejectedValue(error);

      await expect(backupService.getBackupData()).rejects.toThrow('Network Error');
      expect(logger.error).toHaveBeenCalled();
    });
  });

  describe('createBackup', () => {
    it('should create a backup successfully', async () => {
      const result = await backupService.createBackup('FULL');
      
      expect(logger.info).toHaveBeenCalledWith(expect.stringContaining('Creating backup'));
      expect(result).toHaveProperty('backupId');
      expect(result).toHaveProperty('startTime');
    });
  });

  describe('restoreFromBackup', () => {
    it('should start restore successfully', async () => {
      const result = await backupService.restoreFromBackup('bk_123');
      
      expect(logger.info).toHaveBeenCalledWith(expect.stringContaining('Restoring'));
      expect(result).toHaveProperty('restoreStartTime');
    });
  });

  describe('deleteBackup', () => {
    it('should delete backup successfully', async () => {
      const result = await backupService.deleteBackup('bk_123');
      
      expect(logger.info).toHaveBeenCalledWith(expect.stringContaining('Deleting'));
      expect(result).toBe(true);
    });
  });
});
