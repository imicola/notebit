export const createNodePainter = ({ hoverNode, highlightNodes, themeColors }) => {
	return (node, ctx, globalScale) => {
		const isHover = node === hoverNode;
		const isHighlight = highlightNodes.has(node.id);
		const label = node.label;
		const fontSize = 12 / globalScale;
		
		// Obsidian-style: Cleaner sizes
		let radius = Math.sqrt(node.val) * 1.5;
		if (isHover) radius *= 1.2;

		// 1. Draw Node
		ctx.beginPath();
		ctx.arc(node.x, node.y, radius, 0, 2 * Math.PI, false);
		
		// Color logic
		if (isHover) {
			ctx.fillStyle = themeColors.highlight;
		} else if (isHighlight) {
			ctx.fillStyle = themeColors.text; // Highlighted neighbors are bright
		} else {
			ctx.fillStyle = node.color;
		}
		
		// Optional: subtle border for better contrast on dark bg if needed
		// ctx.strokeStyle = 'rgba(0,0,0,0.5)';
		// ctx.lineWidth = 0.5 / globalScale;
		// ctx.stroke();

		ctx.fill();

		// 2. Draw Label
		// Show label if hovered, highlighted, or zoomed in enough, or if it's a major node (concept)
		const showLabel = isHover || isHighlight || globalScale > 1.2 || (node.type === 'concept' && globalScale > 0.8);
		
		if (showLabel) {
			ctx.font = `${fontSize}px Sans-Serif`;
			ctx.textAlign = 'center';
			ctx.textBaseline = 'middle';
			
			const labelY = node.y + radius + (fontSize * 0.8);
			
			// Outline for readability instead of box
			ctx.strokeStyle = 'rgba(0, 0, 0, 0.8)';
			ctx.lineWidth = 3 / globalScale;
			ctx.lineJoin = 'round';
			ctx.strokeText(label, node.x, labelY);
			
			ctx.fillStyle = isHover ? themeColors.highlight : themeColors.text;
			ctx.fillText(label, node.x, labelY);
		}
	};
};

export const createLinkPainter = ({ highlightLinks, themeColors }) => {
	return (link, ctx, globalScale) => {
		const isHighlight = highlightLinks.has(link);
		
		// Don't draw invisible links to save performance
		if (!isHighlight && globalScale < 0.5) return;

		ctx.beginPath();
		ctx.moveTo(link.source.x, link.source.y);
		ctx.lineTo(link.target.x, link.target.y);
		
		if (isHighlight) {
			ctx.strokeStyle = themeColors.highlight; // or themeColors.linkHighlight
			ctx.lineWidth = 1.5 / globalScale;
			ctx.globalAlpha = 0.8;
		} else {
			ctx.strokeStyle = themeColors.link;
			ctx.lineWidth = 0.5 / globalScale; // Thinner lines
			ctx.globalAlpha = 0.6; // themeColors.link should already have low opacity, but ensuring here
		}
		
		ctx.stroke();
		ctx.globalAlpha = 1;
	};
};

export const createNodePointerAreaPaint = () => {
	return (node, color, ctx) => {
		ctx.fillStyle = color;
		// Expanded hit area for easier selection
		const radius = Math.sqrt(node.val) * 1.5;
		ctx.beginPath();
		ctx.arc(node.x, node.y, radius + 4, 0, 2 * Math.PI, false);
		ctx.fill();
	};
};
