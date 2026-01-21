const importService = require('../services/import.service');

// 获取导入页面视图
exports.getImportPage = async (req, res) => {
  try {
    const tasks = await importService.getTasks();
    
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
  } catch (error) {
    console.error('Error loading import page:', error);
    res.status(500).render('error', { error });
  }
};

// API: 获取任务列表 (用于前端轮询刷新)
exports.getTasks = async (req, res) => {
  try {
    const tasks = await importService.getTasks();
    res.json({ success: true, data: tasks });
  } catch (error) {
    res.status(500).json({ success: false, message: error.message });
  }
};

// API: 创建新任务
exports.createTask = async (req, res) => {
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
  
  try {
    await importService.createTask(newTask);
    
    // Simulate task start
    setTimeout(async () => {
       try {
         await importService.updateTask(newTask.id, { 
             status: 'PROCESSING',
             progress: { total: Math.floor(Math.random() * 1000) + 100, processed: 0, failed: 0 }
         });
       } catch (err) {
         console.error('Error updating simulated task:', err);
       }
    }, 2000);

    res.json({ 
      success: true, 
      message: 'Task created successfully',
      task: newTask
    });
  } catch (error) {
    res.status(500).json({ success: false, message: error.message });
  }
};

// API: 任务操作 (Start, Pause, Resume, Retry)
exports.operateTask = async (req, res) => {
  const { taskId, action } = req.body;
  // valid actions: 'start', 'pause', 'resume', 'retry', 'delete'
  
  try {
    const tasks = await importService.getTasks();
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
            await importService.deleteTask(taskId);
            return res.json({ success: true, message: `Task ${taskId} deleted.` });
        default:
            return res.status(400).json({ success: false, message: 'Invalid action' });
    }

    await importService.updateTask(taskId, updates);

    // 模拟操作延迟
    setTimeout(() => {
      res.json({ success: true, message: `Operation ${action} on task ${taskId} executed.` });
    }, 200);
  } catch (error) {
    res.status(500).json({ success: false, message: error.message });
  }
};

// API: 全局操作
exports.operateGlobal = async (req, res) => {
  const { action } = req.body;
  // valid actions: 'pause_all', 'resume_all', 'restart_failed'
  
  try {
    const tasks = await importService.getTasks();
    
    const promises = [];
    if (action === 'pause_all') {
        tasks.forEach(t => {
            if (t.status === 'PROCESSING') {
                promises.push(importService.updateTask(t.id, { status: 'PAUSED' }));
            }
        });
    } else if (action === 'resume_all') {
        tasks.forEach(t => {
            if (t.status === 'PAUSED') {
                promises.push(importService.updateTask(t.id, { status: 'PROCESSING' }));
            }
        });
    } else if (action === 'restart_failed') {
         tasks.forEach(t => {
            if (t.status === 'FAILED' || t.progress.failed > 0) {
                 promises.push(importService.updateTask(t.id, { 
                     status: 'PROCESSING',
                     progress: { ...t.progress, failed: 0 }
                 }));
            }
        });
    } else {
         return res.status(400).json({ success: false, message: 'Invalid global action' });
    }
    
    await Promise.all(promises);
    res.json({ success: true, message: `Global operation ${action} executed.` });
  } catch (error) {
    res.status(500).json({ success: false, message: error.message });
  }
};