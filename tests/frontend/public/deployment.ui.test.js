/**
 * @jest-environment jsdom
 */

// Mock alert
global.alert = jest.fn();

// Mock fetch
global.fetch = jest.fn((url, options) => {
    let body = {};
    if (url.includes('detect-hardware')) {
        body = { status: 'success', platform: 'ascend', details: 'Detected Ascend' };
    } else {
        // Default success for others (test-connection, nodes, etc)
        body = { status: 'success', success: true, data: [] };
    }
    return Promise.resolve({
        json: () => Promise.resolve(body),
    });
});

describe('Deployment Wizard UI Logic', () => {
  let initWizard;

  beforeEach(() => {
    // Reset DOM with SWAPPED steps
    document.body.innerHTML = `
      <form id="wizard-form">
        <!-- Step 1: Basic Config (Inputs) -->
        <div id="step-1" class="step-content active">
           <input type="text" name="mgmt_host" value="1.1.1.1">
           <input type="number" name="mgmt_port" value="3000">
           <textarea name="target_nodes">1.1.1.1:22</textarea>
           
           <button id="verify-ssh-btn" onclick="testConnection('ssh')" type="button">Verify SSH</button>
        </div>
        
        <!-- Step 2: Mode/Hardware (Radios) -->
        <div id="step-2" class="step-content">
           <div id="hardware-detection-status"></div>
           <input type="radio" name="mode" value="new" checked>
           
           <!-- Initially disabled and none checked (or checked=false) for platform -->
           <input type="radio" name="platform" value="nvidia" disabled>
           <input type="radio" name="platform" value="ascend" disabled>
        </div>
        
        <div id="step-3" class="step-content"></div>
        <div id="step-4" class="step-content"></div>
        <div id="step-5" class="step-content">
             <ul id="config-summary"></ul>
        </div>
        
        <div class="step-indicator" data-step="1"><div></div></div>
        <div class="step-indicator" data-step="2"><div></div></div>

        <button id="prev-btn" class="hidden" type="button">Prev</button>
        <button id="next-btn" disabled type="button">Next</button>
      </form>
    `;
    
    // Reload script
    jest.resetModules();
    require('../../../frontend/public/js/deployment.js'); 
    
    document.dispatchEvent(new Event('DOMContentLoaded'));
  });

  test('Step 1: Next button disabled initially', () => {
    const nextBtn = document.getElementById('next-btn');
    expect(nextBtn.disabled).toBe(true);
  });

  test('Step 1 -> Step 2: Detection triggers and auto-selects platform', async () => {
    const nextBtn = document.getElementById('next-btn');
    const verifyBtn = document.getElementById('verify-ssh-btn');
    
    // 1. Verify SSH to enable Next
    verifyBtn.click();
    await new Promise(resolve => setTimeout(resolve, 10));
    expect(nextBtn.disabled).toBe(false);

    // 2. Click Next to go to Step 2
    nextBtn.click();
    await new Promise(resolve => setTimeout(resolve, 10));

    // 3. Check if detectHardware was called (via fetch)
    expect(fetch).toHaveBeenCalledWith(
        expect.stringContaining('detect-hardware'),
        expect.anything()
    );

    // 4. Check if Platform Ascend is selected (as per mock)
    const ascendRadio = document.querySelector('input[value="ascend"]');
    expect(ascendRadio.checked).toBe(true);

    // 5. Check status message
    const statusEl = document.getElementById('hardware-detection-status');
    expect(statusEl.innerHTML).toContain('Detected Ascend');
  });
});