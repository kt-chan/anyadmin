document.addEventListener('DOMContentLoaded', () => {
  initWizard();
  initModelManager();
});

// --- Wizard Logic ---
let currentStep = 1;
const totalSteps = 5;

function initWizard() {
  const nextBtn = document.getElementById('next-btn');
  const prevBtn = document.getElementById('prev-btn');

  nextBtn.addEventListener('click', () => {
    if (currentStep < totalSteps) {
      currentStep++;
      updateWizardUI();
    }
  });

  prevBtn.addEventListener('click', () => {
    if (currentStep > 1) {
      currentStep--;
      updateWizardUI();
    }
  });

  // Mode selection listeners
  const modeInputs = document.querySelectorAll('input[name="mode"]');
  modeInputs.forEach(input => {
    input.addEventListener('change', (e) => {
      // Show/Hide specific sections based on mode if needed
      // e.g., toggle "Test Connection" buttons visibility
      const isIntegrate = e.target.value === 'integrate_existing';
      document.querySelectorAll('.integration-only').forEach(el => {
        el.classList.toggle('hidden', !isIntegrate);
      });
    });
  });
}

function updateWizardUI() {
  // Show active step content
  document.querySelectorAll('.step-content').forEach(el => el.classList.remove('active'));
  document.getElementById(`step-${currentStep}`).classList.add('active');

  // Update progress indicators
  document.querySelectorAll('.step-indicator').forEach(el => {
    const step = parseInt(el.dataset.step);
    const circle = el.querySelector('div');
    const text = el.querySelector('span');

    if (step === currentStep) {
      // Active
      el.classList.remove('text-slate-400', 'text-blue-600', 'text-green-600');
      el.classList.add('text-blue-600');
      circle.className = "w-8 h-8 rounded-full bg-blue-600 text-white flex items-center justify-center font-bold text-sm";
    } else if (step < currentStep) {
      // Completed
      el.classList.remove('text-slate-400', 'text-blue-600');
      el.classList.add('text-green-600');
      circle.className = "w-8 h-8 rounded-full bg-green-100 text-green-600 flex items-center justify-center font-bold text-sm";
      circle.innerHTML = '<i class="fas fa-check"></i>';
    } else {
      // Pending
      el.classList.remove('text-blue-600', 'text-green-600');
      el.classList.add('text-slate-400');
      circle.className = "w-8 h-8 rounded-full bg-white border-2 border-slate-200 text-slate-400 flex items-center justify-center font-bold text-sm";
      circle.innerText = step;
    }
  });

  // Update buttons
  const prevBtn = document.getElementById('prev-btn');
  const nextBtn = document.getElementById('next-btn');
  
  if (currentStep === 1) {
    prevBtn.classList.add('hidden');
  } else {
    prevBtn.classList.remove('hidden');
  }

  if (currentStep === totalSteps) {
    nextBtn.classList.add('hidden');
    updateSummary(); // Populate summary on last step
  } else {
    nextBtn.classList.remove('hidden');
  }
}

function updateSummary() {
  const formData = new FormData(document.getElementById('wizard-form'));
  const summaryList = document.getElementById('config-summary');
  summaryList.innerHTML = '';

  const summaryData = [
    { label: 'Mode', value: formData.get('mode') },
    { label: 'Platform', value: formData.get('platform') },
    { label: 'Model', value: formData.get('model_path') || 'Not set' },
    { label: 'Vector DB', value: formData.get('vector_db') }
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
    
    loadModels(); // Load models when tab is switched
  }
};

// --- API Interactions ---

window.generateDeployment = async function() {
  const form = document.getElementById('wizard-form');
  const formData = new FormData(form);
  const data = Object.fromEntries(formData.entries());

  try {
    const response = await fetch('/deployment/api/generate', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(data)
    });
    
    const result = await response.json();
    if (result.success) {
      alert('Deployment script generated! (Mock)');
      console.log(result.data);
    } else {
      alert('Error generating deployment: ' + result.message);
    }
  } catch (err) {
    console.error(err);
    alert('Failed to contact server.');
  }
};

window.testConnection = async function(type) {
  // Simple mock test logic
  // In real app, gather specific fields based on type
  alert(`Testing connection to ${type}... (Mock Success)`);
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
  document.getElementById('selected-model-name').innerText = model.name;
  // In a real app, populate the form fields with model.params
}
