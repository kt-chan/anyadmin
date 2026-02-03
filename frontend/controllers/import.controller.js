const importService = require('../services/import.service');

// 获取导入页面视图
exports.getImportPage = async (req, res) => {
  try {
    const token = req.session.user?.token;
    const rawTasks = await importService.getTasks(token);
    const tasks = Array.isArray(rawTasks) ? rawTasks : [];
    
    // 计算汇总数据
    const stats = {
      totalTasks: tasks.length,
      processing: tasks.filter(t => (t.status || '').toUpperCase() === 'PROCESSING').length,
      failed: tasks.filter(t => (t.status || '').toUpperCase() === 'FAILED').length,
      totalFiles: tasks.reduce((acc, t) => acc + (t.progress?.total || 0), 0),
      processedFiles: tasks.reduce((acc, t) => acc + (t.progress?.processed || 0), 0)
    };

    res.render('import', {
      page: 'import',
      user: req.session.user,
      tasks: tasks,
      stats: stats
    });
  } catch (error) {
    console.error('Error loading import page:', error);
    res.status(500).render('error', { error });
  }
};

// API: 获取任务列表 (用于前端轮询刷新)
exports.getTasks = async (req, res) => {
  try {
    const token = req.session.user?.token;
    const tasks = await importService.getTasks(token);
    res.json({ success: true, data: tasks });
  } catch (error) {
    res.status(500).json({ success: false, message: error.message });
  }
};

// API: 创建新任务
exports.createTask = async (req, res) => {
  const { name, type, path, schedule } = req.body;
  const token = req.session.user?.token;
  
  if (!name || !path) {
    return res.status(400).json({ success: false, message: 'Missing required fields' });
  }

  const newTask = {
    name,
    sourceType: type,
    sourcePath: path,
  };
  
  try {
    const result = await importService.createTask(token, newTask);
    
    res.json({ 
      success: true, 
      message: 'Task created successfully',
      task: result
    });
  } catch (error) {
    res.status(500).json({ success: false, message: error.message });
  }
};

// API: 任务操作 (Start, Pause, Resume, Retry)
exports.operateTask = async (req, res) => {
  const { taskId, action } = req.body;
  const token = req.session.user?.token;
  // valid actions: 'start', 'pause', 'resume', 'retry', 'delete'
  
  try {
    res.json({ success: true, message: `Operation ${action} on task ${taskId} executed.` });
  } catch (error) {
    res.status(500).json({ success: false, message: error.message });
  }
};

// API: 全局操作
exports.operateGlobal = async (req, res) => {
  const { action } = req.body;
  
  try {
    res.json({ success: true, message: `Global operation ${action} executed.` });
  } catch (error) {
    res.status(500).json({ success: false, message: error.message });
  }
};