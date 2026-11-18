import * as Notifications from 'expo-notifications';
import Constants from 'expo-constants';
import { Platform } from 'react-native';
import { apiFetch } from './apiClient';

Notifications.setNotificationHandler({
    handleNotification: async () => ({
        shouldShowAlert: true,
        shouldPlaySound: true,
        shouldSetBadge: false,
        shouldShowBanner: true,
        shouldShowList: true,
    }),
});

// Registers this device's Expo push token for the reminder jobs; no-ops quietly on failure since reminders are a nice-to-have, not blocking.
export async function registerForPushNotificationsAsync(): Promise<void> {
    try {
        if (Platform.OS === 'android') {
            await Notifications.setNotificationChannelAsync('default', {
                name: 'default',
                importance: Notifications.AndroidImportance.DEFAULT,
            });
        }

        const { status: existingStatus } = await Notifications.getPermissionsAsync();
        let finalStatus = existingStatus;
        if (existingStatus !== 'granted') {
            const { status } = await Notifications.requestPermissionsAsync();
            finalStatus = status;
        }
        if (finalStatus !== 'granted') return;

        const projectId = Constants.expoConfig?.extra?.eas?.projectId;
        const { data: token } = await Notifications.getExpoPushTokenAsync(
            projectId ? { projectId } : undefined
        );
        if (!token) return;

        await apiFetch('/api/notifications/register-token', {
            method: 'POST',
            body: JSON.stringify({ token, platform: Platform.OS }),
        });
    } catch (error) {
        console.log('Push notification registration skipped:', error);
    }
}

// Best-effort removal of this device's push token on sign-out, so a signed-out
// device stops receiving reminders that greet the previous participant by name.
// Must run while the session token is still stored (the endpoint requires auth),
// and never throws — failing to unregister must not block signing out.
export async function unregisterPushNotificationsAsync(): Promise<void> {
    try {
        const { status } = await Notifications.getPermissionsAsync();
        if (status !== 'granted') return; // never registered

        const projectId = Constants.expoConfig?.extra?.eas?.projectId;
        const { data: token } = await Notifications.getExpoPushTokenAsync(
            projectId ? { projectId } : undefined
        );
        if (!token) return;

        await apiFetch('/api/notifications/unregister-token', {
            method: 'POST',
            body: JSON.stringify({ token }),
        });
    } catch (error) {
        console.log('Push notification unregistration skipped:', error);
    }
}
