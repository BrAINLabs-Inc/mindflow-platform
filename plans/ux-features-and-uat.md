# UI/UX Improvements, Feature Roadmap & UAT Plan

_Written 2026-11-18, after the mobile "Bright & Friendly" redesign (branch `feat/mobile-uiux-refresh`). Related: `plans/cutover-and-cleanup.md` (backend cutover) and `FINDINGS_AND_DEVELOPMENT.md` (audit)._

Contents:
1. [UI/UX improvements](#1-uiux-improvements)
2. [Feature roadmap](#2-feature-roadmap)
3. [UAT plan](#3-uat-user-acceptance-testing-plan)

---

## 1. UI/UX improvements

### 1.1 Quick wins (hours each)

- [ ] **Button text contrast (accessibility; do first).** The bright palette puts white text on vivid colours that fail WCAG AA for normal-size text: orange `#FF8A3D` is about 2.3:1 and green `#14A06F` about 3.3:1 (AA needs 4.5:1 at 16px). Add deeper "on-button" shades per colour in `mobile/src/constants/colors.ts` (for example orange `#D9661A`, green `#0E7A57`, teal `#0E8F81`, indigo `#4B43D9`, coral `#D23E37`) and use them only on text-bearing button backgrounds. Icons, tints and charts keep the bright values.
- [ ] **Loading states.** Replace full-screen spinners with grey placeholder shapes ("skeletons") on Dashboard, Journey and Calendar, so the app feels faster and doesn't jump.
- [ ] **Error states.** Many failures only `console.log`. Show a friendly message with a **Try again** button (Dashboard, Journey, Calendar, Profile, and every submit).
- [ ] **Locked roadmap steps.** Show the unlock time on the node itself ("Unlocks Mon", "Next: Dec 18") instead of only in a popup after tapping.
- [ ] **Screen-reader labels.** Add `accessibilityLabel`/`accessibilityRole` to the emoji, scale and option buttons (e.g. "Stress level 5 of 5, highly stressed"). Check every screen at the largest system font size.
- [ ] **Celebration and feedback.** A light tick or confetti animation on the completion screen, and a small haptic tap when choosing an answer (`expo-haptics`; check whether it's installed first).

### 1.2 Medium

- [ ] **One illustration style.** The PNGs in `mobile/assets/` (sign-in, sign-up, About Me, "Great Job") come from different art styles. Commission or pick one consistent set, or redraw them as simple SVG scenes like the dashboard's `CalmScene`. Control-group screens must stay free of mindfulness imagery.
- [ ] **Journey charts.** Add a Week / Month / All toggle; today it's only the last 7 entries. Add short legends explaining the 1–5 scales.
- [ ] **Deep links from reminders.** Tapping the 19:00 "pending tasks" push should open the first pending activity, not just the app.
- [ ] **Dark mode.** Mostly a second set of values in `colors.ts`, plus a `useColorScheme()` switch.
- [ ] **Web app parity.** Bring `web-app/` onto the same design system. Fix its About Me focus bug and add login refresh at the same time (see 2.1).

---

## 2. Feature roadmap

**Study-design check:** anything that changes what participants *see or receive* can act as an intervention. Items marked ⚠️ need research-team approval before building, and may need to be experimental-group (`.ex`) only.

### 2.1 Must-do

- [ ] **Web-app login refresh (bug).** Web participants are logged out about every hour. Port mobile's `apiFetch` refresh logic into `web-app/src/lib/api.ts` and `auth.ts`.
- [ ] **Consent, withdrawal, and "delete / export my data".** A research-ethics requirement; no such flow exists in the app today. It needs a consent screen at signup, a withdraw option in Profile, and a researcher-side process.
- [ ] **Offline saving of submissions.** Queue check-ins locally when offline and send them when back online, so check-ins aren't lost on campus Wi-Fi. It must not create duplicates; the backend's one-per-day upsert already protects against that.

### 2.2 Should-do

- [ ] **Sinhala and Tamil language support.** Use only *validated* translations of PSS-10, WEMWBS-14 and FFMQ-15, not ad-hoc translations.
- [ ] **Researcher "participant timeline"** in web admin: one page showing a participant's daily, weekly and questionnaire history.
- [ ] **Drop-out alerts for researchers:** list participants with no entry in N days.
- [ ] **Add calendar sessions to the phone's calendar.**
- [ ] ⚠️ **Choose your own reminder time** (changes the reminder protocol).

### 2.3 Needs research approval

- [ ] ⚠️ **Show streak and consistency on the dashboard.** The API already returns `streak`/`consistency`, but showing them is feedback that may change behaviour.
- [ ] ⚠️ **Personal trend insights** (e.g. "your stress is lower on days you practise").
- [ ] ⚠️ **Guided breathing timer.** This is an intervention: `.ex` only, or not at all.

---

## 3. UAT (User Acceptance Testing) plan

### 3.1 Scope and environment

- **Backend:** the Go API (`backend-v2`) on **staging**, with a staging Supabase project and staging R2. Not production. (Prerequisite: `plans/cutover-and-cleanup.md` phases 1–2.)
- **Test data:** `make seed` (creates `seed.participant@mindflow.app` / `seed.admin@mindflow.app`), plus fresh accounts created during testing.
- **Devices:**

  | Platform | Minimum |
  |---|---|
  | Android | one low-end/older phone and one recent phone |
  | iOS | one iPhone (current iOS) |
  | Web | Chrome (participant web app and web admin) |

- **Not valid for UAT:** the dev preview (`make dev-mobile-preview`). It bypasses login and uses mock data, so it's only for visual review.

### 3.2 Testers and roles

| Role | How to set up |
|---|---|
| Experimental participant | New account → researcher assigns a `….ex` Research ID |
| Control participant | New account → researcher assigns a `….cg` Research ID |
| Unassigned participant | New account, no Research ID |
| Researcher / admin | `seed.admin@mindflow.app` in web admin |

### 3.3 Entry criteria

- [ ] `database/patches/2026-11-18-security-hardening.sql` applied to staging
- [ ] Go API deployed to staging; `GET /health` returns 200
- [ ] CI green on the branch under test
- [ ] Test build of the mobile app (EAS preview profile) pointing at the staging API

### 3.4 Scenarios

Record a result for each scenario on each platform (Android / iOS / Web): **Pass**, **Fail** (with a screenshot and defect ID), or **N/A**.

| # | Scenario | Steps | Expected result | Android | iOS | Web |
|---|---|---|---|---|---|---|
| 1 | Sign-up & OTP | Sign up with a new email, enter the OTP | Lands on **About Me**, not Home | | | |
| 2 | Partial About Me | Fill part of About Me, close the app, reopen | Answers kept; still routed to About Me | | | |
| 3 | Complete About Me | Fill every required field, submit | Home opens; "Waiting for Group Assignment" shown | | | |
| 4 | Assign experimental group | Researcher sets `….ex`; participant reopens the app | Quotes card, mindfulness roadmap and mindfulness calendar events visible | | | |
| 5 | Assign control group | Researcher sets `….cg`; participant reopens the app | Fun facts instead of quotes; **no** "Mindfulness Session" events; no mindfulness fields in Daily | | | |
| 6 | Daily, experimental, twice | Submit Daily, then submit again the same day | The second submit **updates** the first (one row in admin) | | | |
| 7 | Daily, control, twice | Submit Daily, then try again | The second is **blocked**: "Already submitted for today" | | | |
| 8 | Midnight reset | Submit at ~23:50 Sri Lanka time, then again after 00:00 | Unlocks at 00:00 **Sri Lanka time** (not UTC) | | | |
| 9 | Weekly voice recording | Record 15–45 s, upload, submit | Uploads; week shows completed; the file plays from the admin link | | | |
| 10 | Questionnaire lockouts | Submit Thrive, Stress, Mindful; then retry each | Locked with the next date: Thrive 14 days, Stress 30, Mindful 30 | | | |
| 11 | Session refresh | Stay idle for more than 1 hour, then use the app | Still signed in; no forced logout | | | |
| 12 | Sign out | Sign out | Back to Login; this phone receives no further reminders | | | |
| 13 | Offline | Airplane mode, then try to submit | Clear offline message; no crash; no duplicate after reconnecting | | | |
| 14 | Reminders | Leave tasks pending across 08:00 and 19:00 | **Exactly one** of each; the 19:00 one lists only pending tasks | | | |
| 15 | Password reset | Forgot password → code → new password | New password works; old one rejected | | | |
| 16 | Researcher export | Export each table to CSV in web admin | All of the participant's rows present; text starting with `=` opens as text in Excel | | | |
| 17 | Accessibility | Largest font size + TalkBack/VoiceOver, run 1–7 | Readable, no clipped text, every control announced | | | |
| 18 | Direct-write protection | (Technical tester) PATCH own `research_id` via the anon key | Rejected (error 42501); group unchanged | | | |

### 3.5 Defect severity

| Severity | Meaning | Example |
|---|---|---|
| Critical | Data loss, wrong study arm, security or privacy issue | Control user sees mindfulness content; data saved to the wrong user |
| High | A core flow is blocked, no workaround | Cannot submit Daily; cannot log in |
| Medium | Wrong behaviour, but a workaround exists | Wrong unlock date shown, but the lock works |
| Low | Cosmetic | Misaligned text, colour issue |

### 3.6 Exit criteria and sign-off

- [ ] No open **Critical** or **High** defects
- [ ] Scenarios 1–18 pass on Android, iOS and Web (where applicable)
- [ ] Research lead signs off on the study-integrity scenarios (**4, 5, 6, 7, 8, 16, 18**)
- [ ] Product owner signs off on usability (**1–3, 9–15, 17**)

| Role | Name | Date | Signature |
|---|---|---|---|
| Research lead | | | |
| Product owner | | | |
| Tech lead | | | |
