const request = require('supertest');
const express = require('express');
const session = require('express-session');
const backupRoutes = require('../../../frontend/routes/backup.routes');

// Mock dependencies
jest.mock('../../../frontend/services/backup.service', () => ({
  getBackupData: jest.fn().mockResolvedValue({
    backups: [
      { id: 'backup-001', time: '2024-05-20 10:00:00', type: 'FULL', verified: true, size: '1.2 GB' },
      { id: 'backup-002', time: '2024-05-21 10:00:00', type: 'INC', verified: false, size: '450 MB' }
    ]
  })
}));

const backupService = require('../../../frontend/services/backup.service');

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

app.use('/backup', backupRoutes);

describe('Backup Routes Rendering', () => {
  afterEach(() => {
    jest.clearAllMocks();
  });

  it('GET /backup should render backup view with data', async () => {
    const res = await request(app).get('/backup');
    
    expect(res.statusCode).toBe(200);
    expect(backupService.getBackupData).toHaveBeenCalled();
    // Check for some content in the rendered HTML
    expect(res.text).toContain('备份恢复');
    expect(res.text).toContain('backup-001');
    expect(res.text).toContain('FULL');
    expect(res.text).toContain('1.2 GB');
  });

  it('GET /backup should render empty state when no backups found', async () => {
    backupService.getBackupData.mockResolvedValueOnce({ backups: [] });
    
    const res = await request(app).get('/backup');
    
    expect(res.statusCode).toBe(200);
    expect(res.text).toContain('暂无备份记录');
  });

  it('GET /backup should render empty state when backups is null', async () => {
    backupService.getBackupData.mockResolvedValueOnce({ backups: null });
    
    const res = await request(app).get('/backup');
    
    expect(res.statusCode).toBe(200);
    expect(res.text).toContain('暂无备份记录');
  });
});
