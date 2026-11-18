import React, { memo } from 'react';
import { StyleSheet, View } from 'react-native';
import { Colors } from '../constants/colors';

interface Props {
    isCG: boolean;
    w: number;
    h: number;
}

// Two soft round glows, same treatment as OnboardingScreen's glowYellow/glowBlue —
// colored by research group so the roadmap still reads CG (ochre) vs EX (sage)
// at a glance, without the leaf/bird scene that used to sit behind the nodes.
function RoadmapBgDecorImpl({ isCG, w, h }: Props) {
    const colorPrimary = isCG ? Colors.ochre : Colors.primaryLight;
    const colorSecondary = isCG ? Colors.accentTint : Colors.primaryTint;

    return (
        <View style={[StyleSheet.absoluteFillObject, { width: w, height: h }]} pointerEvents="none">
            <View style={[styles.glowPrimary, { backgroundColor: colorSecondary }]} />
            <View style={[styles.glowSecondary, { backgroundColor: colorPrimary }]} />
        </View>
    );
}

const styles = StyleSheet.create({
    glowPrimary: {
        position: 'absolute',
        top: 20,
        right: -60,
        width: 180,
        height: 180,
        borderRadius: 90,
        opacity: 0.1,
    },
    glowSecondary: {
        position: 'absolute',
        bottom: 120,
        left: -80,
        width: 220,
        height: 220,
        borderRadius: 110,
        opacity: 0.12,
    },
});

export const RoadmapBgDecor = memo(RoadmapBgDecorImpl);
export default RoadmapBgDecor;
