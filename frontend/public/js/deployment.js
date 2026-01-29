document.addEventListener('DOMContentLoaded', () => {
  initWizard();
  initTabs();
});

// --- Tabs ---
let nodesRefreshInterval = null;

function initTabs() {
  const tabs = document.querySelectorAll('.tab-btn');
  tabs.forEach(tab => {
    tab.addEventListener('click', () => {
      const target = tab.dataset.target;
      
      // Update Tab UI
      tabs.forEach(t => {
        t.classList.remove('border-blue-600', 'text-blue-600');
        t.classList.add('border-transparent', 'text-slate-400');
      });
      tab.classList.remove('border-transparent', 'text-slate-400');
      tab.classList.add('border-blue-600', 'text-blue-600');
      
      // Update Content UI
      document.querySelectorAll('.tab-pane').forEach(p => p.classList.add('hidden'));
      document.getElementById(target).classList.remove('hidden');
      
      // Handle Auto-Refresh for Nodes Tab
      if (target === 'tab-nodes-content') {
        refreshNodesDashboard();
        if (!nodesRefreshInterval) {
            nodesRefreshInterval = setInterval(() => {
                // Instead of full refresh which causes flicker, update individual nodes
                const nodes = document.querySelectorAll('[id^="status-badge-"]');
                nodes.forEach(n => {
                    const ip = n.id.replace('status-badge-', '').replace(/-/g, '.');
                    updateNodeStatusInDashboard(ip);
                });
            }, 5000); // Update every 5s for smooth transition to warning
        }
      } else {
        if (nodesRefreshInterval) {
            clearInterval(nodesRefreshInterval);
            nodesRefreshInterval = null;
        }
      }
    });
  });
}

async function refreshNodesDashboard() {
  const grid = document.getElementById('nodes-grid');
  if (!grid) return;
  
  try {
    const res = await fetch('/deployment/api/nodes');
    const result = await res.json();
    if (result.success && result.data) {
        grid.innerHTML = '';
        const nodes = result.data;
        
        for (const node of nodes) {
            const ip = node.split(':')[0];
            const card = document.createElement('div');
            card.className = "bg-white rounded-2xl shadow-sm border border-slate-100 overflow-hidden";
            card.innerHTML = `
                <div class="p-6 border-b border-slate-50 flex justify-between items-center bg-slate-50/50">
                    <div class="flex items-center gap-3">
                        <div class="w-10 h-10 rounded-xl bg-blue-600 text-white flex items-center justify-center">
                            <i class="fas fa-server"></i>
                        </div>
                        <div>
                            <h4 class="font-bold text-slate-800">${ip}</h4>
                            <div class="text-[10px] text-slate-400 font-mono">SSH Port: ${node.split(':')[1] || '22'}</div>
                        </div>
                    </div>
                    <div class="flex items-center gap-3">
                        <div id="status-badge-${ip.replace(/\./g, '-')}" class="px-3 py-1 rounded-full text-[10px] font-bold bg-slate-100 text-slate-400 uppercase">
                            Checking...
                        </div>
                        <div class="flex gap-1">
                            <button onclick="handleAgentAction('${ip}', 'start')" title="Start Agent" class="w-8 h-8 rounded-lg bg-white border border-slate-200 text-green-600 hover:bg-green-50 transition flex items-center justify-center">
                                <i class="fas fa-play text-xs"></i>
                            </button>
                            <button onclick="handleAgentAction('${ip}', 'stop')" title="Stop Agent" class="w-8 h-8 rounded-lg bg-white border border-slate-200 text-yellow-600 hover:bg-yellow-50 transition flex items-center justify-center">
                                <i class="fas fa-stop text-xs"></i>
                            </button>
                            <button onclick="handleAgentAction('${ip}', 'delete')" title="Delete Node" class="w-8 h-8 rounded-lg bg-white border border-slate-200 text-red-600 hover:bg-red-50 transition flex items-center justify-center">
                                <i class="fas fa-trash-alt text-xs"></i>
                            </button>
                        </div>
                    </div>
                </div>
                <div class="p-6 grid grid-cols-1 md:grid-cols-3 gap-6" id="node-details-${ip.replace(/\./g, '-')}">
                    <div class="animate-pulse flex space-x-4">
                        <div class="flex-1 space-y-4 py-1">
                            <div class="h-4 bg-slate-100 rounded w-3/4"></div>
                            <div class="space-y-2">
                                <div class="h-4 bg-slate-100 rounded"></div>
                                <div class="h-4 bg-slate-100 rounded w-5/6"></div>
                            </div>
                        </div>
                    </div>
                </div>
                <div class="px-6 pb-6" id="node-services-${ip.replace(/\./g, '-')}"></div>
                <div id="node-footer-${ip.replace(/\./g, '-')}" class="px-6 pb-6 border-t border-slate-50 hidden"></div>
            `;
            grid.appendChild(card);
            
            // Start individual node update
            updateNodeStatusInDashboard(ip);
        }
    }
  } catch (e) {
    grid.innerHTML = '<div class="p-8 text-red-500">Failed to load nodes.</div>';
  }
}

async function updateNodeStatusInDashboard(ip) {
  const safeIp = ip.replace(/\./g, '-');
  const badge = document.getElementById(`status-badge-${safeIp}`);
  const details = document.getElementById(`node-details-${safeIp}`);
  const services = document.getElementById(`node-services-${safeIp}`);
  
  try {
    const res = await fetch(`/deployment/api/status?ip=${ip}`);
    const result = await res.json();
    
    if (result.success && result.data) {
        const agent = result.data;
        const lastSeen = new Date(agent.last_seen);
        const now = new Date();
        const diffSeconds = Math.floor((now - lastSeen) / 1000);

        let statusClass = "bg-green-100 text-green-600";
        let statusText = "Online";
        let isWarning = false;
        let isDown = false;

        if (diffSeconds > 30) {
            statusClass = "bg-red-100 text-red-600";
            statusText = "Down";
            isDown = true;
        } else if (diffSeconds > 10) {
            statusClass = "bg-yellow-100 text-yellow-600";
            statusText = "Warning";
            isWarning = true;
        }
        
        if (agent.status === 'online' || isWarning || isDown) {
            badge.className = `px-3 py-1 rounded-full text-[10px] font-bold ${statusClass} uppercase`;
            badge.innerText = `${statusText} (${diffSeconds}s ago)`;
            
            details.innerHTML = `
                <div class="p-4 bg-slate-50 rounded-xl border border-slate-100 ${isDown ? 'opacity-50' : ''}">
                    <div class="text-[10px] text-slate-400 font-bold uppercase mb-1">CPU Usage</div>
                    <div class="text-xl font-bold text-slate-800">${agent.cpu_usage.toFixed(1)}%</div>
                    <div class="w-full bg-slate-200 h-1.5 rounded-full mt-2">
                        <div class="bg-blue-500 h-1.5 rounded-full" style="width: ${agent.cpu_usage}%"></div>
                    </div>
                </div>
                <div class="p-4 bg-slate-50 rounded-xl border border-slate-100 ${isDown ? 'opacity-50' : ''}">
                    <div class="text-[10px] text-slate-400 font-bold uppercase mb-1">Memory Usage</div>
                    <div class="text-xl font-bold text-slate-800">${agent.memory_usage.toFixed(1)}%</div>
                    <div class="w-full bg-slate-200 h-1.5 rounded-full mt-2">
                        <div class="bg-indigo-500 h-1.5 rounded-full" style="width: ${agent.memory_usage}%"></div>
                    </div>
                </div>
                <div class="p-4 bg-slate-50 rounded-xl border border-slate-100">
                    <div class="flex justify-between items-center mb-1">
                        <div class="text-[10px] text-slate-400 font-bold uppercase">Docker Engine</div>
                        ${agent.docker_status !== 'active' && !isDown ? `
                            <button onclick="handleAgentAction('${ip}', 'fix-docker')" class="text-[10px] text-blue-600 hover:underline font-bold">
                                <i class="fas fa-wrench mr-1"></i>FIX
                            </button>
                        ` : ''}
                    </div>
                    <div class="text-sm font-bold ${agent.docker_status === 'active' && !isDown ? 'text-green-600' : 'text-red-600'} mt-1">
                        ${isDown ? 'UNKNOWN' : agent.docker_status.toUpperCase()}
                    </div>
                    <div class="text-[10px] text-slate-400 mt-1">${agent.hostname} (${agent.os || 'Linux'})</div>
                </div>
            `;
            
            if (agent.services && agent.services.length > 0) {
                services.innerHTML = `
                    <h5 class="text-[10px] font-bold text-slate-400 uppercase tracking-wider mb-3">Running Services</h5>
                    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3">
                        ${agent.services.map(svc => {
                            let svcState = svc.state;
                            let svcColor = svc.state === 'running' ? 'bg-green-500' : 'bg-red-500';
                            
                            if (isDown) {
                                svcState = 'unknown';
                                svcColor = 'bg-slate-400';
                            } else if (isWarning) {
                                svcColor = 'bg-yellow-500';
                            }

                            return `
                                <div class="flex items-center gap-3 p-3 border border-slate-100 rounded-lg ${isDown ? 'bg-slate-50' : ''}">
                                    <div class="w-8 h-8 rounded bg-slate-50 flex items-center justify-center text-slate-400">
                                        <i class="fas fa-box"></i>
                                    </div>
                                    <div class="flex-1 overflow-hidden">
                                        <div class="text-xs font-bold text-slate-700 truncate">${svc.name}</div>
                                        <div class="text-[10px] text-slate-400 truncate">${isDown ? '---' : svc.uptime}</div>
                                    </div>
                                    <div class="w-2 h-2 rounded-full ${svcColor}"></div>
                                </div>
                            `;
                        }).join('')}
                    </div>
                `;
                
            } else {
                services.innerHTML = '';
            }

            const footer = document.getElementById(`node-footer-${safeIp}`);
            if (agent.deployment_time && footer) {
                const depDate = new Date(agent.deployment_time);
                footer.classList.remove('hidden');
                footer.innerHTML = `
                    <div class="mt-4 p-3 bg-blue-50/50 rounded-lg border border-blue-100/50 flex items-center justify-between">
                        <div class="flex items-center gap-2">
                            <i class="fas fa-history text-blue-400 text-xs"></i>
                            <span class="text-[10px] font-bold text-slate-500 uppercase">Last Agent Deployment</span>
                        </div>
                        <span class="text-[10px] font-mono text-blue-600">${depDate.toLocaleString()}</span>
                    </div>
                `;
            } else if (footer) {
                footer.classList.add('hidden');
            }
        }
    } else {
        badge.className = "px-3 py-1 rounded-full text-[10px] font-bold bg-red-100 text-red-600 uppercase";
        badge.innerText = "Offline";
        details.innerHTML = '<div class="col-span-3 text-center py-4 text-slate-400 italic">No agent detected.</div>';
    }
  } catch (e) {
    badge.innerText = "Error";
    details.innerHTML = '<div class="col-span-3 text-center py-4 text-red-400">Error fetching status.</div>';
  }
}

window.handleAgentAction = async function(ip, action) {
  if (action === 'delete') {
    if (!confirm(`Are you sure you want to remove node ${ip}? This will only remove it from the management list.`)) {
        return;
    }
  }

  // Find the button to show loading state
  const btn = event.currentTarget;
  const originalHtml = btn.innerHTML;
  btn.innerHTML = '<i class="fas fa-spinner fa-spin text-xs"></i>';
  btn.disabled = true;

  try {
    let res;
    if (action === 'delete') {
        res = await fetch(`/deployment/api/nodes?ip=${ip}`, { method: 'DELETE' });
    } else {
        res = await fetch('/deployment/api/agent/control', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ ip, action })
        });
    }

    const result = await res.json();
    if (res.ok && (result.success || result.data || result.message)) {
        if (action === 'delete') {
            refreshNodesDashboard();
        } else {
            // Give it a moment then update status
            setTimeout(() => updateNodeStatusInDashboard(ip), 2000);
        }
    } else {
        alert('Error: ' + (result.error || result.message || 'Action failed'));
    }
  } catch (e) {
    alert('Network error');
  } finally {
    btn.innerHTML = originalHtml;
    btn.disabled = false;
  }
};

// --- State ---
let currentStep = 1;
let sshVerified = false;
let inferenceVerified = false;
const totalSteps = 5;

// --- Node Management ---
async function saveTargetNodes() {
  const targetNodesInput = document.querySelector('textarea[name="target_nodes"]');
  if (!targetNodesInput) return;
  const nodes = targetNodesInput.value.split('\n').map(s => s.trim()).filter(s => s);
  
  try {
    await fetch('/deployment/api/nodes', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ nodes })
    });
  } catch (e) { console.error('Failed to save nodes', e); }
}

window.fetchNodesAndPopulate = async function() {
  try {
    const res = await fetch('/deployment/api/nodes');
    const result = await res.json();
    let nodes = (result.success && Array.isArray(result.data)) ? result.data : [];
    
    // Supplement with current textarea input if in wizard
    const targetNodesInput = document.querySelector('textarea[name="target_nodes"]');
    if (targetNodesInput) {
        const wizardNodes = targetNodesInput.value.split('\n').map(s => s.trim()).filter(s => s);
        wizardNodes.forEach(wn => {
            if (!nodes.includes(wn)) nodes.push(wn);
        });
    }

    const selectors = document.querySelectorAll('select.node-selector');
    selectors.forEach(sel => {
        const currentVal = sel.value;
        sel.innerHTML = '<option value="">Select Target Node</option>';
        nodes.forEach(node => {
            const opt = document.createElement('option');
            // Strip port for service selection (e.g. "172.20.0.10:22" -> "172.20.0.10")
            const ipOnly = node.split(':')[0];
            opt.value = ipOnly;
            opt.textContent = ipOnly;
            sel.appendChild(opt);
        });
        if (nodes.includes(currentVal) || nodes.some(n => n.startsWith(currentVal + ':'))) {
            sel.value = currentVal;
        }
    });
  } catch (e) { console.error('Failed to fetch nodes', e); }
}

// --- Wizard Logic ---

window.refreshModels = async function(isManual = false) {
  const hostSelect = document.getElementById('inference-host-select');
  const portInput = document.getElementById('inference-port-input');
  const modelSelect = document.getElementById('model-select');
  
  if (!hostSelect || !portInput || !modelSelect) return;

  const host = hostSelect.value;
  const port = portInput.value;

  if (!host) {
    if(isManual) alert("Please select a target host first.");
    return;
  }

  const originalContent = modelSelect.innerHTML;
  modelSelect.innerHTML = '<option>Loading models...</option>';
  modelSelect.disabled = true;
  
  // Also rotate icon if called via button
  const btn = document.querySelector('button[onclick="refreshModels()"] i');
  if(btn) btn.classList.add('fa-spin');

  try {
    const response = await fetch('/deployment/api/discover-models', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ host, port })
    });
    
    const result = await response.json();
    
    modelSelect.innerHTML = '<option value="" disabled selected>Select a Model</option>';
    
    if (result.success && result.data && result.data.data) {
        // vLLM returns { object: "list", data: [ { id: "..." } ] }
        result.data.data.forEach(model => {
            const opt = document.createElement('option');
            opt.value = model.id;
            opt.textContent = model.id; // Use ID as display name
            modelSelect.appendChild(opt);
        });
        inferenceVerified = true;
    } else {
         const opt = document.createElement('option');
         opt.value = "";
         opt.textContent = "No models found or connection failed";
         modelSelect.appendChild(opt);
         inferenceVerified = false;
    }
    validateStep();

  } catch (e) {
    console.error(e);
    modelSelect.innerHTML = '<option value="">Error fetching models</option>';
    inferenceVerified = false;
    validateStep();
  } finally {
    modelSelect.disabled = false;
    if(btn) btn.classList.remove('fa-spin');
  }
};

function initWizard() {
  const nextBtn = document.getElementById('next-btn');
  const prevBtn = document.getElementById('prev-btn');

  // --- New Logic for Step 3 Model Selection ---
  // Model name input removed from UI, relying on select value.
  
  // Auto-fetch when host changes in Step 3
  const hostSelect = document.getElementById('inference-host-select');
  if (hostSelect) {
      hostSelect.addEventListener('change', () => {
          inferenceVerified = false;
          validateStep();
          if (currentStep === 3) refreshModels();
      });
  }
  
  const portInput = document.getElementById('inference-port-input');
  if (portInput) {
      portInput.addEventListener('input', () => {
          inferenceVerified = false;
          validateStep();
      });
  }

  if (nextBtn) {
    nextBtn.addEventListener('click', async () => {
        if (currentStep < totalSteps) {
        currentStep++;
        updateWizardUI();

        }
    });
  }

  if (prevBtn) {
    prevBtn.addEventListener('click', () => {
        if (currentStep > 1) {
        currentStep--;
        updateWizardUI();
        }
    });
  }

  // Mode selection listeners
  const modeInputs = document.querySelectorAll('input[name="mode"]');
  modeInputs.forEach(input => {
    input.addEventListener('change', (e) => {
      const isIntegrate = e.target.value === 'integrate_existing';
      document.querySelectorAll('.integration-only').forEach(el => {
        el.classList.toggle('hidden', !isIntegrate);
      });
      validateStep();
    });
  });
  
  // Platform selection listeners (to trigger validation)
  const platformInputs = document.querySelectorAll('input[name="platform"]');
  platformInputs.forEach(input => {
      input.addEventListener('change', validateStep);
  });

  // Validation listeners
  const form = document.getElementById('wizard-form');
  if (form) {
    form.addEventListener('input', (e) => {
        if (currentStep === 1) sshVerified = false;
        validateStep();
    });
    form.addEventListener('change', validateStep);
  }
  
  updateWizardUI(); // Initialize UI state
}

window.toggleSection = function(id, isChecked) {
  const el = document.getElementById(id);
  if (!el) return;
  if (isChecked) {
    el.classList.remove('hidden');
  } else {
    el.classList.add('hidden');
  }
  validateStep();
};

function validateStep() {
  const nextBtn = document.getElementById('next-btn');
  if (!nextBtn) return;

  let isValid = true;
  const currentStepEl = document.getElementById(`step-${currentStep}`);
  
  if (!currentStepEl) return;

  // 1. Check Standard Inputs (text, number, select, textarea)
  // Exclude hidden, radio, checkbox, and buttons
  const inputs = currentStepEl.querySelectorAll('input:not([type="hidden"]):not([type="radio"]):not([type="checkbox"]), select, textarea');
  inputs.forEach(input => {
    // Skip validation if the input is inside a hidden section
    if (input.offsetParent === null) return;

    if (!input.value || input.value.trim() === '') {
      isValid = false;
    }
  });

  // 2. Check Radio Groups (Only if visible)
  const radios = currentStepEl.querySelectorAll('input[type="radio"]');
  if (radios.length > 0) {
    const groups = new Set();
    radios.forEach(r => {
        if (r.offsetParent !== null) groups.add(r.name);
    });
    
    groups.forEach(groupName => {
      const groupRadios = currentStepEl.querySelectorAll(`input[name="${groupName}"]`);
      const isChecked = Array.from(groupRadios).some(r => r.checked);
      if (!isChecked) isValid = false;
    });
  }

  // 3. Special Case: Target Nodes (Step 1)
  if (currentStep === 1) {
     const nodes = document.querySelector('textarea[name="target_nodes"]');
     if (nodes && !nodes.value.trim()) isValid = false;
     if (!sshVerified) isValid = false;
  }
  
  // 4. Special Case: Inference Connection (Step 3)
  if (currentStep === 3) {
      if (!inferenceVerified) isValid = false;
  }

  nextBtn.disabled = !isValid;
  if (isValid) {
    nextBtn.classList.remove('opacity-50', 'cursor-not-allowed');
  } else {
    nextBtn.classList.add('opacity-50', 'cursor-not-allowed');
  }
}

function updateWizardUI() {
  document.querySelectorAll('.step-content').forEach(el => el.classList.remove('active'));
  const currentEl = document.getElementById(`step-${currentStep}`);
  if (currentEl) currentEl.classList.add('active');

  // Fetch nodes from backend for steps 3 & 4
  if (currentStep === 3 || currentStep === 4) {
    fetchNodesAndPopulate();
  }

  // Update progress indicators
  document.querySelectorAll('.step-indicator').forEach(el => {
    const step = parseInt(el.dataset.step);
    const circle = el.querySelector('div');
    
    if (step === currentStep) {
      el.classList.remove('text-slate-400', 'text-blue-600', 'text-green-600');
      el.classList.add('text-blue-600');
      circle.className = "w-8 h-8 rounded-full bg-blue-600 text-white flex items-center justify-center font-bold text-sm";
    } else if (step < currentStep) {
      el.classList.remove('text-slate-400', 'text-blue-600');
      el.classList.add('text-green-600');
      circle.className = "w-8 h-8 rounded-full bg-green-100 text-green-600 flex items-center justify-center font-bold text-sm";
      circle.innerHTML = '<i class="fas fa-check"></i>';
    } else {
      el.classList.remove('text-blue-600', 'text-green-600');
      el.classList.add('text-slate-400');
      circle.className = "w-8 h-8 rounded-full bg-white border-2 border-slate-200 text-slate-400 flex items-center justify-center font-bold text-sm";
      circle.innerText = step;
    }
  });

  validateStep(); 

  const prevBtn = document.getElementById('prev-btn');
  const nextBtn = document.getElementById('next-btn');
  
  if (prevBtn) {
    if (currentStep === 1) {
        prevBtn.classList.add('hidden');
    } else {
        prevBtn.classList.remove('hidden');
    }
  }

  if (nextBtn) {
    if (currentStep === totalSteps) {
        nextBtn.classList.add('hidden');
        updateSummary(); 
    } else {
        nextBtn.classList.remove('hidden');
    }
  }
}

function updateSummary() {
  const form = document.getElementById('wizard-form');
  if (!form) return;
  const formData = new FormData(form);
  const summaryList = document.getElementById('config-summary');
  if (!summaryList) return;
  summaryList.innerHTML = '';

  const summaryData = [
    { label: 'Mode', value: formData.get('mode') },
    { label: 'Platform', value: formData.get('platform') },
    { label: 'Model', value: formData.get('model_path') || 'Not set' },
    { label: 'Vector DB', value: formData.get('enable_vectordb') ? formData.get('vector_db') : 'Disabled' },
    { label: 'Parser', value: formData.get('enable_parser') ? 'Mineru' : 'Disabled' },
    { label: 'RAG App', value: formData.get('enable_rag') ? 'AnythingLLM' : 'Disabled' }
  ];

  summaryData.forEach(item => {
    const li = document.createElement('li');
    li.innerHTML = `<span class="font-bold text-slate-800">${item.label}:</span> ${item.value}`;
    summaryList.appendChild(li);
  });
}

// --- API Interactions ---

function getProcessedConfig() {
  const form = document.getElementById('wizard-form');
  if (!form) return {};
  const formData = new FormData(form);
  const data = Object.fromEntries(formData.entries());

  // 1. Handle model_name and model_path
  // User wants model_name to be the value of model_path, and model_path removed.
  if (data.model_path) {
    data.model_name = data.model_path;
  }
  delete data.model_path;

  // 2. Conditional inclusion based on checkboxes
  // If a section is disabled, remove its related fields
  if (!data.enable_rag) {
    delete data.rag_host;
    delete data.rag_port;
  }
  
  if (!data.enable_vectordb) {
    delete data.vector_db;
    delete data.vectordb_host;
    delete data.vectordb_port;
  }

  if (!data.enable_parser) {
    delete data.parser_host;
    delete data.parser_port;
  }

  // Convert checkbox values to boolean if they exist (FormData puts "on" or nothing)
  data.enable_rag = !!data.enable_rag;
  data.enable_vectordb = !!data.enable_vectordb;
  data.enable_parser = !!data.enable_parser;

  return data;
}

window.generateDeployment = async function() {
  const data = getProcessedConfig();

  const generateBtn = document.querySelector('button[onclick="generateDeployment()"]');
  const originalText = generateBtn ? generateBtn.innerHTML : 'Generate';
  if (generateBtn) {
    generateBtn.innerHTML = '<i class="fas fa-spinner fa-spin mr-2"></i> Generating...';
    generateBtn.disabled = true;
  }

  try {
    const response = await fetch('/deployment/api/generate', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(data)
    });
    
    const result = await response.json();
    if (result.success && result.data) {
      
      const resultsContainer = document.getElementById('generation-results') || createResultsContainer();
      resultsContainer.innerHTML = ''; 
      resultsContainer.classList.remove('hidden');

      // 1. Verification Area
      const verifyArea = document.createElement('div');
      verifyArea.className = "mt-6 pt-6 border-t border-slate-200 text-center";
      verifyArea.innerHTML = `
        <h4 class="text-sm font-bold text-slate-700 mb-2">部署进度验证</h4>
        <p class="text-xs text-slate-500 mb-4">正在等待节点 Agent 建立连接并上报心跳。成功后将自动跳转至仪表板。</p>
        
        <div id="agent-status-container" class="mb-4 p-4 bg-blue-50 border border-blue-100 rounded-lg max-w-md mx-auto">
            <div class="flex items-center justify-center gap-3">
                <i class="fas fa-satellite-dish fa-spin text-blue-600"></i>
                <span class="text-sm font-bold text-blue-800" id="agent-status-text">等待 Agent 心跳...</span>
            </div>
            <p class="text-xs text-blue-600 mt-1" id="agent-target-ip">目标地址: ...</p>
            
            <div id="agent-stats" class="hidden mt-3 grid grid-cols-2 gap-2 border-t border-blue-100 pt-2">
                <div class="text-center">
                    <div class="text-[10px] text-blue-400 uppercase">CPU 使用率</div>
                    <div class="text-sm font-bold text-blue-700" id="agent-cpu">- %</div>
                </div>
                <div class="text-center">
                    <div class="text-[10px] text-blue-400 uppercase">内存使用率</div>
                    <div class="text-sm font-bold text-blue-700" id="agent-mem">- %</div>
                </div>
            </div>
        </div>

        <!-- Services Status Container -->
        <div id="services-status-container" class="hidden mt-6 text-left max-w-2xl mx-auto">
            <div class="flex items-center justify-between mb-3">
                <h5 class="text-xs font-bold text-slate-500 uppercase tracking-wider">运行中的服务 (Agent 已检测)</h5>
                <span id="docker-badge" class="px-2 py-0.5 rounded text-[10px] font-bold"></span>
            </div>
            <div id="services-list" class="grid grid-cols-1 gap-2">
                <p class="text-xs text-slate-400 italic">检测服务中...</p>
            </div>
        </div>
      `;
      resultsContainer.appendChild(verifyArea);

      if (resultsContainer.scrollIntoView) {
          resultsContainer.scrollIntoView({ behavior: 'smooth' });
      }

      // Start polling for agent
      startAgentPolling(data);

    } else {
      alert('Error generating deployment: ' + (result.message || 'Unknown error'));
    }
  } catch (err) {
    console.error(err);
    alert('Failed to contact server.');
  } finally {
    if (generateBtn) {
        generateBtn.innerHTML = originalText;
        generateBtn.disabled = false;
    }
  }
};

window.exportConfiguration = function() {
  const data = getProcessedConfig();
  
  // Format the JSON with 2-space indentation for readability
  const jsonStr = JSON.stringify(data, null, 2);
  
  const blob = new Blob([jsonStr], { type: "application/json" });
  const url = URL.createObjectURL(blob);
  const a = document.createElement("a");
  a.href = url;
  a.download = "deployment_config.json";
  document.body.appendChild(a);
  a.click();
  document.body.removeChild(a);
  URL.revokeObjectURL(url);
};

function startAgentPolling(data) {
  // Extract target IP from data
  // data.target_nodes is "IP:Port\nIP:Port"
  if (!data.target_nodes) return;
  const firstNode = data.target_nodes.split('\n')[0].trim();
  if (!firstNode) return;
  
  const ip = firstNode.split(':')[0];
  document.getElementById('agent-target-ip').innerText = `Target: ${ip}`;
  
  pollAgent(ip);
}

async function pollAgent(ip) {
  const statusText = document.getElementById('agent-status-text');
  const container = document.getElementById('agent-status-container');
  const statsDiv = document.getElementById('agent-stats');
  const servicesContainer = document.getElementById('services-status-container');
  const servicesList = document.getElementById('services-list');
  const dockerBadge = document.getElementById('docker-badge');
  
  let attempts = 0;
  const maxAttempts = 300; // 10 minutes (assuming 2s interval) - give it time for full deployment
  
  const interval = setInterval(async () => {
    attempts++;
    try {
      const res = await fetch(`/deployment/api/status?ip=${ip}`);
      if (res.ok) {
        const result = await res.json();
        if (result.success && result.data) {
            const agent = result.data;
            
            if (agent.status === 'online') {
                // container.className = "mb-4 p-4 bg-green-50 border border-green-100 rounded-lg max-w-md mx-auto";
                statusText.innerText = "Agent 已上线";
                statusText.parentElement.querySelector('i').className = "fas fa-check-circle text-green-600";
                
                // Update Stats
                if (statsDiv) {
                    statsDiv.classList.remove('hidden');
                    document.getElementById('agent-cpu').innerText = `${agent.cpu_usage.toFixed(1)}%`;
                    document.getElementById('agent-mem').innerText = `${agent.memory_usage.toFixed(1)}%`;
                }

                // Update Docker Status
                if (servicesContainer) {
                    servicesContainer.classList.remove('hidden');
                    if (agent.docker_status === 'active') {
                        dockerBadge.className = "px-2 py-0.5 rounded text-[10px] font-bold bg-green-100 text-green-700";
                        dockerBadge.innerText = "DOCKER ACTIVE";
                    } else {
                        dockerBadge.className = "px-2 py-0.5 rounded text-[10px] font-bold bg-red-100 text-red-700";
                        dockerBadge.innerText = "DOCKER INACTIVE";
                    }

                    // Update Services
                    if (agent.services && agent.services.length > 0) {
                        servicesList.innerHTML = '';
                        agent.services.forEach(svc => {
                            const svcEl = document.createElement('div');
                            svcEl.className = "flex items-center justify-between p-3 bg-white border border-slate-200 rounded-lg shadow-sm";
                            
                            const isRunning = svc.state === 'running';
                            const statusColor = isRunning ? 'text-green-500' : 'text-red-500';
                            
                            svcEl.innerHTML = `
                                <div class="flex items-center gap-3">
                                    <div class="w-8 h-8 rounded bg-slate-100 flex items-center justify-center text-slate-500">
                                        <i class="fas fa-box text-xs"></i>
                                    </div>
                                    <div>
                                        <div class="text-sm font-bold text-slate-700">${svc.name}</div>
                                        <div class="text-[10px] text-slate-400 font-mono">${svc.image}</div>
                                    </div>
                                </div>
                                <div class="text-right">
                                    <div class="text-[10px] font-bold ${statusColor} uppercase">${svc.state}</div>
                                    <div class="text-[10px] text-slate-400">${svc.uptime}</div>
                                </div>
                            `;
                            servicesList.appendChild(svcEl);
                        });

                        // All verified, redirect to dashboard
                        clearInterval(interval);
                        statusText.innerText = "部署验证成功！正在跳转仪表板...";
                        setTimeout(() => {
                            window.location.href = '/dashboard';
                        }, 2000);
                    } else if (agent.docker_status === 'active') {
                        servicesList.innerHTML = '<p class="text-xs text-slate-400 italic">尚未检测到目标服务。正在启动中...</p>';
                    }
                }
            }
        }
      }
    } catch (e) {
      console.log("Waiting for agent...", e);
    }
    
    if (attempts >= maxAttempts) {
        clearInterval(interval);
        statusText.innerText = "Agent connection timed out.";
        container.className = "mb-4 p-4 bg-red-50 border border-red-100 rounded-lg max-w-md mx-auto";
    }
  }, 2000);
}

function createResultsContainer() {
  const step5 = document.getElementById('step-5');
  const container = document.createElement('div');
  container.id = 'generation-results';
  container.className = "mt-6 border-t pt-6 hidden";
  step5.appendChild(container);
  return container;
}

window.downloadSSHKey = async function() {
  try {
    const btn = document.querySelector('button[onclick="downloadSSHKey()"]');
    const originalText = btn ? btn.innerHTML : 'Download System SSH Key';
    if(btn) {
       btn.innerHTML = '<i class="fas fa-spinner fa-spin"></i> Downloading...';
       btn.disabled = true;
    }

    const response = await fetch('/deployment/api/ssh-key');
    if (!response.ok) throw new Error('Failed to fetch key');
    const keyContent = await response.text();
    
    const blob = new Blob([keyContent], { type: "text/plain" });
    const url = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url;
    a.download = "id_rsa.pub";
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
  } catch(e) {
    alert('Could not download SSH Key: ' + e.message);
  } finally {
    const btn = document.querySelector('button[onclick="downloadSSHKey()"]');
    if(btn) {
       btn.innerHTML = originalText;
       btn.disabled = false;
    }
  }
};

window.testConnection = async function(type) {
  const form = document.getElementById('wizard-form');
  const formData = new FormData(form);
  
  let payload = { type };
  
  if (type === 'inference') {
    // Use the selected target host from Step 3, not the management host
    payload.host = formData.get('inference_host'); 
    payload.port = formData.get('inference_port');
    if (!payload.host) {
        alert("Please select a Target Host first.");
        btn.innerHTML = originalHtml;
        btn.disabled = false;
        return;
    }
  } else if (type === 'vectordb') {
    payload.host = formData.get('vectordb_host');
    payload.port = formData.get('vectordb_port');
  } else if (type === 'parser') {
    payload.host = formData.get('parser_host');
    payload.port = formData.get('parser_port');
  } else if (type === 'rag_app') {
    payload.host = formData.get('rag_host');
    payload.port = formData.get('rag_port');
  } else if (type === 'ssh') {
    const nodesStr = formData.get('target_nodes');
    if (!nodesStr || !nodesStr.trim()) {
       alert("Please enter Target Nodes to test SSH connection.");
       return;
    }
    payload.host = nodesStr; 
    payload.port = "22";
  }

  const btn = event.target.tagName === 'BUTTON' ? event.target : event.target.closest('button');
  const originalHtml = btn.innerHTML;
  btn.innerHTML = '<i class="fas fa-spinner fa-spin"></i> Testing...';
  btn.disabled = true;

  try {
    const response = await fetch('/deployment/api/test-connection', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload)
    });
    const result = await response.json();
    
    if (result.status === 'success') {
      alert('Success: ' + result.message);
      btn.classList.remove('text-blue-600', 'text-red-600');
      btn.classList.add('text-green-600');
      if (type === 'ssh') {
        sshVerified = true;
        validateStep();
      } else if (type === 'inference') {
        inferenceVerified = true;
        validateStep();
      }
    } else {
      alert('Error: ' + result.message);
      btn.classList.remove('text-blue-600', 'text-green-600');
      btn.classList.add('text-red-600');
      if (type === 'ssh') {
        sshVerified = false;
        validateStep();
      } else if (type === 'inference') {
        inferenceVerified = false;
        validateStep();
      }
    }
  } catch (err) {
    alert('Network Error');
  } finally {
    btn.innerHTML = originalHtml;
    btn.disabled = false;
  }
};

window.verifyDeployment = async function() {
  const btn = document.getElementById('verify-btn');
  const originalText = btn.innerHTML;
  btn.innerHTML = '<i class="fas fa-spinner fa-spin mr-2"></i> Verifying...';
  btn.disabled = true;

  const resultsDiv = document.getElementById('verification-results');
  resultsDiv.innerHTML = '';
  resultsDiv.classList.remove('hidden');

  // Get current config to know what to verify
  const config = getProcessedConfig();
  
  // Define checks
  const checks = [];
  
  // Inference Engine
  if (config.inference_host && config.inference_port) {
      checks.push({ name: 'Inference Engine', type: 'inference', host: config.inference_host, port: config.inference_port });
  }
  
  // RAG App
  if (config.enable_rag && config.rag_host && config.rag_port) {
      checks.push({ name: 'RAG Application', type: 'rag_app', host: config.rag_host, port: config.rag_port });
  }
  
  // Vector DB
  if (config.enable_vectordb && config.vectordb_host && config.vectordb_port) {
      checks.push({ name: 'Vector Database', type: 'vectordb', host: config.vectordb_host, port: config.vectordb_port });
  }
  
  // Parser
  if (config.enable_parser && config.parser_host && config.parser_port) {
      checks.push({ name: 'Document Parser', type: 'parser', host: config.parser_host, port: config.parser_port });
  }

  if (checks.length === 0) {
      resultsDiv.innerHTML = '<p class="text-sm text-slate-500">No additional services to verify.</p>';
  }

  for (const check of checks) {
    const el = document.createElement('div');
    el.className = 'flex justify-between items-center py-2 border-b border-slate-700 last:border-0';
    el.innerHTML = `<span class="text-slate-300 text-sm">${check.name}</span> <span class="text-xs text-yellow-500"><i class="fas fa-circle-notch fa-spin"></i> Checking...</span>`;
    resultsDiv.appendChild(el);
    
    // Perform actual check
    try {
        const response = await fetch('/deployment/api/test-connection', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ type: check.type, host: check.host, port: check.port })
        });
        const result = await response.json();
        
        if (result.status === 'success') {
            el.innerHTML = `<span class="text-slate-300 text-sm">${check.name}</span> <span class="text-xs text-green-400"><i class="fas fa-check-circle"></i> Online</span>`;
        } else {
            el.innerHTML = `<span class="text-slate-300 text-sm">${check.name}</span> <span class="text-xs text-red-400"><i class="fas fa-times-circle"></i> Failed (${result.message})</span>`;
        }
    } catch (e) {
        el.innerHTML = `<span class="text-slate-300 text-sm">${check.name}</span> <span class="text-xs text-red-400"><i class="fas fa-exclamation-triangle"></i> Error</span>`;
    }
  }

  btn.innerHTML = '<i class="fas fa-check-double mr-2"></i> Verified';
  setTimeout(() => {
     btn.innerHTML = originalText;
     btn.disabled = false;
  }, 3000);
};

// Helper
function escapeHtml(text) {
  if (!text) return text;
  return text
      .replace(/&/g, "&amp;")
      .replace(/</g, "&lt;")
      .replace(/>/g, "&gt;")
      .replace(/"/g, "&quot;")
      .replace(/'/g, "&#039;");
}
window.copyContent = function(btn) {
    const pre = btn.closest('.mb-4').querySelector('pre');
    navigator.clipboard.writeText(pre.textContent).then(() => {
        const original = btn.innerHTML;
        btn.innerHTML = '<i class="fas fa-check mr-1"></i> Copied';
        setTimeout(() => btn.innerHTML = original, 2000);
    });
};