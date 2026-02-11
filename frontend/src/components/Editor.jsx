import { useState, useEffect, useRef, useMemo } from 'react';
import { EditorState } from '@codemirror/state';
import { EditorView, keymap, lineNumbers, highlightActiveLine, highlightActiveLineGutter, drawSelection } from '@codemirror/view';
import { defaultKeymap, history, historyKeymap } from '@codemirror/commands';
import { markdown, markdownLanguage } from '@codemirror/lang-markdown';
import { languages } from '@codemirror/language-data';
import { syntaxHighlighting, defaultHighlightStyle, HighlightStyle } from '@codemirror/language';
import { tags } from '@lezer/highlight';
import { MatchDecorator, ViewPlugin, Decoration, WidgetType } from '@codemirror/view';
import MarkdownIt from 'markdown-it';
import markdownItGithubAlerts from 'markdown-it-github-alerts';
import clsx from 'clsx';
import { Split, Eye, Edit3, Save } from 'lucide-react';

// --- Custom Syntax Highlighting (Obsidian-like) ---
const obsidianHighlightStyle = HighlightStyle.define([
  { tag: tags.heading1, class: 'cm-heading-1 text-2xl font-bold text-accent' },
  { tag: tags.heading2, class: 'cm-heading-2 text-xl font-bold text-accent' },
  { tag: tags.heading3, class: 'cm-heading-3 text-lg font-bold text-accent' },
  { tag: tags.strong, class: 'cm-strong font-bold text-obsidian-orange' },
  { tag: tags.emphasis, class: 'cm-emphasis italic text-obsidian-yellow' },
  { tag: tags.link, class: 'cm-link text-obsidian-blue underline' },
  { tag: tags.monospace, class: 'cm-monospace font-mono text-obsidian-pink bg-primary-alt rounded px-1' },
]);

// Decorators for custom syntax (==highlight==, [[wiki]])
const highlightDecorator = new MatchDecorator({
  regexp: /==[^=]+==/g,
  decoration: Decoration.mark({ class: 'bg-obsidian-yellow/30 text-obsidian-yellow rounded px-0.5' }),
});

const wikiLinkDecorator = new MatchDecorator({
  regexp: /\[\[[^\]]+\]\]/g,
  decoration: Decoration.mark({ class: 'text-obsidian-purple hover:underline cursor-pointer' }),
});

const obsidianExtensionsPlugin = ViewPlugin.fromClass(
  class {
    constructor(view) {
      this.highlightDeco = highlightDecorator.createDeco(view);
      this.wikiDeco = wikiLinkDecorator.createDeco(view);
    }
    update(update) {
      this.highlightDeco = highlightDecorator.updateDeco(update, this.highlightDeco);
      this.wikiDeco = wikiLinkDecorator.updateDeco(update, this.wikiDeco);
    }
  },
  {
    decorations: (v) => {
      const highlight = v.highlightDeco || Decoration.none;
      const wiki = v.wikiDeco || Decoration.none;
      return highlight.update({add: wiki.iter(), sort: true}); // Merging decorations is tricky, simplified here
      // Better approach: use multiple plugins or a single one carefully.
      // For now let's just use one plugin for simplicity or separate them in the extensions array.
    },
  }
);

// We need separate plugins to avoid merging issues easily
const highlightPlugin = ViewPlugin.define(
  (view) => ({
    decorations: highlightDecorator.createDeco(view),
    update(u) { this.decorations = highlightDecorator.updateDeco(u, this.decorations); }
  }),
  { decorations: (v) => v.decorations }
);

const wikiPlugin = ViewPlugin.define(
  (view) => ({
    decorations: wikiLinkDecorator.createDeco(view),
    update(u) { this.decorations = wikiLinkDecorator.updateDeco(u, this.decorations); }
  }),
  { decorations: (v) => v.decorations }
);


// --- Theme ---
const editorTheme = EditorView.theme({
  '&': { height: '100%', fontSize: '15px', fontFamily: 'var(--font-text)', backgroundColor: 'var(--background-primary)', color: 'var(--text-normal)' },
  '.cm-scroller': { overflow: 'auto', fontFamily: 'var(--font-text)' },
  '.cm-content': { fontFamily: 'var(--font-text)' },
  '.cm-line': { fontFamily: 'var(--font-text)' },
  '.cm-gutters': { backgroundColor: 'var(--background-primary)', color: 'var(--text-faint)', border: 'none', fontFamily: 'var(--font-text)' },
  '.cm-activeLineGutter': { backgroundColor: 'transparent', color: 'var(--text-muted)' },
  '.cm-activeLine': { backgroundColor: 'var(--background-primary-alt)' },
  '.cm-cursor': { borderLeftColor: 'var(--text-accent)' },
  '.cm-selectionBackground': { backgroundColor: 'var(--background-modifier-border-focus)' },
  '&.cm-focused .cm-selectionBackground': { backgroundColor: 'var(--background-modifier-border-focus)' },
});

// --- Main Component ---
const Editor = ({ content, onChange, onSave, filename, isZenMode }) => {
  const editorRef = useRef(null);
  const viewRef = useRef(null);
  const [viewMode, setViewMode] = useState('split'); // 'edit', 'preview', 'split'
  const [unsaved, setUnsaved] = useState(false);
  
  // Markdown It setup
  const md = useMemo(() => {
    const m = new MarkdownIt({
      html: true,
      linkify: true,
      typographer: true,
    });
    
    // Enable GitHub Alerts (compatible with Obsidian Callouts: > [!NOTE] etc.)
    m.use(markdownItGithubAlerts);
    
    return m;
  }, []);

  // Initialize Editor
  useEffect(() => {
    if (!editorRef.current) return;

    const startState = EditorState.create({
      doc: content,
      extensions: [
        lineNumbers(),
        highlightActiveLineGutter(),
        history(),
        drawSelection(),
        EditorState.allowMultipleSelections.of(true),
        markdown({ base: markdownLanguage, codeLanguages: languages }),
        syntaxHighlighting(obsidianHighlightStyle, { fallback: true }), // Use our custom style
        editorTheme,
        highlightActiveLine(),
        keymap.of([
            ...defaultKeymap, 
            ...historyKeymap,
            { key: "Mod-s", run: () => { handleSave(); return true; } }
        ]),
        highlightPlugin,
        wikiPlugin,
        EditorView.updateListener.of((update) => {
          if (update.docChanged) {
            const newContent = update.state.doc.toString();
            onChange(newContent);
            setUnsaved(true);
          }
        }),
      ],
    });

    const view = new EditorView({
      state: startState,
      parent: editorRef.current,
    });

    viewRef.current = view;

    return () => {
      view.destroy();
    };
  }, []); // Run once on mount

  // Update content if changed externally (e.g. file switch)
  useEffect(() => {
    if (viewRef.current && content !== viewRef.current.state.doc.toString()) {
        const transaction = viewRef.current.state.update({
            changes: { from: 0, to: viewRef.current.state.doc.length, insert: content }
        });
        viewRef.current.dispatch(transaction);
        setUnsaved(false);
    }
  }, [content]); // Only when content prop changes significantly (e.g. file load)

  const handleSave = () => {
    if (viewRef.current) {
      onSave(viewRef.current.state.doc.toString());
      setUnsaved(false);
    }
  };

  return (
    <div className="flex flex-col h-full w-full">
      {/* Toolbar (Hidden in Zen Mode) */}
      {!isZenMode && (
        <div className="flex justify-between items-center px-4 py-2 bg-secondary border-b border-modifier-border h-[40px] shrink-0">
            <div className="flex items-center gap-2">
                <span className="text-sm font-medium text-normal">{filename || 'Untitled'}</span>
                {unsaved && <div className="w-2 h-2 rounded-full bg-accent" title="Unsaved changes" />}
            </div>
            <div className="flex items-center gap-1 bg-primary rounded p-0.5">
                <button 
                    onClick={() => setViewMode('edit')}
                    className={clsx("p-1.5 rounded transition-colors", viewMode === 'edit' ? "bg-modifier-hover text-normal" : "text-muted hover:text-normal")}
                    title="Edit Only"
                >
                    <Edit3 size={14} />
                </button>
                <button 
                    onClick={() => setViewMode('split')}
                    className={clsx("p-1.5 rounded transition-colors", viewMode === 'split' ? "bg-modifier-hover text-normal" : "text-muted hover:text-normal")}
                    title="Split View"
                >
                    <Split size={14} />
                </button>
                <button 
                    onClick={() => setViewMode('preview')}
                    className={clsx("p-1.5 rounded transition-colors", viewMode === 'preview' ? "bg-modifier-hover text-normal" : "text-muted hover:text-normal")}
                    title="Preview Only"
                >
                    <Eye size={14} />
                </button>
            </div>
            <button 
                onClick={handleSave}
                disabled={!unsaved}
                className="text-xs flex items-center gap-1 text-muted hover:text-normal disabled:opacity-30 disabled:hover:text-muted"
            >
                <Save size={14} />
                <span>Save</span>
            </button>
        </div>
      )}

      {/* Main Content */}
      <div className={clsx("flex flex-1 overflow-hidden", isZenMode && "justify-center bg-primary")}>
        <div className={clsx("flex w-full h-full", isZenMode && "max-w-6xl shadow-2xl")}>
            {/* Editor Pane */}
            <div className={clsx("h-full overflow-hidden transition-all duration-300", 
                viewMode === 'preview' ? "hidden" : (viewMode === 'split' ? "w-1/2 border-r border-modifier-border" : "w-full"),
                isZenMode && "bg-primary"
            )}>
                <div ref={editorRef} className="h-full text-base" />
            </div>

            {/* Preview Pane */}
            <div className={clsx("h-full overflow-auto bg-primary p-8 prose font-mono max-w-none transition-all duration-300", 
                viewMode === 'edit' ? "hidden" : (viewMode === 'split' ? "w-1/2" : "w-full"),
                isZenMode && "bg-primary"
            )}>
                <div dangerouslySetInnerHTML={{ __html: md.render(content || '') }} />
            </div>
        </div>
      </div>
    </div>
  );
};

export default Editor;
