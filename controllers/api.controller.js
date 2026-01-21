const dashboardService = require('../services/dashboard.service');
const servicesService = require('../services/services.service');
const backupService = require('../services/backup.service');
const deploymentService = require('../services/deployment.service');
const systemService = require('../services/system.service');
const response = require('../utils/response');

const apiController = {
    // 保存配置
    saveConfig: async (req, res) => {
        try {
            await dashboardService.saveConfig(req.body);
            return response.success(res, { timestamp: new Date().toISOString() }, '配置已保存');
        } catch (err) {
            return response.error(res, '保存配置失败', 500, err);
        }
    },

    // 更新并发数
    updateConcurrency: async (req, res) => {
        try {
            const { value } = req.body;
            const newValue = await dashboardService.updateConcurrency(value);
            return response.success(res, { newValue }, `并发数已更新为 ${value}`);
        } catch (err) {
            return response.error(res, '更新并发数失败', 500, err);
        }
    },

    // 更新Token限制
    updateTokenLimit: async (req, res) => {
        try {
            const { value } = req.body;
            const newValue = await dashboardService.updateTokenLimit(value);
            return response.success(res, { newValue }, `Token限制已更新为 ${value}`);
        } catch (err) {
            return response.error(res, '更新Token限制失败', 500, err);
        }
    },

    // 重启服务
    restartService: async (req, res) => {
        try {
            const { serviceId } = req.body;
            await servicesService.restartService(serviceId);
            return response.success(res, { restartTime: new Date().toLocaleTimeString() }, `服务 ${serviceId} 重启已触发`);
        } catch (err) {
            return response.error(res, '重启服务失败', 500, err);
        }
    },

    // 停止服务
    stopService: async (req, res) => {
        try {
            const { serviceId } = req.body;
            await servicesService.stopService(serviceId);
            return response.success(res, { stopTime: new Date().toLocaleTimeString() }, `服务 ${serviceId} 已停止`);
        } catch (err) {
            return response.error(res, '停止服务失败', 500, err);
        }
    },

    // 获取服务状态
    getServicesStatus: async (req, res) => {
        try {
            const services = await servicesService.getServicesStatus();
            const formattedServices = services.map(service => ({
                id: service.id,
                name: service.name,
                status: service.status,
                lastCheck: new Date().toLocaleTimeString()
            }));
            return response.success(res, { services: formattedServices });
        } catch (err) {
            return response.error(res, '获取服务状态失败', 500, err);
        }
    },

    // 创建备份
    createBackup: async (req, res) => {
        try {
            const { backupType } = req.body;
            const result = await backupService.createBackup(backupType);
            return response.success(res, result, `${backupType}备份任务已开始`);
        } catch (err) {
            return response.error(res, '创建备份失败', 500, err);
        }
    },

    // 从备份恢复
    restoreFromBackup: async (req, res) => {
        try {
            const { backupId } = req.body;
            const result = await backupService.restoreFromBackup(backupId);
            return response.success(res, result, `正在从备份 ${backupId} 恢复系统`);
        } catch (err) {
            return response.error(res, '恢复备份失败', 500, err);
        }
    },

    // 部署下一步
    nextDeploymentStep: async (req, res) => {
        try {
            const { currentStep, data } = req.body;
            const nextStep = await deploymentService.nextDeploymentStep(currentStep, data);
            return response.success(res, { nextStep }, '已进入下一步');
        } catch (err) {
            return response.error(res, '部署步骤处理失败', 500, err);
        }
    },

    // 创建用户
    createUser: async (req, res) => {
        try {
            const { username, role } = req.body;
            const result = await systemService.createUser(username, role);
            return response.success(res, result, `用户 ${username} 创建成功`);
        } catch (err) {
            return response.error(res, '创建用户失败', 500, err);
        }
    },

    // 获取系统指标
    getSystemMetrics: async (req, res) => {
        try {
            const metrics = await dashboardService.getMetrics();
            return response.success(res, { metrics, timestamp: new Date().toISOString() });
        } catch (err) {
            return response.error(res, '获取系统指标失败', 500, err);
        }
    },

    // 删除备份
    deleteBackup: async (req, res) => {
        try {
            const { backupId } = req.body;
            await backupService.deleteBackup(backupId);
            return response.success(res, {}, `备份 ${backupId} 已删除`);
        } catch (err) {
            return response.error(res, '删除备份失败', 500, err);
        }
    },

    // 应用重刷
    appReflash: async (req, res) => {
        try {
            const result = await systemService.appReflash();
            return response.success(res, result, '应用重刷已启动');
        } catch (err) {
            return response.error(res, '应用重刷失败', 500, err);
        }
    },

    // 调试信息
    debugInfo: async (req, res) => {
        const appConfig = require('../config/app.config');
        const debugData = {
            appName: appConfig.appName,
            version: appConfig.version,
            env: appConfig.env,
            uptime: process.uptime(),
            memory: process.memoryUsage(),
            timestamp: new Date().toISOString()
        };
        return response.success(res, debugData, '调试信息获取成功');
    }
};

module.exports = apiController;