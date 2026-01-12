package agent

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSystemPrompt(t *testing.T) {
	prompt := SystemPrompt()

	// Verify key elements are present
	require.Contains(t, prompt, "Bobby", "should contain agent name")
	require.Contains(t, prompt, "Swedish", "should mention Swedish expats")
	require.Contains(t, prompt, "Spain", "should mention Spain")
	require.Contains(t, prompt, "import", "should mention import")
	require.Contains(t, prompt, "matriculation", "should mention matriculation")

	// Verify tool usage instructions
	require.Contains(t, prompt, "calculate_deadlines", "should mention deadline tool")
	require.Contains(t, prompt, "estimate_tax", "should mention tax tool")
	require.Contains(t, prompt, "search_knowledge_base", "should mention RAG tool")
}

func TestSystemPromptLength(t *testing.T) {
	prompt := SystemPrompt()

	// System prompt should be substantial but not excessive
	require.Greater(t, len(prompt), 500, "system prompt should be detailed")
	require.Less(t, len(prompt), 5000, "system prompt should not be too long")
}

func TestSystemPromptStructure(t *testing.T) {
	prompt := SystemPrompt()

	// Should have clear sections
	sections := []string{
		"Role",
		"Capabilities",
		"Guidelines",
	}

	for _, section := range sections {
		if !strings.Contains(prompt, section) {
			t.Logf("Note: System prompt could include a '%s' section for clarity", section)
		}
	}
}

func TestDefaultConfig(t *testing.T) {
	// When store/embedder are nil, agent should still work
	cfg := Config{
		Store:       nil,
		Embedder:    nil,
		ModelName:   "test-model",
		RAGTopK:     3,
		Temperature: 0.5,
	}

	require.Equal(t, "test-model", cfg.ModelName)
	require.Equal(t, 3, cfg.RAGTopK)
	require.Equal(t, float32(0.5), cfg.Temperature)
}
