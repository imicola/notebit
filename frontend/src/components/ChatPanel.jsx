import { useState, useEffect, useRef } from 'react';
import { Send, Loader2, Sparkles } from 'lucide-react';
import { RAGQuery } from '../../wailsjs/go/main/App';
import { EventsOn } from '../../wailsjs/runtime/runtime';
import clsx from 'clsx';

export default function ChatPanel() {
	const [messages, setMessages] = useState([]);
	const [input, setInput] = useState('');
	const [loading, setLoading] = useState(false);
	const messagesEndRef = useRef(null);

	// Track streaming state per message
	const [streamingStates, setStreamingStates] = useState({});

	const scrollToBottom = () => {
		messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
	};

	useEffect(() => {
		scrollToBottom();
	}, [messages, streamingStates, input]);

	// Listen for streaming chunks
	useEffect(() => {
		const cleanup = EventsOn('rag_chunk', (data) => {
			const { messageId, content } = data;
			setStreamingStates(prev => ({ ...prev, [messageId]: true }));

			// Update or append message content
			setMessages(prev => {
				const updated = [...prev];
				const msgIndex = updated.findIndex(m => m.id === messageId);
				if (msgIndex >= 0) {
					if (updated[msgIndex].role === 'assistant') {
						// Append to existing assistant message
						updated[msgIndex] = {
							...updated[msgIndex],
							content: updated[msgIndex].content + content
						};
					} else {
						// Should not happen - handle properly
						updated[msgIndex] = {
							...updated[msgIndex],
							content: content
						};
					}
				} else {
					// Create new message for streaming chunk
					updated.push({
						id: messageId,
						role: 'assistant',
						content: content,
						sources: [],
						timestamp: Date.now(),
					});
				}
				return updated;
			});
		});

		return () => {
			if (cleanup) {
				cleanup();
			}
		};
	}, []);

	const handleSubmit = async (e) => {
		e.preventDefault();
		const trimmed = input.trim();
		if (!trimmed || loading) return;

		// Add user message
		const userMsg = {
			id: Date.now().toString(),
			role: 'user',
			content: trimmed,
			timestamp: Date.now(),
		};
		setMessages(prev => [...prev, userMsg]);
		setInput('');
		setLoading(true);

		try {
			const response = await RAGQuery(trimmed);

			// Find the assistant message and update it
			setMessages(prev => {
				const updated = [...prev];
				const msgIndex = updated.findIndex(m => m.id === userMsg.id);
				if (msgIndex >= 0) {
					updated[msgIndex] = {
						...updated[msgIndex],
						...response,
						timestamp: Date.now(),
					};
				} else {
					updated.push({
						id: userMsg.id,
						role: 'assistant',
						...response,
						timestamp: Date.now(),
					});
				}
				return updated;
			});

			// Process sources
			if (response.sources && response.sources.length > 0) {
				setMessages(prev => {
					const updated = [...prev];
					const msgIndex = updated.findIndex(m => m.id === userMsg.id);
					if (msgIndex >= 0) {
						updated[msgIndex] = {
							...updated[msgIndex],
							sources: response.sources,
						};
					}
					return updated;
				});
			}
		} catch (err) {
			console.error('Query failed:', err);
			setMessages(prev => [...prev, {
				id: Date.now().toString(),
				role: 'system',
				content: 'Error: ' + (err.message || 'Query failed'),
				timestamp: Date.now(),
			}]);
		} finally {
			setLoading(false);
			setStreamingStates({});
		}
	};

	const handleSourceClick = (source) => {
		// Emit event to open file in editor
		window.dispatchEvent(new CustomEvent('open-file', {
			detail: { path: source.path }
		}));
	};

	return (
		<div className="flex flex-col h-full bg-secondary">
			{/* Header */}
			<div className="px-4 py-3 bg-secondary border-b border-modifier-border flex justify-between items-center">
				<div className="flex items-center gap-2">
					<Sparkles className="text-obsidian-purple" size={18} />
					<h2 className="text-sm font-semibold text-normal">Knowledge Chat</h2>
				</div>
			</div>

			{/* Messages */}
			<div className="flex-1 overflow-y-auto p-4">
				{messages.length === 0 ? (
					<div className="flex flex-col items-center justify-center h-full text-center">
						<Sparkles className="text-faint mb-4" size={48} />
						<h3 className="text-lg font-medium text-muted mb-2">Ask Your Notes</h3>
						<p className="text-sm text-muted">
							Ask questions about your notes and get AI-powered answers with citations.
						</p>
					</div>
				) : (
					messages.map((msg) => (
						<div
							key={msg.id}
							className={clsx(
								"flex gap-3",
								msg.role === 'user' ? 'flex-row-reverse' : ''
							)}
						>
							{/* Avatar */}
							<div
								className={clsx(
									"w-8 h-8 rounded-full flex items-center justify-center shrink-0",
									msg.role === 'user'
										? 'bg-obsidian-purple text-white'
										: 'bg-obsidian-purple/10 text-obsidian-purple'
								)}
							>
								{msg.role === 'user' ? 'U' : <Sparkles size={16} />}
							</div>

							{/* Content */}
							<div
								className={clsx(
									"max-w-[80%] rounded-lg px-4 py-2",
									msg.role === 'user'
										? 'bg-obsidian-purple text-white'
										: 'bg-modifier-hover text-normal'
								)}
							>
								<p className="text-sm whitespace-pre-wrap break-words">
									{msg.content}
								</p>

								{/* Sources */}
								{msg.sources && msg.sources.length > 0 && (
									<div className="mt-3 pt-3 border-t border-modifier-border">
										<p className="text-xs font-medium text-muted mb-2">Sources:</p>
										<div className="space-y-1">
											{msg.sources.map((source) => (
												<button
													key={source.chunk_id}
													onClick={() => handleSourceClick(source)}
													className="flex items-start gap-2 p-2 rounded hover:bg-modifier-hover group transition-colors text-left"
												>
													<Sparkles size={14} className="text-faint group-hover:text-obsidian-purple shrink-0 mt-0.5" />
													<div className="flex-1 min-w-0">
														<div className="text-sm font-medium text-normal truncate">
															{source.title}
														</div>
														{source.heading && (
															<div className="text-xs text-muted truncate">
																{source.heading}
															</div>
														)}
														<div className="text-xs text-faint">
															{Math.round(source.similarity * 100)}% relevance
														</div>
													</div>
												</button>
											))}
										</div>
									</div>
								)}

								{/* Tokens */}
								{msg.tokens_used && (
									<div className="mt-1 text-xs text-faint">
										{msg.tokens_used} tokens
									</div>
								)}
							</div>
						</div>
				)))}
			</div>

			{/* Loading indicator */}
			{loading && (
				<div className="flex gap-3 px-4 py-3 bg-secondary border-b border-modifier-border">
					<div className="w-8 h-8 rounded-full flex items-center justify-center bg-secondary">
						<Loader2 size={16} className="animate-spin text-muted" />
					</div>
					<div className="text-sm text-muted">Thinking...</div>
				</div>
			)}

			{/* Input */}
			<form onSubmit={handleSubmit} className="p-4 border-t border-modifier-border">
				<div className="flex gap-2">
					<input
						ref={messagesEndRef}
						type="text"
						value={input}
						onChange={(e) => setInput(e.target.value)}
						onKeyDown={(e) => {
							if (e.key === 'Enter' && !e.shiftKey) {
								e.preventDefault();
								handleSubmit(e);
							}
						}}
						placeholder="Ask a question about your notes..."
						disabled={loading}
						className="flex-1 px-4 py-2 rounded-lg bg-primary-alt border border-modifier-border text-normal focus:border-obsidian-purple focus:outline-none"
					/>
					<button
						type="submit"
						disabled={loading || !input.trim()}
						className="px-4 py-2 bg-obsidian-purple hover:bg-obsidian-purple-hover disabled:opacity-50 disabled:cursor-not-allowed text-white rounded-lg transition-colors"
					>
						<Send size={18} />
					</button>
				</div>
			</form>

			{/* Messages end ref for auto-scroll */}
			<div ref={messagesEndRef} />
		</div>
	);
}
