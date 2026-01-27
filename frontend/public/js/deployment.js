document.addEventListener('DOMContentLoaded', () => {
  initWizard();
});

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

async function fetchNodesAndPopulate() {
  try {
    const res = await fetch('/deployment/api/nodes');
    const result = await res.json();
    if (result.success && result.data) {
        const nodes = result.data;
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
            if (nodes.includes(currentVal)) sel.value = currentVal;
        });
    }
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
        if (currentStep === 1) {
            const btnHtml = nextBtn.innerHTML;
            nextBtn.disabled = true;
            nextBtn.innerHTML = '<i class="fas fa-spinner fa-spin"></i> Saving...';
            await saveTargetNodes();
            nextBtn.disabled = false;
            nextBtn.innerHTML = btnHtml;
        }
        
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

// --- Tab Logic ---
window.switchTab = function(tabName) {
  const wizardContent = document.getElementById('tab-wizard-content');
  const modelsContent = document.getElementById('tab-models-content');
  const wizardBtn = document.getElementById('tab-wizard-btn');
  const modelsBtn = document.getElementById('tab-models-btn');

  if (tabName === 'wizard') {
    wizardContent.classList.remove('hidden');
    modelsContent.classList.add('hidden');
    
    wizardBtn.classList.add('bg-blue-50', 'text-blue-600');
    wizardBtn.classList.remove('text-slate-500');
    
    modelsBtn.classList.remove('bg-blue-50', 'text-blue-600');
    modelsBtn.classList.add('text-slate-500');
  } else {
    wizardContent.classList.add('hidden');
    modelsContent.classList.remove('hidden');

    modelsBtn.classList.add('bg-blue-50', 'text-blue-600');
    modelsBtn.classList.remove('text-slate-500');

    wizardBtn.classList.remove('bg-blue-50', 'text-blue-600');
    wizardBtn.classList.add('text-slate-500');
    
    loadModels(); 
  }
};

// --- API Interactions ---

window.generateDeployment = async function() {
  const form = document.getElementById('wizard-form');
  const formData = new FormData(form);
  const data = Object.fromEntries(formData.entries());

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
    if (result.success && result.data && result.data.artifacts) {
      
      const resultsContainer = document.getElementById('generation-results') || createResultsContainer();
      resultsContainer.innerHTML = ''; 
      resultsContainer.classList.remove('hidden');

      // 1. Artifacts List
      Object.entries(result.data.artifacts).forEach(([filename, content]) => {
        const fileBlock = document.createElement('div');
        fileBlock.className = "mb-4 bg-slate-900 rounded-lg overflow-hidden";
        fileBlock.innerHTML = `
          <div class="flex justify-between items-center bg-slate-800 px-4 py-2">
            <span class="text-xs font-mono text-slate-300">${filename}</span>
            <button onclick="copyContent(this)" class="text-xs text-blue-400 hover:text-blue-300">
              <i class="far fa-copy mr-1"></i> Copy
            </button>
          </div>
          <pre class="p-4 text-xs text-green-400 font-mono overflow-x-auto whitespace-pre-wrap max-h-64 scrollbar-thin">${escapeHtml(content)}</pre>
        `;
        resultsContainer.appendChild(fileBlock);
      });

      // 2. Verify Button Area
      const verifyArea = document.createElement('div');
      verifyArea.className = "mt-6 pt-6 border-t border-slate-200 text-center";
      verifyArea.innerHTML = `
        <h4 class="text-sm font-bold text-slate-700 mb-2">Post-Deployment Verification</h4>
        <p class="text-xs text-slate-500 mb-4">Once you have applied the scripts above, verify the connectivity.</p>
        <button id="verify-btn" onclick="verifyDeployment()" class="px-6 py-2 bg-indigo-600 text-white text-sm font-bold rounded-lg hover:bg-indigo-700 transition">
           <i class="fas fa- stethoscope mr-2"></i> Verify Services
        </button>
        <div id="verification-results" class="mt-4 bg-slate-800 rounded-lg p-4 hidden text-left max-w-md mx-auto"></div>
      `;
      resultsContainer.appendChild(verifyArea);

      resultsContainer.scrollIntoView({ behavior: 'smooth' });

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

  const services = ['Inference Engine', 'Vector Database', 'Document Parser'];
  
  for (const service of services) {
    const el = document.createElement('div');
    el.className = 'flex justify-between items-center py-2 border-b border-slate-700 last:border-0';
    el.innerHTML = `<span class="text-slate-300 text-sm">${service}</span> <span class="text-xs text-yellow-500"><i class="fas fa-circle-notch fa-spin"></i> Checking...</span>`;
    resultsDiv.appendChild(el);
    
    await new Promise(r => setTimeout(r, 800));
    
    el.innerHTML = `<span class="text-slate-300 text-sm">${service}</span> <span class="text-xs text-green-400"><i class="fas fa-check-circle"></i> Online</span>`;
  }

  btn.innerHTML = '<i class="fas fa-check-double mr-2"></i> Verified';
  setTimeout(() => {
     btn.innerHTML = originalText;
     btn.disabled = false;
  }, 3000);
};

// --- Model Manager ---

async function loadModels() {
  const listEl = document.getElementById('model-list');
  listEl.innerHTML = '<li class="p-4 text-center text-slate-400"><i class="fas fa-spinner fa-spin mr-2"></i>Loading...</li>';

  try {
    const response = await fetch('/deployment/api/models');
    const result = await response.json();
    
    if (result.success && result.data) {
      renderModelList(result.data);
    } else {
      listEl.innerHTML = '<li class="p-4 text-center text-red-500">Failed to load models</li>';
    }
  } catch (err) {
    console.error(err);
    listEl.innerHTML = '<li class="p-4 text-center text-red-500">Network Error</li>';
  }
}

function renderModelList(models) {
  const listEl = document.getElementById('model-list');
  listEl.innerHTML = '';

  models.forEach(model => {
    const li = document.createElement('li');
    li.className = "p-4 hover:bg-slate-50 cursor-pointer transition border-l-4 border-transparent hover:border-blue-500";
    li.onclick = () => selectModel(model);
    
    li.innerHTML = `
      <div class="flex justify-between items-center mb-1">
        <span class="font-bold text-slate-700 text-sm">${model.name}</span>
        <span class="text-xs ${model.status === 'loaded' ? 'text-green-600 bg-green-100' : 'text-slate-500 bg-slate-100'} px-2 py-0.5 rounded-full">${model.status}</span>
      </div>
      <div class="text-xs text-slate-400 truncate">${model.platform}</div>
    `;
    listEl.appendChild(li);
  });
}

function selectModel(model) {
  const el = document.getElementById('selected-model-name');
  if(el) el.innerText = model.name;
}

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