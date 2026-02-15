import { useState, useMemo, useEffect, useRef, useCallback, memo } from 'react';
import { ChevronRight, ChevronDown, File, Folder, FolderOpen } from 'lucide-react';
import clsx from 'clsx';

const FileTreeRow = memo(({ node, level, isExpanded, isSelected, isFocused, onToggle, onSelect }) => {
  const handleClick = (e) => {
    e.stopPropagation();
    if (node.isDir) {
      onToggle(node);
    } else {
      onSelect(node);
    }
  };

  const paddingLeft = level * 16 + 8;

  return (
    <div
      className={clsx(
        'flex items-center gap-1.5 py-1.5 pr-2 cursor-pointer select-none text-sm transition-all duration-200 border-l-2 list-item-hover',
        isSelected ? 'bg-primary-alt text-accent font-medium border-[var(--text-accent)] shadow-1' : 'border-transparent',
        isFocused && !isSelected && 'bg-modifier-hover', // Visual indication for focus
        isFocused && isSelected && 'bg-modifier-border-focus shadow-2',
        !isSelected && !isFocused && 'text-muted hover:bg-modifier-hover hover:text-normal',
        node.isDir && 'font-medium'
      )}
      style={{ paddingLeft: `${paddingLeft}px` }}
      onClick={handleClick}
    >
      {node.isDir ? (
        <>
          {isExpanded ? <ChevronDown size={16} className="shrink-0" /> : <ChevronRight size={16} className="shrink-0" />}
          {isExpanded ? <FolderOpen size={16} className="text-muted shrink-0" /> : <Folder size={16} className="text-muted shrink-0" />}
        </>
      ) : (
        <>
          <span style={{ width: 16 }} className="shrink-0" />
          <File size={16} className="text-faint shrink-0" />
        </>
      )}
      <span className="truncate">{node.name}</span>
    </div>
  );
});

const FileTree = ({ tree, onSelect, selectedPath }) => {
  const [expandedPaths, setExpandedPaths] = useState(new Set());
  const [focusedPath, setFocusedPath] = useState(null);
  const containerRef = useRef(null);

  // Initialize expanded state for root
  useEffect(() => {
    if (tree && expandedPaths.size === 0) {
      setExpandedPaths(new Set([tree.path]));
    }
  }, [tree]);

  // Sync focused path with selected path initially or when changed externally
  useEffect(() => {
    if (selectedPath) {
      setFocusedPath(selectedPath);
      // Also ensure parents are expanded? (Optional, maybe too complex for now)
    }
  }, [selectedPath]);

  const toggleExpand = useCallback((node) => {
    setExpandedPaths(prev => {
      const newExpanded = new Set(prev);
      if (newExpanded.has(node.path)) {
        newExpanded.delete(node.path);
      } else {
        newExpanded.add(node.path);
      }
      return newExpanded;
    });
    setFocusedPath(node.path);
  }, []);

  // Flatten the tree based on expanded state
  const visibleNodes = useMemo(() => {
    if (!tree) return [];
    
    const nodes = [];
    const traverse = (node, level) => {
      nodes.push({ ...node, level });
      if (node.isDir && expandedPaths.has(node.path) && node.children) {
        // Sort children: Directories first, then files
        const sortedChildren = [...node.children].sort((a, b) => {
          if (a.isDir === b.isDir) return a.name.localeCompare(b.name);
          return a.isDir ? -1 : 1;
        });
        
        sortedChildren.forEach(child => traverse(child, level + 1));
      }
    };
    
    traverse(tree, 0);
    return nodes;
  }, [tree, expandedPaths]);

  const handleKeyDown = (e) => {
    if (!visibleNodes.length) return;

    const currentIndex = visibleNodes.findIndex(n => n.path === focusedPath);
    let nextIndex = currentIndex;

    switch (e.key) {
      case 'ArrowDown':
        e.preventDefault();
        nextIndex = Math.min(currentIndex + 1, visibleNodes.length - 1);
        break;
      case 'ArrowUp':
        e.preventDefault();
        nextIndex = Math.max(currentIndex - 1, 0);
        break;
      case 'ArrowRight':
        e.preventDefault();
        if (currentIndex !== -1) {
          const node = visibleNodes[currentIndex];
          if (node.isDir) {
            if (!expandedPaths.has(node.path)) {
              toggleExpand(node);
            } else {
              // Move to first child if exists
              nextIndex = Math.min(currentIndex + 1, visibleNodes.length - 1);
            }
          }
        }
        break;
      case 'ArrowLeft':
        e.preventDefault();
        if (currentIndex !== -1) {
          const node = visibleNodes[currentIndex];
          if (node.isDir && expandedPaths.has(node.path)) {
            toggleExpand(node); // Collapse
          } else {
            // Move to parent
            // Find the closest previous node with level < current.level
            for (let i = currentIndex - 1; i >= 0; i--) {
              if (visibleNodes[i].level < node.level) {
                nextIndex = i;
                break;
              }
            }
          }
        }
        break;
      case 'Enter':
        e.preventDefault();
        if (currentIndex !== -1) {
          const node = visibleNodes[currentIndex];
          if (node.isDir) {
            toggleExpand(node);
          } else {
            onSelect(node);
          }
        }
        break;
      default:
        return;
    }

    if (nextIndex !== currentIndex && nextIndex >= 0) {
      setFocusedPath(visibleNodes[nextIndex].path);
      // Ensure element is visible in scroll container
      // This would require refs to each element or calculation
    }
  };

  const handleSelect = useCallback((n) => {
    onSelect(n);
    setFocusedPath(n.path);
  }, [onSelect]);

  if (!tree) {
    return (
      <div className="h-full flex items-center justify-center text-muted text-sm p-5">
        <p>No folder selected</p>
      </div>
    );
  }

  return (
    <div 
      className="h-full overflow-y-auto bg-secondary border-r border-modifier-border outline-none focus:ring-1 focus:ring-accent/50"
      tabIndex={0}
      onKeyDown={handleKeyDown}
      ref={containerRef}
    >
      {visibleNodes.map((node) => (
        <FileTreeRow
          key={node.path}
          node={node}
          level={node.level}
          isExpanded={expandedPaths.has(node.path)}
          isSelected={selectedPath === node.path}
          isFocused={focusedPath === node.path}
          onToggle={toggleExpand}
          onSelect={handleSelect}
        />
      ))}
    </div>
  );
};

export default FileTree;
