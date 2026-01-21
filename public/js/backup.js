// 备份相关JavaScript函数
async function startFullBackup() {
  if (confirm('确定要开始全量备份吗？这可能需要几分钟时间，且会占用系统资源。')) {
    try {
      const response = await fetch('/api/backup/create', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          backupType: 'full',
          description: '手动全量备份'
        })
      });
      
      const result = await response.json();
      if (result.success) {
        alert(`全量备份已开始，备份ID: ${result.backupId}`);
        // 刷新页面或更新备份列表
        location.reload();
      } else {
        alert('备份启动失败: ' + (result.message || '未知错误'));
      }
    } catch (error) {
      console.error('备份错误:', error);
      alert('网络错误，请重试');
    }
  }
}

async function startIncrementalBackup() {
  if (confirm('确定要开始增量备份吗？')) {
    try {
      const response = await fetch('/api/backup/create', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          backupType: 'incremental',
          description: '手动增量备份'
        })
      });
      
      const result = await response.json();
      if (result.success) {
        alert(`增量备份已开始，备份ID: ${result.backupId}`);
        location.reload();
      } else {
        alert('备份启动失败: ' + (result.message || '未知错误'));
      }
    } catch (error) {
      console.error('备份错误:', error);
      alert('网络错误，请重试');
    }
  }
}

async function startAppReflash() {
  if (confirm('⚠️ 警告：应用重刷将会重新部署所有服务并重置数据！\n\n确定要继续吗？')) {
    try {
      const response = await fetch('/api/system/reflash', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          action: 'app_reflash',
          confirm: true
        })
      });
      
      const result = await response.json();
      if (result.success) {
        alert('应用重刷已启动，系统将在几分钟内完成重置。');
      } else {
        alert('重刷启动失败: ' + (result.message || '未知错误'));
      }
    } catch (error) {
      console.error('重刷错误:', error);
      alert('网络错误，请重试');
    }
  }
}

async function restoreBackup(backupId) {
  if (confirm(`确定要从备份 ${backupId} 恢复吗？这将会覆盖当前数据。`)) {
    try {
      const response = await fetch('/api/backup/restore', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ backupId })
      });
      
      const result = await response.json();
      if (result.success) {
        alert(`恢复任务已开始，系统将重启。\n预计完成时间: ${result.estimatedTime || '5分钟'}`);
      } else {
        alert('恢复失败: ' + (result.message || '未知错误'));
      }
    } catch (error) {
      console.error('恢复错误:', error);
      alert('网络错误，请重试');
    }
  }
}

async function deleteBackup(backupId) {
  if (confirm(`确定要删除备份 ${backupId} 吗？此操作不可撤销。`)) {
    try {
      const response = await fetch('/api/backup/delete', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ backupId })
      });
      
      const result = await response.json();
      if (result.success) {
        alert('备份已删除');
        location.reload();
      } else {
        alert('删除失败: ' + (result.message || '未知错误'));
      }
    } catch (error) {
      console.error('删除错误:', error);
      alert('网络错误，请重试');
    }
  }
}

// 页面加载后初始化
document.addEventListener('DOMContentLoaded', function() {
  // 可以在这里添加页面特定的初始化代码
  console.log('备份页面已加载');
});
