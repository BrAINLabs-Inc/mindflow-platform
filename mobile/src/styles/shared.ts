import { ViewStyle } from 'react-native';
import { Colors } from '../constants/colors';

// Two named shadow levels, warm-tinted and very soft, so cards float gently on
// the sand background instead of sitting in hard grey boxes.
export const cardShadow: ViewStyle = {
    shadowColor: Colors.shadow,
    shadowOffset: { width: 0, height: 4 },
    shadowOpacity: 0.05,
    shadowRadius: 14,
    elevation: 2,
};

export const cardShadowElevated: ViewStyle = {
    shadowColor: Colors.shadow,
    shadowOffset: { width: 0, height: 8 },
    shadowOpacity: 0.07,
    shadowRadius: 20,
    elevation: 4,
};
