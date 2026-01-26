# AnyAdmin - AI 基础设施管理平台

[English](README.md) | [中文](README_zh-CN.md)

## 简介
AnyAdmin 是一个专为编排 AI 基础设施而设计的综合管理界面。它提供了一个统一的平台，用于管理计算节点、自动化模型部署（vLLM, MindIE）、配置向量数据库（LanceDB, Milvus）以及监控系统健康状况。该平台支持 NVIDIA GPU 和华为昇腾 NPU 硬件，通过直观的向导式界面简化了复杂的部署流程。

## 核心功能
- **部署向导**：针对推理引擎、向量数据库和文档解析器的多步骤配置。
- **节点管理**：通过自动化的免密 SSH 设置（RSA 密钥生成与分发）管理目标节点。
- **硬件检测**：自动检测目标硬件（昇腾 NPU / NVIDIA GPU）。
- **模型调优**：实时调整推理参数（Temperature, Top P, GPU 显存）。
- **系统监控**：集成仪表盘，用于展示指标和审计日志。

## 项目结构
- `/backend`：基于 Go (Gin 框架) 的 API 服务，处理编排和节点管理。
- `/frontend`：基于 Node.js/Express 的应用程序，使用 Pug 模板作为管理 UI。
- `/docs`：项目文档和需求说明书。

## 设置指南

### 前置要求
- **后端**：[Go](https://go.dev/dl/) 1.25+
- **前端**：[Node.js](https://nodejs.org/) (v16+) 和 npm
- **系统**：Windows (PowerShell) 或 Linux

### 安装

1. **克隆仓库**：
   ```powershell
   git clone https://github.com/your-repo/anyadmin.git
   cd anyadmin
   ```

2. **后端依赖**：
   ```powershell
   cd backend
   go mod download
   ```

3. **前端依赖**：
   ```powershell
   cd ../frontend
   npm install
   ```

## 运行说明

### 1. 启动后端服务
后端处理核心逻辑和节点通信。
```powershell
cd backend
go run cmd/server/main.go
```
API 将在 `http://localhost:8080` 可用。

### 2. 启动前端应用
前端提供管理用户界面。
```powershell
cd frontend
npm start
```
打开浏览器并访问 `http://localhost:3000`。

### 3. 设置 SSH 访问（部署的关键步骤）
1. 转到 **机器节点和部署管理** 页面。
2. 点击 **Download System SSH Key** 下载生成的 `id_rsa.pub`。
3. 将此密钥分发到你的目标节点：
   ```bash
   cat id_rsa.pub >> ~/.ssh/authorized_keys
   ```
4. 在 UI 中输入目标节点 IP，然后点击 **Verify SSH Connectivity** 以确保系统可以管理这些节点。

## 测试
- **后端测试**：`cd backend; go test ./internal/service/...`
- **前端测试**：`npx jest tests/frontend`

## 开发指南 (VS Code)
本项目配备了预配置的 Visual Studio Code 设置，以提供无缝的开发体验。

### 工作区设置
1. 在 VS Code 中打开项目文件夹。
2. 确保已安装推荐的扩展：
   - **Go** (golang.go)
   - **ESLint** (dbaeumer.vscode-eslint)
   - **Pug** (sissel.pug)

### 调试
`.vscode/launch.json` 文件包含了单击即可运行全栈或单个服务的配置。

#### 🚀 运行全栈（推荐）
从“运行和调试”侧边栏 (Ctrl+Shift+D) 选择 **"Run Full Stack"**。
- 这将启动 **Go 后端** 和 **Node.js 前端**。
- 它会自动运行一个任务，在启动前端之前等待后端准备就绪（端口 8080）。
- 断点可同时在 Go 和 JavaScript 代码中工作。

#### 单个服务
- **Run Backend (Go)**：仅以调试模式启动 Go API 服务器。
- **Run Frontend**：启动 Express.js 应用程序（需要手动运行后端或通过复合任务运行）。
- **Debug Frontend Tests (Jest)**：运行并调试前端测试套件。
