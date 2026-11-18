// MindFlow "Bright & Friendly" palette — energetic, saturated colours on clean
// white surfaces (in the spirit of the original MindFlow look), organised as
// tokens so the whole app stays consistent. Body text stays dark navy for
// WCAG AA contrast; the colour lives in accents, icons, buttons and panels.
//
// Use these tokens (and typography/spacing/radius from ./theme) rather than
// inline hex values.

const green = {
  900: '#0E7A57',
  700: '#14A06F', // brand primary: fresh green
  500: '#3CC28F',
  300: '#A3E6CB',
  100: '#E1F7EE',
  50: '#F1FCF7',
};

const ink = {
  900: '#1E2A3A', // headings / body (navy)
  600: '#5A6578', // secondary text
  400: '#96A0B2', // muted text, inactive icons
  300: '#BCC3D0', // placeholders
};

export const Colors = {
  // Surfaces
  background: '#F7F8FC',
  surface: '#FFFFFF',
  surfaceMuted: '#F1F3F9',
  white: '#FFFFFF',

  // Brand
  primary: green[700],
  primaryDark: green[900],
  primarySoft: green[500],
  primaryLight: green[300],
  primaryTint: green[100],
  primaryWash: green[50],
  secondary: green[300],
  accent: '#FF8A3D', // vivid orange: highlights, "today" markers
  accentTint: '#FFF1E5',
  accentBlue: '#3B8FF0',

  // Text
  textPrimary: ink[900],
  textSecondary: ink[600],
  textMuted: ink[400],
  textPlaceholder: ink[300],
  iconMuted: ink[600],
  textOnPrimary: '#FFFFFF',

  // Lines
  border: '#E4E8F1',
  borderLight: '#EEF1F7',
  borderStrong: '#D3D9E6',

  // Bright secondary tones (ink + tint pairs)
  blue: '#3B8FF0',
  blueTint: '#E8F2FE',
  plum: '#6C63FF', // indigo
  plumTint: '#EEEDFF',
  sea: '#14B8A6', // teal
  seaTint: '#DFF7F3',
  ochre: '#F5B523', // sunny yellow
  ochreTint: '#FFF6DB',

  // Functional
  success: green[500],
  successTint: green[100],
  warning: '#F5B523',
  warningTint: '#FFF6DB',
  error: '#F25C54', // coral red
  errorTint: '#FDECEB',

  shadow: '#1E2A3A',
  overlay: 'rgba(30, 42, 58, 0.45)',

  // Roadmap activities: each step has its own bright identity colour.
  tasks: {
    daily: { ink: '#FF8A3D', tint: '#FFF1E5' },   // orange
    weekly: { ink: '#6C63FF', tint: '#EEEDFF' },  // indigo
    thrive: { ink: green[700], tint: green[100] }, // green
    stress: { ink: '#F25C54', tint: '#FDECEB' },  // coral
    mindful: { ink: '#14B8A6', tint: '#DFF7F3' }, // teal
  },

  gradients: {
    primary: [green[500], green[700]] as [string, string],
    calm: ['#E8F2FE', '#FFF6DB'] as [string, string],
    blueMethods: ['#E8F2FE', '#D6E7FD'] as [string, string],
    darkCard: [green[700], green[900]] as [string, string],
  },
};

export type TaskKey = keyof typeof Colors.tasks;
