import React, { useState, useEffect, useRef, useCallback, useMemo } from 'react';
import ForceGraph2D from 'react-force-graph-2d';
import { Loader2, AlertCircle, Globe, Maximize2, Minimize2, ZoomIn, ZoomOut, RefreshCw } from 'lucide-react';
import { graphService } from '../services/graphService';
import { useTheme } from '../hooks/useTheme';
import * as d3 from 'd3';
import { createLinkPainter, createNodePainter, createNodePointerAreaPaint } from './GraphRenderer';

export default function GraphPanel() {
	const { theme } = useTheme();
	const [graphData, setGraphData] = useState({ nodes: [], links: [] });
	const [loading, setLoading] = useState(true);
	const [error, setError] = useState(null);
	const fgRef = useRef();
	const containerRef = useRef();
	const [hoverNode, setHoverNode] = useState(null);
	const [highlightNodes, setHighlightNodes] = useState(new Set());
	const [highlightLinks, setHighlightLinks] = useState(new Set());
	const [dimensions, setDimensions] = useState({ width: 800, height: 600 });
	
	// Double click detection
	const lastClickTime = useRef(0);
	const clickTimeout = useRef(null);

	// --- Theme Configuration ---
	const isDark = theme !== 'light';
	const themeColors = useMemo(() => ({
		background: isDark ? 'rgba(0,0,0,0)' : '#ffffff', // Transparent to let container bg show
		concept: isDark ? '#00D4FF' : '#0066CC', // Aurora Blue / Tech Blue
		note: isDark ? '#00FF88' : '#00AA44',    // Emerald Green / Green
		tag: isDark ? '#FF6B35' : '#FF8C00',     // Neon Orange / Amber
		link: isDark ? 'rgba(0, 212, 255, 0.2)' : 'rgba(0, 102, 204, 0.2)',
		linkHighlight: isDark ? '#00D4FF' : '#0066CC',
		text: isDark ? 'rgba(255,255,255,0.8)' : 'rgba(0,0,0,0.8)',
		particle: isDark ? '#ffffff' : '#000000',
	}), [isDark]);

	// --- Resize Handler ---
	useEffect(() => {
		let resizeTimer;
		const updateDimensions = () => {
			if (containerRef.current) {
				setDimensions({
					width: containerRef.current.clientWidth,
					height: containerRef.current.clientHeight
				});
			}
		};
		const handleResize = () => {
			if (resizeTimer) clearTimeout(resizeTimer);
			resizeTimer = setTimeout(updateDimensions, 150);
		};

		window.addEventListener('resize', handleResize);
		updateDimensions();
		
		// Small delay to ensure container is ready
		setTimeout(updateDimensions, 100);

		return () => {
			window.removeEventListener('resize', handleResize);
			if (resizeTimer) clearTimeout(resizeTimer);
		};
	}, []);

	// --- Load Data ---
	const loadGraph = useCallback(async () => {
		setLoading(true);
		setError(null);
		try {
			const data = await graphService.getGraphData();
			
			// Process data for ForceGraph
			// We need to ensure nodes and links are mutable objects
			const nodes = (data?.nodes || []).map(n => ({
				...n,
				// Assign color based on type
				color: n.type === 'concept' ? themeColors.concept :
					   n.type === 'tag' ? themeColors.tag :
					   themeColors.note,
				val: n.type === 'concept' ? 20 : (n.size * 2 + 5) // Size based on connections
			}));

			const links = (data?.links || []).map(l => ({
				source: l.source,
				target: l.target,
				type: l.type,
				strength: l.strength
			}));

			setGraphData({ nodes, links });
		} catch (err) {
			console.error('Failed to load graph:', err);
			setError(err.message || 'Failed to load graph');
		} finally {
			setLoading(false);
		}
	}, [themeColors]);

	useEffect(() => {
		loadGraph();
	}, [loadGraph]);

	// --- Force Configuration ---
	useEffect(() => {
		if (fgRef.current) {
			// Add collision force to prevent overlap
			fgRef.current.d3Force('collide', d3.forceCollide(node => Math.sqrt(node.val) * 2 + 5));
			
			// Adjust charge force
			fgRef.current.d3Force('charge').strength(-100);
			
			// Adjust link force based on semantic strength
			fgRef.current.d3Force('link').strength(link => link.strength ? link.strength * 0.5 : 0.1);
		}
	}, [graphData]);

	// --- Interactions ---
	const handleNodeHover = (node) => {
		setHoverNode(node || null);
		
		const newHighlightNodes = new Set();
		const newHighlightLinks = new Set();
		
		if (node) {
			newHighlightNodes.add(node.id);
			graphData.links.forEach(link => {
				if (link.source.id === node.id || link.target.id === node.id) {
					newHighlightLinks.add(link);
					newHighlightNodes.add(link.source.id);
					newHighlightNodes.add(link.target.id);
				}
			});
		}

		setHighlightNodes(newHighlightNodes);
		setHighlightLinks(newHighlightLinks);
	};

	const handleNodeClick = (node) => {
		const now = Date.now();
		const timeSinceLastClick = now - lastClickTime.current;
		
		if (timeSinceLastClick < 300) {
			// Double Click
			if (clickTimeout.current) clearTimeout(clickTimeout.current);
			
			// Elastic Animation effect (expand view)
			fgRef.current.zoom(4, 800);
			fgRef.current.centerAt(node.x, node.y, 800);
			
			// Navigate to file
			if (node.path) {
				setTimeout(() => {
					window.dispatchEvent(new CustomEvent('open-file', {
						detail: { path: node.path }
					}));
				}, 500);
			}
		} else {
			// Single Click
			lastClickTime.current = now;
			clickTimeout.current = setTimeout(() => {
				// Focus on node without navigation
				fgRef.current.centerAt(node.x, node.y, 1000);
				fgRef.current.zoom(2, 2000);
			}, 300);
		}
	};

	// --- Custom Rendering ---
	const paintNode = useMemo(
		() => createNodePainter({ hoverNode, highlightNodes, themeColors }),
		[hoverNode, highlightNodes, themeColors]
	);
	const paintLink = useMemo(
		() => createLinkPainter({ highlightLinks, themeColors }),
		[highlightLinks, themeColors]
	);
	const nodePointerAreaPaint = useMemo(() => createNodePointerAreaPaint(), []);

	// --- Controls ---
	const zoomIn = () => {
		fgRef.current.zoom(fgRef.current.zoom() * 1.2, 400);
	};

	const zoomOut = () => {
		fgRef.current.zoom(fgRef.current.zoom() / 1.2, 400);
	};

	const resetZoom = () => {
		fgRef.current.zoomToFit(400);
	};

	return (
		<div className="flex flex-col h-full bg-secondary">
			{/* Header */}
			<div className="px-4 py-3 bg-secondary border-b border-modifier-border flex justify-between items-center z-10">
				<div className="flex items-center gap-2">
					<Globe className="text-obsidian-purple" size={18} />
					<h2 className="text-sm font-semibold text-normal">Knowledge Graph</h2>
				</div>
				<div className="flex items-center gap-2">
					<button onClick={loadGraph} className="p-1 hover:bg-modifier-hover rounded text-muted hover:text-normal transition-colors" title="Reload Graph">
						<RefreshCw size={14} />
					</button>
					<span className="text-xs text-muted border-l border-modifier-border pl-2">
						{graphData.nodes.length} nodes
					</span>
				</div>
			</div>

			{/* Graph Container */}
			<div className="flex-1 relative bg-primary overflow-hidden" ref={containerRef}>
				{loading && (
					<div className="absolute inset-0 flex flex-col items-center justify-center z-20 bg-primary/80 backdrop-blur-sm">
						<Loader2 size={32} className="animate-spin text-obsidian-purple mb-4" />
						<p className="text-sm text-muted">Calculating graph physics...</p>
					</div>
				)}

				{error && (
					<div className="absolute inset-0 flex flex-col items-center justify-center z-20">
						<AlertCircle className="text-obsidian-red mb-4" size={32} />
						<p className="text-sm text-muted">{error}</p>
						<button 
							onClick={loadGraph}
							className="mt-4 px-4 py-2 bg-primary-alt border border-modifier-border rounded hover:bg-modifier-hover text-sm"
						>
							Retry
						</button>
					</div>
				)}

				{!loading && !error && graphData.nodes.length === 0 && (
					<div className="absolute inset-0 flex flex-col items-center justify-center z-20 text-center">
						<Globe className="text-faint mb-4" size={48} />
						<h3 className="text-lg font-medium text-muted mb-2">No Graph Data</h3>
						<p className="text-sm text-muted">
							Create some notes to see them connected here.
						</p>
					</div>
				)}

				{/* Force Graph */}
				<ForceGraph2D
					ref={fgRef}
					width={dimensions.width}
					height={dimensions.height}
					graphData={graphData}
					backgroundColor="transparent" // Use container bg
					
					// Node Styling
					nodeLabel="label"
					nodeRelSize={6}
					nodeCanvasObject={paintNode}
					nodePointerAreaPaint={nodePointerAreaPaint}

					// Link Styling
					linkCanvasObject={paintLink}
					linkDirectionalParticles={2}
					linkDirectionalParticleWidth={2}
					linkDirectionalParticleSpeed={0.005}
					linkDirectionalParticleColor={() => themeColors.particle}
					
					// Forces
					d3AlphaDecay={0.02} // Slower decay = more movement
					d3VelocityDecay={0.3} // Friction
					warmupTicks={100}
					cooldownTicks={100}
					
					// Interaction
					onNodeHover={handleNodeHover}
					onNodeClick={handleNodeClick}
					enableNodeDrag={true}
					enableZoomInteraction={true}
					enablePanInteraction={true}
				/>

				{/* Controls Overlay */}
				<div className="absolute bottom-4 right-4 flex flex-col gap-2 bg-primary-alt/80 backdrop-blur rounded-lg p-2 border border-modifier-border shadow-lg">
					<button onClick={zoomIn} className="p-1.5 hover:bg-modifier-hover rounded text-muted hover:text-normal transition-colors">
						<ZoomIn size={18} />
					</button>
					<button onClick={zoomOut} className="p-1.5 hover:bg-modifier-hover rounded text-muted hover:text-normal transition-colors">
						<ZoomOut size={18} />
					</button>
					<button onClick={resetZoom} className="p-1.5 hover:bg-modifier-hover rounded text-muted hover:text-normal transition-colors">
						<Maximize2 size={18} />
					</button>
				</div>

				{/* Legend Overlay */}
				<div className="absolute top-4 left-4 bg-primary-alt/80 backdrop-blur rounded-lg p-3 border border-modifier-border shadow-lg">
					<div className="flex flex-col gap-2">
						<div className="flex items-center gap-2">
							<div className="w-3 h-3 rounded-full shadow-[0_0_8px_rgba(0,212,255,0.6)]" style={{ backgroundColor: themeColors.concept }}></div>
							<span className="text-xs text-muted">Concept Node</span>
						</div>
						<div className="flex items-center gap-2">
							<div className="w-3 h-3 rounded-full shadow-[0_0_8px_rgba(0,255,136,0.6)]" style={{ backgroundColor: themeColors.note }}></div>
							<span className="text-xs text-muted">Note Node</span>
						</div>
						<div className="flex items-center gap-2">
							<div className="w-3 h-3 rounded-full shadow-[0_0_8px_rgba(255,107,53,0.6)]" style={{ backgroundColor: themeColors.tag }}></div>
							<span className="text-xs text-muted">Tag Node</span>
						</div>
					</div>
				</div>
			</div>
		</div>
	);
}
