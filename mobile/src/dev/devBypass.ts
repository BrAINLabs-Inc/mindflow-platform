/**
 * DEV-ONLY auth bypass + mock API, for previewing the UI without a backend or a
 * Supabase account (e.g. `make dev-mobile-preview`, which runs Expo for web).
 *
 * Enabled only when BOTH are true:
 *   - __DEV__ (a development bundle — Metro sets this to false in every release/EAS
 *     production build, so this whole module is inert and the branches using it
 *     are dead code there), and
 *   - EXPO_PUBLIC_DEV_AUTH_BYPASS=true (opt-in per run; never set it in EAS
 *     build profiles or any deployed environment).
 *
 * When enabled:
 *   - the session token is an in-memory placeholder (SecureStore isn't touched, so
 *     it also works on web where SecureStore isn't available);
 *   - Splash skips straight past login;
 *   - every request to `${API_URL}/api/*` is answered locally from fixtures below,
 *     and submissions mark that roadmap step completed for the rest of the session.
 *
 * EXPO_PUBLIC_DEV_GROUP=ex|cg|unassigned picks the study arm to preview (default ex).
 */
import { API_URL } from '../config/api';

export const DEV_AUTH_BYPASS: boolean =
    __DEV__ && process.env.EXPO_PUBLIC_DEV_AUTH_BYPASS === 'true';

export const DEV_BYPASS_TOKEN = 'dev-bypass-token';

const GROUP = process.env.EXPO_PUBLIC_DEV_GROUP === 'cg' ? 'cg'
    : process.env.EXPO_PUBLIC_DEV_GROUP === 'unassigned' ? '' : 'ex';

const DAY = 24 * 60 * 60 * 1000;
const isoAgo = (days: number) => new Date(Date.now() - days * DAY).toISOString();
const inDays = (days: number) => new Date(Date.now() + days * DAY).toISOString();

// Roadmap steps completed during this session (submissions add to it).
const completed = new Set<string>();
const RESET_DAYS: Record<string, number> = { daily: 1, weekly: 4, thrive: 14, stress: 30, mindful: 30 };
const status = (key: string) => ({
    completed: completed.has(key),
    nextReset: completed.has(key) ? inDays(RESET_DAYS[key]) : null,
});

let aboutMe: Record<string, unknown> = {
    id: 'dev-user', is_completed: true, university_id: 'IT21004512', education_level: 'Third Year',
    faculty: 'Faculty of Computing', major_field_of_study: 'Software Engineering', age: 21,
    living_situation: 'University Hostel / Dormitory', family_background: 'Close-knit family',
    cultural_background: 'Buddhism', hobbies_interests: 'Reading, Sports & Fitness',
    personal_goals: 'Sleep better during exams', why_mindflow: 'To manage exam stress',
};

function calendarEvents() {
    const d = (n: number) => new Date(Date.now() + n * DAY).toISOString().slice(0, 10);
    const events = [
        { id: 1, title: 'Mindfulness Session: Body Scan', description: 'Guided group session', event_date: d(0), event_time: '17:30:00', is_completed: false },
        { id: 2, title: 'Weekly Check-in', description: null, event_date: d(2), event_time: '10:00:00', is_completed: false },
        { id: 3, title: 'Study Info Session', description: 'Faculty auditorium', event_date: d(-3), event_time: null, is_completed: true },
        { id: 4, title: 'Mindfulness Session: Breathing', description: 'Online', event_date: d(6), event_time: '18:00:00', is_completed: false },
    ].map((e) => ({ ...e, created_at: isoAgo(7), updated_at: isoAgo(7) }));
    // Same rule as the backend: the control group never sees mindfulness sessions.
    return GROUP === 'cg' ? events.filter((e) => !e.title.startsWith('Mindfulness Session')) : events;
}

function respond(method: string, path: string, body: unknown): unknown {
    const session = { access_token: DEV_BYPASS_TOKEN, refresh_token: DEV_BYPASS_TOKEN };
    const submit = path.match(/^\/api\/roadmap\/(daily|weekly|thrive|stress|mindful)$/);
    if (method === 'POST' && submit) {
        completed.add(submit[1]);
        return { id: Date.now(), success: true };
    }
    if (method === 'POST') {
        switch (path) {
            case '/api/auth/login':
            case '/api/auth/verify-otp':
            case '/api/auth/refresh':
                return { message: 'ok', session, user: { id: 'dev-user', display_name: 'Amaya' } };
            case '/api/profile/about-me':
                aboutMe = { ...aboutMe, ...(body as object) };
                return aboutMe;
            case '/api/roadmap/weekly/upload': {
                const key = `WeeklyVoice/weekly-dev-${Date.now()}.wav`;
                return { fileKey: key, fileUrl: `https://example.invalid/${key}` };
            }
            case '/api/roadmap/daily/video-progress':
                return { success: true, video_play_seconds: 140 };
            default:
                return { success: true, message: 'ok' };
        }
    }

    const done = { dailyDone: completed.has('daily'), weeklyDone: completed.has('weekly'), monthlyDone: false };
    switch (path) {
        case '/api/profile':
            return { username: 'Amaya', research_id: GROUP ? `MF-DEV-001.${GROUP}` : null, email: 'amaya@example.com' };
        case '/api/profile/about-me':
            return aboutMe;
        case '/api/dashboard/summary':
            return { streak: 6, consistency: 73, weeklyProgress: 31, totalCompleted: 22, status: done, group: GROUP };
        case '/api/journey/status':
            return {
                daily: { ...status('daily'), videoPlaySeconds: 140, group: GROUP },
                weekly: { ...status('weekly'), week: 47, year: 2026 },
                thrive: status('thrive'), stress: status('stress'), mindful: status('mindful'),
            };
        case '/api/roadmap/daily/status':
            return { ...status('daily'), videoPlaySeconds: 140, group: GROUP };
        case '/api/roadmap/weekly/status':
            return { ...status('weekly'), week: 47, year: 2026 };
        case '/api/roadmap/thrive/status':
            return status('thrive');
        case '/api/roadmap/stress/status':
            return status('stress');
        case '/api/roadmap/mindful/status':
            return status('mindful');
        case '/api/roadmap/weekly/video':
            return { id: 1, week_no: 47, title: 'Noticing the Breath', youtube_id: 'inpok4MKVLM', description: 'A short guided practice.', target_group: GROUP || null };
        case '/api/calendar/events':
        case '/api/roadmap/calendar/events':
            return calendarEvents();
        case '/api/journey':
            return {
                daily: Array.from({ length: 18 }, (_, i) => ({
                    id: i, stress_level: 1 + ((i * 7) % 5), sleep_quality: 1 + ((i * 3 + 2) % 5),
                    calm_before: 2 + (i % 3), calm_after: 3 + (i % 3), created_at: isoAgo(i),
                })),
                weekly: Array.from({ length: 5 }, (_, i) => ({ id: i, week_number: 47 - i, year: 2026, duration: 30 + i, created_at: isoAgo(i * 7) })),
                research: {
                    pss10: [{ id: 1, created_at: isoAgo(9) }],
                    ffmq15: [{ id: 1, created_at: isoAgo(11) }],
                    wemwbs14: [{ id: 1, created_at: isoAgo(5) }, { id: 2, created_at: isoAgo(19) }],
                },
            };
        default:
            return {};
    }
}

let installed = false;

/** Routes `${API_URL}/api/*` requests to the fixtures above. No-op unless enabled. */
export function installDevBypass(): void {
    if (!DEV_AUTH_BYPASS || installed) return;
    installed = true;
    console.warn('[MindFlow] DEV AUTH BYPASS is ON: no real login, all API data is mocked.');

    const realFetch = globalThis.fetch.bind(globalThis);
    globalThis.fetch = async (input: RequestInfo | URL, init?: RequestInit) => {
        const href = typeof input === 'string' ? input : input instanceof URL ? input.href : input.url;
        if (!href.startsWith(`${API_URL}/api/`)) return realFetch(input, init);

        const url = new URL(href);
        const method = (init?.method ?? 'GET').toUpperCase();
        let body: unknown = undefined;
        if (typeof init?.body === 'string') {
            try { body = JSON.parse(init.body); } catch { body = undefined; }
        }
        // A short delay so loading states are visible, as with a real network.
        await new Promise((resolve) => setTimeout(resolve, 250));
        return new Response(JSON.stringify(respond(method, url.pathname, body)), {
            status: 200,
            headers: { 'Content-Type': 'application/json' },
        });
    };
}
