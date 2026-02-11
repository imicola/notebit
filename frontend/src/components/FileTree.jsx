import { useState } from 'react';
import { ChevronRight, ChevronDown, File, Folder, FolderOpen } from 'lucide-react';
import './FileTree.css';

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
        className={`file-node ${isSelected ? 'selected' : ''} ${node.isDir ? 'directory' : 'file'}`}
        style={{ paddingLeft: `${paddingLeft}px` }}
        onClick={handleClick}
      >
        {node.isDir ? (
          <>
            {isExpanded ? <ChevronDown size={16} /> : <ChevronRight size={16} />}
            {isExpanded ? <FolderOpen size={16} /> : <Folder size={16} />}
          </>
        ) : (
          <>
            <span style={{ width: 16 }} />
            <File size={16} />
          </>
        )}
        <span className="file-name">{node.name}</span>
      </div>
      {node.isDir && isExpanded && node.children && (
        <div className="children">
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
      <div className="file-tree empty">
        <p className="empty-message">No folder selected</p>
      </div>
    );
  }

  return (
    <div className="file-tree">
      <FileTreeNode
        node={tree}
        onSelect={onSelect}
        selectedPath={selectedPath}
      />
    </div>
  );
};

export default FileTree;
