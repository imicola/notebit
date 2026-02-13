/**
 * useTheme Hook
 * Manages eye protection themes, color temperature, and automatic theme switching
 */
import { useState, useEffect, useCallback } from 'react';

const STORAGE_KEY = 'notebit-theme-settings';

const DEFAULT_SETTINGS = {
  theme: 'default', // default, amber, forest, space
  colorTemp: 6500, // Color temperature in Kelvin (2700-6500)
  autoSwitch: false, // Automatic day/night theme switching
  autoSwitchDayTheme: 'default',
  autoSwitchNightTheme: 'amber',
};

/**
 * Map color temperature to lightness adjustment
 * Warmer temperatures (lower K) reduce blue light and increase warmth
 * @param {number} temp - Color temperature in Kelvin
 * @returns {number} Lightness adjustment percentage
 */
function mapColorTempToLightness(temp) {
  // 6500K (daylight) = 0% adjustment
  // 2700K (warm) = -10% adjustment (slightly darker for eye comfort)
  const normalized = (temp - 2700) / (6500 - 2700); // 0 to 1
  return Math.round((normalized - 1) * 10); // -10 to 0
}

/**
 * Check if current time is night (18:00-06:00)
 * @returns {boolean} True if night time
 */
function isNightTime() {
  const hour = new Date().getHours();
  return hour < 6 || hour >= 18;
}

export function useTheme() {
  const [settings, setSettings] = useState(() => {
    try {
      const stored = localStorage.getItem(STORAGE_KEY);
      return stored ? { ...DEFAULT_SETTINGS, ...JSON.parse(stored) } : DEFAULT_SETTINGS;
    } catch (error) {
      console.error('Failed to load theme settings:', error);
      return DEFAULT_SETTINGS;
    }
  });

  // Apply theme to DOM
  const applyTheme = useCallback((themeName) => {
    document.documentElement.setAttribute('data-theme', themeName);
  }, []);

  // Apply color temperature adjustments
  const applyColorTemp = useCallback((temp) => {
    const adjustment = mapColorTempToLightness(temp);
    document.documentElement.style.setProperty('--color-temp-adjust', `${adjustment}%`);
  }, []);

  // Update theme
  const setTheme = useCallback((themeName) => {
    setSettings((prev) => {
      const newSettings = { ...prev, theme: themeName };
      localStorage.setItem(STORAGE_KEY, JSON.stringify(newSettings));
      return newSettings;
    });
    applyTheme(themeName);
  }, [applyTheme]);

  // Update color temperature
  const setColorTemp = useCallback((temp) => {
    const clampedTemp = Math.max(2700, Math.min(6500, temp));
    setSettings((prev) => {
      const newSettings = { ...prev, colorTemp: clampedTemp };
      localStorage.setItem(STORAGE_KEY, JSON.stringify(newSettings));
      return newSettings;
    });
    applyColorTemp(clampedTemp);
  }, [applyColorTemp]);

  // Toggle automatic theme switching
  const setAutoSwitch = useCallback((enabled, dayTheme = 'default', nightTheme = 'amber') => {
    setSettings((prev) => {
      const newSettings = {
        ...prev,
        autoSwitch: enabled,
        autoSwitchDayTheme: dayTheme,
        autoSwitchNightTheme: nightTheme,
      };
      localStorage.setItem(STORAGE_KEY, JSON.stringify(newSettings));
      return newSettings;
    });
  }, []);

  // Update auto-switch themes
  const setAutoSwitchThemes = useCallback((dayTheme, nightTheme) => {
    setSettings((prev) => {
      const newSettings = {
        ...prev,
        autoSwitchDayTheme: dayTheme,
        autoSwitchNightTheme: nightTheme,
      };
      localStorage.setItem(STORAGE_KEY, JSON.stringify(newSettings));
      return newSettings;
    });
  }, []);

  // Apply theme on mount
  useEffect(() => {
    if (settings.autoSwitch) {
      const targetTheme = isNightTime() ? settings.autoSwitchNightTheme : settings.autoSwitchDayTheme;
      applyTheme(targetTheme);
    } else {
      applyTheme(settings.theme);
    }
    
    applyColorTemp(settings.colorTemp);
  }, [settings, applyTheme, applyColorTemp]);

  // Auto-switch timer (check every minute)
  useEffect(() => {
    if (!settings.autoSwitch) return;

    const checkAndSwitch = () => {
      const targetTheme = isNightTime() ? settings.autoSwitchNightTheme : settings.autoSwitchDayTheme;
      applyTheme(targetTheme);
    };

    // Check immediately
    checkAndSwitch();

    // Set up interval to check every minute
    const interval = setInterval(checkAndSwitch, 60000);

    return () => clearInterval(interval);
  }, [settings.autoSwitch, settings.autoSwitchDayTheme, settings.autoSwitchNightTheme, applyTheme]);

  return {
    theme: settings.theme,
    colorTemp: settings.colorTemp,
    autoSwitch: settings.autoSwitch,
    autoSwitchDayTheme: settings.autoSwitchDayTheme,
    autoSwitchNightTheme: settings.autoSwitchNightTheme,
    setTheme,
    setColorTemp,
    setAutoSwitch,
    setAutoSwitchThemes,
  };
}

export default useTheme;
