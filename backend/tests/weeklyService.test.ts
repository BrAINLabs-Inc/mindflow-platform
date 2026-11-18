import { createSupabaseMock } from './helpers/supabaseMock';

function mockFor(researchId: string | null, recordings: any[]) {
    return createSupabaseMock({
        profiles: { data: { research_id: researchId }, error: null },
        weekly_recordings: { data: recordings, error: null },
    });
}

describe('WeeklyService.getWeeklyVideo group selection', () => {
    it('returns the experimental-group video when the user is in .ex', async () => {
        jest.resetModules();
        jest.doMock('../src/config/supabase', () => ({
            supabase: mockFor('MF-2026-001.ex', [
                { id: 1, week_no: 26, youtube_id: 'ssss7V1_eyA', target_group: 'ex' },
                { id: 2, week_no: 26, youtube_id: 'ZSt9tm3RoUU', target_group: 'cg' },
            ]),
        }));
        const { WeeklyService } = require('../src/services/weeklyService');

        const video = await new WeeklyService().getWeeklyVideo('user-1');

        expect(video.youtube_id).toBe('ssss7V1_eyA');
    });

    it('returns the control-group video when the user is in .cg', async () => {
        jest.resetModules();
        jest.doMock('../src/config/supabase', () => ({
            supabase: mockFor('MF-2026-002.cg', [
                { id: 1, week_no: 26, youtube_id: 'ssss7V1_eyA', target_group: 'ex' },
                { id: 2, week_no: 26, youtube_id: 'ZSt9tm3RoUU', target_group: 'cg' },
            ]),
        }));
        const { WeeklyService } = require('../src/services/weeklyService');

        const video = await new WeeklyService().getWeeklyVideo('user-2');

        expect(video.youtube_id).toBe('ZSt9tm3RoUU');
    });

    it('falls back to a group-agnostic video when no group-specific match exists', async () => {
        jest.resetModules();
        jest.doMock('../src/config/supabase', () => ({
            supabase: mockFor(null, [
                { id: 1, week_no: 1, youtube_id: 'wAIoe992Qak', target_group: null },
            ]),
        }));
        const { WeeklyService } = require('../src/services/weeklyService');

        const video = await new WeeklyService().getWeeklyVideo('user-3');

        expect(video.youtube_id).toBe('wAIoe992Qak');
    });

    it('returns null when nothing is published for the week', async () => {
        jest.resetModules();
        jest.doMock('../src/config/supabase', () => ({
            supabase: mockFor('MF-2026-003.ex', []),
        }));
        const { WeeklyService } = require('../src/services/weeklyService');

        const video = await new WeeklyService().getWeeklyVideo('user-4');

        expect(video).toBeNull();
    });
});

describe('WeeklyService.submitWeeklyEntry file verification', () => {
    const userId = '11111111-1111-1111-1111-111111111111';

    function load() {
        jest.resetModules();
        const supabase = createSupabaseMock({
            voice_recordings: { data: { id: 1 }, error: null },
        });
        jest.doMock('../src/config/supabase', () => ({ supabase }));
        jest.doMock('../src/config/r2', () => ({
            r2: { send: jest.fn() },
            BUCKET_NAME: 'bucket',
            PUBLIC_URL_BASE: 'https://cdn.example.com',
        }));
        const { WeeklyService } = require('../src/services/weeklyService');
        const { getISOWeekNumber } = require('../src/utils/date');
        const [year, week] = getISOWeekNumber(new Date());
        const fileKey = `WeeklyVoice/weekly-${year}-W${String(week).padStart(2, '0')}-${userId}.wav`;
        return { service: new WeeklyService(), fileKey };
    }

    it('accepts the URL uploadAudio returned for this key', async () => {
        const { service, fileKey } = load();
        await expect(service.submitWeeklyEntry(userId, {
            file_key: fileKey,
            file_url: `https://cdn.example.com/${fileKey}`,
        })).resolves.toEqual({ id: 1 });
    });

    it.each([
        'javascript:alert(document.cookie)',
        'https://phish.example.net/recording.wav',
        'https://cdn.example.com/WeeklyVoice/other.wav',
    ])('rejects a file_url that does not match the key: %s', async (fileUrl) => {
        const { service, fileKey } = load();
        await expect(service.submitWeeklyEntry(userId, { file_key: fileKey, file_url: fileUrl }))
            .rejects.toThrow('WEEKLY_FILE_KEY_MISMATCH');
    });
});

describe('storedAudioContentType', () => {
    const { storedAudioContentType } = require('../src/middlewares/uploadMiddleware');

    it.each([
        ['audio/wav', 'rec.wav', 'audio/wav'],
        ['audio/webm', 'rec.webm', 'audio/webm'],
        ['text/html', 'x.wav', 'audio/wav'],
        ['image/svg+xml', 'x.mp3', 'audio/mpeg'],
        ['application/octet-stream', 'blob', 'application/octet-stream'],
    ])('declared %s + %s is stored as %s', (mime, name, expected) => {
        expect(storedAudioContentType(mime, name)).toBe(expected);
    });
});
