/* eslint-disable @typescript-eslint/no-require-imports -- jest.isolateModules needs a fresh require() of the module per scenario. */
describe('dev auth bypass', () => {
    const original = process.env.EXPO_PUBLIC_DEV_AUTH_BYPASS;
    afterEach(() => {
        process.env.EXPO_PUBLIC_DEV_AUTH_BYPASS = original;
        jest.resetModules();
    });

    it('is off unless explicitly enabled', () => {
        delete process.env.EXPO_PUBLIC_DEV_AUTH_BYPASS;
        jest.isolateModules(() => {
            expect(require('../devBypass').DEV_AUTH_BYPASS).toBe(false);
        });
    });

    it('is off outside development bundles even when the flag is set', () => {
        process.env.EXPO_PUBLIC_DEV_AUTH_BYPASS = 'true';
        const g = global as unknown as { __DEV__: boolean };
        const dev = g.__DEV__;
        g.__DEV__ = false;
        try {
            jest.isolateModules(() => {
                expect(require('../devBypass').DEV_AUTH_BYPASS).toBe(false);
            });
        } finally {
            g.__DEV__ = dev;
        }
    });

    it('serves mock API data and remembers completed steps when enabled', async () => {
        process.env.EXPO_PUBLIC_DEV_AUTH_BYPASS = 'true';
        const realFetch = global.fetch;
        global.fetch = jest.fn(async () => new Response('real')) as unknown as typeof fetch;
        try {
            await jest.isolateModulesAsync(async () => {
                const { installDevBypass, DEV_AUTH_BYPASS } = require('../devBypass');
                const { API_URL } = require('../../config/api');
                expect(DEV_AUTH_BYPASS).toBe(true);
                installDevBypass();

                const before = await (await fetch(`${API_URL}/api/journey/status`)).json();
                expect(before.daily.completed).toBe(false);
                await fetch(`${API_URL}/api/roadmap/daily`, { method: 'POST', body: '{}' });
                const after = await (await fetch(`${API_URL}/api/journey/status`)).json();
                expect(after.daily.completed).toBe(true);

                // Non-API requests still go to the real network.
                expect(await (await fetch('https://example.com/x')).text()).toBe('real');
            });
        } finally {
            global.fetch = realFetch;
        }
    });
});
