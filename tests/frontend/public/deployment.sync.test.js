/**
 * @jest-environment jsdom
 */

global.alert = jest.fn();
global.fetch = jest.fn();

// Mock requestSubmit which is missing in JSDOM
if (!HTMLFormElement.prototype.requestSubmit) {
    HTMLFormElement.prototype.requestSubmit = function() {
        this.dispatchEvent(new Event('submit', { bubbles: true, cancelable: true }));
    };
}

describe('Deployment Wizard Sync and Population', () => {
    beforeEach(() => {
        document.body.innerHTML = `
          <form id="wizard-form">
              <div id="step-1" class="step-content active">
                  <textarea name="target_nodes">192.168.1.100\n192.168.1.101:2222</textarea>
              </div>
              <div id="step-3" class="step-content">
                  <select id="inference-host-select" name="inference_host" class="node-selector">
                      <option value="">Select Target Node</option>
                  </select>
                  <input id="inference-port-input" type="number" name="inference_port" value="8000">
                  <select id="model-select" name="model_path">
                      <option value="">Select a Model</option>
                  </select>
              </div>
              <div id="step-4" class="step-content">
                  <select name="rag_host" class="node-selector">
                      <option value="">Select Target Node</option>
                  </select>
              </div>
              <div id="step-5" class="step-content">
                  <ul id="config-summary"></ul>
                  <button id="generate-btn" onclick="generateDeployment()">Deploy</button>
              </div>
              <button id="next-btn">Next</button>
          </form>
        `;
        
        // Mock fetch response for GET /deployment/api/nodes
        global.fetch.mockImplementation((url) => {
            if (url === '/deployment/api/nodes') {
                return Promise.resolve({
                    json: () => Promise.resolve({ success: true, data: ['172.20.0.10'] })
                });
            }
            return Promise.resolve({
                json: () => Promise.resolve({ success: true, data: {} })
            });
        });

        jest.resetModules();
        require('../../../frontend/public/js/deployment.js');
    });

    test('fetchNodesAndPopulate includes nodes from textarea', async () => {
        // Trigger population
        await window.fetchNodesAndPopulate();

        const selectors = document.querySelectorAll('select.node-selector');
        selectors.forEach(sel => {
            const options = Array.from(sel.options).map(opt => opt.value);
            // From API
            expect(options).toContain('172.20.0.10');
            // From textarea
            expect(options).toContain('192.168.1.100');
            expect(options).toContain('192.168.1.101');
        });
    });

    test('Step 1 Next button does NOT call saveNodes', async () => {
        const nextBtn = document.getElementById('next-btn');
        
        // Reset fetch mocks to track calls
        global.fetch.mockClear();
        global.fetch.mockImplementation(() => Promise.resolve({
            json: () => Promise.resolve({ success: true, data: {} })
        }));

        nextBtn.click();
        
        // Should NOT have called /deployment/api/nodes with POST
        const saveNodesCall = global.fetch.mock.calls.find(call => 
            call[0] === '/deployment/api/nodes' && call[1]?.method === 'POST'
        );
        expect(saveNodesCall).toBeUndefined();
    });

    test('Final Deployment calls generate with full config', async () => {
        global.fetch.mockResolvedValue({
            json: () => Promise.resolve({ success: true, data: { artifacts: {} } })
        });

        // Set some values
        document.querySelector('textarea[name="target_nodes"]').value = "10.0.0.1";
        const hostSel = document.getElementById('inference-host-select');
        
        // Need to populate first so we can select
        await window.fetchNodesAndPopulate();
        hostSel.value = "10.0.0.1";
        
        await window.generateDeployment();

        expect(global.fetch).toHaveBeenCalledWith('/deployment/api/generate', expect.objectContaining({
            method: 'POST',
            body: expect.stringContaining('"target_nodes":"10.0.0.1"')
        }));
    });
});
