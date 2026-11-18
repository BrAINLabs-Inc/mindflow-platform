import React from 'react';
import AppNavigator from './src/navigation/AppNavigator';
import { SafeAreaProvider } from 'react-native-safe-area-context';
import { ErrorBoundary } from './src/components/ErrorBoundary';
import { OfflineNotice } from './src/components/OfflineNotice';
import { DevBypassBadge } from './src/dev/DevBypassBadge';

export default function App() {
  return (
    <ErrorBoundary>
      <SafeAreaProvider>
        <AppNavigator />
        <OfflineNotice />
        <DevBypassBadge />
      </SafeAreaProvider>
    </ErrorBoundary>
  );
}
