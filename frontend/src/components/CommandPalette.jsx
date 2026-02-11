import { Fragment, useState, useEffect } from 'react';
import { Combobox, Dialog, Transition } from '@headlessui/react';
import { Search, File, Command, FolderOpen, Save, Monitor } from 'lucide-react';
import Fuse from 'fuse.js';
import clsx from 'clsx';

const CommandPalette = ({ isOpen, setIsOpen, files, onFileSelect, commands }) => {
  const [query, setQuery] = useState('');

  // Flatten file tree for search
  const flattenFiles = (node, acc = []) => {
    if (!node) return acc;
    if (!node.isDir) {
      acc.push(node);
    }
    if (node.children) {
      node.children.forEach(child => flattenFiles(child, acc));
    }
    return acc;
  };

  const flatFiles = files ? flattenFiles(files) : [];

  const allItems = [
    ...commands.map(c => ({ ...c, type: 'command' })),
    ...flatFiles.map(f => ({ ...f, type: 'file', id: f.path, label: f.name }))
  ];

  const fuse = new Fuse(allItems, {
    keys: ['label', 'name'],
    threshold: 0.3,
  });

  const filteredItems = query === '' ? allItems.slice(0, 10) : fuse.search(query).map(result => result.item).slice(0, 10);

  const handleSelect = (item) => {
    if (item.type === 'command') {
      item.action();
    } else {
      onFileSelect(item);
    }
    setIsOpen(false);
    setQuery('');
  };

  return (
    <Transition.Root show={isOpen} as={Fragment} afterLeave={() => setQuery('')}>
      <Dialog as="div" className="relative z-50" onClose={setIsOpen}>
        <Transition.Child
          as={Fragment}
          enter="ease-out duration-300"
          enterFrom="opacity-0"
          enterTo="opacity-100"
          leave="ease-in duration-200"
          leaveFrom="opacity-100"
          leaveTo="opacity-0"
        >
          <div className="fixed inset-0 bg-black/50 transition-opacity" />
        </Transition.Child>

        <div className="fixed inset-0 z-10 overflow-y-auto p-4 sm:p-6 md:p-20">
          <Transition.Child
            as={Fragment}
            enter="ease-out duration-300"
            enterFrom="opacity-0 scale-95"
            enterTo="opacity-100 scale-100"
            leave="ease-in duration-200"
            leaveFrom="opacity-100 scale-100"
            leaveTo="opacity-0 scale-95"
          >
            <Dialog.Panel className="mx-auto max-w-xl transform overflow-hidden rounded-xl bg-secondary shadow-2xl ring-1 ring-black ring-opacity-5 transition-all border border-modifier-border">
              <Combobox onChange={handleSelect}>
                <div className="relative">
                  <Search
                    className="pointer-events-none absolute top-3.5 left-4 h-5 w-5 text-muted"
                    aria-hidden="true"
                  />
                  <Combobox.Input
                    className="h-12 w-full border-0 bg-transparent pl-11 pr-4 text-normal placeholder-muted focus:ring-0 sm:text-sm"
                    placeholder="Search files or commands..."
                    onChange={(event) => setQuery(event.target.value)}
                    displayValue={(item) => item?.label}
                  />
                </div>

                {filteredItems.length > 0 && (
                  <Combobox.Options static className="max-h-80 scroll-py-2 overflow-y-auto py-2 text-sm text-muted bg-secondary">
                    {filteredItems.map((item) => (
                      <Combobox.Option
                        key={item.id || item.path}
                        value={item}
                        className={({ active }) =>
                          clsx(
                            'cursor-default select-none px-4 py-2',
                            active && 'bg-modifier-hover text-normal'
                          )
                        }
                      >
                        {({ active }) => (
                          <div className="flex items-center gap-3">
                            {item.type === 'command' ? (
                                item.icon ? <item.icon size={16} /> : <Command size={16} />
                            ) : (
                                <File size={16} />
                            )}
                            <span className={clsx('flex-auto truncate', active ? 'text-normal' : 'text-muted')}>
                              {item.label || item.name}
                            </span>
                            {item.shortcut && (
                                <span className="text-xs text-faint">{item.shortcut}</span>
                            )}
                          </div>
                        )}
                      </Combobox.Option>
                    ))}
                  </Combobox.Options>
                )}

                {query !== '' && filteredItems.length === 0 && (
                  <p className="p-4 text-sm text-muted">No results found.</p>
                )}
              </Combobox>
            </Dialog.Panel>
          </Transition.Child>
        </div>
      </Dialog>
    </Transition.Root>
  );
};

export default CommandPalette;
