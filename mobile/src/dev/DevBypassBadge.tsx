import React from 'react';
import { StyleSheet, Text, View } from 'react-native';
import { useSafeAreaInsets } from 'react-native-safe-area-context';
import { Colors } from '../constants/colors';
import { DEV_AUTH_BYPASS } from './devBypass';

// Small always-visible marker so a bypassed dev preview is never mistaken for a
// real, logged-in session. Renders nothing unless the bypass is enabled.
export function DevBypassBadge() {
    const insets = useSafeAreaInsets();
    if (!DEV_AUTH_BYPASS) return null;
    return (
        <View pointerEvents="none" style={[styles.wrap, { top: insets.top + 2 }]}>
            <View style={styles.badge}>
                <Text style={styles.text}>Dev preview</Text>
            </View>
        </View>
    );
}

const styles = StyleSheet.create({
    wrap: { position: 'absolute', right: 10, alignItems: 'flex-end' },
    badge: {
        backgroundColor: Colors.accent,
        paddingHorizontal: 8,
        paddingVertical: 2,
        borderRadius: 999,
        opacity: 0.85,
    },
    text: { color: Colors.white, fontSize: 10, fontWeight: '600' },
});
