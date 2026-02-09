document.addEventListener('DOMContentLoaded', () => {
    // Tab Switching
    const tabs = document.querySelectorAll('.tab-btn');
    const contents = document.querySelectorAll('.tab-content');

    tabs.forEach(tab => {
        tab.addEventListener('click', () => {
            tabs.forEach(t => t.classList.remove('active', 'border-blue-600', 'text-blue-600'));
            tabs.forEach(t => t.classList.add('border-transparent', 'text-slate-500'));
            
            tab.classList.add('active', 'border-blue-600', 'text-blue-600');
            tab.classList.remove('border-transparent', 'text-slate-500');
            
            contents.forEach(c => c.classList.add('hidden'));
            document.getElementById(tab.dataset.target).classList.remove('hidden');
        });
    });

    // Configuration Data (injected from server)
    const configData = window.SERVER_CONFIG_DATA || {};

    // --- System Settings ---
    const systemForm = document.getElementById('systemForm');
    if (systemForm) {
        systemForm.addEventListener('submit', async (e) => {
            e.preventDefault();
            const formData = new FormData(systemForm);
            const data = Object.fromEntries(formData.entries());
            
            try {
                const res = await fetch('/api/v1/configs/system', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify(data)
                });
                if (res.ok) {
                    showToast('Success', 'System configuration saved successfully', 'success');
                } else {
                    throw new Error('Failed to save');
                }
            } catch (err) {
                showToast('Error', err.message, 'error');
            }
        });
    }

    // --- Agent Config ---
    window.editAgentConfig = (nodeIP) => {
        const node = configData.nodes.find(n => n.node_ip === nodeIP);
        if (!node || !node.agent_config) {
            showToast('Error', 'No agent config found for this node', 'error');
            return;
        }

        const form = document.getElementById('agentConfigForm');
        // Populate form
        // Using a generic approach or specific fields?
        // Let's use specific fields based on the example JSON
        document.getElementById('agent_node_ip_hidden').value = nodeIP;
        
        // Populate specific fields
        const fields = ['mgmt_host', 'mgmt_port', 'log_file']; // Add others as needed
        fields.forEach(f => {
            const input = form.querySelector(`[name="${f}"]`);
            if (input) input.value = node.agent_config[f] || '';
        });

        // Show Modal
        showModal('agentConfigModal');
    };

    const agentForm = document.getElementById('agentConfigForm');
    if (agentForm) {
        agentForm.addEventListener('submit', async (e) => {
            e.preventDefault();
            const formData = new FormData(agentForm);
            const data = Object.fromEntries(formData.entries());
            const nodeIP = data.node_ip_hidden;
            delete data.node_ip_hidden;

            // Merge with existing to keep other fields?
            // For now just send what we have
            
            try {
                const res = await fetch('/api/v1/configs/agent', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ target_node_ip: nodeIP, config: data })
                });
                if (res.ok) {
                    showToast('Success', 'Agent configuration saved', 'success');
                    hideModal('agentConfigModal');
                    // Reload to reflect?
                    setTimeout(() => window.location.reload(), 1000);
                } else {
                    throw new Error('Failed to save');
                }
            } catch (err) {
                showToast('Error', err.message, 'error');
            }
        });
    }

    // --- Service Config (vLLM / AnythingLLM) ---
    window.editServiceConfig = (nodeIP, serviceName, type) => {
        let config;

        if (nodeIP) {
            const node = configData.nodes.find(n => n.node_ip === nodeIP);
            if (!node) return;
            if (type === 'vllm') {
                config = node.inference_cfgs.find(c => c.name === serviceName);
            } else if (type === 'rag') {
                config = node.rag_app_cfgs.find(c => c.name === serviceName);
            }
        } else {
            // Global Config: Find first instance to populate defaults
            if (configData.grouped_services && configData.grouped_services[serviceName]) {
                const instances = configData.grouped_services[serviceName];
                if (instances && instances.length > 0) {
                    config = instances[0].config;
                }
            }
        }

        if (type === 'vllm') {
            openVllmModal(config, nodeIP);
        } else if (type === 'rag') {
            openRagModal(config, nodeIP);
        }
    };

    function openVllmModal(config, nodeIP) {
        const modal = document.getElementById('vllmConfigModal');
        const form = document.getElementById('vllmConfigForm');
        if (!config) return;

        // Populate basic fields
        form.querySelector('[name="name"]').value = config.name;
        form.querySelector('[name="node_ip"]').value = nodeIP; // Hidden
        form.querySelector('[name="model_name"]').value = config.model_name;

        // Set default values from data.json if available
        const modeSelect = document.getElementById('vllm-optimization-mode');
        if (modeSelect) {
            modeSelect.value = config.mode || 'balanced';
        }

        form.querySelector('[name="gpu_memory_size"]').value = config.gpu_memory_size || 24;
        form.querySelector('[name="gpu_memory_utilization"]').value = config.gpu_memory_utilization || 0.85;

        // Set other parameters from config (which might be 0 if not set yet, recalculation will fill them)
        form.querySelector('[name="max_model_len"]').value = config.max_model_len || '';
        form.querySelector('[name="max_num_seqs"]').value = config.max_num_seqs || '';
        form.querySelector('[name="max_num_batched_tokens"]').value = config.max_num_batched_tokens || '';

        // Initialize model selection
        refreshVllmModels(config.model_name);

        showModal('vllmConfigModal');
    }

    window.refreshVllmModels = async function(selectedModel) {
        const form = document.getElementById('vllmConfigForm');
        const modelSelect = document.getElementById('vllm-model-select');
        const nodeIP = form.querySelector('[name="node_ip"]').value;
        const syncIcon = document.querySelector('button[onclick="refreshVllmModels()"] i');

        if (!modelSelect) return;

        modelSelect.disabled = true;
        if (syncIcon) syncIcon.classList.add('fa-spin');
        modelSelect.innerHTML = '<option value="" disabled selected>Loading models...</option>';

        try {
            // Use current config or default if nodeIP is missing (global config)
            const host = nodeIP || '127.0.0.1'; 
            const payload = { host, port: '8000', mode: 'new_deployment' }; // Default to new_deployment for local discovery
            
            const response = await fetch('/deployment/api/discover-models', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(payload)
            });
            
            const result = await response.json();
            modelSelect.innerHTML = '<option value="" disabled>Select a Model</option>';
            
            if (result.success && result.data && result.data.data) {
                result.data.data.forEach(model => {
                    const opt = document.createElement('option');
                    opt.value = model.id;
                    opt.textContent = model.id;
                    if (model.id === selectedModel) opt.selected = true;
                    modelSelect.appendChild(opt);
                });
            } else {
                const opt = document.createElement('option');
                opt.value = "";
                opt.textContent = "No models found";
                modelSelect.appendChild(opt);
            }
        } catch (error) {
            console.error('Error refreshing models:', error);
            modelSelect.innerHTML = '<option value="">Error loading models</option>';
        } finally {
            modelSelect.disabled = false;
            if (syncIcon) syncIcon.classList.remove('fa-spin');
        }
    };

    // Model selection change handler
    const vllmModelSelect = document.getElementById('vllm-model-select');
    if (vllmModelSelect) {
        vllmModelSelect.addEventListener('change', function() {
            const form = document.getElementById('vllmConfigForm');
            const modelNameInput = form.querySelector('[name="model_name"]');
            modelNameInput.value = this.value;
            
            // Trigger recalculation
            const mode = document.getElementById('vllm-optimization-mode').value;
            const nodeIP = form.querySelector('[name="node_ip"]').value;
            const gpuMemorySize = parseFloat(form.querySelector('[name="gpu_memory_size"]').value);
            const gpuUtilization = parseFloat(form.querySelector('[name="gpu_memory_utilization"]').value);
            calculateVllmSuggestionsForServices(mode, this.value, nodeIP, gpuMemorySize, gpuUtilization);
        });
    }

    async function calculateVllmSuggestionsForServices(mode, modelName, nodeIP, gpuMemorySize, gpuUtilization) {
        const form = document.getElementById('vllmConfigForm');
        if (!form) return;

        try {
            const response = await fetch('/api/v1/configs/vllm-calculate', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ 
                    model_name: modelName, 
                    node_ip: nodeIP, 
                    mode: mode,
                    gpu_memory_size: gpuMemorySize,
                    gpu_utilization: gpuUtilization
                })
            });

            if (!response.ok) throw new Error(`API Error: ${response.status}`);
            const data = await response.json();
            
            // Update the calculated parameters
            form.querySelector('[name="max_model_len"]').value = data.vllm_config.max_model_len;
            form.querySelector('[name="max_num_seqs"]').value = data.vllm_config.max_num_seqs;
            form.querySelector('[name="max_num_batched_tokens"]').value = data.vllm_config.max_num_batched_tokens;
            // Note: We don't overwrite gpu_memory_utilization or size here if they were inputs

        } catch (error) {
            console.error('Calculation failed:', error);
            showToast('Error', 'Failed to calculate optimized parameters', 'error');
        }
    }

    // Attach listener for the mode dropdown in modal
    const vllmModeSelect = document.getElementById('vllm-optimization-mode');
    if (vllmModeSelect) {
        vllmModeSelect.addEventListener('change', function() {
            const form = document.getElementById('vllmConfigForm');
            const modelName = form.querySelector('[name="model_name"]').value;
            const nodeIP = form.querySelector('[name="node_ip"]').value;
            const gpuMemorySize = parseFloat(form.querySelector('[name="gpu_memory_size"]').value);
            const gpuUtilization = parseFloat(form.querySelector('[name="gpu_memory_utilization"]').value);
            calculateVllmSuggestionsForServices(this.value, modelName, nodeIP, gpuMemorySize, gpuUtilization);
        });
    }

    // Also add listeners for memory size and utilization changes to recalculate
    const vllmCalcInputs = ['gpu_memory_size', 'gpu_memory_utilization'];
    vllmCalcInputs.forEach(name => {
        const input = document.querySelector(`#vllmConfigForm [name="${name}"]`);
        if (input) {
            input.addEventListener('change', function() {
                const form = document.getElementById('vllmConfigForm');
                const mode = document.getElementById('vllm-optimization-mode').value;
                const modelName = form.querySelector('[name="model_name"]').value;
                const nodeIP = form.querySelector('[name="node_ip"]').value;
                const gpuMemorySize = parseFloat(form.querySelector('[name="gpu_memory_size"]').value);
                const gpuUtilization = parseFloat(form.querySelector('[name="gpu_memory_utilization"]').value);
                calculateVllmSuggestionsForServices(mode, modelName, nodeIP, gpuMemorySize, gpuUtilization);
            });
        }
    });

    function openRagModal(config, nodeIP) {
        const modal = document.getElementById('ragConfigModal');
        const form = document.getElementById('ragConfigForm');
        if (!config) return;

        // Populate
        form.querySelector('[name="name"]').value = config.name;
        form.querySelector('[name="host"]').value = nodeIP; // Hidden/Readonly (it calls it host in struct)
        
        // Map fields
        const map = {
            'storage_dir': 'storage_dir',
            'llm_provider': 'llm_provider',
            'generic_openai_base_path': 'generic_openai_base_path',
            'generic_openai_model_pref': 'generic_openai_model_pref',
            'generic_openai_model_token_limit': 'generic_openai_model_token_limit',
            'generic_openai_max_tokens': 'generic_openai_max_tokens',
            'generic_openai_api_key': 'generic_openai_api_key',
            'vector_db': 'vector_db'
        };

        for (const [key, val] of Object.entries(map)) {
            const input = form.querySelector(`[name="${key}"]`);
            if (input) input.value = config[val] || '';
        }

        showModal('ragConfigModal');
    }

    // --- Shared Save & Restart Logic ---
    async function saveConfigAndRestart(formId, apiUrl, payloadBuilder, modalId, serviceType) {
        const form = document.getElementById(formId);
        if (!form) return;

        form.addEventListener('submit', async (e) => {
            e.preventDefault();
            const formData = new FormData(form);
            const rawData = Object.fromEntries(formData.entries());
            
            // Build payload (custom logic per form)
            let data;
            try {
                data = payloadBuilder(rawData);
            } catch (err) {
                showToast('Error', 'Invalid form data: ' + err.message, 'error');
                return;
            }

            const submitBtn = form.querySelector('button[type="submit"]');
            const originalText = submitBtn.innerText;
            submitBtn.disabled = true;

            try {
                // 1. Save Config
                const res = await fetch(apiUrl, {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify(data)
                });
                
                if (!res.ok) {
                    const errData = await res.json();
                    throw new Error(errData.message || 'Failed to save config');
                }

                // 2. Restart Service
                submitBtn.innerText = '正在重启服务...';
                
                // Determine restart params
                // For vLLM: name=data.name, node_ip=data.ip, type='Container'
                // For RAG: name=data.name, node_ip=data.host, type='Container' (AnythingLLM runs as container too)
                const restartParams = {
                    name: data.name,
                    node_ip: serviceType === 'vLLM' ? data.ip : data.host, // vLLM uses 'ip', RAG uses 'host'
                    type: 'Container'
                };

                const restartRes = await fetch('/api/service/restart', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify(restartParams)
                });

                if (restartRes.ok) {
                    showToast('Success', '配置已保存并触发服务重启', 'success');
                    hideModal(modalId);
                    setTimeout(() => window.location.reload(), 2000);
                } else {
                    throw new Error('Config saved but restart failed');
                }

            } catch (err) {
                showToast('Error', err.message, 'error');
                submitBtn.innerText = originalText;
                submitBtn.disabled = false;
            }
        });
    }

    // Initialize vLLM Save Handler
    saveConfigAndRestart(
        'vllmConfigForm',
        '/api/v1/configs/inference',
        (rawData) => ({
            ...rawData,
            max_model_len: parseInt(rawData.max_model_len),
            max_num_seqs: parseInt(rawData.max_num_seqs),
            max_num_batched_tokens: parseInt(rawData.max_num_batched_tokens),
            gpu_memory_utilization: parseFloat(rawData.gpu_memory_utilization),
            gpu_utilization: parseFloat(rawData.gpu_memory_utilization),
            gpu_memory_size: parseFloat(rawData.gpu_memory_size),
            mode: rawData.optimization_mode,
            ip: rawData.node_ip
        }),
        'vllmConfigModal',
        'vLLM'
    );

    // Initialize RAG (AnythingLLM) Save Handler
    saveConfigAndRestart(
        'ragConfigForm',
        '/api/v1/configs/rag',
        (rawData) => ({
            ...rawData,
            generic_openai_model_token_limit: parseInt(rawData.generic_openai_model_token_limit),
            generic_openai_max_tokens: parseInt(rawData.generic_openai_max_tokens)
        }),
        'ragConfigModal',
        'RAG'
    );

    // --- Add Node Logic ---
    window.downloadSSHKey = async function() {
        try {
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
        }
    };

    window.testNodeSSH = async function() {
        const form = document.getElementById('addNodeForm');
        const formData = new FormData(form);
        const nodesStr = formData.get('target_nodes');
        
        if (!nodesStr || !nodesStr.trim()) {
            alert("Please enter Target Nodes to test SSH connection.");
            return;
        }

        const btn = event.target;
        const originalHtml = btn.innerHTML;
        btn.innerHTML = '<i class="fas fa-spinner fa-spin"></i> Testing...';
        btn.disabled = true;

        try {
            const response = await fetch('/deployment/api/test-connection', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ type: 'ssh', host: nodesStr, port: '22' })
            });
            const result = await response.json();
            
            if (result.status === 'success') {
                alert('Success: ' + result.message);
            } else {
                alert('Error: ' + result.message);
            }
        } catch (err) {
            alert('Network Error');
        } finally {
            btn.innerHTML = originalHtml;
            btn.disabled = false;
        }
    };

    const addNodeForm = document.getElementById('addNodeForm');
    if (addNodeForm) {
        addNodeForm.addEventListener('submit', async (e) => {
            e.preventDefault();
            const formData = new FormData(addNodeForm);
            const data = Object.fromEntries(formData.entries());
            
            // Mode is required for DeployService
            data.mode = 'new_deployment';
            // Disable other services for just node registration
            data.enable_rag = false;
            data.enable_vectordb = false;
            data.enable_parser = false;

            const submitBtn = addNodeForm.querySelector('button[type="submit"]');
            const originalText = submitBtn.textContent;
            submitBtn.innerHTML = '<i class="fas fa-spinner fa-spin mr-2"></i> 注册中...';
            submitBtn.disabled = true;

            try {
                const res = await fetch('/deployment/api/generate', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify(data)
                });
                
                if (res.ok) {
                    showToast('Success', 'Nodes registered and Agent deployment started.', 'success');
                    hideModal('addNodeModal');
                    setTimeout(() => window.location.reload(), 2000);
                } else {
                    const result = await res.json();
                    throw new Error(result.error || result.message || 'Registration failed');
                }
            } catch (err) {
                showToast('Error', err.message, 'error');
            } finally {
                submitBtn.innerHTML = originalText;
                submitBtn.disabled = false;
            }
        });
    }
});

function showToast(title, message, type) {
    // Simple alert for now or implement a toast UI
    alert(`${title}: ${message}`);
}
