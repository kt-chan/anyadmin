const modelsService = require('../services/models.service');
const logger = require('../utils/logger');

// Display the models management page
exports.getModelsPage = async (req, res) => {
  try {
    const token = req.session.user?.token;
    const data = await modelsService.getModels(token);
    const models = data.models || [];

    res.render('models', {
      page: 'models',
      user: req.session.user,
      models: models
    });
  } catch (error) {
    logger.error('Error loading models page:', error);
    res.status(500).render('error', { error });
  }
};

// API: Get models list
exports.getModels = async (req, res) => {
  try {
    const token = req.session.user?.token;
    const data = await modelsService.getModels(token);
    res.json({ success: true, data: data.models });
  } catch (error) {
    res.status(500).json({ success: false, message: error.message });
  }
};

// API: Init Upload
exports.initUpload = async (req, res) => {
  try {
    const token = req.session.user?.token;
    const result = await modelsService.initUpload(token, req.body);
    res.json(result); // { upload_id, offset, ... }
  } catch (error) {
    res.status(500).json({ error: error.message });
  }
};

// API: Upload Chunk
exports.uploadChunk = async (req, res) => {
  try {
    const token = req.session.user?.token;
    const uploadId = req.body.upload_id;
    const chunkFile = req.file; // From multer single('chunk')

    if (!uploadId || !chunkFile) {
        return res.status(400).json({ error: "Missing upload_id or chunk" });
    }

    const result = await modelsService.uploadChunk(token, uploadId, chunkFile);
    
    // Clean up temp file
    const fs = require('fs');
    fs.unlink(chunkFile.path, () => {});

    res.json(result);
  } catch (error) {
    res.status(500).json({ error: error.message });
  }
};

// API: Abort Upload
exports.abortUpload = async (req, res) => {
  try {
    const token = req.session.user?.token;
    const result = await modelsService.abortUpload(token, req.body.upload_id);
    res.json(result);
  } catch (error) {
    res.status(500).json({ error: error.message });
  }
};

// API: Finalize Upload
exports.finalizeUpload = async (req, res) => {
  try {
    const token = req.session.user?.token;
    const result = await modelsService.finalizeUpload(token, req.body);
    res.json(result);
  } catch (error) {
    res.status(500).json({ error: error.message });
  }
};

// API: Delete model
exports.deleteModel = async (req, res) => {
  try {
    const token = req.session.user?.token;
    const { name } = req.params;

    await modelsService.deleteModel(token, name);
    res.json({ success: true, message: 'Model deleted successfully' });
  } catch (error) {
    res.status(500).json({ success: false, message: error.message });
  }
};