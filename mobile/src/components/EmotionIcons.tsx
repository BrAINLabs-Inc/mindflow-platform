import React from 'react';
import { Colors } from '../constants/colors';
import Svg, { Circle, Path, Ellipse } from 'react-native-svg';

interface EmotionIconProps {
    size?: number;
    color?: string;
}

// STRESS LEVEL ICONS (1-5, Low to High)

export const StressLevel1 = ({ size = 40 }: EmotionIconProps) => (
    <Svg width={size} height={size} viewBox="0 0 40 40" fill="none">
        <Circle cx="20" cy="20" r="18" fill={Colors.primaryTint} stroke={Colors.primary} strokeWidth="2" />
        <Circle cx="14" cy="16" r="2.5" fill={Colors.primary} />
        <Circle cx="26" cy="16" r="2.5" fill={Colors.primary} />
        <Path d="M12 25 Q20 32, 28 25" stroke={Colors.primary} strokeWidth="2.5" strokeLinecap="round" fill="none" />
    </Svg>
);

export const StressLevel2 = ({ size = 40 }: EmotionIconProps) => (
    <Svg width={size} height={size} viewBox="0 0 40 40" fill="none">
        <Circle cx="20" cy="20" r="18" fill={Colors.primaryTint} stroke={Colors.primarySoft} strokeWidth="2" />
        <Circle cx="14" cy="16" r="2.5" fill={Colors.primarySoft} />
        <Circle cx="26" cy="16" r="2.5" fill={Colors.primarySoft} />
        <Path d="M13 25 Q20 29, 27 25" stroke={Colors.primarySoft} strokeWidth="2.5" strokeLinecap="round" fill="none" />
    </Svg>
);

export const StressLevel3 = ({ size = 40 }: EmotionIconProps) => (
    <Svg width={size} height={size} viewBox="0 0 40 40" fill="none">
        <Circle cx="20" cy="20" r="18" fill={Colors.surfaceMuted} stroke={Colors.textMuted} strokeWidth="2" />
        <Circle cx="14" cy="16" r="2.5" fill={Colors.textMuted} />
        <Circle cx="26" cy="16" r="2.5" fill={Colors.textMuted} />
        <Path d="M13 26 L27 26" stroke={Colors.textMuted} strokeWidth="2.5" strokeLinecap="round" />
    </Svg>
);

export const StressLevel4 = ({ size = 40 }: EmotionIconProps) => (
    <Svg width={size} height={size} viewBox="0 0 40 40" fill="none">
        <Circle cx="20" cy="20" r="18" fill={Colors.ochreTint} stroke={Colors.ochre} strokeWidth="2" />
        <Circle cx="14" cy="16" r="2.5" fill={Colors.ochre} />
        <Circle cx="26" cy="16" r="2.5" fill={Colors.ochre} />
        <Path d="M13 28 Q20 24, 27 28" stroke={Colors.ochre} strokeWidth="2.5" strokeLinecap="round" fill="none" />
    </Svg>
);

export const StressLevel5 = ({ size = 40 }: EmotionIconProps) => (
    <Svg width={size} height={size} viewBox="0 0 40 40" fill="none">
        <Circle cx="20" cy="20" r="18" fill={Colors.accentTint} stroke={Colors.accent} strokeWidth="2" />
        <Ellipse cx="14" cy="15" rx="3" ry="4" fill={Colors.accent} />
        <Ellipse cx="26" cy="15" rx="3" ry="4" fill={Colors.accent} />
        <Path d="M12 29 Q20 22, 28 29" stroke={Colors.accent} strokeWidth="2.5" strokeLinecap="round" fill="none" />
        <Path d="M30 12 Q32 8, 34 12 Q32 14, 30 12" fill={Colors.blue} />
    </Svg>
);

// SLEEP QUALITY ICONS (1-5, Poor to Great)

export const SleepLevel1 = ({ size = 40 }: EmotionIconProps) => (
    <Svg width={size} height={size} viewBox="0 0 40 40" fill="none">
        <Circle cx="20" cy="20" r="18" fill={Colors.accentTint} stroke={Colors.accent} strokeWidth="2" />
        <Path d="M11 17 L17 15" stroke={Colors.accent} strokeWidth="2" strokeLinecap="round" />
        <Path d="M23 15 L29 17" stroke={Colors.accent} strokeWidth="2" strokeLinecap="round" />
        <Circle cx="14" cy="18" r="2" fill={Colors.accent} />
        <Circle cx="26" cy="18" r="2" fill={Colors.accent} />
        <Path d="M14 27 Q20 24, 26 27" stroke={Colors.accent} strokeWidth="2" strokeLinecap="round" fill="none" />
    </Svg>
);

export const SleepLevel2 = ({ size = 40 }: EmotionIconProps) => (
    <Svg width={size} height={size} viewBox="0 0 40 40" fill="none">
        <Circle cx="20" cy="20" r="18" fill={Colors.ochreTint} stroke={Colors.ochre} strokeWidth="2" />
        <Circle cx="14" cy="16" r="2.5" fill={Colors.ochre} />
        <Circle cx="26" cy="16" r="2.5" fill={Colors.ochre} />
        <Ellipse cx="20" cy="26" rx="4" ry="3" fill={Colors.ochre} />
    </Svg>
);

export const SleepLevel3 = ({ size = 40 }: EmotionIconProps) => (
    <Svg width={size} height={size} viewBox="0 0 40 40" fill="none">
        <Circle cx="20" cy="20" r="18" fill={Colors.surfaceMuted} stroke={Colors.textMuted} strokeWidth="2" />
        <Circle cx="14" cy="16" r="2.5" fill={Colors.textMuted} />
        <Circle cx="26" cy="16" r="2.5" fill={Colors.textMuted} />
        <Path d="M13 26 L27 26" stroke={Colors.textMuted} strokeWidth="2.5" strokeLinecap="round" />
    </Svg>
);

export const SleepLevel4 = ({ size = 40 }: EmotionIconProps) => (
    <Svg width={size} height={size} viewBox="0 0 40 40" fill="none">
        <Circle cx="20" cy="20" r="18" fill={Colors.primaryTint} stroke={Colors.primarySoft} strokeWidth="2" />
        <Circle cx="14" cy="16" r="2.5" fill={Colors.primarySoft} />
        <Circle cx="26" cy="16" r="2.5" fill={Colors.primarySoft} />
        <Path d="M13 24 Q20 28, 27 24" stroke={Colors.primarySoft} strokeWidth="2.5" strokeLinecap="round" fill="none" />
    </Svg>
);

export const SleepLevel5 = ({ size = 40 }: EmotionIconProps) => (
    <Svg width={size} height={size} viewBox="0 0 40 40" fill="none">
        <Circle cx="20" cy="20" r="18" fill={Colors.primaryTint} stroke={Colors.primary} strokeWidth="2" />
        <Path d="M11 16 Q14 14, 17 16" stroke={Colors.primary} strokeWidth="2" strokeLinecap="round" fill="none" />
        <Path d="M23 16 Q26 14, 29 16" stroke={Colors.primary} strokeWidth="2" strokeLinecap="round" fill="none" />
        <Path d="M13 24 Q20 30, 27 24" stroke={Colors.primary} strokeWidth="2.5" strokeLinecap="round" fill="none" />
        <Path d="M32 8 L33 10 L35 10 L33.5 11.5 L34 14 L32 12.5 L30 14 L30.5 11.5 L29 10 L31 10 Z" fill={Colors.ochre} />
    </Svg>
);

// RELAXATION LEVEL ICONS (1-5, Tense to Calm)

export const RelaxationLevel1 = ({ size = 40 }: EmotionIconProps) => (
    <Svg width={size} height={size} viewBox="0 0 40 40" fill="none">
        <Circle cx="20" cy="20" r="18" fill={Colors.accentTint} stroke={Colors.accent} strokeWidth="2" />
        <Path d="M10 12 L16 14" stroke={Colors.accent} strokeWidth="2" strokeLinecap="round" />
        <Path d="M24 14 L30 12" stroke={Colors.accent} strokeWidth="2" strokeLinecap="round" />
        <Circle cx="14" cy="17" r="2.5" fill={Colors.accent} />
        <Circle cx="26" cy="17" r="2.5" fill={Colors.accent} />
        <Path d="M14 28 Q20 23, 26 28" stroke={Colors.accent} strokeWidth="2" strokeLinecap="round" fill="none" />
    </Svg>
);

export const RelaxationLevel2 = ({ size = 40 }: EmotionIconProps) => (
    <Svg width={size} height={size} viewBox="0 0 40 40" fill="none">
        <Circle cx="20" cy="20" r="18" fill={Colors.ochreTint} stroke={Colors.ochre} strokeWidth="2" />
        <Circle cx="14" cy="16" r="2.5" fill={Colors.ochre} />
        <Circle cx="26" cy="16" r="2.5" fill={Colors.ochre} />
        <Path d="M13 27 Q20 24, 27 27" stroke={Colors.ochre} strokeWidth="2" strokeLinecap="round" fill="none" />
    </Svg>
);

export const RelaxationLevel3 = ({ size = 40 }: EmotionIconProps) => (
    <Svg width={size} height={size} viewBox="0 0 40 40" fill="none">
        <Circle cx="20" cy="20" r="18" fill={Colors.surfaceMuted} stroke={Colors.textMuted} strokeWidth="2" />
        <Circle cx="14" cy="16" r="2.5" fill={Colors.textMuted} />
        <Circle cx="26" cy="16" r="2.5" fill={Colors.textMuted} />
        <Path d="M13 26 L27 26" stroke={Colors.textMuted} strokeWidth="2.5" strokeLinecap="round" />
    </Svg>
);

export const RelaxationLevel4 = ({ size = 40 }: EmotionIconProps) => (
    <Svg width={size} height={size} viewBox="0 0 40 40" fill="none">
        <Circle cx="20" cy="20" r="18" fill={Colors.primaryTint} stroke={Colors.primarySoft} strokeWidth="2" />
        <Path d="M11 16 Q14 15, 17 16" stroke={Colors.primarySoft} strokeWidth="2" strokeLinecap="round" fill="none" />
        <Path d="M23 16 Q26 15, 29 16" stroke={Colors.primarySoft} strokeWidth="2" strokeLinecap="round" fill="none" />
        <Path d="M13 24 Q20 28, 27 24" stroke={Colors.primarySoft} strokeWidth="2.5" strokeLinecap="round" fill="none" />
    </Svg>
);

export const RelaxationLevel5 = ({ size = 40 }: EmotionIconProps) => (
    <Svg width={size} height={size} viewBox="0 0 40 40" fill="none">
        <Circle cx="20" cy="20" r="18" fill={Colors.primaryTint} stroke={Colors.primary} strokeWidth="2" />
        <Path d="M11 16 Q14 14, 17 16" stroke={Colors.primary} strokeWidth="2" strokeLinecap="round" fill="none" />
        <Path d="M23 16 Q26 14, 29 16" stroke={Colors.primary} strokeWidth="2" strokeLinecap="round" fill="none" />
        <Path d="M13 24 Q20 30, 27 24" stroke={Colors.primary} strokeWidth="2.5" strokeLinecap="round" fill="none" />
        <Path d="M6 20 Q8 18, 6 16" stroke={Colors.primary} strokeWidth="1" strokeLinecap="round" fill="none" opacity="0.5" />
        <Path d="M34 20 Q32 18, 34 16" stroke={Colors.primary} strokeWidth="1" strokeLinecap="round" fill="none" opacity="0.5" />
    </Svg>
);

// Export arrays for easy mapping
export const StressIcons = [StressLevel1, StressLevel2, StressLevel3, StressLevel4, StressLevel5];
export const SleepIcons = [SleepLevel1, SleepLevel2, SleepLevel3, SleepLevel4, SleepLevel5];
export const RelaxationIcons = [RelaxationLevel1, RelaxationLevel2, RelaxationLevel3, RelaxationLevel4, RelaxationLevel5];
