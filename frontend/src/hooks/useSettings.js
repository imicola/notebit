/**
 * useSettings Hook
 * Manages application settings with localStorage persistence
 */
import { useState, useEffect, useCallback } from 'react';
import { STORAGE_KEYS, DEFAULT_SETTINGS } from '../constants';

/**
 * Custom hook for managing settings
 * @returns {Object} Settings state and handlers
 */
export const useSettings = () => {
  const [settings, setSettings] = useState(DEFAULT_SETTINGS);

  // Load settings from localStorage on mount
  useEffect(() => {
    const saved = localStorage.getItem(STORAGE_KEYS.SETTINGS);
    if (saved) {
      try {
        const parsed = JSON.parse(saved);
        setSettings(prev => ({ ...prev, ...parsed }));
        applyFontSettings(parsed);
      } catch (e) {
        console.error('Failed to load settings:', e);
      }
    }
  }, []);

  // Update a single setting
  const updateSetting = useCallback((key, value) => {
    setSettings(prev => {
      const newSettings = { ...prev, [key]: value };
      localStorage.setItem(STORAGE_KEYS.SETTINGS, JSON.stringify(newSettings));
      return newSettings;
    });

    // Apply CSS variable for font changes
    applyFontVariable(key, value);
  }, []);

  // Update multiple settings at once
  const updateSettings = useCallback((newSettings) => {
    setSettings(prev => {
      const merged = { ...prev, ...newSettings };
      localStorage.setItem(STORAGE_KEYS.SETTINGS, JSON.stringify(merged));
      return merged;
    });

    Object.entries(newSettings).forEach(([key, value]) => {
      applyFontVariable(key, value);
    });
  }, []);

  // Reset to defaults
  const resetSettings = useCallback(() => {
    localStorage.setItem(STORAGE_KEYS.SETTINGS, JSON.stringify(DEFAULT_SETTINGS));
    setSettings(DEFAULT_SETTINGS);
    applyFontSettings(DEFAULT_SETTINGS);
  }, []);

  return {
    settings,
    updateSetting,
    updateSettings,
    resetSettings
  };
};

/**
 * Apply font settings to CSS variables
 */
const applyFontSettings = (settings) => {
  if (settings.fontInterface) {
    applyFontVariable('fontInterface', settings.fontInterface);
  }
  if (settings.fontText) {
    applyFontVariable('fontText', settings.fontText);
  }
};

/**
 * Apply a single font variable to document
 */
const applyFontVariable = (key, value) => {
  if (key === 'fontInterface') {
    document.documentElement.style.setProperty('--font-interface', value);
  } else if (key === 'fontText') {
    document.documentElement.style.setProperty('--font-text', value);
  }
};

export default useSettings;
