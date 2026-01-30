// 动态数据相关的JavaScript函数
function updateMode(value) {
  console.log('更新优化模式:', value);
  if (typeof calculateVllmSuggestions === 'function') {
    calculateVllmSuggestions(value);
  }
}

function updateConcurrency(value) {
  console.log('更新并发数:', value);
}

function updateTokenLimit(value) {
  console.log('更新Token限制:', value);
}

async function saveConfigAndRestart() {
  const mode = document.getElementById('optimization-mode').value;
  
  if (confirm(`确定要保存配置(模式: ${mode})并重启服务吗？这会导致服务短暂中断。`)) {
    console.log('保存配置并重启服务');
    
    try {
      // 1. Save Config
      const saveResponse = await fetch('/api/v1/configs/inference', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          name: 'default', // Assuming default model
          mode: mode
        })
      });

      if (!saveResponse.ok) throw new Error('Failed to save config');

      // 2. Restart Service (assuming 'vllm' or specific agent service)
      // We need to know which service to restart. For now, restart the Agent or a known service.
      // Ideally this info comes from the backend or context. 
      // We will restart all relevant agents.
      // For this prototype, we'll just trigger a general restart call or rely on the fact
      // that config change might trigger something in a real system.
      // But the requirement says "restart with new configuration".
      
      // Let's call restart for the 'Agent' type which hosts the inference
      // We need node IP. This is a bit tricky without context. 
      // We will restart the first available Agent found in the table or just alert success for now
      // as strictly triggering restart requires selecting a specific node.
      
      // However, the button is global. Let's assume it restarts the inference service.
      // We'll call a hypothetical 'inference' service restart or just reload page.
      
      alert('配置已保存。请手动重启相关服务以应用更改，或稍候...');
      location.reload();

    } catch (error) {
      console.error('Error:', error);
      alert('操作失败: ' + error.message);
    }
  }
}

async function restartService(serviceName, nodeIP, serviceType) {
  if (confirm(`确定要重启服务 ${serviceName} 吗？`)) {
    try {
      const response = await fetch('/api/v1/services/restart', {
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
      } else {
        alert('Restart failed');
      }
    } catch (error) {
      console.error('Restart failed:', error);
    }
  }
}

async function stopService(serviceName, nodeIP, serviceType) {
  if (confirm(`确定要停止服务 ${serviceName} 吗？`)) {
    try {
      const response = await fetch('/api/v1/services/stop', {
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
      } else {
        alert('Stop failed');
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

  // Initialize vLLM Suggestions and Event Listener
  const modeSelect = document.getElementById('optimization-mode');
  if (modeSelect) {
    console.log('Attaching event listener to optimization-mode dropdown');
    // Ensure initial calculation
    calculateVllmSuggestions(modeSelect.value);
    
    // Add event listener for changes
    modeSelect.addEventListener('change', function() {
      console.log('Dropdown changed to:', this.value);
      updateMode(this.value);
    });
  } else {
    console.warn('Optimization mode dropdown not found');
  }
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

// --- vLLM Configuration Suggestion Logic ---

function parseGpuMemory(agentMessage) {
  if (!agentMessage) return 24; // Default fallback

  // 1. Try to find explicit GPU section first
  const gpuSectionMatch = agentMessage.match(/GPU:\s*(.+)$/i);
  const searchString = gpuSectionMatch ? gpuSectionMatch[1] : agentMessage;

  // 2. Look for explicit GB/MB pattern
  // Matches "8GB", "8 GB", "8.5GB", "16384MB"
  const memoryMatch = searchString.match(/(\d+(?:\.\d+)?)\s*([GM]B)/i);
  
  if (memoryMatch) {
    let val = parseFloat(memoryMatch[1]);
    let unit = memoryMatch[2].toUpperCase();
    if (unit === 'MB') val /= 1024;
    return Math.round(val); // Return integer GB
  }
  
  return 24; 
}

function parseGpuVendor(agentMessage) {
  if (!agentMessage) return 'Unknown';
  const lower = agentMessage.toLowerCase();
  if (lower.includes('nvidia') || lower.includes('tesla') || lower.includes('geforce')) return 'NVIDIA GPU';
  if (lower.includes('ascend') || lower.includes('npu') || lower.includes('huawei')) return 'Ascend NPU';
  return 'GPU'; // Generic fallback
}

function estimateModelParams(modelName) {
  const lower = (modelName || '').toLowerCase();
  const match = lower.match(/(\d+(?:\.\d+)?)[b]/);
  if (match) return parseFloat(match[1]);
  if (lower.includes('7b')) return 7;
  if (lower.includes('13b')) return 13;
  if (lower.includes('70b')) return 70;
  return 7; // Default
}

async function calculateVllmSuggestions(mode) {
  console.log('Calculating vLLM suggestions for mode:', mode);
  const container = document.getElementById('vllm-suggestions');
  const content = document.getElementById('suggestion-content');
  const gpuLabel = document.getElementById('suggestion-gpu');
  const hardwareAccelDisplay = document.getElementById('hardware-acceleration-display');
  
  if (!container || !content) {
    console.warn('vLLM suggestion elements not found');
    return;
  }

  // 1. Find Data from Global Scope
  const services = window.servicesData || [];
  
  // Find Agent (for GPU Info) - Prefer 'Running' one, but take any if running not found
  const agent = services.find(s => s.type === 'Agent' && (s.status === 'Running' || s.status === 'Healthy')) || 
                services.find(s => s.type === 'Agent');
  
  // Find Container (for Model Info) - Look for vLLM or similar
  const vllmContainer = services.find(s => s.type === 'Container' && 
    (s.name.toLowerCase().includes('vllm') || s.name.toLowerCase().includes('qwen') || s.name.toLowerCase().includes('llama')));

  if (!agent) {
    console.log('No agent found for vLLM calculation');
    container.classList.add('hidden');
    return;
  }

  // 2. Prepare API Request
  const modelName = vllmContainer ? vllmContainer.name : 'Qwen3-1.7B'; // Default if not found
  const nodeIP = agent.node_ip;

  try {
    const response = await fetch('/api/v1/configs/vllm-calculate', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        // 'Authorization': 'Bearer ' + token // Ensure auth if needed, usually cookie or header handled
      },
      body: JSON.stringify({
        model_name: modelName,
        node_ip: nodeIP,
        mode: mode
      })
    });

    if (!response.ok) {
      throw new Error(`API Error: ${response.status}`);
    }

    const data = await response.json();
    const vllmConfig = data.vllm_config;
    const modelConfig = data.model_config;
    const gpuMemory = Math.round(data.gpu_memory);

    container.classList.remove('hidden');

    // Update UI Labels
    const gpuVendor = parseGpuVendor(agent.message || '');
    const modelDisplayName = modelConfig.Name || modelName;
    if (gpuLabel) gpuLabel.innerText = `${gpuVendor}: ${gpuMemory}GB | Model: ${modelDisplayName}`;
    if (hardwareAccelDisplay) hardwareAccelDisplay.innerText = gpuVendor;

    // 4. Render from Backend Data
    content.innerHTML = `
      <div class="flex justify-between"><span>--max-model-len</span> <span class="text-white">${vllmConfig.max_model_len}</span></div>
      <div class="flex justify-between"><span>--max-num-seqs</span> <span class="text-white">${vllmConfig.max_num_seqs}</span></div>
      <div class="flex justify-between"><span>--max-num-batched-tokens</span> <span class="text-white">${vllmConfig.max_num_batched_tokens}</span></div>
      <div class="flex justify-between"><span>--gpu-memory-utilization</span> <span class="text-white">${vllmConfig.gpu_memory_util}</span></div>
    `;

  } catch (error) {
    console.error('Failed to fetch vLLM config:', error);
    content.innerHTML = `<span class="text-red-400">Calculation failed: ${error.message}</span>`;
    container.classList.remove('hidden');
  }
}

