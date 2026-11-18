import React from 'react';
import { StyleSheet, View, Text } from 'react-native';
import { Colors } from '../constants/colors';

interface LogoBlockProps {
    /** Compact: wordmark only, left-aligned, for in-app headers. */
    compact?: boolean;
}

// Shared brand mark: a quiet single-colour "MindFlow" wordmark with a soft leaf dot,
// plus a sentence-case tagline. Used as the hero block on the auth screens and,
// compact, in the dashboard header.
export function LogoBlock({ compact = false }: LogoBlockProps) {
    return (
        <View style={[styles.logoBlock, compact && styles.logoBlockCompact]}>
            <View style={styles.logoRow}>
                <View style={[styles.dot, compact && styles.dotCompact]} />
                <Text style={[styles.wordmark, compact && styles.wordmarkCompact]}>
                    <Text style={styles.mind}>Mind</Text>Flow
                </Text>
            </View>
            {!compact && <Text style={styles.tagline}>Find your inner peace</Text>}
        </View>
    );
}

const styles = StyleSheet.create({
    logoBlock: { alignItems: 'center', marginBottom: 4 },
    logoBlockCompact: { alignItems: 'flex-start', marginBottom: 0 },
    logoRow: { flexDirection: 'row', alignItems: 'center', gap: 8 },
    dot: { width: 14, height: 14, borderRadius: 7, borderTopRightRadius: 2, backgroundColor: Colors.accent },
    dotCompact: { width: 11, height: 11, borderRadius: 6, borderTopRightRadius: 2 },
    wordmark: { fontSize: 28, fontWeight: '700', color: Colors.primary, letterSpacing: -0.5 },
    wordmarkCompact: { fontSize: 21 },
    mind: { fontWeight: '400', color: Colors.sea },
    tagline: { fontSize: 13, color: Colors.textMuted, marginTop: 4, fontWeight: '500', letterSpacing: 0.2 },
});
