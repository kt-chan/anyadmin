document.addEventListener('DOMContentLoaded', () => {
    // --- Chunked Upload Logic ---
    const CHUNK_SIZE = 5 * 1024 * 1024; // 5MB
    let isPaused = false;
    let isUploading = false;
    
    // State to track upload progress for each file
    let uploadState = {
        tar: { file: null, id: null, offset: 0, total: 0 },
        sum: { file: null, id: null, offset: 0, total: 0 }
    };

    const uploadForm = document.getElementById('uploadModelForm');
    const submitBtn = document.getElementById('uploadSubmitBtn');
    const pauseBtn = document.getElementById('pauseBtn');
    const resumeBtn = document.getElementById('resumeBtn');
    const abortBtn = document.getElementById('abortBtn');

    // --- Form Validation ---
    const checkFormValidity = () => {
        const name = document.getElementById('modelName').value;
        const type = document.getElementById('modelType').value;
        const tarFile = document.getElementById('tarFile').files.length > 0;
        const sumFile = document.getElementById('sumFile').files.length > 0;
        
        submitBtn.disabled = !(name && type && tarFile && sumFile) || isUploading;
    };

    ['modelName', 'modelType', 'tarFile', 'sumFile'].forEach(id => {
        const el = document.getElementById(id);
        if (el) el.addEventListener('input', checkFormValidity);
        if (el) el.addEventListener('change', checkFormValidity);
    });

    if (uploadForm) {
        uploadForm.addEventListener('submit', async (e) => {
            e.preventDefault();
            if (isUploading) return;

            const name = document.getElementById('modelName').value;
            const tarFile = document.getElementById('tarFile').files[0];
            const sumFile = document.getElementById('sumFile').files[0];

            if (!name || !tarFile || !sumFile) {
                showToast("错误", "请填写所有必填项", "error");
                return;
            }

            // Initialize State
            isPaused = false;
            isUploading = true;
            uploadState.tar = { file: tarFile, id: null, offset: 0, total: tarFile.size };
            uploadState.sum = { file: sumFile, id: null, offset: 0, total: sumFile.size };

            // UI Update
            toggleInputs(false);
            showProgressUI(true);
            updateButtons('uploading');

            try {
                // 1. Initialize Uploads (Get IDs and Resume Offsets)
                await initUploadSession('tar', tarFile);
                if (isUploading) await initUploadSession('sum', sumFile);

                // 2. Start Upload Loop
                const modelType = document.getElementById('modelType').value;
                if (isUploading) await processUploads(name, modelType);

            } catch (err) {
                if (isUploading) {
                    console.error(err);
                    showToast("上传失败", err.message, "error");
                    resetUI();
                }
            }
        });
    }

    if (pauseBtn) {
        pauseBtn.addEventListener('click', () => {
            isPaused = true;
            updateButtons('paused');
        });
    }

    if (resumeBtn) {
        resumeBtn.addEventListener('click', () => {
            isPaused = false;
            updateButtons('uploading');
            const name = document.getElementById('modelName').value;
            processUploads(name); // Resume loop
        });
    }

    if (abortBtn) {
        abortBtn.addEventListener('click', async () => {
            if (confirm("确定要中止上传吗？已上传的数据将被彻底删除。")) {
                const tarId = uploadState.tar.id;
                const sumId = uploadState.sum.id;
                
                isUploading = false;
                isPaused = false;
                resetUI();

                try {
                    // Call abort API for both sessions
                    if (tarId) await fetch('/models/api/upload/abort', {
                        method: 'POST',
                        headers: { 'Content-Type': 'application/json' },
                        body: JSON.stringify({ upload_id: tarId })
                    });
                    if (sumId) await fetch('/models/api/upload/abort', {
                        method: 'POST',
                        headers: { 'Content-Type': 'application/json' },
                        body: JSON.stringify({ upload_id: sumId })
                    });
                    showNotification("上传已中止，临时文件已清理", "info");
                } catch (e) {
                    console.error("Failed to abort on server:", e);
                    showNotification("上传已中止，但部分临时文件可能清理失败", "warning");
                }
            }
        });
    }

    async function initUploadSession(type, file) {
        if (!isUploading) return;
        const response = await fetch('/models/api/upload/init', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ filename: file.name, total_size: file.size })
        });
        
        if (!response.ok) throw new Error(`Init failed for ${type}`);
        const data = await response.json();
        
        uploadState[type].id = data.upload_id;
        uploadState[type].offset = data.offset || 0;
        
        updateProgress(type, uploadState[type].offset, uploadState[type].total);
    }

    async function processUploads(modelName, modelType) {
        // Upload loop for both files. simpler to do one then the other or parallel?
        // Let's do parallel chunks for speed? No, simpler sequential or interleaved.
        // Let's just loop until both are done.

        while (isUploading && (uploadState.tar.offset < uploadState.tar.total || uploadState.sum.offset < uploadState.sum.total) && !isPaused) {
            
            // Upload Tar Chunk
            if (isUploading && uploadState.tar.offset < uploadState.tar.total) {
                await uploadChunk('tar');
            }

            // Upload Sum Chunk
            if (isUploading && uploadState.sum.offset < uploadState.sum.total) {
                await uploadChunk('sum');
            }
        }

        if (isUploading && !isPaused) {
            // Both done
            finalizeUpload(modelName, modelType);
        }
    }

    async function uploadChunk(type) {
        if (!isUploading) return;
        const state = uploadState[type];
        const start = state.offset;
        const end = Math.min(start + CHUNK_SIZE, state.total);
        const chunk = state.file.slice(start, end);

        const formData = new FormData();
        formData.append('upload_id', state.id);
        formData.append('chunk', chunk);

        try {
            const res = await fetch('/models/api/upload/chunk', {
                method: 'POST',
                body: formData
            });

            if (!res.ok) throw new Error(`Chunk upload failed for ${type}`);
            const data = await res.json();
            
            if (isUploading) {
                state.offset = data.offset; // Update offset from server response
                updateProgress(type, state.offset, state.total);
            }

        } catch (e) {
            if (isUploading) {
                console.error(`Chunk error ${type}:`, e);
                // If connection failure, maybe pause and ask user to resume?
                // Or simple retry logic?
                // For now, pause.
                isPaused = true;
                updateButtons('paused');
                showToast("连接中断", "上传中断，请检查网络后点击继续", "warning");
                throw e; // Break loop
            }
        }
    }

    async function finalizeUpload(modelName, modelType) {
        if (!isUploading) return;
        updateButtons('finalizing');
        document.getElementById(`tarStatus`).innerText = "正在校验并保存...";
        document.getElementById(`sumStatus`).innerText = "正在校验并保存...";
        
        try {
            const res = await fetch('/models/api/finalize', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    model_name: modelName,
                    model_type: modelType,
                    tar_upload_id: uploadState.tar.id,
                    checksum_upload_id: uploadState.sum.id
                })
            });

            const data = await res.json();
            
            if (res.ok && isUploading) {
                showNotification("模型上传并验证成功", "success");
                setTimeout(() => window.location.href = '/models', 1000);
            } else if (isUploading) {
                throw new Error(data.error || "Finalization failed");
            }
        } catch (e) {
            if (isUploading) {
                showNotification("验证失败: " + e.message, "error");
                resetUI();
            }
        }
    }

    // --- UI Helpers ---

    function toggleInputs(enabled) {
        document.getElementById('modelName').disabled = !enabled;
        document.getElementById('modelType').disabled = !enabled;
        document.getElementById('tarFile').disabled = !enabled;
        document.getElementById('sumFile').disabled = !enabled;
        document.querySelector('button[onclick*="hideModal"]').disabled = !enabled;
    }

    function showProgressUI(show) {
        const method = show ? 'remove' : 'add';
        document.getElementById('tarProgressContainer').classList[method]('hidden');
        document.getElementById('sumProgressContainer').classList[method]('hidden');
    }

    function updateProgress(type, loaded, total) {
        const percent = Math.round((loaded / total) * 100);
        document.getElementById(`${type}Percent`).innerText = `${percent}%`;
        document.getElementById(`${type}ProgressBar`).style.width = `${percent}%`;
        
        let status = "上传中...";
        if (loaded >= total) status = "已上传，等待验证";
        document.getElementById(`${type}Status`).innerText = status;
    }

    function updateButtons(state) {
        // states: idle, uploading, paused, finalizing
        submitBtn.classList.add('hidden');
        pauseBtn.classList.add('hidden');
        resumeBtn.classList.add('hidden');
        abortBtn.classList.add('hidden');

        if (state === 'uploading') {
            pauseBtn.classList.remove('hidden');
            abortBtn.classList.remove('hidden');
        } else if (state === 'paused') {
            resumeBtn.classList.remove('hidden');
            abortBtn.classList.remove('hidden');
        } else if (state === 'finalizing') {
            // Show loading spinner on submit btn maybe?
            submitBtn.classList.remove('hidden');
            submitBtn.disabled = true;
            submitBtn.innerHTML = '<i class="fas fa-cog fa-spin mr-2"></i>验证保存中...';
        } else {
            submitBtn.classList.remove('hidden');
            submitBtn.disabled = false;
            submitBtn.innerHTML = '开始上传';
            checkFormValidity();
        }
    }

    function resetUI() {
        isUploading = false;
        isPaused = false;
        toggleInputs(true);
        showProgressUI(false);
        updateButtons('idle');
    }

    // --- Model Types Management ---
    window.manageModelTypes = () => {
        showModal('modelTypesModal');
        loadModelTypesList();
    };

    async function loadModelTypesList() {
        const list = document.getElementById('modelTypesList');
        if (!list) return;
        
        try {
            const res = await fetch('/models/api/model-types');
            const types = await res.json();
            
            list.innerHTML = '';
            types.forEach(t => {
                const tag = document.createElement('div');
                tag.className = 'flex items-center gap-2 bg-white border border-slate-200 px-3 py-1 rounded-full text-sm font-bold text-slate-700 shadow-sm';
                tag.innerHTML = `
                    ${t}
                    <button onclick="deleteModelType('${t}')" class="text-slate-400 hover:text-red-500 transition">
                        <i class="fas fa-times-circle"></i>
                    </button>
                `;
                list.appendChild(tag);
            });

            // Also update the dropdowns in modals
            const selects = ['modelType', 'editModelType'];
            selects.forEach(id => {
                const select = document.getElementById(id);
                if (select) {
                    const currentVal = select.value;
                    select.innerHTML = id === 'modelType' ? '<option value="" disabled selected>选择模型类型</option>' : '';
                    types.forEach(t => {
                        const opt = document.createElement('option');
                        opt.value = t;
                        opt.textContent = t.toUpperCase();
                        select.appendChild(opt);
                    });
                    if (types.includes(currentVal)) select.value = currentVal;
                }
            });
        } catch (e) {
            console.error("Failed to load model types:", e);
        }
    }

    window.addModelType = async () => {
        const input = document.getElementById('newModelType');
        const type = input.value.trim().toLowerCase();
        if (!type) return;

        try {
            const res = await fetch('/models/api/model-types', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ type })
            });
            if (res.ok) {
                input.value = '';
                loadModelTypesList();
            }
        } catch (e) {
            console.error("Failed to add model type:", e);
        }
    };

    window.deleteModelType = async (type) => {
        if (!confirm(`确定要删除类型 "${type}" 吗？`)) return;

        try {
            const res = await fetch(`/models/api/model-types/${type}`, {
                method: 'DELETE'
            });
            if (res.ok) {
                loadModelTypesList();
            }
        } catch (e) {
            console.error("Failed to delete model type:", e);
        }
    };

    // Load initial model types
    loadModelTypesList();

    window.editModel = (name, currentType) => {
        document.getElementById('editModelName').value = name;
        document.getElementById('displayModelName').innerText = name;
        document.getElementById('editModelType').value = currentType;
        showModal('editModelModal');
    };

    const editForm = document.getElementById('editModelForm');
    if (editForm) {
        editForm.addEventListener('submit', async (e) => {
            e.preventDefault();
            const name = document.getElementById('editModelName').value;
            const model_type = document.getElementById('editModelType').value;

            try {
                const response = await fetch(`/models/api/${name}`, {
                    method: 'PUT',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ name, model_type })
                });

                const result = await response.json();
                if (result.success) {
                    showNotification('模型配置已更新', 'success');
                    hideModal('editModelModal');
                    setTimeout(() => window.location.reload(), 1000);
                } else {
                    throw new Error(result.message || '更新失败');
                }
            } catch (error) {
                showNotification('更新失败: ' + error.message, 'error');
            }
        });
    }

    // Keep delete function
    window.deleteModel = async (name) => {
        if (!confirm(`确定要删除模型 "${name}" 吗？此操作不可撤销。`)) {
            return;
        }

        try {
            const response = await fetch(`/models/api/${name}`, {
                method: 'DELETE'
            });

            const result = await response.json();
            if (result.success) {
                showNotification('模型已删除', 'success');
                setTimeout(() => window.location.href = '/models', 1000);
            } else {
                throw new Error(result.message || '删除失败');
            }
        } catch (error) {
            showNotification('删除失败: ' + error.message, 'error');
        }
    };
});

function showToast(title, message, type) {
    if (window.showNotification) {
        window.showNotification(message, type);
    } else {
        alert(`${title}: ${message}`);
    }
}