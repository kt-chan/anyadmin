from typing import List, Dict, Any, Optional
import datetime

# 用户数据
users = [
  { "username": 'admin', "password": 'password', "role": 'admin' },
  { "username": 'operator_01', "password": 'password', "role": 'operator' }
]

# 仪表板数据
def get_dashboard_metrics():
    return {
        "runningServices": {
            "current": 8,
            "total": 8,
            "onlineRate": '100% Online'
        },
        "computeLoad": {
            "percentage": '64.2%',
            "type": 'NPU (Ascend)'
        },
        "memoryUsage": {
            "used": '14.2',
            "total": '32',
            "percentage": '44.3%'
        },
        "taskQueue": {
            "count": 12,
            "status": 'Normal Queue'
        }
    }

def get_dashboard_services():
    return [
        { "id": 'service1', "name": '推理引擎 (llama-3-8b)', "type": 'vLLM / MindIE', "status": 'running' },
        { "id": 'service2', "name": '向量库 (Milvus)', "type": 'Vector DB', "status": 'healthy' },
        { "id": 'service3', "name": '文档解析引擎', "type": 'Doc Parser', "status": 'processing' },
        { "id": 'service4', "name": 'bge-large-zh-v1.5', "type": 'Embedding', "status": 'running' }
    ]

def get_backup_info():
    return {
        "lastBackup": {
            "time": '2024-05-20 02:00:00',
            "type": '增量备份'
        },
        "availablePoints": 12
    }

def get_dashboard_config():
    return {
        "concurrency": 64,
        "tokenOptions": [
            { "value": '4096', "label": '4,096 Tokens', "selected": False },
            { "value": '8192', "label": '8,192 Tokens', "selected": True },
            { "value": '32768', "label": '32,768 Tokens', "selected": False }
        ],
        "dynamicBatching": True,
        "hardwareAcceleration": '昇腾 MindIE'
    }

def get_dashboard_audit_logs():
    return [
        { 
            "user": 'Admin', 
            "action": '修改了推理并发数', 
            "time": '10分钟前', 
            "details": '终端: 192.168.1.102',
            "type": 'user'
        },
        { 
            "user": '系统自检', 
            "action": '全量备份完成', 
            "time": '今天 02:00', 
            "details": '自动化任务',
            "type": 'system'
        }
    ]

# 服务管理页面数据
def get_services_data():
    return [
        { "name": 'llama-3-8b-instruct', "type": 'Inference (MindIE)', "status": 'RUNNING', "endpoint": 'http://10.0.1.5:8000' },
        { "name": 'bge-large-zh-v1.5', "type": 'Embedding', "status": 'RUNNING', "endpoint": 'http://10.0.1.5:8001' },
        { "name": 'milvus-standalone', "type": 'Vector DB', "status": 'STOPPED', "endpoint": 'http://10.0.1.8:19530' }
    ]

# 备份页面数据
def get_backups_data():
    return [
        { "id": 'bk_20240520_full', "time": 'Today, 02:00 AM', "type": 'FULL', "verified": True, "totalSize" : "100GB"},
        { "id": 'bk_20240519_inc', "time": 'Yesterday, 02:00 AM', "type": 'INC', "verified": True, "totalSize" : "10GB" }
    ]

# 系统管理页面数据
def get_system_users_data():
    return [
        { "username": 'admin', "role": 'ADMINISTRATOR', "status": 'Active', "lastLogin": 'Just now' },
        { "username": 'operator_01', "role": 'OPERATOR', "status": 'Active', "lastLogin": '2 days ago' }
    ]

def get_system_audit_logs():
    return [
        { "time": '14:20:05', "action": '用户 admin 登录系统', "details": 'IP: 192.168.1.102 | Method: JWT Auth' },
        { "time": '10:00:00', "action": '系统自动执行全量备份', "details": 'Backup ID: bk_20240520_full | Status: Success' }
    ]

# 文件导入任务数据 (Mutable)
import_tasks = [
  {
    "id": 'task_001',
    "name": '文档全量同步',
    "sourceType": 'NFS',
    "sourcePath": '/mnt/nfs/docs/v1',
    "status": 'PROCESSING', # PENDING, PROCESSING, PAUSED, COMPLETED, FAILED
    "progress": {
      "total": 15000,
      "processed": 8432,
      "failed": 12
    },
    "schedule": 'HOURLY', # REALTIME, HOURLY, DAILY, WEEKLY, MONTHLY, MANUAL
    "lastScan": '2024-05-21 10:00:00',
    "nextScan": '2024-05-21 11:00:00'
  },
  {
    "id": 'task_002',
    "name": '图片资源归档',
    "sourceType": 'S3',
    "sourcePath": 's3://company-assets/images',
    "status": 'PAUSED',
    "progress": {
      "total": 5000,
      "processed": 2100,
      "failed": 0
    },
    "schedule": 'DAILY',
    "lastScan": '2024-05-20 02:00:00',
    "nextScan": '2024-05-22 02:00:00'
  },
  {
    "id": 'task_003',
    "name": '临时数据导入',
    "sourceType": 'LOCAL',
    "sourcePath": '/tmp/upload_buffer',
    "status": 'FAILED',
    "progress": {
      "total": 100,
      "processed": 45,
      "failed": 55
    },
    "schedule": 'MANUAL',
    "lastScan": '2024-05-21 09:30:00',
    "nextScan": '-'
  }
]

def get_import_tasks():
    return import_tasks

def add_import_task(task: Dict[str, Any]):
    import_tasks.append(task)
    return task

def update_import_task(task_id: str, updates: Dict[str, Any]):
    for i, task in enumerate(import_tasks):
        if task["id"] == task_id:
            import_tasks[i].update(updates)
            return True
    return False

def delete_import_task(task_id: str):
    global import_tasks
    import_tasks = [t for t in import_tasks if t["id"] != task_id]
    return True
