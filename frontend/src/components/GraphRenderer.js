export const createNodePainter = ({ hoverNode, highlightNodes, themeColors }) => {
	return (node, ctx, globalScale) => {
		const isHover = node === hoverNode;
		const isHighlight = highlightNodes.has(node.id);
		const label = node.label;
		const fontSize = 12 / globalScale;
		const radius = Math.sqrt(node.val) * 2;

		if (isHover) {
			const time = Date.now();
			const rippleRadius = radius + (Math.sin(time / 200) + 1) * 5;
			ctx.beginPath();
			ctx.arc(node.x, node.y, rippleRadius, 0, 2 * Math.PI, false);
			ctx.strokeStyle = node.color;
			ctx.globalAlpha = 0.3;
			ctx.stroke();
			ctx.globalAlpha = 1;
		}

		ctx.shadowColor = node.color;
		ctx.shadowBlur = isHover ? 15 : 5;
		
		ctx.beginPath();
		ctx.arc(node.x, node.y, radius, 0, 2 * Math.PI, false);
		ctx.fillStyle = node.color;

		const gradientKey = `${node.color}:${radius}`;
		if (node.__gradientKey !== gradientKey || !node.__gradient) {
			const gradient = ctx.createRadialGradient(node.x, node.y, 0, node.x, node.y, radius);
			gradient.addColorStop(0, 'rgba(255,255,255,0.8)');
			gradient.addColorStop(0.5, node.color);
			gradient.addColorStop(1, node.color);
			node.__gradient = gradient;
			node.__gradientKey = gradientKey;
		}
		ctx.fillStyle = node.__gradient || node.color;
		
		ctx.fill();
		
		ctx.shadowBlur = 0;

		if (isHover || isHighlight || globalScale > 1.5 || node.type === 'concept') {
			ctx.font = `${fontSize}px Sans-Serif`;
			ctx.textAlign = 'center';
			ctx.textBaseline = 'middle';
			ctx.fillStyle = themeColors.text;
			
			const textWidth = ctx.measureText(label).width;
			const bckgDimensions = [textWidth, fontSize].map(n => n + fontSize * 0.2);
			const labelY = node.y + radius + 2;
			node.__bckgDimensions = bckgDimensions;
			ctx.fillStyle = 'rgba(0, 0, 0, 0.6)';
			ctx.fillRect(node.x - bckgDimensions[0] / 2, labelY, bckgDimensions[0], bckgDimensions[1]);
			
			ctx.fillStyle = themeColors.text;
			ctx.fillText(label, node.x, labelY + fontSize / 2);
		} else {
			node.__bckgDimensions = null;
		}
	};
};

export const createLinkPainter = ({ highlightLinks, themeColors }) => {
	return (link, ctx, globalScale) => {
		const isHighlight = highlightLinks.has(link);
		
		ctx.beginPath();
		ctx.moveTo(link.source.x, link.source.y);
		ctx.lineTo(link.target.x, link.target.y);
		
		ctx.strokeStyle = isHighlight ? themeColors.linkHighlight : themeColors.link;
		ctx.lineWidth = isHighlight ? 2 / globalScale : 1 / globalScale;
		ctx.globalAlpha = isHighlight ? 0.8 : 0.4;
		
		ctx.stroke();
		ctx.globalAlpha = 1;
	};
};

export const createNodePointerAreaPaint = () => {
	return (node, color, ctx) => {
		ctx.fillStyle = color;
		const bckgDimensions = node.__bckgDimensions;
		if (bckgDimensions) {
			const radius = Math.sqrt(node.val) * 2;
			const labelY = node.y + radius + 2;
			ctx.fillRect(node.x - bckgDimensions[0] / 2, labelY, bckgDimensions[0], bckgDimensions[1]);
		}
		ctx.beginPath();
		ctx.arc(node.x, node.y, Math.sqrt(node.val) * 2 + 2, 0, 2 * Math.PI, false);
		ctx.fill();
	};
};
