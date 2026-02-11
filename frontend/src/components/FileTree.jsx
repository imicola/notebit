import { useState } from 'react';
import { ChevronRight, ChevronDown, File, Folder, FolderOpen } from 'lucide-react';
import clsx from 'clsx';

const FileTreeNode = ({ node, level = 0, onSelect, selectedPath }) => {
  const [isExpanded, setIsExpanded] = useState(false);

  const handleClick = () => {
    if (node.isDir) {
      setIsExpanded(!isExpanded);
    } else {
      onSelect(node);
    }
  };

  const isSelected = selectedPath === node.path;
  const paddingLeft = level * 16 + 8;

  return (
    <div>
      <div
        className={clsx(
          'flex items-center gap-1.5 py-1.5 pr-2 cursor-pointer select-none text-sm transition-all duration-150 border-l-2',
          'text-muted hover:bg-modifier-hover hover:text-normal',
          isSelected ? 'bg-primary-alt text-accent font-medium border-[var(--text-accent)]' : 'border-transparent',
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
      {node.isDir && isExpanded && node.children && (
        <div>
          {node.children.map((child) => (
            <FileTreeNode
              key={child.path}
              node={child}
              level={level + 1}
              onSelect={onSelect}
              selectedPath={selectedPath}
            />
          ))}
        </div>
      )}
    </div>
  );
};

const FileTree = ({ tree, onSelect, selectedPath }) => {
  if (!tree) {
    return (
      <div className="h-full flex items-center justify-center text-muted text-sm p-5">
        <p>No folder selected</p>
      </div>
    );
  }

  return (
    <div className="h-full overflow-y-auto bg-secondary border-r border-modifier-border">
      <FileTreeNode
        node={tree}
        onSelect={onSelect}
        selectedPath={selectedPath}
      />
    </div>
  );
};

export default FileTree;
