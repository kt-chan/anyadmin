/**
 * @jest-environment jsdom
 */

// Mock fetch
global.fetch = jest.fn(() => 
  Promise.resolve({
    json: () => Promise.resolve({ success: true, data: { data: [{id: 'local-model'}] } }),
    ok: true
  })
);

// Mock alert
global.alert = jest.fn();

describe('Deployment Wizard - Step 2 Enrichment', () => {
  beforeEach(() => {
    // Reset mocks
    fetch.mockClear();
    
    // Setup DOM
    document.body.innerHTML = `
      <form id="wizard-form">
        <!-- Step 1 (Hidden) -->
        <div id="step-1" class="step-content">
           <textarea name="target_nodes">1.1.1.1</textarea>
        </div>

        <!-- Step 2 (Active) -->
        <div id="step-2" class="step-content active">
           <input type="radio" name="mode" value="new_deployment" checked>
           <input type="radio" name="mode" value="integrate_existing">
           
           <div class="integration-only hidden">
             <button onclick="testConnection('inference')">Test</button>
           </div>

           <select id="inference-host-select">
             <option value="1.1.1.1">1.1.1.1</option>
           </select>
           <input id="inference-port-input" value="8000">
           
           <select id="model-select">
             <option value="">Select a Model</option>
           </select>
           
           <input type="radio" name="platform" value="nvidia" checked>
        </div>
        
        <button id="next-btn" type="button">Next</button>
      </form>
    `;

    // Load script
    jest.resetModules();
    require('../../../frontend/public/js/deployment.js');
    
    // Trigger init
    document.dispatchEvent(new Event('DOMContentLoaded'));
  });

  test('should fetch models automatically when in new_deployment mode', async () => {
    // Manually trigger refreshModels via mode change simulation or just checking if init triggered it
    // In my code: 
    // const defaultMode = document.querySelector('input[name="mode"]:checked');
    // if (defaultMode && defaultMode.value === 'new_deployment') { refreshModels(); }
    // This happens in initWizard.
    
    // Wait for async operations
    await new Promise(resolve => setTimeout(resolve, 0));

    expect(fetch).toHaveBeenCalledWith(
        '/deployment/api/discover-models',
        expect.objectContaining({
            body: expect.stringContaining('"mode":"new_deployment"')
        })
    );
  });

  test('should toggle integration-only section and clear models when switching to integrate_existing', async () => {
    // Wait for initial load
    await new Promise(resolve => setTimeout(resolve, 0));
    fetch.mockClear();

    const integrateRadio = document.querySelector('input[value="integrate_existing"]');
    integrateRadio.checked = true;
    integrateRadio.dispatchEvent(new Event('change'));

    // Check visibility
    const integrationOnly = document.querySelector('.integration-only');
    expect(integrationOnly.classList.contains('hidden')).toBe(false);

    // Check model select cleared
    const modelSelect = document.getElementById('model-select');
    // It should have reset to "Select a Model" (disabled selected)
    expect(modelSelect.innerHTML).toContain('Select a Model');
    
    // Check it didn't fetch automatically (since host might not be selected/verified or just strictly clear first)
    // My code says: 
    // if (!isIntegrate) { refreshModels(); } else { clear models }
    expect(fetch).not.toHaveBeenCalled();
  });

  test('should fetch with remote mode when refreshing in integrate_existing mode', async () => {
    const integrateRadio = document.querySelector('input[value="integrate_existing"]');
    integrateRadio.checked = true;
    integrateRadio.dispatchEvent(new Event('change'));
    
    // Manually trigger refresh (e.g. via button or calling function if accessible)
    // Since window.refreshModels is attached to window
    await window.refreshModels();

    expect(fetch).toHaveBeenCalledWith(
        '/deployment/api/discover-models',
        expect.objectContaining({
            body: expect.stringContaining('"mode":"integrate_existing"')
        })
    );
  });
});
