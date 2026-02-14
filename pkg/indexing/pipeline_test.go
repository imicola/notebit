package indexing

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"notebit/pkg/ai"
	"notebit/pkg/config"
	"notebit/pkg/database"
	"notebit/pkg/files"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestIndexingPipeline_ConcurrentSamePath(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "notebit-indexing-pipeline-*")
	if err != nil {
		t.Fatalf("create temp dir failed: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/embeddings" {
			http.NotFound(w, r)
			return
		}

		var req struct {
			Model string `json:"model"`
			Input string `json:"input"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		sum := 0
		for _, ch := range req.Input {
			sum += int(ch)
		}

		resp := map[string]interface{}{
			"embedding": []float32{float32(len(req.Input)%13 + 1), float32(sum%17 + 1), 0.5},
			"model":     req.Model,
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	database.Reset()
	dbManager := database.GetInstance()
	if err := dbManager.Init(tmpDir); err != nil {
		t.Fatalf("database init failed: %v", err)
	}
	defer func() {
		_ = dbManager.Close()
		database.Reset()
	}()

	fm := files.NewManager()
	if err := fm.SetBasePath(tmpDir); err != nil {
		t.Fatalf("set base path failed: %v", err)
	}

	path := "same.md"
	content := "# Title\n\n" + strings.Repeat("alpha beta gamma ", 40)
	if err := os.WriteFile(filepath.Join(tmpDir, path), []byte(content), 0644); err != nil {
		t.Fatalf("write file failed: %v", err)
	}

	cfg := config.New()
	cfg.SetOllamaConfig(server.URL, "nomic-embed-text", 3)
	cfg.SetProvider("ollama")
	cfg.SetEmbeddingModel("nomic-embed-text")

	aiService := ai.NewService(cfg)
	if err := aiService.Initialize(); err != nil {
		t.Fatalf("ai initialize failed: %v", err)
	}

	pipeline := NewPipeline(aiService, dbManager.Repository(), fm)
	pipeline.Start()
	defer pipeline.Stop()

	opts := IndexOptions{
		SkipIfUnchanged:       true,
		FallbackToMetadataOnly: true,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var wg sync.WaitGroup
	errCh := make(chan error, 12)
	for i := 0; i < 12; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			errCh <- pipeline.IndexFile(ctx, path, opts)
		}()
	}
	wg.Wait()
	close(errCh)

	for err := range errCh {
		if err != nil {
			t.Fatalf("concurrent IndexFile failed: %v", err)
		}
	}

	repo := pipeline.Repository()
	indexedFile, err := repo.GetFileByPath(path)
	if err != nil {
		t.Fatalf("GetFileByPath failed: %v", err)
	}

	chunks, err := repo.GetChunksByFileID(indexedFile.ID)
	if err != nil {
		t.Fatalf("GetChunksByFileID failed: %v", err)
	}
	if len(chunks) == 0 {
		t.Fatalf("expected chunks after concurrent indexing")
	}

	needsReindex, err := repo.FileNeedsIndexing(path, content)
	if err != nil {
		t.Fatalf("FileNeedsIndexing failed: %v", err)
	}
	if needsReindex {
		t.Fatalf("expected no reindex needed after successful concurrent indexing")
	}
}
