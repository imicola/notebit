import { useState, useEffect, useRef } from 'react';
import { Network } from 'vis-network/standalone';
import { Loader2, AlertCircle, Globe } from 'lucide-react';
import { graphService } from '../services/graphService';

export default function GraphPanel() {
	const [graphData, setGraphData] = useState({ nodes: [], links: [] });
	const [loading, setLoading] = useState(true);
	const [error, setError] = useState(null);
	const networkRef = useRef(null);
	const containerRef = useRef(null);

	useEffect(() => {
		loadGraph();
	}, []);

	const loadGraph = async () => {
		setLoading(true);
		setError(null);
		try {
			const data = await graphService.getGraphData();
			setGraphData(data);
		} catch (err) {
			console.error('Failed to load graph:', err);
			setError(err.message || 'Failed to load graph');
		} finally {
			setLoading(false);
		}
	};

	useEffect(() => {
		if (!loading && !error && graphData.nodes.length > 0 && !networkRef.current && containerRef.current) {
			// Convert to vis-network format
			const nodes = graphData.nodes.map(node => ({
				id: node.id,
				label: node.label,
				title: `${node.type}: ${node.label}`,
				value: node.val || 1,
				font: { size: node.size * 2 + 12 },
				shape: 'dot',
				color: {
					background: 'var(--obsidian-purple)',
					border: 'var(--obsidian-purple)',
					highlight: {
						background: 'var(--obsidian-purple)',
						border: 'var(--obsidian-purple)',
					},
				},
			}));

			const links = graphData.links.map(link => ({
				from: link.source,
				to: link.target,
				title: `${link.type} - ${Math.round(link.strength * 100)}%`,
				arrows: {
					to: { enabled: true, strokeWidth: 1.5, type: 'arrow' },
				},
				color: {
					color: link.type === 'explicit'
						? 'var(--obsidian-purple)'
						: 'var(--obsidian-cyan)',
					opacity: link.type === 'explicit' ? 0.6 : 0.3,
					dashes: link.type !== 'explicit',
				},
				width: link.strength * 3,
			}));

			if (containerRef.current) {
				const data = { nodes, edges: links };
				const options = {
					nodes: {
						borderWidth: 2,
						size: 30,
					},
					edges: {
						width: 2,
						smooth: { type: 'continuous' },
					},
					physics: {
						stabilization: true,
						barnesHut: {
							gravitationalConstant: -2000,
							springConstant: 0.04,
							springLength: 95,
						},
					},
					interaction: {
						hover: true,
						tooltipDelay: 200,
					},
				};
				networkRef.current = new Network(containerRef.current, data, options);
					
					// Force a redraw after a short delay to ensure correct sizing
					setTimeout(() => {
						if (networkRef.current) {
							networkRef.current.fit();
							networkRef.current.redraw();
						}
					}, 100);

					networkRef.current.fit();

				// Add node click handler
				networkRef.current.on('click', (params) => {
					if (params.nodes.length > 0) {
						const nodeId = params.nodes[0];
						const node = graphData.nodes.find(n => n.id === nodeId);
						if (node) {
							// Emit event to open file in editor
							window.dispatchEvent(new CustomEvent('open-file', {
								detail: { path: node.path }
							}));
						}
					}
				});
			}
		}
	}, [graphData, loading, error]); // Removed networkRef from deps as it's a ref

	// Cleanup on unmount
	useEffect(() => {
		return () => {
			if (networkRef.current) {
				networkRef.current.destroy();
				networkRef.current = null;
			}
		};
	}, []);

	return (
		<div className="flex flex-col h-full bg-secondary">
			{/* Header */}
			<div className="px-4 py-3 bg-secondary border-b border-modifier-border flex justify-between items-center">
				<div className="flex items-center gap-2">
					<Globe className="text-obsidian-purple" size={18} />
					<h2 className="text-sm font-semibold text-normal">Knowledge Graph</h2>
				</div>
				<div className="text-xs text-muted">
					{graphData.nodes.length} nodes, {graphData.links.length} links
				</div>
			</div>

			{/* Loading */}
			{loading && (
				<div className="flex flex-col items-center justify-center h-full">
					<Loader2 size={32} className="animate-spin text-muted mb-4" />
					<p className="text-sm text-muted mt-2">Loading graph...</p>
				</div>
			)}

			{/* Error */}
			{error && (
				<div className="flex flex-col items-center justify-center h-full">
					<AlertCircle className="text-obsidian-red mb-4" size={32} />
					<p className="text-sm text-muted mt-2">{error}</p>
				</div>
			)}

			{/* Legend */}
			{!loading && !error && graphData.nodes.length > 0 && (
				<div className="px-4 py-2 bg-primary-alt/50 border-b border-modifier-border flex gap-4">
					<div className="flex items-center gap-2">
						<div className="w-3 h-3 rounded-full bg-obsidian-purple"></div>
						<span className="text-xs text-muted">Notes</span>
					</div>
					<div className="flex items-center gap-2">
						<div className="w-3 h-0.5 border-t-2 border-obsidian-purple"></div>
						<span className="text-xs text-muted">Wiki Links</span>
					</div>
					<div className="flex items-center gap-2">
						<div className="w-3 h-0.5 border-t-2 border-dashed border-obsidian-cyan"></div>
						<span className="text-xs text-muted">Semantic Similarity</span>
					</div>
				</div>
			)}

			{/* Empty state */}
			{!loading && !error && graphData.nodes.length === 0 && (
				<div className="flex flex-col items-center justify-center h-full text-center">
					<Globe className="text-faint mb-4" size={48} />
					<h3 className="text-lg font-medium text-muted mb-2">No Graph Data</h3>
					<p className="text-sm text-muted">
						Open and index some files to see the knowledge graph.
					</p>
				</div>
			)}

			{/* Network Container */}
			<div
				ref={containerRef}
				className="flex-1 bg-primary min-h-[400px]"
				style={{ height: '100%', width: '100%' }}
			/>
		</div>
	);
}
