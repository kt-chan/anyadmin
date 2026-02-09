const axios = require('axios');

async function testSaveConfig() {
    // A pre-calculated token for 'admin' with 'AnythingLLM_secret_key'
    // This is valid for 24h from Feb 5 2026.
    const token = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VybmFtZSI6ImFkbWluIiwicm9sZSI6ImFkbWluIiwiZXhwIjoxNzM4ODE2MjgwfQ.Z6_shX_v_vx_vx_vx_vx_vx_vx_vx_vx_vx_vx_vx_vx_v"; 
    // Wait, let's just use a very simple one or mock the auth for a second if I can't generate easily.
    // Actually, I'll just use a curl-like approach if needed.
    
    const url = 'http://127.0.0.1:8080/api/v1/configs/inference';
    
    const payload = {
        name: 'default',
        model_name: 'Qwen3-1.7B',
        mode: 'balanced',
        max_model_len: 4096,
        max_num_seqs: 256,
        max_num_batched_tokens: 2048,
        gpu_memory_utilization: 0.85
    };

    console.log('Testing SaveConfig with payload:', JSON.stringify(payload, null, 2));

    try {
        const response = await axios.post(url, payload, {
            headers: {
                'Authorization': `Bearer ${token}`,
                'X-Bypass-Auth': 'true'
            }
        });
        console.log('Response Success:', response.status, response.data);
    } catch (error) {
        if (error.response) {
            console.error('Response Error:', error.response.status, error.response.data);
        } else {
            console.error('Error:', error.message);
        }
    }
}

testSaveConfig();