const request = require('supertest');
const express = require('express');
const session = require('express-session');
const path = require('path');

// Mock dependencies
jest.mock('../../../frontend/services/models.service', () => ({
  getModels: jest.fn().mockResolvedValue({
    models: [
      { name: 'Qwen3-1.7B', size: 3435973836, updated_at: '2026-02-09T00:00:00Z' }
    ]
  }),
  uploadModel: jest.fn().mockResolvedValue({ status: 'success' }),
  deleteModel: jest.fn().mockResolvedValue({ status: 'success' })
}));

const modelsService = require('../../../frontend/services/models.service');
const modelsRoutes = require('../../../frontend/routes/models.routes');

const app = express();
app.set('views', path.join(__dirname, '../../../frontend/views'));
app.set('view engine', 'pug');

app.use(express.json());
app.use(session({
    secret: 'test-secret',
    resave: false,
    saveUninitialized: false
}));

app.use((req, res, next) => {
    req.session.user = { username: 'admin', role: 'admin' };
    next();
});

app.use('/models', modelsRoutes);

describe('Models Routes', () => {
  it('GET /models should render models view', async () => {
    const res = await request(app).get('/models');
    
    expect(res.statusCode).toBe(200);
    expect(modelsService.getModels).toHaveBeenCalled();
  });

  it('GET /models/api should return models list', async () => {
    const res = await request(app).get('/models/api');
    
    expect(res.statusCode).toBe(200);
    expect(res.body.success).toBe(true);
    expect(res.body.data).toHaveLength(1);
  });

  it('DELETE /models/api/:name should delete model', async () => {
    const res = await request(app).delete('/models/api/test-model');
    
    expect(res.statusCode).toBe(200);
    expect(res.body.success).toBe(true);
    expect(modelsService.deleteModel).toHaveBeenCalledWith(undefined, 'test-model');
  });
});
