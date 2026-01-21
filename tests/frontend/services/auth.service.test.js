const authService = require('../../../frontend/services/auth.service');
const apiClient = require('../../../frontend/utils/apiClient');

// Mock apiClient
jest.mock('../../../frontend/utils/apiClient');

describe('Auth Service', () => {
  afterEach(() => {
    jest.clearAllMocks();
  });

  it('should authenticate user successfully', async () => {
    const mockUser = { username: 'admin', role: 'admin' };
    apiClient.post.mockResolvedValue({ data: mockUser });

    const result = await authService.authenticate('admin', 'password');
    
    expect(apiClient.post).toHaveBeenCalledWith('/auth/login', { username: 'admin', password: 'password' });
    expect(result).toEqual(mockUser);
  });

  it('should return null on failed authentication', async () => {
    // Mock axios error
    const error = new Error('Authentication failed');
    error.response = { status: 401, statusText: 'Unauthorized' };
    apiClient.post.mockRejectedValue(error);

    const result = await authService.authenticate('admin', 'wrongpassword');
    
    expect(apiClient.post).toHaveBeenCalledWith('/auth/login', { username: 'admin', password: 'wrongpassword' });
    expect(result).toBeNull();
  });
});
