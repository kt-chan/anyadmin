# AnyAdmin - AI Infrastructure Management Platform

## Abstract
AnyAdmin is a comprehensive management interface designed for orchestrating AI infrastructure. it provides a unified platform for managing compute nodes, automating model deployments (vLLM, MindIE), configuring vector databases (LanceDB, Milvus), and monitoring system health. The platform simplifies complex deployment workflows through an intuitive wizard-based interface, supporting both NVIDIA GPU and Huawei Ascend NPU hardware.

## Key Features
- **Deployment Wizard**: Multi-step configuration for inference engines, vector DBs, and document parsers.
- **Node Management**: Manage target nodes with automated passwordless SSH setup (RSA key generation and distribution).
- **Hardware Detection**: Automatic detection of target hardware (Ascend NPU / NVIDIA GPU).
- **Model Tuning**: Real-time adjustment of inference parameters (Temperature, Top P, GPU Memory).
- **System Monitoring**: Integrated dashboard for metrics and audit logs.

## Project Structure
- `/backend`: Go-based API service (Gin framework) handling orchestration and node management.
- `/frontend`: Node.js/Express application with Pug templates for the management UI.
- `/docs`: Project documentation and requirement specifications.

## Setup Guide

### Prerequisites
- **Backend**: [Go](https://go.dev/dl/) 1.25+
- **Frontend**: [Node.js](https://nodejs.org/) (v16+) and npm
- **System**: Windows (PowerShell) or Linux

### Installation

1. **Clone the repository**:
   ```powershell
   git clone https://github.com/your-repo/anyadmin.git
   cd anyadmin
   ```

2. **Backend Dependencies**:
   ```powershell
   cd backend
   go mod download
   ```

3. **Frontend Dependencies**:
   ```powershell
   cd ../frontend
   npm install
   ```

## Run Instructions

### 1. Start the Backend Service
The backend handles the core logic and node communication.
```powershell
cd backend
go run cmd/server/main.go
```
The API will be available at `http://localhost:8080`.

### 2. Start the Frontend Application
The frontend provides the management UI.
```powershell
cd frontend
npm start
```
Open your browser and navigate to `http://localhost:3000`.

### 3. Setup SSH Access (Crucial for Deployment)
1. Go to the **机器节点和部署管理** (Deployment Management) page.
2. Click **Download System SSH Key** to get the generated `id_rsa.pub`.
3. Distribute this key to your target nodes:
   ```bash
   cat id_rsa.pub >> ~/.ssh/authorized_keys
   ```
4. Enter target node IPs in the UI and click **Verify SSH Connectivity** to ensure the system can manage the nodes.

## Testing
- **Backend Tests**: `cd backend; go test ./internal/service/...`
- **Frontend Tests**: `npx jest tests/frontend`

## Development Guide (VS Code)
This project comes with pre-configured Visual Studio Code settings for a seamless development experience.

### Workspace Setup
1. Open the project folder in VS Code.
2. Ensure you have the recommended extensions installed:
   - **Go** (golang.go)
   - **ESLint** (dbaeumer.vscode-eslint)
   - **Pug** (sissel.pug)

### Debugging
The `.vscode/launch.json` file includes configurations to run the full stack or individual services with a single click.

#### 🚀 Run Full Stack (Recommended)
Select **"Run Full Stack"** from the Run and Debug side panel (Ctrl+Shift+D).
- This launches **both** the Go backend and Node.js frontend.
- It automatically runs a task to wait for the backend to be ready (port 8080) before starting the frontend.
- Breakpoints work in both Go and JavaScript code simultaneously.

#### Individual Services
- **Run Backend (Go)**: Starts only the Go API server in debug mode.
- **Run Frontend**: Starts the Express.js application (requires backend to be running manually or via the compound task).
- **Debug Frontend Tests (Jest)**: Runs and debugs the frontend test suite.
