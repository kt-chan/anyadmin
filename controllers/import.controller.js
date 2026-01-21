const mockData = require('../data/mockData');

// 获取导入页面视图
exports.getImportPage = (req, res) => {
  const tasks = mockData.getImportTasks();
  
  // 计算汇总数据
  const stats = {
    totalTasks: tasks.length,
    processing: tasks.filter(t => t.status === 'PROCESSING').length,
    failed: tasks.filter(t => t.status === 'FAILED').length,
    totalFiles: tasks.reduce((acc, t) => acc + t.progress.total, 0),
    processedFiles: tasks.reduce((acc, t) => acc + t.progress.processed, 0)
  };

  res.render('import', {
    page: 'import',
    user: req.session.user,
    tasks: tasks,
    stats: stats
  });
};

// API: 获取任务列表 (用于前端轮询刷新)
exports.getTasks = (req, res) => {
  const tasks = mockData.getImportTasks();
  res.json({ success: true, data: tasks });
};

// API: 创建新任务
exports.createTask = (req, res) => {
  const { name, type, path, schedule } = req.body;
  
  if (!name || !path) {
    return res.status(400).json({ success: false, message: 'Missing required fields' });
  }

  const newTask = {
    id: `task_${Date.now()}`,
    name,
    sourceType: type,
    sourcePath: path,
    status: 'PENDING',
    progress: {
      total: 0,
      processed: 0,
      failed: 0
    },
    schedule: schedule || 'MANUAL',
    lastScan: '-',
    nextScan: '-'
  };
  
  mockData.addImportTask(newTask);
  
  // Simulate task start
  setTimeout(() => {
     mockData.updateImportTask(newTask.id, { 
         status: 'PROCESSING',
         progress: { total: Math.floor(Math.random() * 1000) + 100, processed: 0, failed: 0 }
     });
  }, 2000);

  res.json({ 
    success: true, 
    message: 'Task created successfully',
    task: newTask
  });
};

// API: 任务操作 (Start, Pause, Resume, Retry)
exports.operateTask = (req, res) => {
  const { taskId, action } = req.body;
  // valid actions: 'start', 'pause', 'resume', 'retry', 'delete'
  
  const tasks = mockData.getImportTasks();
  const task = tasks.find(t => t.id === taskId);
  
  if (!task) {
      return res.status(404).json({ success: false, message: 'Task not found' });
  }

  let updates = {};
  
  switch (action) {
      case 'pause':
          updates = { status: 'PAUSED' };
          break;
      case 'resume':
      case 'start':
          updates = { status: 'PROCESSING' };
          break;
      case 'retry':
          updates = { 
              status: 'PROCESSING',
              progress: { ...task.progress, failed: 0 } 
          };
          break;
      case 'delete':
          mockData.deleteImportTask(taskId);
           return res.json({ success: true, message: `Task ${taskId} deleted.` });
      default:
          return res.status(400).json({ success: false, message: 'Invalid action' });
  }

  mockData.updateImportTask(taskId, updates);

  // 模拟操作延迟
  setTimeout(() => {
    res.json({ success: true, message: `Operation ${action} on task ${taskId} executed.` });
  }, 200);
};

// API: 全局操作
exports.operateGlobal = (req, res) => {
  const { action } = req.body;
  // valid actions: 'pause_all', 'resume_all', 'restart_failed'
  
  const tasks = mockData.getImportTasks();
  
  if (action === 'pause_all') {
      tasks.forEach(t => {
          if (t.status === 'PROCESSING') {
              mockData.updateImportTask(t.id, { status: 'PAUSED' });
          }
      });
  } else if (action === 'resume_all') {
      tasks.forEach(t => {
          if (t.status === 'PAUSED') {
              mockData.updateImportTask(t.id, { status: 'PROCESSING' });
          }
      });
  } else if (action === 'restart_failed') {
       tasks.forEach(t => {
          if (t.status === 'FAILED' || t.progress.failed > 0) {
               mockData.updateImportTask(t.id, { 
                   status: 'PROCESSING',
                   progress: { ...t.progress, failed: 0 }
               });
          }
      });
  } else {
       return res.status(400).json({ success: false, message: 'Invalid global action' });
  }
  
  res.json({ success: true, message: `Global operation ${action} executed.` });
};