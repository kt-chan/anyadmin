from fastapi import FastAPI, HTTPException, Body
from fastapi.middleware.cors import CORSMiddleware
from typing import List, Dict, Any, Optional
try:
    from . import mock_data
except ImportError:
    import mock_data

app = FastAPI()

# Configure CORS
origins = [
    "http://localhost",
    "http://localhost:3000",
    "http://127.0.0.1:3000",
    "http://localhost:8080", # Typical Express ports
    "http://127.0.0.1:8080",
    "*" # Allow all for dev
]

app.add_middleware(
    CORSMiddleware,
    allow_origins=origins,
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

@app.get("/")
def read_root():
    return {"message": "AnyAdmin Backend API"}

# Auth / Users
@app.get("/users")
def get_users():
    return mock_data.users

@app.post("/auth/login")
def login(credentials: Dict[str, str] = Body(...)):
    username = credentials.get("username")
    password = credentials.get("password")
    for user in mock_data.users:
        if user["username"] == username and user["password"] == password:
            return {"username": user["username"], "role": user["role"]}
    raise HTTPException(status_code=401, detail="Invalid credentials")

# Dashboard
@app.get("/dashboard/metrics")
def get_dashboard_metrics():
    return mock_data.get_dashboard_metrics()

@app.get("/dashboard/services")
def get_dashboard_services():
    return mock_data.get_dashboard_services()

@app.get("/dashboard/config")
def get_dashboard_config():
    return mock_data.get_dashboard_config()

@app.get("/dashboard/audit-logs")
def get_dashboard_audit_logs():
    return mock_data.get_dashboard_audit_logs()

@app.get("/backup/info")
def get_backup_info():
    return mock_data.get_backup_info()

# Services
@app.get("/services")
def get_services_data():
    return mock_data.get_services_data()

# Backups
@app.get("/backups")
def get_backups_data():
    return mock_data.get_backups_data()

# System
@app.get("/system/users")
def get_system_users_data():
    return mock_data.get_system_users_data()

@app.get("/system/audit-logs")
def get_system_audit_logs():
    return mock_data.get_system_audit_logs()

# Import Tasks
@app.get("/import/tasks")
def get_import_tasks():
    return mock_data.get_import_tasks()

@app.post("/import/tasks")
def add_import_task(task: Dict[str, Any] = Body(...)):
    mock_data.add_import_task(task)
    return task

@app.put("/import/tasks/{task_id}")
def update_import_task(task_id: str, updates: Dict[str, Any] = Body(...)):
    success = mock_data.update_import_task(task_id, updates)
    if not success:
        raise HTTPException(status_code=404, detail="Task not found")
    return {"status": "success"}

@app.delete("/import/tasks/{task_id}")
def delete_import_task(task_id: str):
    mock_data.delete_import_task(task_id)
    return {"status": "success"}

if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8000)
