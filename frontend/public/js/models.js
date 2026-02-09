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
                await initUploadSession('sum', sumFile);

                // 2. Start Upload Loop
                await processUploads(name);

            } catch (err) {
                console.error(err);
                showToast("上传失败", err.message, "error");
                resetUI();
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

    async function initUploadSession(type, file) {
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

    async function processUploads(modelName) {
        // Upload loop for both files. simpler to do one then the other or parallel?
        // Let's do parallel chunks for speed? No, simpler sequential or interleaved.
        // Let's just loop until both are done.

        while ((uploadState.tar.offset < uploadState.tar.total || uploadState.sum.offset < uploadState.sum.total) && !isPaused) {
            
            // Upload Tar Chunk
            if (uploadState.tar.offset < uploadState.tar.total) {
                await uploadChunk('tar');
            }

            // Upload Sum Chunk
            if (uploadState.sum.offset < uploadState.sum.total) {
                await uploadChunk('sum');
            }
        }

        if (!isPaused) {
            // Both done
            finalizeUpload(modelName);
        }
    }

    async function uploadChunk(type) {
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
            
            state.offset = data.offset; // Update offset from server response
            updateProgress(type, state.offset, state.total);

        } catch (e) {
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

    async function finalizeUpload(modelName) {
        updateButtons('finalizing');
        document.getElementById(`tarStatus`).innerText = "正在校验并保存...";
        document.getElementById(`sumStatus`).innerText = "正在校验并保存...";
        
        try {
            const res = await fetch('/models/api/finalize', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    model_name: modelName,
                    tar_upload_id: uploadState.tar.id,
                    checksum_upload_id: uploadState.sum.id
                })
            });

            const data = await res.json();
            
            if (res.ok) {
                showToast("成功", "模型上传并验证成功", "success");
                setTimeout(() => window.location.reload(), 1000);
            } else {
                throw new Error(data.error || "Finalization failed");
            }
        } catch (e) {
            showToast("验证失败", e.message, "error");
            resetUI();
        }
    }

    // --- UI Helpers ---

    function toggleInputs(disabled) {
        document.getElementById('modelName').disabled = disabled;
        document.getElementById('tarFile').disabled = disabled;
        document.getElementById('sumFile').disabled = disabled;
        document.querySelector('button[onclick*="hideModal"]').disabled = disabled;
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

        if (state === 'uploading') {
            pauseBtn.classList.remove('hidden');
        } else if (state === 'paused') {
            resumeBtn.classList.remove('hidden');
        } else if (state === 'finalizing') {
            // Show loading spinner on submit btn maybe?
            submitBtn.classList.remove('hidden');
            submitBtn.disabled = true;
            submitBtn.innerHTML = '<i class="fas fa-cog fa-spin mr-2"></i>验证解压中...';
        } else {
            submitBtn.classList.remove('hidden');
            submitBtn.disabled = false;
            submitBtn.innerHTML = '开始上传';
        }
    }

    function resetUI() {
        isUploading = false;
        isPaused = false;
        toggleInputs(false);
        showProgressUI(false);
        updateButtons('idle');
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
                showToast('成功', '模型已删除', 'success');
                setTimeout(() => window.location.reload(), 1000);
            } else {
                throw new Error(result.message || '删除失败');
            }
        } catch (error) {
            showToast('错误', error.message, 'error');
        }
    };
});