# 使用说明

前端使用vite/node，后端使用go

## 前端：
1. frontend目录下的.env配置了后端的IP:PORT，默认是本机的8080端口提供服务
2. 启动前端： cd frontend; npm run dev

## 后端
1. core目录下的config.yaml配置了后端监听的端口
2. 启动后端：
    - cd core; go run cmd/server/main.go
    - 或者: 先把后端编译成可执行程序再执行: cd core; go build -o anyzearch-admin cmd/server/main.go; ./anyzearch-admin

## 登录：
1. 默认端口：http://localhost:5173/dashboard
2. 默认管理员用户名密码：admin/admin

