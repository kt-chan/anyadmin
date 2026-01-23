from fastapi.testclient import TestClient
from backend.main import app

client = TestClient(app)

def test_read_main():
    response = client.get("/")
    assert response.status_code == 200
    assert response.json() == {"message": "AnyAdmin Backend API"}

def test_get_users():
    response = client.get("/users")
    assert response.status_code == 200
    users = response.json()
    assert isinstance(users, list)
    assert len(users) > 0
    assert "username" in users[0]

def test_login_success():
    response = client.post("/auth/login", json={"username": "admin", "password": "password"})
    assert response.status_code == 200
    data = response.json()
    assert data["username"] == "admin"
    assert data["role"] == "admin"

def test_login_failure():
    response = client.post("/auth/login", json={"username": "admin", "password": "wrongpassword"})
    assert response.status_code == 401

# Dashboard Tests
def test_get_dashboard_metrics():
    response = client.get("/dashboard/metrics")
    assert response.status_code == 200
    data = response.json()
    assert "runningServices" in data
    assert "computeLoad" in data

def test_get_dashboard_services():
    response = client.get("/dashboard/services")
    assert response.status_code == 200
    assert isinstance(response.json(), list)

def test_get_dashboard_config():
    response = client.get("/dashboard/config")
    assert response.status_code == 200
    data = response.json()
    assert "concurrency" in data

def test_get_dashboard_audit_logs():
    response = client.get("/dashboard/audit-logs")
    assert response.status_code == 200
    assert isinstance(response.json(), list)

# Backup Tests
def test_get_backup_info():
    response = client.get("/backup/info")
    assert response.status_code == 200
    data = response.json()
    assert "lastBackup" in data

def test_get_backups_data():
    response = client.get("/backups")
    assert response.status_code == 200
    assert isinstance(response.json(), list)

# Services Tests
def test_get_services_data():
    response = client.get("/services")
    assert response.status_code == 200
    assert isinstance(response.json(), list)

# System Tests
def test_get_system_users_data():
    response = client.get("/system/users")
    assert response.status_code == 200
    assert isinstance(response.json(), list)

def test_get_system_audit_logs():
    response = client.get("/system/audit-logs")
    assert response.status_code == 200
    assert isinstance(response.json(), list)

# Import Tasks Tests
def test_import_tasks_crud():
    # 1. Get initial tasks
    response = client.get("/import/tasks")
    assert response.status_code == 200
    initial_tasks = response.json()
    
    # 2. Add new task
    new_task = {
        "id": "test_task_123",
        "name": "Test Task",
        "sourceType": "LOCAL",
        "sourcePath": "/tmp/test",
        "status": "PENDING",
        "progress": {"total": 0, "processed": 0, "failed": 0},
        "schedule": "MANUAL",
        "lastScan": "-",
        "nextScan": "-"
    }
    response = client.post("/import/tasks", json=new_task)
    assert response.status_code == 200
    assert response.json()["id"] == "test_task_123"

    # 3. Update task
    updates = {"status": "PROCESSING"}
    response = client.put("/import/tasks/test_task_123", json=updates)
    assert response.status_code == 200
    
    # Verify update
    response = client.get("/import/tasks")
    tasks = response.json()
    updated_task = next(t for t in tasks if t["id"] == "test_task_123")
    assert updated_task["status"] == "PROCESSING"

    # 4. Delete task
    response = client.delete("/import/tasks/test_task_123")
    assert response.status_code == 200

    # Verify deletion
    response = client.get("/import/tasks")
    tasks = response.json()
    assert not any(t["id"] == "test_task_123" for t in tasks)

    # 5. Update non-existent task
    response = client.put("/import/tasks/non_existent_id", json={"status": "FAILED"})
    assert response.status_code == 404

# Deployment Tests
def test_deploy_generate():
    config = {
        "mode": "new_deployment",
        "platform": "nvidia",
        "components": {
            "model": "llama-2",
            "vector_db": "milvus"
        }
    }
    response = client.post("/deploy/generate", json=config)
    assert response.status_code == 200
    data = response.json()
    assert data["status"] == "success"
    assert "artifacts" in data
    assert "deploy_ubuntu.sh" in data["artifacts"]

def test_deploy_generate_multinode():
    config = {
        "mode": "new_deployment",
        "platform": "nvidia",
        "target_nodes": "10.0.0.1\n10.0.0.2",
        "inference_host": "node-1",
        "components": {}
    }
    response = client.post("/deploy/generate", json=config)
    assert response.status_code == 200
    data = response.json()
    artifacts = data["artifacts"]
    
    assert "hosts.ini" in artifacts
    assert "10.0.0.1" in artifacts["hosts.ini"]
    
    assert "nodeSelector" in artifacts["k8s_deployment.yaml"]
    assert "node-1" in artifacts["k8s_deployment.yaml"]
    
    assert "Multi-node setup detected" in artifacts["deploy_ubuntu.sh"]

def test_deploy_nodes():
    nodes = ["192.168.1.10", "192.168.1.11"]
    
    # Save
    response = client.post("/deploy/nodes", json={"nodes": nodes})
    assert response.status_code == 200
    assert response.json()["status"] == "success"
    
    # Get
    response = client.get("/deploy/nodes")
    assert response.status_code == 200
    assert response.json()["nodes"] == nodes

def test_deploy_test_connection():
    # Success case - Generic
    response = client.post("/deploy/test-connection", json={"host": "127.0.0.1", "port": 8000, "type": "inference"})
    assert response.status_code == 200
    assert response.json()["status"] == "success"

    # Success case - Parser
    response = client.post("/deploy/test-connection", json={"host": "127.0.0.1", "port": 8888, "type": "parser"})
    assert response.status_code == 200
    assert "Mineru" in response.json()["message"]

    # Fail case
    response = client.post("/deploy/test-connection", json={"host": "fail-host", "port": 8000, "type": "inference"})
    assert response.status_code == 200
    assert response.json()["status"] == "error"

def test_deploy_test_connection_ssh():
    # Success case
    response = client.post("/deploy/test-connection", json={"host": "192.168.1.1\n192.168.1.2", "port": 22, "type": "ssh"})
    assert response.status_code == 200
    json_resp = response.json()
    assert json_resp["status"] == "success"
    assert "192.168.1.1" in json_resp["message"]

    # Fail case
    response = client.post("/deploy/test-connection", json={"host": "192.168.1.1\nfail-node", "port": 22, "type": "ssh"})
    assert response.status_code == 200
    json_resp = response.json()
    assert json_resp["status"] == "error"
    assert "fail-node" in json_resp["message"]


def test_get_models():
    response = client.get("/models")
    assert response.status_code == 200
    models = response.json()
    assert isinstance(models, list)
    assert len(models) > 0
    assert "name" in models[0]

def test_save_model_config():
    config = {
        "name": "NewModel",
        "path": "/models/new",
        "platform": "ascend",
        "params": {"max_tokens": 1024}
    }
    response = client.post("/models", json=config)
    assert response.status_code == 200
    assert response.json()["status"] == "success"