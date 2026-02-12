package ai

import (
	"strings"
	"unicode"
)

// FixedSizeChunker implements fixed-size window chunking with overlap
type FixedSizeChunker struct {
	chunkSize    int
	chunkOverlap int
	minChunkSize int
}

// NewFixedSizeChunker creates a new fixed-size chunker
func NewFixedSizeChunker(chunkSize, chunkOverlap, minChunkSize int) *FixedSizeChunker {
	return &FixedSizeChunker{
		chunkSize:    chunkSize,
		chunkOverlap: chunkOverlap,
		minChunkSize: minChunkSize,
	}
}

// Chunk splits text into fixed-size chunks with overlap
func (c *FixedSizeChunker) Chunk(text string) ([]TextChunk, error) {
	if text == "" {
		return nil, nil
	}

	var chunks []TextChunk
	runes := []rune(text)
	textLen := len(runes)

	if textLen <= c.chunkSize {
		return []TextChunk{{Content: text, Index: 0}}, nil
	}

	start := 0
	overlap := c.chunkOverlap

	for start < textLen {
		end := start + c.chunkSize
		if end > textLen {
			end = textLen
		}

		chunk := string(runes[start:end])

		// Only add if it meets minimum size requirement
		if len([]rune(chunk)) >= c.minChunkSize || end == textLen {
			chunks = append(chunks, TextChunk{
				Content: chunk,
				Index:   len(chunks),
			})
		}

		// Move start position, accounting for overlap
		start = end - overlap
		if start < 0 {
			start = 0
		}

		// Prevent infinite loop when overlap equals chunk size
		if start >= end {
			start = end
		}
	}

	return chunks, nil
}

// Name returns the strategy name
func (c *FixedSizeChunker) Name() string {
	return "fixed"
}

// HeadingChunker implements heading-based chunking
// Splits text at markdown headings, preserving heading context
type HeadingChunker struct {
	maxChunkSize     int
	minChunkSize     int
	preserveHeading  bool
	headingSeparator string
}

// NewHeadingChunker creates a new heading-based chunker
func NewHeadingChunker(maxChunkSize, minChunkSize int, preserveHeading bool, headingSeparator string) *HeadingChunker {
	if headingSeparator == "" {
		headingSeparator = "\n\n"
	}

	return &HeadingChunker{
		maxChunkSize:     maxChunkSize,
		minChunkSize:     minChunkSize,
		preserveHeading:  preserveHeading,
		headingSeparator: headingSeparator,
	}
}

// Chunk splits text based on markdown headings
func (c *HeadingChunker) Chunk(text string) ([]TextChunk, error) {
	if text == "" {
		return nil, nil
	}

	// Find all heading boundaries
	chunks := c.extractHeadingChunks(text)

	var result []TextChunk
	currentChunk := &TextChunk{Index: 0}
	var currentContent strings.Builder
	var currentHeading string

	for _, chunk := range chunks {
		// If this chunk has a heading, store it
		if chunk.Heading != "" {
			currentHeading = chunk.Heading
		}

		// Check if adding this content would exceed max size
		proposedSize := currentContent.Len() + len(chunk.Content)
		if proposedSize > c.maxChunkSize && currentContent.Len() > 0 {
			// Save current chunk and start a new one
			if c.contentMeetsMinimum(currentContent.String()) {
				currentChunk.Content = currentContent.String()
				if c.preserveHeading {
					currentChunk.Heading = currentHeading
				}
				result = append(result, *currentChunk)
			}

			// Start new chunk
			currentChunk = &TextChunk{Index: len(result)}
			currentContent.Reset()

			// Include heading in new chunk if preserving
			if c.preserveHeading && currentHeading != "" {
				currentContent.WriteString(currentHeading)
				currentContent.WriteString(c.headingSeparator)
			}
		} else if currentContent.Len() == 0 && c.preserveHeading && currentHeading != "" {
			// First content in a new chunk with heading
			currentContent.WriteString(currentHeading)
			currentContent.WriteString(c.headingSeparator)
		}

		currentContent.WriteString(chunk.Content)
		if chunk.Content != "" && !strings.HasSuffix(chunk.Content, "\n") {
			currentContent.WriteString("\n")
		}
	}

	// Don't forget the last chunk
	if currentContent.Len() > 0 {
		if c.contentMeetsMinimum(currentContent.String()) {
			currentChunk.Content = currentContent.String()
			if c.preserveHeading {
				currentChunk.Heading = currentHeading
			}
			result = append(result, *currentChunk)
		}
	}

	// Handle edge case where no chunks were created
	if len(result) == 0 && len(text) > 0 {
		result = append(result, TextChunk{
			Content: text,
			Index:   0,
		})
	}

	return result, nil
}

// headingChunk represents a segment with optional heading
type headingChunk struct {
	Heading string
	Content string
}

// extractHeadingChunks splits text into segments by headings
func (c *HeadingChunker) extractHeadingChunks(text string) []headingChunk {
	var chunks []headingChunk
	var currentContent strings.Builder
	var currentHeading string
	lineIndex := 0
	start := 0
	for start <= len(text) {
		end := strings.IndexByte(text[start:], '\n')
		var line string
		if end == -1 {
			line = text[start:]
		} else {
			line = text[start : start+end]
		}

		trimmed := strings.TrimLeft(line, " \t")

		// Check if this is a markdown heading (#, ##, ###, etc.)
		if len(trimmed) >= 2 && trimmed[0] == '#' && (trimmed[1] == ' ' || trimmed[1] == '#') {
			// Save previous chunk
			if currentContent.Len() > 0 {
				chunks = append(chunks, headingChunk{
					Heading: currentHeading,
					Content: strings.TrimRight(currentContent.String(), "\n"),
				})
				currentContent.Reset()
			}

			// Set new heading (get the full heading line)
			currentHeading = line
		} else {
			// Add line to current content
			if lineIndex > 0 {
				currentContent.WriteString("\n")
			}
			currentContent.WriteString(line)
		}

		if end == -1 {
			break
		}
		start = start + end + 1
		lineIndex++
	}

	// Don't forget the last chunk
	if currentContent.Len() > 0 {
		chunks = append(chunks, headingChunk{
			Heading: currentHeading,
			Content: strings.TrimRight(currentContent.String(), "\n"),
		})
	}

	// If no headings were found, treat entire text as one chunk
	if len(chunks) == 0 || (len(chunks) == 1 && chunks[0].Heading == "" && chunks[0].Content == "") {
		return []headingChunk{{Content: text}}
	}

	return chunks
}

// contentMeetsMinimum checks if content meets minimum size requirement
func (c *HeadingChunker) contentMeetsMinimum(content string) bool {
	return len([]rune(content)) >= c.minChunkSize
}

// Name returns the strategy name
func (c *HeadingChunker) Name() string {
	return "heading"
}

// SlidingWindowChunker implements sliding window chunking
// Creates overlapping chunks that slide through the text
type SlidingWindowChunker struct {
	windowSize   int
	step         int
	minChunkSize int
}

// NewSlidingWindowChunker creates a new sliding window chunker
func NewSlidingWindowChunker(windowSize, step, minChunkSize int) *SlidingWindowChunker {
	return &SlidingWindowChunker{
		windowSize:   windowSize,
		step:         step,
		minChunkSize: minChunkSize,
	}
}

// Chunk splits text using a sliding window
func (c *SlidingWindowChunker) Chunk(text string) ([]TextChunk, error) {
	if text == "" {
		return nil, nil
	}

	runes := []rune(text)
	textLen := len(runes)

	if textLen <= c.windowSize {
		return []TextChunk{{Content: text, Index: 0}}, nil
	}

	var chunks []TextChunk

	for start := 0; start < textLen; start += c.step {
		end := start + c.windowSize
		if end > textLen {
			end = textLen
		}

		chunk := string(runes[start:end])

		// Only add if it meets minimum size requirement
		if len([]rune(chunk)) >= c.minChunkSize || end == textLen {
			chunks = append(chunks, TextChunk{
				Content: chunk,
				Index:   len(chunks),
			})
		}

		// Stop if we've reached the end
		if end == textLen {
			break
		}
	}

	return chunks, nil
}

// Name returns the strategy name
func (c *SlidingWindowChunker) Name() string {
	return "sliding"
}

// SentenceChunker implements sentence-based chunking
// Splits text into chunks containing complete sentences
type SentenceChunker struct {
	maxChunkSize     int
	minChunkSize     int
	overlapSentences int // Number of sentences to overlap between chunks
}

// NewSentenceChunker creates a new sentence-based chunker
func NewSentenceChunker(maxChunkSize, minChunkSize, overlapSentences int) *SentenceChunker {
	return &SentenceChunker{
		maxChunkSize:     maxChunkSize,
		minChunkSize:     minChunkSize,
		overlapSentences: overlapSentences,
	}
}

// Chunk splits text into sentence-based chunks
func (c *SentenceChunker) Chunk(text string) ([]TextChunk, error) {
	if text == "" {
		return nil, nil
	}

	// Split into sentences
	sentences := c.splitSentences(text)
	if len(sentences) == 0 {
		return nil, nil
	}

	var chunks []TextChunk
	var currentChunk strings.Builder
	var overlapSentences []string // Sentences to carry over to next chunk

	for i, sentence := range sentences {
		testSize := currentChunk.Len() + len(sentence)
		shouldBreak := testSize > c.maxChunkSize && currentChunk.Len() > 0

		if shouldBreak {
			// Save current chunk if it meets minimum
			content := strings.TrimSpace(currentChunk.String())
			if len([]rune(content)) >= c.minChunkSize {
				chunks = append(chunks, TextChunk{
					Content: content,
					Index:   len(chunks),
				})
			}

			// Start new chunk with overlap
			currentChunk.Reset()
			for _, os := range overlapSentences {
				currentChunk.WriteString(os)
				currentChunk.WriteString(" ")
			}
		}

		// Add current sentence
		currentChunk.WriteString(sentence)
		currentChunk.WriteString(" ")

		// Track sentences for overlap (keep last N sentences)
		if i < len(sentences)-1 { // Don't include the last sentence in overlap
			overlapSentences = append(overlapSentences, sentence)
			if len(overlapSentences) > c.overlapSentences {
				overlapSentences = overlapSentences[1:]
			}
		}
	}

	// Don't forget the last chunk
	if currentChunk.Len() > 0 {
		content := strings.TrimSpace(currentChunk.String())
		if len([]rune(content)) >= c.minChunkSize {
			chunks = append(chunks, TextChunk{
				Content: content,
				Index:   len(chunks),
			})
		}
	}

	return chunks, nil
}

// splitSentences splits text into sentences
func (c *SentenceChunker) splitSentences(text string) []string {
	var sentences []string
	runes := []rune(text)
	start := 0

	for i := 0; i < len(runes); i++ {
		r := runes[i]
		if r != '.' && r != '!' && r != '?' {
			continue
		}

		boundary, nextNonSpace := isSentenceBoundary(runes, i)
		if !boundary {
			continue
		}

		token := tokenBeforePunctuation(runes, i)
		if token != "" {
			lower := strings.ToLower(token)
			if _, ok := sentenceAbbreviations[lower]; ok {
				continue
			}
			if len([]rune(token)) == 1 && nextNonSpace != 0 && unicode.IsUpper(nextNonSpace) {
				continue
			}
		}

		sentence := strings.TrimSpace(string(runes[start : i+1]))
		if sentence != "" {
			sentences = append(sentences, sentence)
		}
		start = i + 1
	}

	if start < len(runes) {
		sentence := strings.TrimSpace(string(runes[start:]))
		if sentence != "" {
			sentences = append(sentences, sentence)
		}
	}

	return sentences
}

var sentenceAbbreviations = map[string]struct{}{
	"mr":   {},
	"mrs":  {},
	"ms":   {},
	"dr":   {},
	"prof": {},
	"sr":   {},
	"jr":   {},
	"vs":   {},
	"etc":  {},
	"e.g":  {},
	"i.e":  {},
	"fig":  {},
	"no":   {},
	"vol":  {},
	"al":   {},
	"u.s":  {},
}

func isSentenceBoundary(runes []rune, idx int) (bool, rune) {
	if idx+1 >= len(runes) {
		return true, 0
	}

	j := idx + 1
	for j < len(runes) {
		r := runes[j]
		if r == '"' || r == '\'' || r == ')' || r == ']' || r == '}' {
			j++
			continue
		}
		break
	}

	if j >= len(runes) {
		return true, 0
	}

	if unicode.IsSpace(runes[j]) {
		k := j
		for k < len(runes) && unicode.IsSpace(runes[k]) {
			k++
		}
		if k >= len(runes) {
			return true, 0
		}
		return true, runes[k]
	}

	return false, 0
}

func tokenBeforePunctuation(runes []rune, idx int) string {
	start := idx - 1
	for start >= 0 && !unicode.IsSpace(runes[start]) {
		start--
	}
	start++
	if start >= idx {
		return ""
	}
	token := strings.Trim(string(runes[start:idx]), "\"'()[]{}")
	return strings.Trim(token, ".")
}

// Name returns the strategy name
func (c *SentenceChunker) Name() string {
	return "sentence"
}
