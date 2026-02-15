package main

import (
	"context"
	"fmt"
	"notebit/pkg/config"
	"notebit/pkg/database"
	"notebit/pkg/graph"
	"strings"
)

// ============ SEMANTIC SEARCH API METHODS ============

// SimilarNote represents a note with similarity score for semantic search results
type SimilarNote struct {
	Path       string  `json:"path"`
	Title      string  `json:"title"`
	Content    string  `json:"content"`
	Heading    string  `json:"heading"`
	Similarity float32 `json:"similarity"`
	ChunkID    uint    `json:"chunk_id"`
}

// FindSimilar finds semantically similar notes based on content
func (a *App) FindSimilar(content string, limit int) ([]SimilarNote, error) {
	if a.ks == nil {
		return nil, fmt.Errorf("knowledge service not initialized - please open a folder first")
	}

	results, err := a.ks.FindSimilar(content, limit)
	if err != nil {
		return nil, err
	}

	// Convert to local struct to maintain API compatibility
	notes := make([]SimilarNote, len(results))
	for i, r := range results {
		notes[i] = SimilarNote{
			Path:       r.Path,
			Title:      r.Title,
			Content:    r.Content,
			Heading:    r.Heading,
			Similarity: r.Similarity,
			ChunkID:    r.ChunkID,
		}
	}
	return notes, nil
}

// GetSimilarityStatus returns the availability status of semantic search
func (a *App) GetSimilarityStatus() (map[string]interface{}, error) {
	if a.ks == nil {
		vectorEngine := "unknown"
		if a.dbm.IsInitialized() && a.dbm.Repository() != nil {
			engine := a.dbm.Repository().GetVectorEngine()
			if engine != "" {
				vectorEngine = engine
			}
		}
		return map[string]interface{}{
			"available":      false,
			"db_initialized": a.dbm.IsInitialized(),
			"ai_healthy":     false,
			"indexed_chunks": 0,
			"total_chunks":   0,
			"vector_engine":  vectorEngine,
		}, nil
	}

	return a.ks.GetSimilarityStatus()
}

// GetVectorSearchEngine returns current vector search engine and available options.
func (a *App) GetVectorSearchEngine() (map[string]interface{}, error) {
	if !a.dbm.IsInitialized() {
		return map[string]interface{}{
			"current":   "",
			"available": []string{database.VectorEngineBruteForce, database.VectorEngineSQLiteVec},
		}, nil
	}

	return map[string]interface{}{
		"current":   a.dbm.Repository().GetVectorEngine(),
		"available": []string{database.VectorEngineBruteForce, database.VectorEngineSQLiteVec},
	}, nil
}

// SetVectorSearchEngine updates vector search engine with fallback behavior.
func (a *App) SetVectorSearchEngine(engine string) (map[string]interface{}, error) {
	if !a.dbm.IsInitialized() {
		return nil, fmt.Errorf("database not initialized")
	}

	effective := a.dbm.Repository().SetVectorEngine(engine)
	a.cfg.SetVectorSearchEngine(effective)
	if err := a.cfg.Save(); err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"requested": engine,
		"effective": effective,
	}, nil
}

// ============ RAG CHAT API METHODS ============

// RAGQuery performs a RAG query
func (a *App) RAGQuery(query string) (map[string]interface{}, error) {
	if a.chatSvc == nil {
		return nil, fmt.Errorf("chat service not initialized")
	}
	defaultSession, err := a.chatSvc.EnsureDefaultSession()
	if err != nil {
		return nil, err
	}
	return a.RAGQueryWithSession(defaultSession.ID, query)
}

// RAGQueryWithSession performs a RAG query and persists the chat in a given session
func (a *App) RAGQueryWithSession(sessionID, query string) (map[string]interface{}, error) {
	if a.rag == nil {
		return nil, fmt.Errorf("RAG service not initialized")
	}
	if a.chatSvc == nil {
		return nil, fmt.Errorf("chat service not initialized")
	}
	if strings.TrimSpace(sessionID) == "" {
		return nil, fmt.Errorf("session id cannot be empty")
	}
	if strings.TrimSpace(query) == "" {
		return nil, fmt.Errorf("query cannot be empty")
	}

	if _, err := a.chatSvc.AppendMessage(sessionID, "user", query, nil, nil, "sent"); err != nil {
		return nil, err
	}

	response, err := a.rag.Query(context.Background(), query)
	if err != nil {
		_, _ = a.chatSvc.AppendMessage(sessionID, "system", "Error: "+err.Error(), nil, nil, "error")
		return nil, err
	}

	var tokensUsed *int
	if response.TokensUsed != nil {
		used := *response.TokensUsed
		tokensUsed = &used
	}
	if _, err := a.chatSvc.AppendMessage(sessionID, "assistant", response.Content, response.Sources, tokensUsed, "done"); err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"session_id":  sessionID,
		"message_id":  response.MessageID,
		"content":     response.Content,
		"sources":     response.Sources,
		"tokens_used": response.TokensUsed,
	}, nil
}

// GetRAGStatus returns the status of the RAG service
func (a *App) GetRAGStatus() (map[string]interface{}, error) {
	if a.rag == nil {
		return map[string]interface{}{
			"available":      false,
			"llm_provider":   "",
			"llm_model":      "",
			"database_ready": a.dbm.IsInitialized(),
		}, nil
	}

	status := a.rag.GetStatus()

	return map[string]interface{}{
		"available":      status.Available,
		"llm_provider":   status.LLMProvider,
		"llm_model":      status.LLMModel,
		"database_ready": status.DatabaseReady,
	}, nil
}

// ============ GRAPH API METHODS ============

// GetGraphData returns the knowledge graph data
func (a *App) GetGraphData() (*graph.GraphData, error) {
	if a.graph == nil {
		return &graph.GraphData{Nodes: []graph.Node{}, Links: []graph.Link{}}, nil
	}

	return a.graph.BuildGraph()
}

// GetGraphConfig returns the graph configuration
func (a *App) GetGraphConfig() (config.GraphConfig, error) {
	return a.cfg.GetGraphConfig(), nil
}

// SetGraphConfig sets the graph configuration
func (a *App) SetGraphConfig(minSimilarityThreshold float32, maxNodes int, showImplicitLinks bool) error {
	cfg := config.GraphConfig{
		MinSimilarityThreshold: minSimilarityThreshold,
		MaxNodes:               maxNodes,
		ShowImplicitLinks:      showImplicitLinks,
	}
	a.cfg.SetGraphConfig(cfg)
	return a.cfg.Save()
}
