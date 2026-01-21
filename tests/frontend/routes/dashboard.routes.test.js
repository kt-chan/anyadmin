const request = require('supertest');
const express = require('express');
const session = require('express-session');
const dashboardRoutes = require('../../../frontend/routes/dashboard.routes');

// Mock dependencies
jest.mock('../../../frontend/services/dashboard.service', () => ({
  getOverviewData: jest.fn().mockResolvedValue({
    metrics: {
      runningServices: { onlineRate: '100%', current: 5, total: 5 },
      computeLoad: { type: 'GPU', percentage: '50%' },
      memoryUsage: { used: 8, total: 16, percentage: '50%' },
      taskQueue: { status: 'Idle', count: 0 }
    },
    services: [],
    backupInfo: {
      lastBackup: { time: 'Yesterday', type: 'Full' },
      availablePoints: 3
    },
    config: {
      concurrency: 10,
      tokenOptions: [],
      dynamicBatching: true,
      hardwareAcceleration: 'CUDA'
    },
    auditLogs: []
  })
}));

const dashboardService = require('../../../frontend/services/dashboard.service');

// Create test app
const app = express();
// Adjust path for test execution context (running from root)
app.set('views', './frontend/views'); 
app.set('view engine', 'pug');

// Mock session middleware
app.use(session({
    secret: 'test-secret',
    resave: false,
    saveUninitialized: false
}));

// Mock authentication middleware (if needed, but usually handled in route or app level)
// For this test, we simulate an authenticated user via session in the test request if possible, 
// OR we mock the middleware if it's applied on the router. 
// Looking at routes/index.js, middleware might be applied there.
// For unit testing the route specifically, we can mount it directly.

// We need to simulate req.session.user
app.use((req, res, next) => {
    req.session.user = { username: 'admin', role: 'admin' };
    next();
});

app.use('/dashboard', dashboardRoutes);

describe('Dashboard Routes', () => {
  it('GET /dashboard should render dashboard view', async () => {
    const res = await request(app).get('/dashboard');
    
    expect(res.statusCode).toBe(200);
    expect(dashboardService.getOverviewData).toHaveBeenCalled();
  });
});
