import { Fragment, useState } from 'react'
import { Dialog, Transition } from '@headlessui/react'
import { X, Type, Brain } from 'lucide-react'
import { FONTS_INTERFACE, FONTS_TEXT } from '../constants'
import AISettings from './AISettings'

export default function SettingsModal({ isOpen, onClose, settings, onUpdateSettings }) {
  const [activeTab, setActiveTab] = useState('appearance');
  const [forceCustom, setForceCustom] = useState(false);
  const isPreset = FONTS_TEXT.some(f => f.value === settings.fontText);
  const showCustomInput = forceCustom || !isPreset;

  const tabs = [
    { id: 'appearance', label: 'Appearance', icon: Type },
    { id: 'intelligence', label: 'Intelligence', icon: Brain },
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
