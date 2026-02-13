import { Fragment, useState } from 'react'
import { Dialog, Transition } from '@headlessui/react'
import { X, Type, Brain, Eye, Sun, Moon } from 'lucide-react'
import { FONTS_INTERFACE, FONTS_TEXT } from '../constants'
import AISettings from './AISettings'
import { useTheme } from '../hooks'

export default function SettingsModal({ isOpen, onClose, settings, onUpdateSettings }) {
  const [activeTab, setActiveTab] = useState('appearance');
  const [forceCustom, setForceCustom] = useState(false);
  const isPreset = FONTS_TEXT.some(f => f.value === settings.fontText);
  const showCustomInput = forceCustom || !isPreset;

  // Theme hook
  const {
    theme,
    colorTemp,
    autoSwitch,
    autoSwitchDayTheme,
    autoSwitchNightTheme,
    setTheme,
    setColorTemp,
    setAutoSwitch,
    setAutoSwitchThemes,
  } = useTheme();

  const tabs = [
    { id: 'appearance', label: 'Appearance', icon: Type },
    { id: 'intelligence', label: 'Intelligence', icon: Brain },
  ];

  const themes = [
    { id: 'default', name: 'Default (Blue)', description: 'Standard dark theme' },
    { id: 'amber', name: 'Amber Protection', description: 'Warm eye protection theme' },
    { id: 'forest', name: 'Forest Protection', description: 'Green eye protection theme' },
    { id: 'space', name: 'Space Protection', description: 'Dark gray eye protection theme' },
  ];

  return (
    <Transition appear show={isOpen} as={Fragment}>
      <Dialog as="div" className="relative z-50" onClose={onClose}>
        <Transition.Child
          as={Fragment}
          enter="ease-out duration-300"
          enterFrom="opacity-0"
          enterTo="opacity-100"
          leave="ease-in duration-200"
          leaveFrom="opacity-100"
          leaveTo="opacity-0"
        >
          <div className="fixed inset-0 bg-black/50 backdrop-blur-sm" />
        </Transition.Child>

        <div className="fixed inset-0 overflow-y-auto">
          <div className="flex min-h-full items-center justify-center p-4 text-center">
            <Transition.Child
              as={Fragment}
              enter="ease-out duration-300"
              enterFrom="opacity-0 scale-95"
              enterTo="opacity-100 scale-100"
              leave="ease-in duration-200"
              leaveFrom="opacity-100 scale-100"
              leaveTo="opacity-0 scale-95"
            >
              <Dialog.Panel className="w-full max-w-4xl transform overflow-hidden rounded-2xl bg-secondary border border-modifier-border p-6 text-left align-middle shadow-xl transition-all">
                <div className="flex justify-between items-center mb-6 border-b border-modifier-border pb-4">
                  <Dialog.Title
                    as="h3"
                    className="text-lg font-medium leading-6 text-normal"
                  >
                    Settings
                  </Dialog.Title>
                  <button
                    onClick={onClose}
                    className="rounded-md p-1 hover:bg-modifier-hover text-muted hover:text-normal transition-colors"
                  >
                    <X size={20} />
                  </button>
                </div>

                <div className="flex gap-6 h-[500px]">
                  {/* Sidebar / Tabs */}
                  <div className="w-48 shrink-0 flex flex-col gap-1">
                    {tabs.map((tab) => (
                        <button
                            key={tab.id}
                            onClick={() => setActiveTab(tab.id)}
                            className={`flex items-center gap-2 px-3 py-2 rounded-md font-medium text-sm text-left transition-colors ${
                                activeTab === tab.id
                                    ? 'bg-modifier-hover text-normal'
                                    : 'text-muted hover:text-normal hover:bg-modifier-hover/50'
                            }`}
                        >
                            <tab.icon size={16} />
                            {tab.label}
                        </button>
                    ))}
                  </div>

                  {/* Content */}
                  <div className="flex-1 overflow-y-auto pr-2">
                    {activeTab === 'appearance' && (
                        <div className="space-y-6">
                            {/* Eye Protection Theme */}
                            <div>
                                <label className="flex items-center gap-2 text-sm font-medium text-normal mb-2">
                                    <Eye size={16} />
                                    Eye Protection Theme
                                </label>
                                <p className="text-xs text-muted mb-3">
                                    Choose a theme optimized for reducing eye strain during extended writing sessions.
                                </p>
                                <div className="space-y-2">
                                    {themes.map((themeOption) => (
                                        <button
                                            key={themeOption.id}
                                            onClick={() => setTheme(themeOption.id)}
                                            className={`w-full text-left px-4 py-3 rounded-lg border transition-all ${
                                                theme === themeOption.id
                                                    ? 'border-obsidian-purple bg-obsidian-purple/10 text-normal'
                                                    : 'border-modifier-border bg-primary-alt text-muted hover:border-obsidian-purple/50 hover:text-normal'
                                            }`}
                                        >
                                            <div className="font-medium text-sm">{themeOption.name}</div>
                                            <div className="text-xs opacity-75 mt-1">{themeOption.description}</div>
                                        </button>
                                    ))}
                                </div>
                            </div>

                            {/* Color Temperature */}
                            <div>
                                <label className="block text-sm font-medium text-normal mb-2">
                                    Color Temperature
                                </label>
                                <p className="text-xs text-muted mb-3">
                                    Adjust warmth (lower = warmer, reduces blue light). Current: {colorTemp}K
                                </p>
                                <div className="flex items-center gap-4">
                                    <span className="text-xs text-muted">2700K<br/>(Warm)</span>
                                    <input
                                        type="range"
                                        min="2700"
                                        max="6500"
                                        step="100"
                                        value={colorTemp}
                                        onChange={(e) => setColorTemp(parseInt(e.target.value))}
                                        className="flex-1 h-2 bg-modifier-border rounded-lg appearance-none cursor-pointer slider"
                                    />
                                    <span className="text-xs text-muted">6500K<br/>(Cool)</span>
                                </div>
                            </div>

                            {/* Auto Theme Switching */}
                            <div>
                                <label className="flex items-center justify-between text-sm font-medium text-normal mb-2">
                                    <span className="flex items-center gap-2">
                                        <Sun size={16} className="text-yellow" />
                                        /
                                        <Moon size={16} className="text-blue" />
                                        Auto Theme Switching
                                    </span>
                                    <button
                                        onClick={() => setAutoSwitch(!autoSwitch)}
                                        className={`relative inline-flex h-6 w-11 items-center rounded-full transition-colors ${
                                            autoSwitch ? 'bg-obsidian-purple' : 'bg-modifier-border'
                                        }`}
                                    >
                                        <span
                                            className={`inline-block h-4 w-4 transform rounded-full bg-white transition-transform ${
                                                autoSwitch ? 'translate-x-6' : 'translate-x-1'
                                            }`}
                                        />
                                    </button>
                                </label>
                                <p className="text-xs text-muted mb-3">
                                    Automatically switch theme based on time of day (Night: 18:00-06:00).
                                </p>
                                {autoSwitch && (
                                    <div className="space-y-3 pt-3 border-t border-modifier-border">
                                        <div>
                                            <label className="block text-xs font-medium text-muted mb-1">
                                                Day Theme (06:00-18:00)
                                            </label>
                                            <select
                                                value={autoSwitchDayTheme}
                                                onChange={(e) => setAutoSwitchThemes(e.target.value, autoSwitchNightTheme)}
                                                className="block w-full rounded-md border border-modifier-border bg-primary-alt py-2 px-3 text-sm text-normal shadow-sm focus:border-obsidian-purple focus:outline-none focus:ring-1 focus:ring-obsidian-purple"
                                            >
                                                {themes.map((t) => (
                                                    <option key={t.id} value={t.id}>
                                                        {t.name}
                                                    </option>
                                                ))}
                                            </select>
                                        </div>
                                        <div>
                                            <label className="block text-xs font-medium text-muted mb-1">
                                                Night Theme (18:00-06:00)
                                            </label>
                                            <select
                                                value={autoSwitchNightTheme}
                                                onChange={(e) => setAutoSwitchThemes(autoSwitchDayTheme, e.target.value)}
                                                className="block w-full rounded-md border border-modifier-border bg-primary-alt py-2 px-3 text-sm text-normal shadow-sm focus:border-obsidian-purple focus:outline-none focus:ring-1 focus:ring-obsidian-purple"
                                            >
                                                {themes.map((t) => (
                                                    <option key={t.id} value={t.id}>
                                                        {t.name}
                                                    </option>
                                                ))}
                                            </select>
                                        </div>
                                    </div>
                                )}
                            </div>

                            {/* Divider */}
                            <div className="pt-4 border-t border-modifier-border" />

                            {/* Interface Font */}
                            <div>
                                <label className="block text-sm font-medium text-normal mb-2">
                                    Interface Font
                                </label>
                                <p className="text-xs text-muted mb-3">
                                    Used for UI elements like sidebar, menus, and dialogs.
                                </p>
                                <select
                                    value={settings.fontInterface}
                                    onChange={(e) => onUpdateSettings('fontInterface', e.target.value)}
                                    className="block w-full rounded-md border border-modifier-border bg-primary-alt py-2 px-3 text-normal shadow-sm focus:border-obsidian-purple focus:outline-none focus:ring-1 focus:ring-obsidian-purple sm:text-sm"
                                >
                                    {FONTS_INTERFACE.map((font) => (
                                        <option key={font.name} value={font.value}>
                                            {font.name}
                                        </option>
                                    ))}
                                </select>
                            </div>

                            {/* Text Font */}
                            <div>
                                <label className="block text-sm font-medium text-normal mb-2">
                                    Text Font
                                </label>
                                <p className="text-xs text-muted mb-3">
                                    Used for the editor and markdown preview.
                                </p>
                                <select
                                    value={showCustomInput ? 'custom' : settings.fontText}
                                    onChange={(e) => {
                                        if (e.target.value === 'custom') {
                                            setForceCustom(true);
                                        } else {
                                            setForceCustom(false);
                                            onUpdateSettings('fontText', e.target.value);
                                        }
                                    }}
                                    className="block w-full rounded-md border border-modifier-border bg-primary-alt py-2 px-3 text-normal shadow-sm focus:border-obsidian-purple focus:outline-none focus:ring-1 focus:ring-obsidian-purple sm:text-sm"
                                >
                                    {FONTS_TEXT.map((font) => (
                                        <option key={font.name} value={font.value}>
                                            {font.name}
                                        </option>
                                    ))}
                                    <option value="custom">Custom...</option>
                                </select>

                                {showCustomInput && (
                                    <input
                                        type="text"
                                        value={settings.fontText}
                                        onChange={(e) => onUpdateSettings('fontText', e.target.value)}
                                        placeholder='e.g. "Fira Code", monospace'
                                        className="mt-2 block w-full rounded-md border border-modifier-border bg-primary-alt py-2 px-3 text-normal shadow-sm focus:border-obsidian-purple focus:outline-none focus:ring-1 focus:ring-obsidian-purple sm:text-sm"
                                    />
                                )}
                            </div>

                            <div className="pt-4 border-t border-modifier-border">
                                <p className="text-xs text-muted">
                                    Changes are saved automatically and applied immediately.
                                </p>
                            </div>
                        </div>
                    )}

                    {activeTab === 'intelligence' && (
                        <AISettings />
                    )}
                  </div>
                </div>

              </Dialog.Panel>
            </Transition.Child>
          </div>
        </div>
      </Dialog>
    </Transition>
  )
}
