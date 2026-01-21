/**
 * 知识库管理平台 - 主JavaScript文件
 * 包含全局功能、页面切换和交互逻辑
 */

// 等待DOM加载完成
document.addEventListener('DOMContentLoaded', function() {
    console.log('知识库管理平台已初始化');
    
    // 初始化页面
    initPage();
    
    // 绑定全局事件
    bindGlobalEvents();
    
    // 检查登录状态
    checkAuthStatus();
});

/**
 * 初始化页面
 */
function initPage() {
    // 设置当前活跃页面
    setActivePage();
    
    // 初始化模态框
    initModals();
    
    // 初始化表单
    initForms();
    
    // 初始化工具提示（如果有的话）
    initTooltips();
}

/**
 * 绑定全局事件
 */
function bindGlobalEvents() {
    // 绑定模态框关闭事件
    document.addEventListener('click', function(event) {
        if (event.target.classList.contains('modal')) {
            event.target.classList.remove('open');
        }
    });
    
    // 绑定ESC键关闭模态框
    document.addEventListener('keydown', function(event) {
        if (event.key === 'Escape') {
            const openModals = document.querySelectorAll('.modal.open');
            openModals.forEach(modal => {
                modal.classList.remove('open');
            });
        }
    });
    
    // 绑定表单提交事件
    const forms = document.querySelectorAll('form:not(#login-form)');
    forms.forEach(form => {
        form.addEventListener('submit', handleFormSubmit);
    });
}

/**
 * 设置活跃页面
 */
function setActivePage() {
    // 从URL获取当前页面
    const path = window.location.pathname;
    const pageMap = {
        '/dashboard': 'dashboard',
        '/deployment': 'deployment',
        '/services': 'services',
        '/backup': 'backup',
        '/system': 'system',
        '/import': 'import'
    };
    
    const currentPage = pageMap[path] || 'dashboard';
    
    // 更新导航菜单
    document.querySelectorAll('.nav-item').forEach(item => {
        item.classList.remove('nav-active');
        if (item.getAttribute('data-page') === currentPage) {
            item.classList.add('nav-active');
        }
    });
    
    // 显示当前页面内容
    document.querySelectorAll('.page-content').forEach(page => {
        page.classList.remove('active');
    });
    
    const currentPageElement = document.getElementById(`page-${currentPage}`);
    if (currentPageElement) {
        currentPageElement.classList.add('active');
    }
}

/**
 * 初始化模态框
 */
function initModals() {
    // 模态框打开按钮
    const modalTriggers = document.querySelectorAll('[data-modal-target]');
    modalTriggers.forEach(trigger => {
        trigger.addEventListener('click', function() {
            const modalId = this.getAttribute('data-modal-target');
            showModal(modalId);
        });
    });
    
    // 模态框关闭按钮
    const modalClosers = document.querySelectorAll('[data-modal-close]');
    modalClosers.forEach(closer => {
        closer.addEventListener('click', function() {
            const modalId = this.getAttribute('data-modal-close');
            hideModal(modalId);
        });
    });
}

/**
 * 初始化表单
 */
function initForms() {
    // 范围输入实时显示
    const rangeInputs = document.querySelectorAll('input[type="range"]');
    rangeInputs.forEach(input => {
        const display = input.nextElementSibling;
        if (display && display.classList.contains('range-value')) {
            input.addEventListener('input', function() {
                display.textContent = this.value;
            });
        }
    });
    
    // 表单验证
    const validateInputs = document.querySelectorAll('input[required], select[required]');
    validateInputs.forEach(input => {
        input.addEventListener('blur', validateField);
    });
}

/**
 * 初始化工具提示
 */
function initTooltips() {
    // 如果有工具提示元素，可以在这里初始化
    // 例如使用tippy.js或自定义工具提示
}

/**
 * 显示模态框
 * @param {string} modalId - 模态框ID
 */
function showModal(modalId) {
    const modal = document.getElementById(modalId);
    if (modal) {
        modal.classList.add('open');
        document.body.style.overflow = 'hidden'; // 防止背景滚动
    }
}

/**
 * 隐藏模态框
 * @param {string} modalId - 模态框ID
 */
function hideModal(modalId) {
    const modal = document.getElementById(modalId);
    if (modal) {
        modal.classList.remove('open');
        document.body.style.overflow = ''; // 恢复滚动
    }
}

/**
 * 表单字段验证
 * @param {Event} event - 事件对象
 */
function validateField(event) {
    const input = event.target;
    const isValid = input.checkValidity();
    
    if (!isValid) {
        input.classList.add('border-red-500');
        // 显示错误消息
        const errorId = `${input.id}-error`;
        let errorElement = document.getElementById(errorId);
        if (!errorElement) {
            errorElement = document.createElement('div');
            errorElement.id = errorId;
            errorElement.className = 'text-red-500 text-xs mt-1';
            input.parentNode.appendChild(errorElement);
        }
        errorElement.textContent = input.validationMessage;
    } else {
        input.classList.remove('border-red-500');
        const errorId = `${input.id}-error`;
        const errorElement = document.getElementById(errorId);
        if (errorElement) {
            errorElement.remove();
        }
    }
}

/**
 * 处理表单提交
 * @param {Event} event - 事件对象
 */
async function handleFormSubmit(event) {
    event.preventDefault();
    
    const form = event.target;
    const formId = form.id || 'form';
    const endpoint = form.getAttribute('data-endpoint');
    
    if (!endpoint) {
        console.warn('表单未指定提交端点');
        return;
    }
    
    // 收集表单数据
    const formData = new FormData(form);
    const data = Object.fromEntries(formData.entries());
    
    // 显示加载状态
    const submitBtn = form.querySelector('button[type="submit"]');
    const originalText = submitBtn.textContent;
    submitBtn.disabled = true;
    submitBtn.innerHTML = '<i class="fas fa-spinner fa-spin mr-2"></i>处理中...';
    
    try {
        const response = await fetch(endpoint, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'X-Requested-With': 'XMLHttpRequest'
            },
            body: JSON.stringify(data)
        });
        
        const result = await response.json();
        
        if (result.success) {
            // 成功处理
            showNotification(result.message || '操作成功', 'success');
            
            // 如果有重定向
            if (result.redirect) {
                setTimeout(() => {
                    window.location.href = result.redirect;
                }, 1500);
            }
            
            // 如果有回调函数
            if (result.callback) {
                window[result.callback](result.data);
            }
        } else {
            // 错误处理
            showNotification(result.message || '操作失败', 'error');
        }
    } catch (error) {
        console.error('表单提交错误:', error);
        showNotification('网络错误，请重试', 'error');
    } finally {
        // 恢复按钮状态
        submitBtn.disabled = false;
        submitBtn.textContent = originalText;
    }
}

/**
 * 显示通知
 * @param {string} message - 消息内容
 * @param {string} type - 消息类型 (success, error, warning, info)
 */
function showNotification(message, type = 'info') {
    // 创建通知元素
    const notification = document.createElement('div');
    notification.className = `fixed top-4 right-4 z-50 px-6 py-4 rounded-lg shadow-lg transform transition-all duration-300 translate-x-full`;
    
    // 根据类型设置样式
    const typeClasses = {
        success: 'bg-green-50 text-green-800 border-l-4 border-green-500',
        error: 'bg-red-50 text-red-800 border-l-4 border-red-500',
        warning: 'bg-yellow-50 text-yellow-800 border-l-4 border-yellow-500',
        info: 'bg-blue-50 text-blue-800 border-l-4 border-blue-500'
    };
    
    notification.classList.add(...typeClasses[type].split(' '));
    
    // 添加内容和图标
    const icons = {
        success: 'check-circle',
        error: 'exclamation-circle',
        warning: 'exclamation-triangle',
        info: 'info-circle'
    };
    
    notification.innerHTML = `
        <div class="flex items-center">
            <i class="fas fa-${icons[type]} mr-3"></i>
            <span>${message}</span>
            <button class="ml-4 text-gray-400 hover:text-gray-600" onclick="this.parentElement.parentElement.remove()">
                <i class="fas fa-times"></i>
            </button>
        </div>
    `;
    
    // 添加到页面
    document.body.appendChild(notification);
    
    // 显示动画
    setTimeout(() => {
        notification.classList.remove('translate-x-full');
        notification.classList.add('translate-x-0');
    }, 10);
    
    // 自动移除
    setTimeout(() => {
        notification.classList.remove('translate-x-0');
        notification.classList.add('translate-x-full');
        setTimeout(() => {
            if (notification.parentNode) {
                notification.parentNode.removeChild(notification);
            }
        }, 300);
    }, 5000);
}

/**
 * 检查认证状态
 */
function checkAuthStatus() {
    // 如果有登录页面，检查用户是否已登录
    const loginScreen = document.getElementById('login-screen');
    const appContainer = document.getElementById('app-container');
    
    if (loginScreen && appContainer) {
        // 模拟检查本地存储中的登录状态
        const isLoggedIn = localStorage.getItem('kb_user_logged_in');
        
        if (isLoggedIn === 'true') {
            loginScreen.style.display = 'none';
            appContainer.style.opacity = '1';
        }
    }
}

/**
 * 处理登录
 */
async function handleLogin() {
    const form = document.getElementById('login-form');
    if (!form) return;
    
    const formData = new FormData(form);
    const data = Object.fromEntries(formData.entries());
    
    // 显示加载状态
    const submitBtn = form.querySelector('button[type="button"]');
    const originalText = submitBtn.textContent;
    submitBtn.disabled = true;
    submitBtn.innerHTML = '<i class="fas fa-spinner fa-spin mr-2"></i>登录中...';
    
    try {
        // 模拟登录请求
        const response = await fetch('/login', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(data)
        });
        
        const result = await response.json();
        
        if (result.success) {
            // 保存登录状态到本地存储
            localStorage.setItem('kb_user_logged_in', 'true');
            
            // 显示成功消息
            showNotification('登录成功', 'success');
            
            // 跳转到仪表板
            setTimeout(() => {
                window.location.href = result.redirect || '/dashboard';
            }, 1000);
        } else {
            showNotification(result.message || '登录失败', 'error');
        }
    } catch (error) {
        console.error('登录错误:', error);
        showNotification('网络错误，请重试', 'error');
    } finally {
        // 恢复按钮状态
        submitBtn.disabled = false;
        submitBtn.textContent = originalText;
    }
}

/**
 * 处理注销
 */
function handleLogout() {
    if (confirm('确定要退出登录吗？')) {
        // 清除本地存储
        localStorage.removeItem('kb_user_logged_in');
        
        // 跳转到登录页面
        window.location.href = '/logout';
    }
}

/**
 * 导航到页面
 * @param {string} pageId - 页面ID
 */
function navTo(pageId) {
    // 更新URL
    history.pushState({}, '', '/' + pageId);
    
    // 更新导航菜单
    document.querySelectorAll('.nav-item').forEach(el => {
        el.classList.remove('nav-active');
    });
    const navItem = document.querySelector(`[data-page="${pageId}"]`);
    if (navItem) {
        navItem.classList.add('nav-active');
    }
    
    // 更新页面内容
    document.querySelectorAll('.page-content').forEach(el => {
        el.classList.remove('active');
    });
    
    setTimeout(() => {
        const target = document.getElementById('page-' + pageId);
        if (target) {
            target.classList.add('active');
        }
    }, 100);
}

/**
 * 模拟服务状态检查
 */
function checkServiceStatus() {
    // 模拟服务状态更新
    const statusElements = document.querySelectorAll('.service-status');
    statusElements.forEach(element => {
        // 模拟状态变化
        const statuses = ['running', 'healthy', 'processing', 'stopped'];
        const randomStatus = statuses[Math.floor(Math.random() * statuses.length)];
        
        // 更新状态显示
        if (randomStatus === 'running') {
            element.innerHTML = '<span class="text-green-500 flex items-center gap-2 font-medium"><i class="fas fa-check-circle text-xs"></i> 运行中</span>';
        } else if (randomStatus === 'processing') {
            element.innerHTML = '<span class="text-orange-500 flex items-center gap-2 font-medium"><i class="fas fa-spinner fa-spin text-xs"></i> 任务处理中</span>';
        } else if (randomStatus === 'stopped') {
            element.innerHTML = '<span class="text-red-500 flex items-center gap-2 font-medium"><i class="fas fa-times-circle text-xs"></i> 已停止</span>';
        }
    });
}

// 暴露全局函数供内联脚本调用
window.showModal = showModal;
window.hideModal = hideModal;
window.handleLogin = handleLogin;
window.handleLogout = handleLogout;
window.navTo = navTo;
window.showNotification = showNotification;

// 如果页面需要定期更新状态
if (document.querySelector('.service-status')) {
    // 每30秒检查一次服务状态
    setInterval(checkServiceStatus, 30000);
}