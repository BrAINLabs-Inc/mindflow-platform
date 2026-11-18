import React, { memo, useEffect, useRef } from 'react';
import { Animated, Easing, StyleSheet, View } from 'react-native';
import { LinearGradient } from 'expo-linear-gradient';
import Svg, { Circle, Path } from 'react-native-svg';
import { Colors } from '../constants/colors';

interface CalmSceneProps {
    width: number;
    height?: number;
    /** Warmer, ochre sun for the control group; sage-and-clay for experimental. */
    tone?: 'sage' | 'warm';
}

// A quiet landscape: a soft sun with slowly "breathing" halos over layered hills.
// Deliberately neutral (no meditating figure), so it suits both study arms. The
// control group must not see mindfulness-themed imagery (see CLAUDE.md).
function CalmSceneImpl({ width, height = 150, tone = 'sage' }: CalmSceneProps) {
    const breath = useRef(new Animated.Value(0)).current;

    useEffect(() => {
        // ~4s in, ~4s out: a slow breathing rhythm, not a busy loading pulse.
        const loop = Animated.loop(
            Animated.sequence([
                Animated.timing(breath, { toValue: 1, duration: 4000, easing: Easing.inOut(Easing.sin), useNativeDriver: true }),
                Animated.timing(breath, { toValue: 0, duration: 4000, easing: Easing.inOut(Easing.sin), useNativeDriver: true }),
            ]),
        );
        loop.start();
        return () => loop.stop();
    }, [breath]);

    const sun = tone === 'warm' ? Colors.ochre : Colors.accent;
    const sunTint = tone === 'warm' ? '#FFE9A8' : '#FFD9BD';
    const w = width;
    const h = height;
    const sunX = w * 0.68;
    const sunY = h * 0.42;
    const haloSize = h * 0.9;

    return (
        <View style={[styles.wrap, { width: w, height: h }]}>
            {/* Layers, back to front: sky, breathing halo, sun + hills. */}
            <LinearGradient colors={[Colors.blueTint, Colors.ochreTint]} style={StyleSheet.absoluteFill} />
            <Animated.View
                pointerEvents="none"
                style={[
                    styles.halo,
                    {
                        width: haloSize,
                        height: haloSize,
                        borderRadius: haloSize / 2,
                        left: sunX - haloSize / 2,
                        top: sunY - haloSize / 2,
                        backgroundColor: sunTint,
                        opacity: breath.interpolate({ inputRange: [0, 1], outputRange: [0.45, 0.85] }),
                        transform: [{ scale: breath.interpolate({ inputRange: [0, 1], outputRange: [0.6, 0.85] }) }],
                    },
                ]}
            />
            <Svg width={w} height={h} viewBox={`0 0 ${w} ${h}`} style={StyleSheet.absoluteFill}>
                <Circle cx={sunX} cy={sunY} r={h * 0.17} fill={sun} opacity={0.85} />
                {/* Far, middle and near hills, lightest to darkest. */}
                <Path
                    d={`M0 ${h * 0.7} C ${w * 0.18} ${h * 0.5}, ${w * 0.36} ${h * 0.52}, ${w * 0.52} ${h * 0.64} S ${w * 0.84} ${h * 0.5}, ${w} ${h * 0.6} L ${w} ${h} L 0 ${h} Z`}
                    fill={Colors.primaryLight}
                    opacity={0.6}
                />
                <Path
                    d={`M0 ${h * 0.82} C ${w * 0.22} ${h * 0.66}, ${w * 0.42} ${h * 0.7}, ${w * 0.6} ${h * 0.8} S ${w * 0.88} ${h * 0.7}, ${w} ${h * 0.76} L ${w} ${h} L 0 ${h} Z`}
                    fill={Colors.primarySoft}
                    opacity={0.6}
                />
                <Path
                    d={`M0 ${h * 0.92} C ${w * 0.3} ${h * 0.82}, ${w * 0.62} ${h * 0.86}, ${w} ${h * 0.9} L ${w} ${h} L 0 ${h} Z`}
                    fill={Colors.primary}
                    opacity={0.55}
                />
            </Svg>
        </View>
    );
}

const styles = StyleSheet.create({
    wrap: {
        overflow: 'hidden',
        borderRadius: 28,
    },
    halo: {
        position: 'absolute',
    },
});

export const CalmScene = memo(CalmSceneImpl);
export default CalmScene;
