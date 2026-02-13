/**
 * Wiki Link Service
 * Handles parsing and navigation for Obsidian-style wiki links
 * Supports: [[Note]], [[Note|Alias]], [[Note#Heading]]
 */

/**
 * Parse wiki link syntax
 * @param {string} linkText - Raw wiki link text (including [[ ]])
 * @returns {object|null} Parsed link object or null if invalid
 * @property {string} noteName - Target note name
 * @property {string|null} alias - Display alias (if present)
 * @property {string|null} heading - Heading anchor (if present)
 */
export function parseWikiLink(linkText) {
  // Remove [[ and ]]
  const cleanText = linkText.replace(/^\[\[|\]\]$/g, '');
  if (!cleanText) return null;

  // Pattern: [[noteName|alias#heading]]
  // Support combinations:
  // - [[noteName]]
  // - [[noteName|alias]]
  // - [[noteName#heading]]
  // - [[noteName|alias#heading]]
  const pattern = /^([^|\]#]+)(?:\|([^\]#]+))?(?:#([^\]]+))?$/;
  const match = cleanText.match(pattern);

  if (!match) return null;

  return {
    noteName: match[1].trim(),
    alias: match[2] ? match[2].trim() : null,
    heading: match[3] ? match[3].trim() : null,
    raw: linkText,
  };
}

/**
 * Find file path from note name
 * @param {string} noteName - Note name (without .md extension)
 * @param {object} fileTree - File tree from fileService.listFiles()
 * @returns {string|null} Relative path to file or null if not found
 */
export function findFilePathByName(noteName, fileTree) {
  if (!fileTree) return null;

  // Normalize note name (remove .md if present)
  const normalizedName = noteName.endsWith('.md') ? noteName : `${noteName}.md`;

  // Recursive search through file tree
  const search = (node) => {
    if (!node) return null;

    // Check if current node matches
    if (node.name === normalizedName && !node.isDir) {
      return node.path;
    }

    // Search children
    if (node.children) {
      for (const child of node.children) {
        const result = search(child);
        if (result) return result;
      }
    }

    return null;
  };

  return search(fileTree);
}

/**
 * Generate list of all notes for autocomplete
 * @param {object} fileTree - File tree from fileService.listFiles()
 * @returns {Array<{name: string, path: string}>} List of notes
 */
export function extractAllNotes(fileTree) {
  const notes = [];

  const traverse = (node) => {
    if (!node) return;

    if (!node.isDir && node.name.endsWith('.md')) {
      notes.push({
        name: node.name.replace(/\.md$/, ''),
        path: node.path,
      });
    }

    if (node.children) {
      node.children.forEach(traverse);
    }
  };

  traverse(fileTree);
  return notes;
}

/**
 * Find heading position in markdown content
 * @param {string} content - Markdown content
 * @param {string} heading - Heading text to find
 * @returns {number} Line number (0-indexed) or -1 if not found
 */
export function findHeadingLine(content, heading) {
  if (!content || !heading) return -1;

  const lines = content.split('\n');
  const normalizedHeading = heading.toLowerCase().trim();

  for (let i = 0; i < lines.length; i++) {
    const line = lines[i].trim();
    // Match any level heading: # Heading
    const headingMatch = line.match(/^#{1,6}\s+(.+)$/);
    if (headingMatch) {
      const headingText = headingMatch[1].toLowerCase().trim();
      if (headingText === normalizedHeading) {
        return i;
      }
    }
  }

  return -1;
}

/**
 * Scroll to heading in CodeMirror editor
 * @param {EditorView} view - CodeMirror EditorView instance
 * @param {string} heading - Heading text to scroll to
 */
export function scrollToHeading(view, heading) {
  if (!view || !heading) return;

  const content = view.state.doc.toString();
  const lineNumber = findHeadingLine(content, heading);

  if (lineNumber >= 0 && lineNumber < view.state.doc.lines) {
    const line = view.state.doc.line(lineNumber + 1); // CodeMirror uses 1-indexed
    const lineBlock = view.lineBlockAt(line.from);

    // Scroll to line with smooth animation
    view.scrollDOM.scrollTo({
      top: lineBlock.top,
      behavior: 'smooth',
    });

    // Highlight the line temporarily
    highlightLine(view, line.from);
  }
}

/**
 * Highlight a line temporarily
 * @param {EditorView} view - CodeMirror EditorView instance
 * @param {number} pos - Position in document
 */
function highlightLine(view, pos) {
  // Add temporary highlight class
  const line = view.state.doc.lineAt(pos);
  const lineElement = view.domAtPos(line.from).node;
  
  if (lineElement && lineElement.parentElement) {
    const lineDiv = lineElement.parentElement.closest('.cm-line');
    if (lineDiv) {
      lineDiv.classList.add('wiki-link-highlight');
      
      // Remove highlight after animation
      setTimeout(() => {
        lineDiv.classList.remove('wiki-link-highlight');
      }, 2000);
    }
  }
}

/**
 * Create autocomplete source for wiki links
 * @param {Array<{name: string, path: string}>} notes - List of notes
 * @returns {Function} Autocomplete source function for CodeMirror
 */
export function createWikiLinkAutocomplete(notes) {
  return (context) => {
    // Check if we're inside [[ ]]
    const before = context.matchBefore(/\[\[[^\]]*$/);
    if (!before) return null;

    // Extract partial input
    const partial = before.text.slice(2).toLowerCase(); // Remove [[

    // Filter matching notes
    const options = notes
      .filter((note) => note.name.toLowerCase().includes(partial))
      .map((note) => ({
        label: note.name,
        type: 'text',
        apply: (view, completion, from, to) => {
          // Insert note name and close brackets
          const text = `${note.name}]]`;
          view.dispatch({
            changes: { from: before.from + 2, to, insert: text },
          });
        },
      }));

    return {
      from: before.from + 2,
      options,
      validFor: /^[^\]]*$/,
    };
  };
}

export default {
  parseWikiLink,
  findFilePathByName,
  extractAllNotes,
  findHeadingLine,
  scrollToHeading,
  createWikiLinkAutocomplete,
};
