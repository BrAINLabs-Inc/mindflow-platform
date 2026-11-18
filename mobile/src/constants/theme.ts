import { TextStyle } from 'react-native';
import { Colors } from './colors';

// Spacing on a 4pt grid.
export const Spacing = {
  xs: 4,
  sm: 8,
  md: 12,
  lg: 16,
  xl: 20,
  xxl: 24,
  xxxl: 32,
  screen: 20, // horizontal screen gutter
} as const;

export const Radius = {
  sm: 10,
  md: 14,
  lg: 20,
  xl: 28,
  pill: 999,
} as const;

// Sentence-case, softly weighted type: calm and readable rather than shouty
// all-caps headings. System font everywhere (no extra font downloads).
export const Typography: Record<string, TextStyle> = {
  display: { fontSize: 30, lineHeight: 38, fontWeight: '700', color: Colors.textPrimary, letterSpacing: -0.4 },
  title: { fontSize: 24, lineHeight: 31, fontWeight: '700', color: Colors.textPrimary, letterSpacing: -0.3 },
  heading: { fontSize: 19, lineHeight: 26, fontWeight: '600', color: Colors.textPrimary, letterSpacing: -0.2 },
  subheading: { fontSize: 16, lineHeight: 22, fontWeight: '600', color: Colors.textPrimary },
  body: { fontSize: 15, lineHeight: 22, fontWeight: '400', color: Colors.textSecondary },
  bodyStrong: { fontSize: 15, lineHeight: 22, fontWeight: '600', color: Colors.textPrimary },
  caption: { fontSize: 13, lineHeight: 18, fontWeight: '500', color: Colors.textMuted },
  // Small label above a section: muted, lightly spaced, NOT all-caps shouting.
  eyebrow: { fontSize: 13, lineHeight: 18, fontWeight: '600', color: Colors.primarySoft, letterSpacing: 0.3 },
  button: { fontSize: 16, lineHeight: 22, fontWeight: '600', letterSpacing: 0.2 },
};
