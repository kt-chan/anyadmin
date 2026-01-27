/**
 * @jest-environment jsdom
 */

global.alert = jest.fn();
global.fetch = jest.fn();

describe('Deployment RAG Functions', () => {
    beforeEach(() => {
        document.body.innerHTML = `
          <form id="wizard-form">
              <input type="checkbox" id="enable-rag" name="enable_rag" checked>
              <div id="rag-config">
                  <input type="text" name="rag_host" value="127.0.0.1">
                  <input type="number" name="rag_port" value="3000">
              </div>

              <input type="checkbox" id="enable-vectordb" name="enable_vectordb">
              <div id="vectordb-config" class="hidden">
                  <input type="text" name="vectordb_host" value="127.0.0.1">
              </div>

              <input type="checkbox" id="enable-parser" name="enable_parser">
              <div id="parser-config" class="hidden">
                  <input type="text" name="parser_host" value="127.0.0.1">
              </div>

              <!-- Need a button for event target resolution -->
              <button id="test-btn">Test</button>
          </form>
          <div id="step-1"></div>
          <button id="next-btn"></button>
        `;
        jest.resetModules();
        require('../../../frontend/public/js/deployment.js');
    });

    test('initial state matches requirements', () => {
        expect(document.getElementById('enable-rag').checked).toBe(true);
        expect(document.getElementById('rag-config').classList.contains('hidden')).toBe(false);

        expect(document.getElementById('enable-vectordb').checked).toBe(false);
        expect(document.getElementById('vectordb-config').classList.contains('hidden')).toBe(true);

        expect(document.getElementById('enable-parser').checked).toBe(false);
        expect(document.getElementById('parser-config').classList.contains('hidden')).toBe(true);
    });

    test('toggleSection shows/hides element', () => {
        const config = document.getElementById('rag-config');
        
        window.toggleSection('rag-config', true);
        expect(config.classList.contains('hidden')).toBe(false);

        window.toggleSection('rag-config', false);
        expect(config.classList.contains('hidden')).toBe(true);
    });

    test('testConnection sends correct payload for rag_app', async () => {
        global.fetch.mockResolvedValue({
            json: () => Promise.resolve({ status: 'success', message: 'ok' })
        });

        const btn = document.getElementById('test-btn');
        // Mock global event
        global.event = { target: btn };

        await window.testConnection('rag_app');

        expect(global.fetch).toHaveBeenCalledWith('/deployment/api/test-connection', expect.objectContaining({
            method: 'POST',
            body: JSON.stringify({
                type: 'rag_app',
                host: '127.0.0.1',
                port: '3000'
            })
        }));
    });
});
