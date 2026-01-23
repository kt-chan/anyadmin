from typing import List, Dict, Any, Optional
import datetime
import json
import os

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

# --- Deployment & Models Mock Data ---

models_data = [
    {
        "id": "model_1",
        "name": "Llama-2-7b-chat",
        "platform": "nvidia",
        "status": "loaded",
        "params": {"max_tokens": 4096, "temperature": 0.7}
    },
    {
        "id": "model_2",
        "name": "ChatGLM3-6B",
        "platform": "ascend",
        "status": "stopped",
        "params": {"max_tokens": 8192, "temperature": 0.5}
    }
]

def get_models():
    return models_data

def save_model_config(config: Dict[str, Any]):
    # Mock save - in memory only
    # Check if exists to update, else append
    for i, m in enumerate(models_data):
        if m["name"] == config.get("name"):
            models_data[i].update(config)
            return {"status": "success", "message": f"Updated {config.get('name')}"}
    
    # New mock model
    new_model = {
        "id": f"model_{len(models_data) + 1}",
        "status": "stopped",
        **config
    }
    models_data.append(new_model)
    return {"status": "success", "message": f"Saved {config.get('name')}"}

# --- Node Management (File Persistence) ---
NODES_FILE = "mock_deployment_nodes.json"

def save_target_nodes(nodes: List[str]):
    try:
        with open(NODES_FILE, "w") as f:
            json.dump({"nodes": nodes}, f)
        return {"status": "success", "message": f"Saved {len(nodes)} nodes."}
    except Exception as e:
        return {"status": "error", "message": str(e)}

def get_target_nodes():
    if not os.path.exists(NODES_FILE):
        return []
    try:
        with open(NODES_FILE, "r") as f:
            data = json.load(f)
            return data.get("nodes", [])
    except:
        return []

def generate_deployment_artifacts(config: Dict[str, Any]):
    mode = config.get("mode", "unknown")
    platform = config.get("platform", "unknown")
    target_nodes_str = (config.get("target_nodes") or "").strip()
    inference_host = (config.get("inference_host") or "").strip()
    
    artifacts = {}
    
    if mode == "new_deployment":
        # 0. Hosts Inventory (if applicable)
        if target_nodes_str:
            nodes = [n.strip() for n in target_nodes_str.split('\n') if n.strip()]
            if nodes:
                hosts_ini = "[ai_nodes]\n" + "\n".join(nodes) + "\n\n[master]\nlocalhost ansible_connection=local"
                artifacts["hosts.ini"] = hosts_ini

        # 1. Bash Script for Ubuntu 22.04
        bash_script = """#!/bin/bash
# Auto-generated deployment script for AnyAdmin (One-Click Wizard)
# Platform: {platform}
# Target OS: Ubuntu 22.04 LTS

set -e

echo ">>> Starting system preparation for {platform}..."
"""
        if target_nodes_str:
             bash_script += "echo '>>> Multi-node setup detected. See hosts.ini for inventory.'\n"
             
        bash_script += "apt-get update && apt-get install -y docker.io docker-compose\n"

        if platform == "nvidia":
            bash_script += """
echo ">>> Installing NVIDIA Container Toolkit..."
distribution=$(. /etc/os-release;echo $ID$VERSION_ID)
curl -s -L https://nvidia.github.io/libnvidia-container/gpgkey | sudo apt-key add -
curl -s -L https://nvidia.github.io/libnvidia-container/$distribution/libnvidia-container.list | sudo tee /etc/apt/sources.list.d/libnvidia-container.list
apt-get update && apt-get install -y nvidia-container-toolkit
sudo systemctl restart docker
"""
        elif platform == "ascend":
             bash_script += """
echo ">>> Installing Ascend Driver & Firmware..."
# Placeholder for Ascend driver installation
# ./Ascend-hdk-910b-npu-driver_23.0.rc3_linux-aarch64.run --full
"""

        bash_script += """
echo ">>> Pulling Docker images..."
docker pull vectordb/lancedb:latest
docker pull mineru/parser:latest
"""
        if platform == "nvidia":
            bash_script += "docker pull vllm/vllm-openai:latest\n"
        elif platform == "ascend":
            bash_script += "docker pull mindspore/mindie:1.0.0\n"

        bash_script += "\necho 'Deployment preparation complete. Run docker-compose up -d to start.'"
        artifacts["deploy_ubuntu.sh"] = bash_script

        # 2. Kubernetes YAML
        k8s_yaml = f"""apiVersion: v1
kind: Namespace
metadata:
  name: anyadmin-ai
---
# Vector Database (LanceDB)
apiVersion: apps/v1
kind: Deployment
metadata:
  name: lancedb
  namespace: anyadmin-ai
spec:
  replicas: 1
  selector:
    matchLabels:
      app: lancedb
  template:
    metadata:
      labels:
        app: lancedb
    spec:
      containers:
      - name: lancedb
        image: vectordb/lancedb:latest
        ports:
        - containerPort: 8080
---
# Document Parser (Mineru)
apiVersion: apps/v1
kind: Deployment
metadata:
  name: mineru-parser
  namespace: anyadmin-ai
spec:
  replicas: 1
  selector:
    matchLabels:
      app: mineru
  template:
    metadata:
      labels:
        app: mineru
    spec:
      containers:
      - name: parser
        image: mineru/parser:latest
        ports:
        - containerPort: 8888
---
# Inference Engine ({platform})
apiVersion: apps/v1
kind: Deployment
metadata:
  name: inference-engine
  namespace: anyadmin-ai
spec:
  replicas: 1
  selector:
    matchLabels:
      app: inference
  template:
    metadata:
      labels:
        app: inference
    spec:
"""
        if inference_host and inference_host != "localhost":
             k8s_yaml += f"      nodeSelector:\n        kubernetes.io/hostname: {inference_host}\n"
        
        k8s_yaml += "      containers:\n      - name: engine\n"

        if platform == "nvidia":
            k8s_yaml += """
        image: vllm/vllm-openai:latest
        resources:
          limits:
            nvidia.com/gpu: 1
"""
        elif platform == "ascend":
            k8s_yaml += """
        image: mindspore/mindie:1.0.0
        resources:
          limits:
            huawei.com/Ascend910: 1
"""
        artifacts["k8s_deployment.yaml"] = k8s_yaml

    else: # Integrate Existing
        install_script = f"""#!/bin/bash
# Integration Agent Installer
# Connects this machine to AnyAdmin Control Plane

echo ">>> Verifying connectivity..."
curl -v {config.get('mgmt_host', 'localhost')}:{config.get('mgmt_port', 3000)}

echo ">>> Installing Agent..."
# pip install anyadmin-agent
echo "Agent installed and connected."
"""
        artifacts["install_agent.sh"] = install_script

    return {
        "status": "success",
        "artifacts": artifacts
    }

def test_service_connection(service: Dict[str, Any]):
    # Mock connection test logic
    service_type = service.get("type", "unknown")
    host = service.get("host", "localhost")
    port = service.get("port", 80)
    
    # Simulate failures for specific 'fail' hosts
    if "fail" in str(host):
         return {
            "status": "error", 
            "message": f"Connection refused to {service_type} at {host}:{port}"
        }
    
    # Specific mock delays or logic per type could go here
    if service_type == "vectordb":
        return {
            "status": "success",
            "message": f"Connected to Vector DB ({host}:{port}). Collections: 4"
        }
    elif service_type == "parser":
        return {
            "status": "success",
            "message": f"Connected to Mineru Parser ({host}:{port}). Workers: 2"
        }
    elif service_type == "ssh":
        nodes = [n.strip() for n in str(host).split('\n') if n.strip()]
        failed_nodes = [n for n in nodes if "fail" in n]
        if failed_nodes:
             return {
                "status": "error",
                "message": f"SSH Connection failed for nodes: {', '.join(failed_nodes)}. Check keys."
             }
        return {
            "status": "success",
            "message": f"SSH Connection successful for {len(nodes)} nodes: {', '.join(nodes)}"
        }

    return {
        "status": "success", 
        "message": f"Successfully connected to {service_type} at {host}:{port}. Latency: 12ms"
    }

def detect_system_hardware(nodes: List[str]):
    # Mock detection logic
    # Real world: SSH to nodes, run `nvidia-smi` or `npu-smi`
    detected_platform = "nvidia"
    details = "Detected NVIDIA GPU (Tesla T4) via nvidia-smi"
    
    for node in nodes:
        # Mock: if any node string contains 'ascend', we simulate Ascend detection
        if "ascend" in node.lower():
            detected_platform = "ascend"
            details = "Detected Huawei Ascend 910 via npu-smi"
            break
            
    return {
        "status": "success",
        "platform": detected_platform,
        "details": details
    }
