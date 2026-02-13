/**
 * Indent Guides Plugin for CodeMirror 6
 * Renders visual indent guides for improved code/text readability
 */
import { ViewPlugin, Decoration, EditorView } from '@codemirror/view';
import { RangeSetBuilder } from '@codemirror/state';

/**
 * Count leading whitespace indentation
 * @param {string} text - Line text
 * @param {number} tabSize - Tab size (default 2)
 * @returns {number} Indentation level
 */
function getIndentLevel(text, tabSize = 2) {
  let indent = 0;
  for (let i = 0; i < text.length; i++) {
    if (text[i] === ' ') {
      indent++;
    } else if (text[i] === '\t') {
      indent += tabSize;
    } else {
      break;
    }
  }
  return Math.floor(indent / tabSize);
}

/**
 * Create line decoration for indent guides
 * @param {number} level - Indentation level
 * @returns {Decoration} Line decoration
 */
function createIndentDecoration(level) {
  return Decoration.line({
    attributes: {
      class: `indent-guide-line`,
      style: `--indent-level: ${level}`,
    },
  });
}

/**
 * Indent guides plugin
 * Adds visual indent guide lines to the editor
 */
export const indentGuidesPlugin = ViewPlugin.fromClass(
  class {
    decorations;

    constructor(view) {
      this.decorations = this.buildDecorations(view);
    }

    update(update) {
      if (update.docChanged || update.viewportChanged) {
        this.decorations = this.buildDecorations(update.view);
      }
    }

    buildDecorations(view) {
      const builder = new RangeSetBuilder();
      const tabSize = view.state.facet(EditorView.tabSize) || 2;

      for (let { from, to } of view.visibleRanges) {
        for (let pos = from; pos <= to; ) {
          const line = view.state.doc.lineAt(pos);
          const lineText = line.text;
          const indentLevel = getIndentLevel(lineText, tabSize);

          if (indentLevel > 0) {
            builder.add(
              line.from,
              line.from,
              createIndentDecoration(indentLevel)
            );
          }

          pos = line.to + 1;
        }
      }

      return builder.finish();
    }
  },
  {
    decorations: (v) => v.decorations,
  }
);

/**
 * Indent guides theme
 * CSS styles for indent guide lines
 */
export const indentGuidesTheme = EditorView.baseTheme({
  '.indent-guide-line': {
    position: 'relative',
  },
  '.indent-guide-line::before': {
    content: '""',
    position: 'absolute',
    left: 'calc(var(--indent-level) * 2ch)',
    top: '0',
    bottom: '0',
    width: '1px',
    backgroundColor: 'var(--background-modifier-border)',
    opacity: '0.4',
    pointerEvents: 'none',
  },
});

export default indentGuidesPlugin;
