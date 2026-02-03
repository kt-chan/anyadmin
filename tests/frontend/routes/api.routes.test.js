const request = require('supertest');
const express = require('express');
const session = require('express-session');
const apiRoutes = require('../../../frontend/routes/api.routes');

// Mock dependencies
jest.mock('../../../frontend/services/services.service', () => ({
  restartService: jest.fn().mockResolvedValue(true),
  stopService: jest.fn().mockResolvedValue(true),
  controlAgent: jest.fn().mockResolvedValue(true),
  getServicesStatus: jest.fn().mockResolvedValue([])
}));

jest.mock('../../../frontend/services/dashboard.service', () => ({
  saveConfig: jest.fn().mockResolvedValue(true)
}));

const servicesService = require('../../../frontend/services/services.service');
const dashboardService = require('../../../frontend/services/dashboard.service');

// Create test app
const app = express();
app.use(express.json());

// Mock session middleware
app.use(session({
    secret: 'test-secret',
    resave: false,
    saveUninitialized: false
}));

// Simulate an authenticated user
app.use((req, res, next) => {
    req.session.user = { username: 'admin', role: 'admin', token: 'fake-token' };
    next();
});

// Mock requireLogin middleware (it's used in api.routes.js)
// Since we are mocking the session above, requireLogin should pass if it checks req.session.user
// But we can also mock it if needed. Let's see if it works without mocking.

app.use('/api', apiRoutes);

describe('API Routes (Service Operations)', () => {
  afterEach(() => {
    jest.clearAllMocks();
  });

  it('POST /api/service/restart should call servicesService.restartService', async () => {
    const payload = { name: 'vllm', node_ip: '172.20.0.10', type: 'Container' };
    const res = await request(app)
      .post('/api/service/restart')
      .send(payload);
    
    expect(res.statusCode).toBe(200);
    expect(res.body.success).toBe(true);
    expect(servicesService.restartService).toHaveBeenCalledWith('vllm', '172.20.0.10', 'fake-token', 'Container');
  });

  it('POST /api/service/stop should call servicesService.stopService', async () => {
    const payload = { name: 'vllm', node_ip: '172.20.0.10', type: 'Container' };
    const res = await request(app)
      .post('/api/service/stop')
      .send(payload);
    
    expect(res.statusCode).toBe(200);
    expect(res.body.success).toBe(true);
    expect(servicesService.stopService).toHaveBeenCalledWith('vllm', '172.20.0.10', 'fake-token', 'Container');
  });

  it('POST /api/agent/control should call servicesService.controlAgent', async () => {
    const payload = { ip: '172.20.0.10', action: 'stop' };
    const res = await request(app)
      .post('/api/agent/control')
      .send(payload);
    
    expect(res.statusCode).toBe(200);
    expect(res.body.success).toBe(true);
    expect(servicesService.controlAgent).toHaveBeenCalledWith('172.20.0.10', 'stop', 'fake-token');
  });

  it('POST /api/config/save should call dashboardService.saveConfig', async () => {
    const payload = { name: 'default', mode: 'balanced' };
    const res = await request(app)
      .post('/api/config/save')
      .send(payload);
    
    expect(res.statusCode).toBe(200);
    expect(res.body.success).toBe(true);
    expect(dashboardService.saveConfig).toHaveBeenCalledWith('fake-token', payload);
  });

  it('GET /api/services/status should call servicesService.getServicesStatus', async () => {
    const mockServices = [{ id: 1, name: 'vllm', status: 'Running' }];
    servicesService.getServicesStatus.mockResolvedValue(mockServices);

    const res = await request(app)
      .get('/api/services/status');
    
    expect(res.statusCode).toBe(200);
    expect(res.body.success).toBe(true);
    expect(res.body.data.services).toHaveLength(1);
    expect(res.body.data.services[0].name).toBe('vllm');
    expect(servicesService.getServicesStatus).toHaveBeenCalledWith('fake-token');
  });
});
