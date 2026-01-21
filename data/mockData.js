// 用户数据
const users = [
  { username: 'admin', password: 'password', role: 'admin' },
  { username: 'operator_01', password: 'password', role: 'operator' }
];

// 仪表板数据
const getDashboardMetrics = () => ({
  runningServices: {
    current: 8,
    total: 8,
    onlineRate: '100% Online'
  },
  computeLoad: {
    percentage: '64.2%',
    type: 'NPU (Ascend)'
  },
  memoryUsage: {
    used: '14.2',
    total: '32',
    percentage: '44.3%'
  },
  taskQueue: {
    count: 12,
    status: 'Normal Queue'
  }
});

const getDashboardServices = () => [
  { id: 'service1', name: '推理引擎 (llama-3-8b)', type: 'vLLM / MindIE', status: 'running' },
  { id: 'service2', name: '向量库 (Milvus)', type: 'Vector DB', status: 'healthy' },
  { id: 'service3', name: '文档解析引擎', type: 'Doc Parser', status: 'processing' },
  { id: 'service4', name: 'bge-large-zh-v1.5', type: 'Embedding', status: 'running' }
];

const getBackupInfo = () => ({
  lastBackup: {
    time: '2024-05-20 02:00:00',
    type: '增量备份'
  },
  availablePoints: 12
});

const getDashboardConfig = () => ({
  concurrency: 64,
  tokenOptions: [
    { value: '4096', label: '4,096 Tokens', selected: false },
    { value: '8192', label: '8,192 Tokens', selected: true },
    { value: '32768', label: '32,768 Tokens', selected: false }
  ],
  dynamicBatching: true,
  hardwareAcceleration: '昇腾 MindIE'
});

const getDashboardAuditLogs = () => ([
  { 
    user: 'Admin', 
    action: '修改了推理并发数', 
    time: '10分钟前', 
    details: '终端: 192.168.1.102',
    type: 'user'
  },
  { 
    user: '系统自检', 
    action: '全量备份完成', 
    time: '今天 02:00', 
    details: '自动化任务',
    type: 'system'
  }
]);

// 服务管理页面数据
const getServicesData = () => [
  { name: 'llama-3-8b-instruct', type: 'Inference (MindIE)', status: 'RUNNING', endpoint: 'http://10.0.1.5:8000' },
  { name: 'bge-large-zh-v1.5', type: 'Embedding', status: 'RUNNING', endpoint: 'http://10.0.1.5:8001' },
  { name: 'milvus-standalone', type: 'Vector DB', status: 'STOPPED', endpoint: 'http://10.0.1.8:19530' }
];

// 备份页面数据
const getBackupsData = () => [
  { id: 'bk_20240520_full', time: 'Today, 02:00 AM', type: 'FULL', verified: true, totalSize : "100GB"},
  { id: 'bk_20240519_inc', time: 'Yesterday, 02:00 AM', type: 'INC', verified: true, totalSize : "10GB" }
];

// 系统管理页面数据
const getSystemUsersData = () => [
  { username: 'admin', role: 'ADMINISTRATOR', status: 'Active', lastLogin: 'Just now' },
  { username: 'operator_01', role: 'OPERATOR', status: 'Active', lastLogin: '2 days ago' }
];

const getSystemAuditLogs = () => [
  { time: '14:20:05', action: '用户 admin 登录系统', details: 'IP: 192.168.1.102 | Method: JWT Auth' },
  { time: '10:00:00', action: '系统自动执行全量备份', details: 'Backup ID: bk_20240520_full | Status: Success' }
];

// 文件导入任务数据 (Mutable)
let importTasks = [
  {
    id: 'task_001',
    name: '文档全量同步',
    sourceType: 'NFS',
    sourcePath: '/mnt/nfs/docs/v1',
    status: 'PROCESSING', // PENDING, PROCESSING, PAUSED, COMPLETED, FAILED
    progress: {
      total: 15000,
      processed: 8432,
      failed: 12
    },
    schedule: 'HOURLY', // REALTIME, HOURLY, DAILY, WEEKLY, MONTHLY, MANUAL
    lastScan: '2024-05-21 10:00:00',
    nextScan: '2024-05-21 11:00:00'
  },
  {
    id: 'task_002',
    name: '图片资源归档',
    sourceType: 'S3',
    sourcePath: 's3://company-assets/images',
    status: 'PAUSED',
    progress: {
      total: 5000,
      processed: 2100,
      failed: 0
    },
    schedule: 'DAILY',
    lastScan: '2024-05-20 02:00:00',
    nextScan: '2024-05-22 02:00:00'
  },
  {
    id: 'task_003',
    name: '临时数据导入',
    sourceType: 'LOCAL',
    sourcePath: '/tmp/upload_buffer',
    status: 'FAILED',
    progress: {
      total: 100,
      processed: 45,
      failed: 55
    },
    schedule: 'MANUAL',
    lastScan: '2024-05-21 09:30:00',
    nextScan: '-'
  }
];

const getImportTasks = () => importTasks;
const addImportTask = (task) => importTasks.push(task);
const updateImportTask = (taskId, updates) => {
  const taskIndex = importTasks.findIndex(t => t.id === taskId);
  if (taskIndex !== -1) {
    importTasks[taskIndex] = { ...importTasks[taskIndex], ...updates };
    return true;
  }
  return false;
};
const deleteImportTask = (taskId) => {
    importTasks = importTasks.filter(t => t.id !== taskId);
};


module.exports = {
  users,
  getDashboardMetrics,
  getDashboardServices,
  getBackupInfo,
  getDashboardConfig,
  getDashboardAuditLogs,
  getServicesData,
  getBackupsData,
  getSystemUsersData,
  getSystemAuditLogs,
  getImportTasks,
  addImportTask,
  updateImportTask,
  deleteImportTask
};