// Import Page Scripts

function toggleGlobalPause(event) {
  const btn = event.currentTarget;
  const isPaused = btn.innerText.includes('启动');
  const action = isPaused ? 'resume_all' : 'pause_all';
  
  fetch('/import/api/global', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ action })
  })
  .then(res => res.json())
  .then(data => {
    if (data.success) {
      // Update UI button state
      if (isPaused) {
        btn.innerHTML = '<i class="fas fa-pause-circle mr-2 text-orange-500"></i>全局暂停';
      } else {
        btn.innerHTML = '<i class="fas fa-play-circle mr-2 text-green-500"></i>全局启动';
      }
      refreshTasks();
    }
  });
}

function restartFailedTasks() {
    fetch('/import/api/global', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ action: 'restart_failed' })
  })
  .then(res => res.json())
  .then(data => {
    if (data.success) {
      refreshTasks();
    }
  });
}

function operateTask(taskId, action) {
  if (action === 'delete' && !confirm('Are you sure you want to delete this task?')) {
      return;
  }

  fetch('/import/api/operate', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ taskId, action })
  })
  .then(res => res.json())
  .then(data => {
    if (data.success) {
      // 简单起见，操作后刷新整个列表
      refreshTasks();
    } else {
        alert(data.message);
    }
  });
}

function submitNewTask() {
  const name = document.getElementById('newTaskName').value;
  const type = document.querySelector('input[name="sourceType"]:checked').value;
  const path = document.getElementById('newTaskPath').value;
  const schedule = document.getElementById('newTaskSchedule').value;

  if (!name || !path) {
    alert('请填写完整信息');
    return;
  }

  fetch('/import/api/create', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ name, type, path, schedule })
  })
  .then(res => res.json())
  .then(data => {
    if (data.success) {
      hideImportModal('createTaskModal');
      refreshTasks();
      // Clear form
      document.getElementById('newTaskName').value = '';
      document.getElementById('newTaskPath').value = '';
    }
  });
}

function refreshTasks() {
  // Reload the page to refresh server-side rendered data
  // In a full SPA this would fetch JSON and re-render the table
  window.location.reload();
}

// Auto refresh every 30 seconds
setInterval(() => {
  // refreshTasks(); // Uncomment to enable auto-refresh
}, 30000);

// Modal functions for Import page
function showImportModal(id) { 
  const el = document.getElementById(id);
  if (el) el.classList.remove('hidden'); 
}

function hideImportModal(id) { 
  const el = document.getElementById(id);
  if (el) el.classList.add('hidden'); 
}