import { Network } from 'lucide-react';

/**
 * Graph configuration tab for AISettings
 */
export default function GraphTab({ graphConfig, setGraphConfig }) {
  return (
    <div className="space-y-6">
      <section>
        <div className="flex items-center gap-2 mb-4">
          <Network className="text-obsidian-purple" size={20} />
          <h3 className="text-lg font-medium text-normal">Knowledge Graph Configuration</h3>
        </div>

        <div className="space-y-4 bg-primary-alt/30 p-4 rounded-lg border border-modifier-border">
          <div>
            <label className="block text-sm font-medium text-normal mb-1">Similarity Threshold</label>
            <input
              type="number"
              step="0.01"
              min="0"
              max="1"
              value={graphConfig.min_similarity_threshold}
              onChange={(e) => setGraphConfig({...graphConfig, min_similarity_threshold: parseFloat(e.target.value)})}
              className="w-full rounded-md border border-modifier-border bg-primary-alt px-3 py-2 text-sm text-normal focus:border-obsidian-purple focus:outline-none"
            />
            <p className="text-xs text-muted mt-1">
              Minimum similarity score (0-1) for creating implicit links between semantically similar files.
            </p>
          </div>

          <div>
            <label className="block text-sm font-medium text-normal mb-1">Max Nodes</label>
            <input
              type="number"
              min="10"
              max="500"
              value={graphConfig.max_nodes}
              onChange={(e) => setGraphConfig({...graphConfig, max_nodes: parseInt(e.target.value)})}
              className="w-full rounded-md border border-modifier-border bg-primary-alt px-3 py-2 text-sm text-normal focus:border-obsidian-purple focus:outline-none"
            />
            <p className="text-xs text-muted mt-1">
              Maximum number of nodes to display in the graph (for performance).
            </p>
          </div>

          <div className="flex items-center gap-2">
            <input
              type="checkbox"
              id="show-implicit-links"
              checked={graphConfig.show_implicit_links}
              onChange={(e) => setGraphConfig({...graphConfig, show_implicit_links: e.target.checked})}
              className="rounded border-modifier-border bg-primary-alt"
            />
            <label htmlFor="show-implicit-links" className="text-sm text-normal cursor-pointer">
              Show implicit (semantic similarity) links
            </label>
            <p className="text-xs text-muted">
              When enabled, shows dashed lines for semantically similar files.
            </p>
          </div>
        </div>
      </section>
    </div>
  );
}
