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
        document.getElementById('agent_node_ip_hidden').value = nodeIP;
        
        const fields = ['mgmt_host', 'mgmt_port', 'log_file']; 
        fields.forEach(f => {
            const input = form.querySelector(`[name="${f}"]`);
            if (input) input.value = node.agent_config[f] || '';
        });

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

            try {
                const res = await fetch('/api/v1/configs/agent', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ target_node_ip: nodeIP, config: data })
                });
                if (res.ok) {
                    showToast('Success', 'Agent configuration saved', 'success');
                    hideModal('agentConfigModal');
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
        const form = document.getElementById('vllmConfigForm');
        if (!config) return;

        form.querySelector('[name="name"]').value = config.name;
        form.querySelector('[name="node_ip"]').value = nodeIP; 
        form.querySelector('[name="model_name"]').value = config.model_name;

        const modeSelect = document.getElementById('vllm-optimization-mode');
        if (modeSelect) {
            modeSelect.value = config.mode || 'balanced';
        }

        form.querySelector('[name="gpu_memory_size"]').value = config.gpu_memory_size || 24;
        form.querySelector('[name="gpu_memory_utilization"]').value = config.gpu_memory_utilization || 0.85;

        form.querySelector('[name="max_model_len"]').value = config.max_model_len || '';
        form.querySelector('[name="max_num_seqs"]').value = config.max_num_seqs || '';
        form.querySelector('[name="max_num_batched_tokens"]').value = config.max_num_batched_tokens || '';

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
            const host = nodeIP || '127.0.0.1'; 
            const payload = { host, port: '8000', mode: 'new_deployment' }; 
            
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
                modelSelect.innerHTML = '<option value="">No models found</option>';
            }
        } catch (error) {
            console.error('Error refreshing models:', error);
            modelSelect.innerHTML = '<option value="">Error loading models</option>';
        } finally {
            modelSelect.disabled = false;
            if (syncIcon) syncIcon.classList.remove('fa-spin');
        }
    };

    window.refreshRagModels = async function(selectedModel) {
        const form = document.getElementById('ragConfigForm');
        const modelSelect = document.getElementById('rag-model-select');
        const host = form.querySelector('[name="host"]').value;
        const syncIcon = document.querySelector('button[onclick="refreshRagModels()"] i');

        if (!modelSelect) return;

        modelSelect.disabled = true;
        if (syncIcon) syncIcon.classList.add('fa-spin');
        modelSelect.innerHTML = '<option value="" disabled selected>Loading models...</option>';

        try {
            const targetHost = host || '127.0.0.1'; 
            const payload = { host: targetHost, port: '8000', mode: 'new_deployment' }; 
            
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
                modelSelect.innerHTML = '<option value="">No models found</option>';
            }
        } catch (error) {
            console.error('Error refreshing models:', error);
            modelSelect.innerHTML = '<option value="">Error loading models</option>';
        } finally {
            modelSelect.disabled = false;
            if (syncIcon) syncIcon.classList.remove('fa-spin');
        }
    };

    const vllmModelSelect = document.getElementById('vllm-model-select');
    if (vllmModelSelect) {
        vllmModelSelect.addEventListener('change', function() {
            const form = document.getElementById('vllmConfigForm');
            const modelNameInput = form.querySelector('[name="model_name"]');
            modelNameInput.value = this.value;
            
            const mode = document.getElementById('vllm-optimization-mode').value;
            const nodeIP = form.querySelector('[name="node_ip"]').value;
            const gpuMemorySize = parseFloat(form.querySelector('[name="gpu_memory_size"]').value);
            const gpuUtilization = parseFloat(form.querySelector('[name="gpu_memory_utilization"]').value);
            calculateVllmSuggestionsForServices(mode, this.value, nodeIP, gpuMemorySize, gpuUtilization);
        });
    }

    const ragModelSelect = document.getElementById('rag-model-select');
    if (ragModelSelect) {
        ragModelSelect.addEventListener('change', function() {
            const form = document.getElementById('ragConfigForm');
            const modelNameInput = form.querySelector('[name="generic_openai_model_pref"]');
            if (modelNameInput) modelNameInput.value = this.value;
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
            
            form.querySelector('[name="max_model_len"]').value = data.vllm_config.max_model_len;
            form.querySelector('[name="max_num_seqs"]').value = data.vllm_config.max_num_seqs;
            form.querySelector('[name="max_num_batched_tokens"]').value = data.vllm_config.max_num_batched_tokens;

        } catch (error) {
            console.error('Calculation failed:', error);
            showToast('Error', 'Failed to calculate optimized parameters', 'error');
        }
    }

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
        const form = document.getElementById('ragConfigForm');
        if (!config) return;

        form.querySelector('[name="name"]').value = config.name;
        form.querySelector('[name="host"]').value = nodeIP; 
        
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

        refreshRagModels(config.generic_openai_model_pref);
        showModal('ragConfigModal');
    }

    async function saveConfigAndRestart(formId, apiUrl, payloadBuilder, modalId, serviceType) {
        const form = document.getElementById(formId);
        if (!form) return;

        form.addEventListener('submit', async (e) => {
            e.preventDefault();
            const formData = new FormData(form);
            const rawData = Object.fromEntries(formData.entries());
            
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
                const res = await fetch(apiUrl, {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify(data)
                });
                
                if (!res.ok) {
                    const errData = await res.json();
                    throw new Error(errData.message || 'Failed to save config');
                }

                submitBtn.innerText = '正在重启服务...';
                
                const restartParams = {
                    name: data.name,
                    node_ip: serviceType === 'vLLM' ? data.ip : data.host, 
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
            data.mode = 'new_deployment';
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

    // --- Connect Service Logic ---
    window.toggleServiceConnectMode = function(mode) {
        const testBtnGroup = document.getElementById('test-connection-group');
        const managedFields = document.querySelectorAll('.managed-only');
        const externalFields = document.getElementById('external-fields');
        
        if (mode === 'integrate_existing') {
            if (testBtnGroup) testBtnGroup.classList.remove('hidden');
            if (externalFields) externalFields.classList.remove('hidden');
            managedFields.forEach(el => el.classList.add('hidden'));
        } else {
            if (testBtnGroup) testBtnGroup.classList.add('hidden');
            if (externalFields) externalFields.classList.add('hidden');
            managedFields.forEach(el => el.classList.remove('hidden'));
        }
    };

    const populateModelData = async () => {
        try {
            const typeRes = await fetch('/models/api/model-types');
            const types = await typeRes.json();
            
            const typeSelects = ['connect-model-type', 'external-model-type'];
            typeSelects.forEach(id => {
                const sel = document.getElementById(id);
                if (!sel) return;
                sel.innerHTML = id === 'connect-model-type' ? '<option value="" disabled selected>选择模型类型</option>' : '<option value="" disabled selected>选择外部模型类型</option>';
                types.forEach(t => {
                    const opt = document.createElement('option');
                    opt.value = t;
                    opt.textContent = t.toUpperCase();
                    sel.appendChild(opt);
                });
            });

            const modelRes = await fetch('/models/api');
            const modelData = await modelRes.json();
            const models = modelData.data || [];

            const typeSelect = document.getElementById('connect-model-type');
            const nameSelect = document.getElementById('connect-model-name');
            
            if (typeSelect && nameSelect) {
                typeSelect.addEventListener('change', () => {
                    const selectedType = typeSelect.value;
                    nameSelect.innerHTML = '<option value="" disabled selected>选择模型名称</option>';
                    
                    const filtered = models.filter(m => m.model_type === selectedType);
                    if (filtered.length > 0) {
                        filtered.forEach(m => {
                            const opt = document.createElement('option');
                            opt.value = m.name;
                            opt.textContent = m.name;
                            nameSelect.appendChild(opt);
                        });
                    } else {
                        const opt = document.createElement('option');
                        opt.value = "";
                        opt.textContent = "该类型下暂无模型";
                        nameSelect.appendChild(opt);
                    }
                });
            }

        } catch (e) {
            console.error("Failed to load model data for connect modal", e);
        }
    };

    const populateConnectNodes = async () => {
        const select = document.getElementById('connect-node-select');
        if (!select) return;
        
        try {
            const res = await fetch('/deployment/api/nodes');
            const result = await res.json();
            const nodes = result.data || [];
            
            select.innerHTML = '<option value="" disabled selected>选择目标节点</option>';
            nodes.forEach(node => {
                const opt = document.createElement('option');
                const nodeIP = typeof node === 'string' ? node : node.node_ip;
                opt.value = nodeIP;
                opt.textContent = nodeIP;
                select.appendChild(opt);
            });
        } catch(e) {
            console.error("Failed to load nodes", e);
            select.innerHTML = '<option value="" disabled>加载节点失败</option>';
        }
    };

    window.testServiceConnect = async function() {
        const form = document.getElementById('connectServiceForm');
        const formData = new FormData(form);
        const host = formData.get('target_node');
        const port = formData.get('port');
        const serviceType = formData.get('service_type');

        if (!host || !port) {
            alert("请选择节点并填写端口");
            return;
        }

        const btn = document.querySelector('#test-connection-group button');
        const originalHtml = btn.innerHTML;
        btn.innerHTML = '<i class="fas fa-spinner fa-spin mr-2"></i>Testing...';
        btn.disabled = true;

        try {
            let connType = 'tcp';
            if (serviceType === 'inference') connType = 'inference';
            else if (serviceType === 'rag') connType = 'rag_app';

            const response = await fetch('/deployment/api/test-connection', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ type: connType, host, port })
            });
            const result = await response.json();
            
            if (result.status === 'success') {
                alert('Success: ' + result.message);
            } else {
                alert('Error: ' + result.message);
            }
        } catch (err) {
            alert('网络错误，测试失败');
        } finally {
            btn.innerHTML = originalHtml;
            btn.disabled = false;
        }
    };

    const connectServiceForm = document.getElementById('connectServiceForm');
    if (connectServiceForm) {
        connectServiceForm.addEventListener('submit', async (e) => {
            e.preventDefault();
            const formData = new FormData(connectServiceForm);
            
            const payload = {
                mode: formData.get('mode'),
                platform: 'nvidia', 
                mgmt_host: configData.mgmt_host || '172.20.0.1',
                mgmt_port: configData.mgmt_port || '8080',
                target_nodes: formData.get('target_node'),
                enable_rag: false,
                enable_vectordb: false,
                enable_parser: false,
            };

            const svcType = formData.get('service_type');
            const mode = formData.get('mode');

            if (svcType === 'inference') {
                payload.inference_host = formData.get('target_node');
                payload.inference_port = formData.get('port');
                
                if (mode === 'new_deployment') {
                    payload.model_type = formData.get('model_type_select');
                    payload.model_name = formData.get('model_name_select');
                } else {
                    payload.model_type = formData.get('external_model_type');
                    payload.model_name = formData.get('external_model_name');
                    payload.api_key = formData.get('api_key');
                    payload.base_url = formData.get('base_url');
                }
            } else if (svcType === 'rag') {
                payload.enable_rag = true;
                payload.rag_host = formData.get('target_node');
                payload.rag_port = formData.get('port');
            } else if (svcType === 'vectordb') {
                payload.enable_vectordb = true;
                payload.vectordb_host = formData.get('target_node');
                payload.vectordb_port = formData.get('port');
                payload.vector_db = 'lancedb'; 
            } else if (svcType === 'parser') {
                payload.enable_parser = true;
                payload.parser_host = formData.get('target_node');
                payload.parser_port = formData.get('port');
            }

            const submitBtn = connectServiceForm.querySelector('button[type="submit"]');
            const originalText = submitBtn.textContent;
            submitBtn.innerHTML = '<i class="fas fa-spinner fa-spin mr-2"></i> 处理中...';
            submitBtn.disabled = true;

            try {
                const res = await fetch('/deployment/api/generate', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify(payload)
                });
                
                if (res.ok) {
                    showToast('Success', '服务配置已保存并开始部署流程', 'success');
                    hideModal('connectServiceModal');
                    setTimeout(() => window.location.reload(), 1500);
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

    populateModelData();
    populateConnectNodes();
});

function showToast(title, message, type) {
    if (window.showNotification) {
        window.showNotification(message, type);
    } else {
        alert(`${title}: ${message}`);
    }
}
