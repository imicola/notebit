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
	mu             sync.RWMutex
	db             *database.Manager
	cfg            *config.Config
	cachedGraph    *GraphData
	cachedRevision uint64
	cachedConfig   config.GraphConfig
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
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.db.IsInitialized() {
		return &GraphData{Nodes: []Node{}, Links: []Link{}}, nil
	}

	repo := s.db.Repository()
	graphConfig := s.cfg.GetGraphConfig()
	revision := repo.GetRevision()
	if s.cachedGraph != nil && s.cachedRevision == revision && s.cachedConfig == graphConfig {
		return s.cachedGraph, nil
	}

	// Get all files with embeddings
	files, err := repo.ListFilesWithChunks()
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

	data := &GraphData{
		Nodes: nodes,
		Links: links,
	}
	s.cachedGraph = data
	s.cachedRevision = revision
	s.cachedConfig = graphConfig
	return data, nil
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
			lines := strings.Split(chunk.Content, "\n")
			for _, line := range lines {
				matches := tagRegex.FindAllStringSubmatchIndex(line, -1)
				if len(matches) == 0 {
					continue
				}

				trimmedLine := strings.TrimLeft(line, " \t")
				leading := len(line) - len(trimmedLine)

				for _, match := range matches {
					if len(match) < 4 {
						continue
					}

					tagStart := match[0]
					if tagStart == leading {
						continue
					}

					tagName := line[match[2]:match[3]]
					if seenTags[tagName] {
						continue
					}
					seenTags[tagName] = true

					link := Link{
						Source:   generateNodeID("file", file.Path),
						Target:   generateNodeID("tag", tagName),
						Type:     "tag",
						Strength: 1.0,
					}

					links = append(links, link)
				}
			}
		}
	}
	return links
}

// extractWikiLinks parses markdown for [[wiki]] links
func (s *Service) extractWikiLinks(files []database.File) []Link {
	var links []Link

	for _, file := range files {
		for _, chunk := range file.Chunks {
			matches := wikiLinkRegex.FindAllStringSubmatch(chunk.Content, -1)

			for _, match := range matches {
				if len(match) < 2 {
					continue
				}

				targetName := match[1]
				if idx := strings.Index(targetName, "|"); idx >= 0 {
					targetName = targetName[:idx]
				}
				if idx := strings.Index(targetName, "#"); idx >= 0 {
					targetName = targetName[:idx]
				}
				targetName = strings.TrimSpace(targetName)
				if targetName == "" {
					continue
				}
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

	queryVectors := make([][]float32, 0, len(files))
	queryPaths := make([]string, 0, len(files))
	for _, file := range files {
		if len(file.Chunks) == 0 {
			continue
		}

		var embedding []float32
		for _, chunk := range file.Chunks {
			emb := chunk.GetEmbedding()
			if len(emb) > 0 {
				embedding = emb
				break
			}
		}

		if len(embedding) == 0 {
			continue
		}

		queryVectors = append(queryVectors, embedding)
		queryPaths = append(queryPaths, file.Path)
	}

	if len(queryVectors) == 0 {
		return links
	}

	batchResults, err := repo.SearchSimilarBatch(queryVectors, 10)
	if err != nil {
		return links
	}

	for i, similar := range batchResults {
		sourcePath := queryPaths[i]
		for _, sim := range similar {
			if sim.File == nil {
				continue
			}

			if sim.File.Path == sourcePath || sim.Similarity < threshold {
				continue
			}

			link := Link{
				Source:   generateNodeID("file", sourcePath),
				Target:   generateNodeID("file", sim.File.Path),
				Type:     "implicit",
				Strength: sim.Similarity,
			}

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
