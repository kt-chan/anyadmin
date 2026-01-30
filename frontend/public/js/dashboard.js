// 动态数据相关的JavaScript函数
function updateConcurrency(value) {
  console.log('更新并发数:', value);
  // 这里可以添加AJAX请求来更新后端配置
}

function updateTokenLimit(value) {
  console.log('更新Token限制:', value);
  // 这里可以添加AJAX请求来更新后端配置
}

function saveConfigAndRestart() {
  if (confirm('确定要保存配置并重启服务吗？这会导致服务短暂中断。')) {
    console.log('保存配置并重启服务');
    // 这里可以添加AJAX请求来保存配置并重启服务
    alert('配置已保存，服务正在重启...');
  }
}

async function restartService(serviceName, nodeIP, serviceType) {
  if (confirm(`确定要重启服务 ${serviceName} 吗？`)) {
    try {
      const response = await fetch('/api/service/restart', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          name: serviceName,
          node_ip: nodeIP,
          type: serviceType
        })
      });
      
      if (response.ok) {
        // Quiet success, reload after delay
        setTimeout(() => location.reload(), 2000);
      }
    } catch (error) {
      console.error('Restart failed:', error);
    }
  }
}

async function stopService(serviceName, nodeIP, serviceType) {
  if (confirm(`确定要停止服务 ${serviceName} 吗？`)) {
    try {
      const response = await fetch('/api/service/stop', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          name: serviceName,
          node_ip: nodeIP,
          type: serviceType
        })
      });
      
      if (response.ok) {
        // Quiet success, reload after delay
        setTimeout(() => location.reload(), 2000);
      }
    } catch (error) {
      console.error('Stop failed:', error);
    }
  }
}

function startFullBackup() {
  if (confirm('确定要开始全量备份吗？这可能会影响系统性能。')) {
    console.log('开始全量备份');
    // 这里可以添加AJAX请求来开始备份
    alert('全量备份任务已开始');
  }
}

function restoreFromBackup() {
  if (confirm('确定要从最近备份点恢复吗？这将会覆盖当前数据。')) {
    console.log('从备份恢复');
    // 这里可以添加AJAX请求来执行恢复
    alert('系统恢复任务已开始');
  }
}

// 自动刷新数据（可选）
let refreshInterval = null;

function startAutoRefresh() {
  if (refreshInterval) clearInterval(refreshInterval);
  refreshInterval = setInterval(() => {
    console.log('自动刷新数据...');
    // 这里可以添加AJAX请求来获取最新数据并更新页面
    // 或者直接刷新页面: location.reload();
  }, 15000); // 15秒刷新一次
}

function stopAutoRefresh() {
  if (refreshInterval) {
    clearInterval(refreshInterval);
    refreshInterval = null;
    console.log('已停止自动刷新');
  }
}

// 页面加载完成后启动自动刷新和过滤器
document.addEventListener('DOMContentLoaded', function() {
  // 根据用户选择决定是否启动自动刷新
  const refreshBtn = document.querySelector('button:has(.fa-sync-alt)');
  if (refreshBtn) {
    refreshBtn.addEventListener('click', function() {
      if (refreshInterval) {
        stopAutoRefresh();
        this.innerHTML = '<i class="fas fa-sync-alt"></i> 启动自动刷新';
      } else {
        startAutoRefresh();
        this.innerHTML = '<i class="fas fa-sync-alt"></i> 停止自动刷新 (15s)';
      }
    });
  }
  
  // 初始启动自动刷新
  startAutoRefresh();

  // 初始化过滤器
  initFilters();
});

function initFilters() {
  const nameFilter = document.getElementById('filter-name');
  const typeFilter = document.getElementById('filter-type');
  const statusFilter = document.getElementById('filter-status');

  if (nameFilter && typeFilter && statusFilter) {
    const applyFilters = () => {
      const nameVal = nameFilter.value.toLowerCase();
      const typeVal = typeFilter.value;
      const statusVal = statusFilter.value;

      const rows = document.querySelectorAll('#services-table tbody tr');
      let visibleCount = 0;

      rows.forEach(row => {
        const name = row.getAttribute('data-name');
        const type = row.getAttribute('data-type');
        const status = row.getAttribute('data-status');

        const matchesName = name.includes(nameVal);
        const matchesType = !typeVal || type === typeVal;
        const matchesStatus = !statusVal || status === statusVal || 
                             (statusVal === 'Running' && status === 'healthy');

        if (matchesName && matchesType && matchesStatus) {
          row.style.display = '';
          visibleCount++;
        } else {
          row.style.display = 'none';
        }
      });

      const countEl = document.getElementById('service-count');
      if (countEl) {
        countEl.innerText = `共 ${visibleCount} 个服务`;
      }
    };

    nameFilter.addEventListener('input', applyFilters);
    typeFilter.addEventListener('change', applyFilters);
    statusFilter.addEventListener('change', applyFilters);
  }
}

// 页面离开时清理定时器
window.addEventListener('beforeunload', function() {
  stopAutoRefresh();
});
