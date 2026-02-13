/**
 * Line Mapper Utility
 * Maps editor line positions to preview block elements for precise vertical synchronization
 */

/**
 * Build source line mapping from markdown content
 * Parses markdown and assigns line numbers to each block element
 * @param {string} markdownText - Raw markdown content
 * @returns {Array<{type: string, line: number, text: string}>} Block mapping array
 */
export function buildLineMapping(markdownText) {
  const lines = markdownText.split('\n');
  const mapping = [];
  let currentBlock = null;
  let inCodeBlock = false;
  let inTable = false;

  lines.forEach((line, lineNumber) => {
    const trimmed = line.trim();

    // Code block detection
    if (trimmed.startsWith('```')) {
      if (inCodeBlock) {
        // Closing code block
        if (currentBlock) {
          currentBlock.endLine = lineNumber;
          mapping.push(currentBlock);
          currentBlock = null;
        }
        inCodeBlock = false;
      } else {
        // Opening code block
        inCodeBlock = true;
        currentBlock = {
          type: 'code',
          line: lineNumber,
          startLine: lineNumber,
          text: line,
        };
      }
      return;
    }

    // Inside code block, accumulate content
    if (inCodeBlock) {
      if (currentBlock) {
        currentBlock.text += '\n' + line;
      }
      return;
    }

    // Table detection
    if (trimmed.match(/^\|.+\|$/)) {
      if (!inTable) {
        inTable = true;
        currentBlock = {
          type: 'table',
          line: lineNumber,
          startLine: lineNumber,
          text: line,
        };
      } else if (currentBlock) {
        currentBlock.text += '\n' + line;
      }
      return;
    } else if (inTable) {
      // End of table
      if (currentBlock) {
        currentBlock.endLine = lineNumber - 1;
        mapping.push(currentBlock);
        currentBlock = null;
      }
      inTable = false;
    }

    // Heading detection
    const headingMatch = trimmed.match(/^(#{1,6})\s+(.+)/);
    if (headingMatch) {
      mapping.push({
        type: 'heading',
        level: headingMatch[1].length,
        line: lineNumber,
        text: headingMatch[2],
      });
      return;
    }

    // List detection
    if (trimmed.match(/^(\*|-|\+|\d+\.)\s+/)) {
      mapping.push({
        type: 'list',
        line: lineNumber,
        text: trimmed,
        indent: line.search(/\S/),
      });
      return;
    }

    // Blockquote detection
    if (trimmed.startsWith('>')) {
      mapping.push({
        type: 'blockquote',
        line: lineNumber,
        text: trimmed.substring(1).trim(),
      });
      return;
    }

    // Paragraph (non-empty lines)
    if (trimmed.length > 0) {
      mapping.push({
        type: 'paragraph',
        line: lineNumber,
        text: trimmed,
      });
    }
  });

  // Close any remaining open blocks
  if (currentBlock) {
    currentBlock.endLine = lines.length - 1;
    mapping.push(currentBlock);
  }

  return mapping;
}

/**
 * Inject data-source-line attributes into HTML
 * Used by markdown-it plugin to add line number metadata
 * @param {object} md - MarkdownIt instance
 */
export function injectSourceLinePlugin(md) {
  // Store line mapping during rendering
  const lineMapping = new Map();

  // Override default renderers to inject data attributes
  const defaultRender = md.renderer.render.bind(md.renderer);

  md.renderer.render = function (tokens, options, env) {
    // Enhance tokens with line info
    tokens.forEach((token, idx) => {
      if (token.map && token.map.length >= 2) {
        const startLine = token.map[0];
        lineMapping.set(idx, startLine);
        
        // Inject data-source-line attribute
        if (!token.attrGet('data-source-line')) {
          token.attrPush(['data-source-line', String(startLine)]);
        }
      }
    });

    return defaultRender(tokens, options, env);
  };

  return md;
}

/**
 * Find preview element corresponding to editor line
 * @param {HTMLElement} previewContainer - Preview container element
 * @param {number} editorLine - Line number in editor (0-indexed)
 * @returns {HTMLElement|null} Corresponding preview element
 */
export function findPreviewElementByLine(previewContainer, editorLine) {
  if (!previewContainer) return null;

  // Find element with matching or closest preceding data-source-line
  const allElements = Array.from(
    previewContainer.querySelectorAll('[data-source-line]')
  );

  if (allElements.length === 0) return null;

  // Find closest element with line <= editorLine
  let closest = null;
  let minDiff = Infinity;

  allElements.forEach((el) => {
    const sourceLine = parseInt(el.getAttribute('data-source-line'), 10);
    if (sourceLine <= editorLine) {
      const diff = editorLine - sourceLine;
      if (diff < minDiff) {
        minDiff = diff;
        closest = el;
      }
    }
  });

  return closest || allElements[0];
}

/**
 * Calculate scroll position based on line mapping
 * @param {number} editorScrollTop - Current editor scroll position
 * @param {number} editorScrollHeight - Total editor scroll height
 * @param {number} editorClientHeight - Visible editor height
 * @param {HTMLElement} editorView - CodeMirror view
 * @param {HTMLElement} previewContainer - Preview container
 * @returns {number} Target preview scroll position
 */
export function calculatePreviewScrollPosition(
  editorScrollTop,
  editorScrollHeight,
  editorClientHeight,
  editorView,
  previewContainer
) {
  if (!editorView || !previewContainer) return 0;

  try {
    // Get current visible position in editor
    const visibleTop = editorScrollTop;
    
    // Find the line at the top of the viewport
    // Using lineBlockAt to get accurate line position
    const topPos = editorView.lineBlockAt(visibleTop);
    if (!topPos) return 0;

    const editorLine = editorView.state.doc.lineAt(topPos.from).number - 1; // 0-indexed

    // Find corresponding preview element
    const targetElement = findPreviewElementByLine(previewContainer, editorLine);
    if (!targetElement) return 0;

    // Calculate precise scroll position
    const elementTop = targetElement.offsetTop;
    const elementHeight = targetElement.offsetHeight;

    // Calculate offset within the editor line
    const lineOffset = (visibleTop - topPos.top) / topPos.height;
    const previewOffset = lineOffset * elementHeight;

    return elementTop + previewOffset;
  } catch (error) {
    console.warn('Error calculating preview scroll position:', error);
    // Fallback to percentage-based scrolling
    const percentage = editorScrollTop / (editorScrollHeight - editorClientHeight);
    return percentage * (previewContainer.scrollHeight - previewContainer.clientHeight);
  }
}

/**
 * Calculate editor scroll position from preview scroll
 * @param {number} previewScrollTop - Current preview scroll position
 * @param {HTMLElement} previewContainer - Preview container
 * @param {HTMLElement} editorView - CodeMirror view
 * @param {number} editorScrollHeight - Total editor scroll height
 * @param {number} editorClientHeight - Visible editor height
 * @returns {number} Target editor scroll position
 */
export function calculateEditorScrollPosition(
  previewScrollTop,
  previewContainer,
  editorView,
  editorScrollHeight,
  editorClientHeight
) {
  if (!editorView || !previewContainer) return 0;

  try {
    // Find visible preview element at scroll position
    const visibleElements = Array.from(
      previewContainer.querySelectorAll('[data-source-line]')
    ).filter((el) => {
      const rect = el.getBoundingClientRect();
      const containerRect = previewContainer.getBoundingClientRect();
      return rect.top >= containerRect.top && rect.top <= containerRect.bottom;
    });

    if (visibleElements.length === 0) {
      // Fallback to percentage
      const percentage = previewScrollTop / (previewContainer.scrollHeight - previewContainer.clientHeight);
      return percentage * (editorScrollHeight - editorClientHeight);
    }

    const topElement = visibleElements[0];
    const sourceLine = parseInt(topElement.getAttribute('data-source-line'), 10);

    // Find corresponding editor line
    if (sourceLine >= 0 && sourceLine < editorView.state.doc.lines) {
      const line = editorView.state.doc.line(sourceLine + 1); // 1-indexed
      const lineBlock = editorView.lineBlockAt(line.from);
      
      // Calculate offset within preview element
      const elementRect = topElement.getBoundingClientRect();
      const containerRect = previewContainer.getBoundingClientRect();
      const elementOffset = (containerRect.top - elementRect.top) / elementRect.height;
      
      return lineBlock.top + elementOffset * lineBlock.height;
    }

    return 0;
  } catch (error) {
    console.warn('Error calculating editor scroll position:', error);
    // Fallback to percentage-based scrolling
    const percentage = previewScrollTop / (previewContainer.scrollHeight - previewContainer.clientHeight);
    return percentage * (editorScrollHeight - editorClientHeight);
  }
}
