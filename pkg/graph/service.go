package graph

import (
	"fmt"
	"regexp"
	"strings"
	"sync"

	"notebit/pkg/config"
	"notebit/pkg/database"
)

var wikiLinkRegex = regexp.MustCompile(`\[\[([^\]]+)\]\]`)
var tagRegex = regexp.MustCompile(`#([\w\p{L}-]+)`)

// Service handles knowledge graph operations
type Service struct {
	mu  sync.RWMutex
	db  *database.Manager
	cfg *config.Config
}

// Node represents a node in the knowledge graph
type Node struct {
	ID    string  `json:"id"`
	Label string  `json:"label"`
	Type  string  `json:"type"` // "file"
	Path  string  `json:"path"` // File path (for navigation)
	Size  int     `json:"size"` // Number of connections
	Val   float64 `json:"val"`  // Centrality/importance
}

// Link represents a link between nodes
type Link struct {
	Source   string  `json:"source"`
	Target   string  `json:"target"`
	Type     string  `json:"type"`     // "explicit" (wiki link), "implicit" (semantic)
	Strength float32 `json:"strength"` // Similarity score for implicit links
}

// GraphData represents the complete graph structure
type GraphData struct {
	Nodes []Node `json:"nodes"`
	Links []Link `json:"links"`
}

// NewService creates a new graph service
func NewService(db *database.Manager, cfg *config.Config) *Service {
	return &Service{
		db:  db,
		cfg: cfg,
	}
}

// BuildGraph constructs the knowledge graph
func (s *Service) BuildGraph() (*GraphData, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.db.IsInitialized() {
		return &GraphData{Nodes: []Node{}, Links: []Link{}}, nil
	}

	repo := s.db.Repository()
	graphConfig := s.cfg.GetGraphConfig()

	// Get all files with embeddings
	files, err := repo.ListFiles()
	if err != nil {
		return nil, fmt.Errorf("failed to list files: %w", err)
	}

	// Apply max nodes limit
	if graphConfig.MaxNodes > 0 && len(files) > graphConfig.MaxNodes {
		files = files[:graphConfig.MaxNodes]
	}

	nodes := s.buildNodes(files)
	links := s.buildLinks(files, repo, graphConfig)

	// Calculate node sizes based on connections
	nodeSizeMap := s.calculateNodeSizes(links)

	// Add tag nodes found in links
	tagNodes := make(map[string]bool)
	for _, link := range links {
		if strings.HasPrefix(link.Target, "tag:") {
			tagID := link.Target
			if _, exists := tagNodes[tagID]; !exists {
				tagName := strings.TrimPrefix(tagID, "tag:")
				nodes = append(nodes, Node{
					ID:    tagID,
					Label: "#" + tagName,
					Type:  "tag",
					Path:  "",
					Size:  0,
					Val:   1.0,
				})
				tagNodes[tagID] = true
			}
		}
	}

	for i := range nodes {
		nodes[i].Size = nodeSizeMap[nodes[i].ID]
		// Identify Concept nodes (high connectivity files)
		if nodes[i].Type == "file" && nodes[i].Size >= 5 {
			nodes[i].Type = "concept"
		}
	}

	return &GraphData{
		Nodes: nodes,
		Links: links,
	}, nil
}

// buildNodes creates nodes from files
func (s *Service) buildNodes(files []database.File) []Node {
	nodes := make([]Node, 0, len(files))

	for _, file := range files {
		nodes = append(nodes, Node{
			ID:    generateNodeID("file", file.Path),
			Label: file.Title,
			Type:  "file",
			Path:  file.Path,
			Size:  0, // Will be calculated based on links
			Val:   1.0,
		})
	}

	return nodes
}

// buildLinks creates links between nodes
func (s *Service) buildLinks(files []database.File, repo *database.Repository, graphConfig config.GraphConfig) []Link {
	var links []Link

	// 1. Extract explicit links (wiki-style [[links]])
	wikiLinks := s.extractWikiLinks(files)
	links = append(links, wikiLinks...)

	// 2. Extract tag links
	tagLinks := s.extractTagLinks(files)
	links = append(links, tagLinks...)

	// 3. Extract implicit links (semantic similarity)
	if graphConfig.ShowImplicitLinks {
		implicitLinks := s.extractImplicitLinks(files, repo, graphConfig.MinSimilarityThreshold)
		links = append(links, implicitLinks...)
	}

	return links
}

// extractTagLinks parses markdown for #tags
func (s *Service) extractTagLinks(files []database.File) []Link {
	var links []Link

	for _, file := range files {
		seenTags := make(map[string]bool)
		for _, chunk := range file.Chunks {
			matches := tagRegex.FindAllStringSubmatch(chunk.Content, -1)

			for _, match := range matches {
				if len(match) < 2 {
					continue
				}

				tagName := match[1]
				if seenTags[tagName] {
					continue
				}
				seenTags[tagName] = true

				link := Link{
					Source:   generateNodeID("file", file.Path),
					Target:   generateNodeID("tag", tagName), // Helper handles prefix? No, generateNodeID does.
					Type:     "tag",
					Strength: 1.0,
				}

				links = append(links, link)
			}
		}
	}
	return links
}

// extractWikiLinks parses markdown for [[wiki]] links
func (s *Service) extractWikiLinks(files []database.File) []Link {
	// Build path -> file map for quick lookup
	fileMap := make(map[string]*database.File)
	for i := range files {
		fileMap[files[i].Path] = &files[i]
	}

	var links []Link

	for _, file := range files {
		for _, chunk := range file.Chunks {
			matches := wikiLinkRegex.FindAllStringSubmatch(chunk.Content, -1)

			for _, match := range matches {
				if len(match) < 2 {
					continue
				}

				targetName := match[1]
				// Try to find matching file by title or path
				for _, targetFile := range files {
					if s.filesMatch(targetName, &targetFile) {
						link := Link{
							Source:   generateNodeID("file", file.Path),
							Target:   generateNodeID("file", targetFile.Path),
							Type:     "explicit",
							Strength: 1.0,
						}

						// Avoid duplicate links
						if !linkExists(links, link) {
							links = append(links, link)
						}
						break
					}
				}
			}
		}
	}

	return links
}

// extractImplicitLinks finds semantically similar files
func (s *Service) extractImplicitLinks(files []database.File, repo *database.Repository, threshold float32) []Link {
	var links []Link

	// For each file, find semantically similar files
	for _, file := range files {
		if len(file.Chunks) == 0 {
			continue
		}

		// Get first chunk as representative (or find the chunk with embedding)
		var embedding []float32
		for _, chunk := range file.Chunks {
			if len(chunk.Embedding) > 0 {
				embedding = chunk.Embedding
				break
			}
		}

		if len(embedding) == 0 {
			continue
		}

		// Search for similar files
		similar, err := repo.SearchSimilar(embedding, 10)
		if err != nil {
			continue
		}

		for _, sim := range similar {
			// Skip self and low similarity
			if sim.File.Path == file.Path || sim.Similarity < threshold {
				continue
			}

			link := Link{
				Source:   generateNodeID("file", file.Path),
				Target:   generateNodeID("file", sim.File.Path),
				Type:     "implicit",
				Strength: sim.Similarity,
			}

			// Avoid duplicate links
			if !linkExists(links, link) {
				links = append(links, link)
			}
		}
	}

	return links
}

// filesMatch checks if a target name matches a file (by title or path)
func (s *Service) filesMatch(targetName string, file *database.File) bool {
	// Exact title match
	if file.Title == targetName {
		return true
	}

	// Path contains match (case insensitive)
	if strings.Contains(strings.ToLower(file.Path), strings.ToLower(targetName)) {
		return true
	}

	// Filename match
	if strings.Contains(strings.ToLower(file.Path), strings.ToLower(targetName)) {
		return true
	}

	return false
}

// linkExists checks if a link already exists in the list
func linkExists(links []Link, link Link) bool {
	for _, existing := range links {
		if existing.Source == link.Source && existing.Target == link.Target {
			return true
		}
	}
	return false
}

// calculateNodeSizes calculates the size (number of connections) for each node
func (s *Service) calculateNodeSizes(links []Link) map[string]int {
	sizeMap := make(map[string]int)

	// Count connections for each node
	for _, link := range links {
		sizeMap[link.Source]++
		sizeMap[link.Target]++
	}

	return sizeMap
}

// generateNodeID creates a unique node ID
func generateNodeID(typ, path string) string {
	return fmt.Sprintf("%s:%s", typ, path)
}
