import { registerRootComponent } from 'expo';

import App from './App';
import { installDevBypass } from './src/dev/devBypass';

// registerRootComponent calls AppRegistry.registerComponent('main', () => App);
// It also ensures that whether you load the app in Expo Go or in a native build,
// the environment is set up appropriately
// DEV-ONLY: mocked API + auth bypass when EXPO_PUBLIC_DEV_AUTH_BYPASS=true in a dev
// bundle (see src/dev/devBypass.ts). Inert in release builds.
installDevBypass();

registerRootComponent(App);
