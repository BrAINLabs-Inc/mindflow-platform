// MindFlow "Sage & Sand" palette — calm, warm and low-contrast-on-purpose for a
// mindfulness study, while keeping body text at WCAG AA contrast on the
// backgrounds it's used on. One brand colour (deep sage) carries every primary
// action; the five roadmap activities each get a muted tone from the same
// family instead of five unrelated bright accents.
//
// Use these tokens (and typography/spacing/radius from ./theme) rather than
// inline hex values, so the whole app stays on one palette.

const sage = {
  900: '#2C4B3F',
  700: '#3F6B5A', // brand primary
  500: '#6E9583',
  300: '#A9C3B6',
  100: '#E3ECE6',
  50: '#F1F6F2',
};

const sand = {
  50: '#FAF7F2', // app background
  100: '#F3EDE3', // soft surface
  200: '#E8E1D6', // borders
  300: '#D9CFBF',
};

const ink = {
  900: '#26302C', // headings / body
  600: '#5F6B66', // secondary text
  400: '#97A09B', // muted text, inactive icons
  300: '#B9BFBB', // placeholders
};

export const Colors = {
  // Surfaces
  background: sand[50],
  surface: '#FFFFFF',
  surfaceMuted: sand[100],
  white: '#FFFFFF',

  // Brand
  primary: sage[700],
  primaryDark: sage[900],
  primarySoft: sage[500],
  primaryLight: sage[300],
  primaryTint: sage[100],
  primaryWash: sage[50],
  secondary: sage[300],
  accent: '#C47B5A', // soft clay: highlights, "today" markers
  accentTint: '#F6EADF',
  accentBlue: '#5A7D9A',

  // Text
  textPrimary: ink[900],
  textSecondary: ink[600],
  textMuted: ink[400],
  textPlaceholder: ink[300],
  iconMuted: ink[600],
  textOnPrimary: '#FFFFFF',

  // Lines
  border: sand[200],
  borderLight: '#EFE9DF',
  borderStrong: sand[300],

  // Muted secondary tones (ink + tint pairs) — one per kind of content
  blue: '#5A7D9A',
  blueTint: '#E5EDF3',
  plum: '#7D6F96',
  plumTint: '#ECE8F2',
  sea: '#4E8A86',
  seaTint: '#E2EFEC',
  ochre: '#B08A45',
  ochreTint: '#F4ECDA',

  // Functional (muted, not alarm-red)
  success: sage[500],
  successTint: sage[100],
  warning: '#B08A45',
  warningTint: '#F4ECDA',
  error: '#B85C4E',
  errorTint: '#F7E4DF',

  shadow: '#2B2620',
  overlay: 'rgba(38, 48, 44, 0.45)',

  // Roadmap activities: each step keeps its own identity, all from one family.
  tasks: {
    daily: { ink: sage[700], tint: sage[100] },
    weekly: { ink: '#5A7D9A', tint: '#E5EDF3' },
    thrive: { ink: '#B08A45', tint: '#F4ECDA' },
    stress: { ink: '#C47B5A', tint: '#F6EADF' },
    mindful: { ink: '#7D6F96', tint: '#ECE8F2' },
  },

  gradients: {
    primary: [sage[700], sage[900]] as [string, string],
    calm: [sage[50], sand[50]] as [string, string],
    blueMethods: ['#E5EDF3', '#D6E2EC'] as [string, string],
    darkCard: [sage[700], sage[900]] as [string, string],
  },
};

export type TaskKey = keyof typeof Colors.tasks;
