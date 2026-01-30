/**
 * @jest-environment jsdom
 */
const fs = require('fs');
const path = require('path');
const vm = require('vm');

describe('Dashboard UI Logic', () => {

  beforeEach(() => {
    // Reset DOM
    document.body.innerHTML = `
      <select id="optimization-mode">
        <option value="balanced">Balanced</option>
      </select>
      <div id="vllm-suggestions" class="hidden">
        <div id="suggestion-gpu"></div>
        <div id="suggestion-content"></div>
      </div>
      <div id="hardware-acceleration-display"></div>
    `;

    // Mock fetch
    global.fetch = jest.fn((url) => {
      if (url.includes('/api/v1/configs/vllm-calculate')) {
        return Promise.resolve({
          ok: true,
          json: () => Promise.resolve({
            vllm_config: {
              max_model_len: 4096,
              max_num_seqs: 32,
              max_num_batched_tokens: 32768,
              gpu_memory_util: 0.9
            },
            model_config: { Name: 'test-model' },
            gpu_memory: 24
          }),
        });
      }
      return Promise.resolve({ ok: true, json: () => Promise.resolve({}) });
    });

    // Mock window.servicesData
    window.servicesData = [
      { type: 'Agent', node_ip: '1.1.1.1', status: 'Running', message: 'NVIDIA GPU 24GB' }
    ];

    // Read and execute the script in the jsdom context
    const scriptPath = path.resolve(__dirname, '../../../frontend/public/js/dashboard.js');
    const scriptContent = fs.readFileSync(scriptPath, 'utf8');
    
    // We need to provide the DOM environment to the script
    const context = {
      window: window,
      document: document,
      navigator: navigator,
      console: console,
      fetch: global.fetch,
      setTimeout: setTimeout,
      setInterval: setInterval,
      clearInterval: clearInterval,
      confirm: jest.fn(() => true),
      location: { reload: jest.fn() },
      alert: jest.fn()
    };
    
    vm.runInNewContext(scriptContent, context);
    
    // Assign the functions to global for the test to access
    global.calculateVllmSuggestions = context.calculateVllmSuggestions;
  });

  test('calculateVllmSuggestions should include parameter hints in the output', async () => {
    const calculateVllmSuggestions = global.calculateVllmSuggestions;
    
    expect(typeof calculateVllmSuggestions).toBe('function');
    
    await calculateVllmSuggestions('balanced');

    const content = document.getElementById('suggestion-content').innerHTML;
    
    expect(content).toContain('(context windows per request)');
    expect(content).toContain('(batch size / concurrency)');
    expect(content).toContain('(total context windows)');
    expect(content).toContain('(GPU HBM used for LLM)');
    expect(content).toContain('--max-model-len');
    expect(content).toContain('4096');
  });
});