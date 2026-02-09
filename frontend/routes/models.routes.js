const express = require('express');
const router = express.Router();
const modelsController = require('../controllers/models.controller');
const { requireLogin } = require('../middleware/auth.middleware');
const multer = require('multer');
const path = require('path');
const fs = require('fs');

// Configure multer for temporary storage
const uploadDir = path.join(__dirname, '../../.gemini/tmp/uploads');
if (!fs.existsSync(uploadDir)) {
    fs.mkdirSync(uploadDir, { recursive: true });
}

const storage = multer.diskStorage({
  destination: function (req, file, cb) {
    cb(null, uploadDir);
  },
  filename: function (req, file, cb) {
    cb(null, Date.now() + '-' + file.originalname);
  }
});

const upload = multer({ storage: storage });

// All routes require authentication
router.use(requireLogin);

// View routes
router.get('/', modelsController.getModelsPage);

// API routes
router.get('/api', modelsController.getModels);
router.post('/api/upload/init', modelsController.initUpload);
router.post('/api/upload/chunk', upload.single('chunk'), modelsController.uploadChunk);
router.post('/api/upload/abort', modelsController.abortUpload);
router.post('/api/finalize', modelsController.finalizeUpload);
router.delete('/api/:name', modelsController.deleteModel);

module.exports = router;