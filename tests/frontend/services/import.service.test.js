const importService = require('../../../frontend/services/import.service');
const apiClient = require('../../../frontend/utils/apiClient');
const logger = require('../../../frontend/utils/logger');

jest.mock('../../../frontend/utils/apiClient');
jest.mock('../../../frontend/utils/logger');

describe('Import Service', () => {
  afterEach(() => {
    jest.clearAllMocks();
  });

  describe('getTasks', () => {
    it('should fetch tasks successfully', async () => {
      const mockTasks = [{ id: 1, name: 'Task 1' }];
      apiClient.get.mockResolvedValue({ data: mockTasks });

      const result = await importService.getTasks();

      expect(apiClient.get).toHaveBeenCalledWith('/import/tasks');
      expect(result).toEqual(mockTasks);
    });

    it('should throw error on failure', async () => {
      apiClient.get.mockRejectedValue(new Error('API Error'));
      await expect(importService.getTasks()).rejects.toThrow('API Error');
      expect(logger.error).toHaveBeenCalled();
    });
  });

  describe('createTask', () => {
    it('should create task successfully', async () => {
      const task = { name: 'New Task' };
      apiClient.post.mockResolvedValue({ data: task });

      const result = await importService.createTask(task);

      expect(apiClient.post).toHaveBeenCalledWith('/import/tasks', task);
      expect(result).toEqual(task);
    });

    it('should throw error on failure', async () => {
      apiClient.post.mockRejectedValue(new Error('API Error'));
      await expect(importService.createTask({})).rejects.toThrow('API Error');
      expect(logger.error).toHaveBeenCalled();
    });
  });

  describe('updateTask', () => {
    it('should update task successfully', async () => {
      const updates = { status: 'DONE' };
      apiClient.put.mockResolvedValue({ data: { id: 1, ...updates } });

      const result = await importService.updateTask(1, updates);

      expect(apiClient.put).toHaveBeenCalledWith('/import/tasks/1', updates);
      expect(result).toEqual({ id: 1, status: 'DONE' });
    });

    it('should throw error on failure', async () => {
      apiClient.put.mockRejectedValue(new Error('API Error'));
      await expect(importService.updateTask(1, {})).rejects.toThrow('API Error');
      expect(logger.error).toHaveBeenCalled();
    });
  });

  describe('deleteTask', () => {
    it('should delete task successfully', async () => {
      apiClient.delete.mockResolvedValue({ data: { success: true } });

      const result = await importService.deleteTask(1);

      expect(apiClient.delete).toHaveBeenCalledWith('/import/tasks/1');
      expect(result).toEqual({ success: true });
    });

    it('should throw error on failure', async () => {
      apiClient.delete.mockRejectedValue(new Error('API Error'));
      await expect(importService.deleteTask(1)).rejects.toThrow('API Error');
      expect(logger.error).toHaveBeenCalled();
    });
  });
});
