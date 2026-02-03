const request = require('supertest');
const express = require('express');
const session = require('express-session');
const importRoutes = require('../../../frontend/routes/import.routes');

// Mock dependencies
jest.mock('../../../frontend/services/import.service', () => ({
  getTasks: jest.fn().mockResolvedValue([
    { 
      id: 'task-001', 
      name: 'Import 1', 
      status: 'PROCESSING', 
      sourceType: 'LOCAL', 
      sourcePath: '/data',
      progress: { total: 100, processed: 50, failed: 0 },
      schedule: 'Daily',
      nextScan: 'Tomorrow'
    }
  ]),
  createTask: jest.fn()
}));

const importService = require('../../../frontend/services/import.service');

// Create test app
const app = express();
app.set('views', './frontend/views');
app.set('view engine', 'pug');

// Mock session middleware
app.use(session({
    secret: 'test-secret',
    resave: false,
    saveUninitialized: false
}));

// Simulate authenticated user
app.use((req, res, next) => {
    req.session.user = { username: 'admin', role: 'admin', token: 'fake-token' };
    next();
});

app.use('/import', importRoutes);

describe('Import Routes Rendering', () => {
  afterEach(() => {
    jest.clearAllMocks();
  });

  it('GET /import should render import view with data', async () => {
    const res = await request(app).get('/import');
    
    expect(res.statusCode).toBe(200);
    expect(importService.getTasks).toHaveBeenCalled();
    expect(res.text).toContain('批量文件导入');
    expect(res.text).toContain('Import 1');
    expect(res.text).toContain('task-001');
  });

  it('GET /import should handle null tasks from service', async () => {
    importService.getTasks.mockResolvedValueOnce(null);
    
    const res = await request(app).get('/import');
    
    expect(res.statusCode).toBe(200);
    expect(res.text).toContain('暂无导入任务');
  });

  it('GET /import should handle empty array from service', async () => {
    importService.getTasks.mockResolvedValueOnce([]);
    
    const res = await request(app).get('/import');
    
    expect(res.statusCode).toBe(200);
    expect(res.text).toContain('暂无导入任务');
  });
});
