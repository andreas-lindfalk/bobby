package agent

import (
	"context"
	"sync"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
)

// SessionInput represents input with a session ID for memory.
type SessionInput struct {
	SessionID string `json:"session_id"`
	Message   string `json:"message"`
}

// SessionOutput represents output with session context.
type SessionOutput struct {
	SessionID string `json:"session_id"`
	Response  string `json:"response"`
}

// sessionStore holds conversation history per session (in-memory).
type sessionStore struct {
	mu       sync.RWMutex
	sessions map[string][]Message
}

var sessions = &sessionStore{
	sessions: make(map[string][]Message),
}

func (s *sessionStore) get(id string) []Message {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if history, ok := s.sessions[id]; ok {
		// Return a copy to avoid race conditions
		result := make([]Message, len(history))
		copy(result, history)
		return result
	}
	return nil
}

func (s *sessionStore) append(id string, userMsg, modelMsg string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessions[id] = append(s.sessions[id],
		Message{Role: "user", Content: userMsg},
		Message{Role: "model", Content: modelMsg},
	)
}

// ChatInput represents user input to the chat flow.
type ChatInput struct {
	Message string `json:"message"`
	// Optional context from previous turns
	History []Message `json:"history,omitempty"`
}

// Message represents a single message in the conversation.
type Message struct {
	Role    string `json:"role"` // "user" or "model"
	Content string `json:"content"`
}

// ChatOutput represents the agent's response.
type ChatOutput struct {
	Response string    `json:"response"`
	History  []Message `json:"history"`
}

// DefineCarImportChatFlow defines the main conversational flow for car import assistance.
// It creates a streaming flow that can be used for real-time responses.
func (a *Agent) DefineCarImportChatFlow() {
	// Define the streaming chat flow
	genkit.DefineStreamingFlow(a.g, "carImportChat",
		func(ctx context.Context, input ChatInput, sendChunk func(context.Context, string) error) (*ChatOutput, error) {
			// Build message history
			messages := buildMessages(input.History, input.Message)

			// Create streaming callback
			var streamCallback ai.ModelStreamCallback
			if sendChunk != nil {
				streamCallback = func(ctx context.Context, chunk *ai.ModelResponseChunk) error {
					return sendChunk(ctx, chunk.Text())
				}
			}

			// Generate response with tools
			resp, err := genkit.Generate(ctx, a.g,
				ai.WithModelName(a.config.ModelName),
				ai.WithSystem(SystemPrompt()),
				ai.WithMessages(messages...),
				ai.WithTools(a.tools...),
				ai.WithStreaming(streamCallback),
			)
			if err != nil {
				return nil, err
			}

			// Build output
			output := &ChatOutput{
				Response: resp.Text(),
				History:  appendToHistory(input.History, input.Message, resp.Text()),
			}

			return output, nil
		},
	)
}

// DefineSimpleQueryFlow defines a non-streaming flow for simple queries (stateless).
func (a *Agent) DefineSimpleQueryFlow() {
	genkit.DefineFlow(a.g, "carImportQuery",
		func(ctx context.Context, query string) (string, error) {
			resp, err := genkit.Generate(ctx, a.g,
				ai.WithModelName(a.config.ModelName),
				ai.WithSystem(SystemPrompt()),
				ai.WithPrompt(query),
				ai.WithTools(a.tools...),
			)
			if err != nil {
				return "", err
			}
			return resp.Text(), nil
		},
	)
}

// DefineSessionFlow defines a flow with server-side session memory.
// Use session_id to maintain conversation context across requests.
func (a *Agent) DefineSessionFlow() {
	genkit.DefineFlow(a.g, "chat",
		func(ctx context.Context, input SessionInput) (*SessionOutput, error) {
			// Get existing history for this session
			history := sessions.get(input.SessionID)

			// Build messages from history
			messages := buildMessages(history, input.Message)

			// Generate response with tools
			resp, err := genkit.Generate(ctx, a.g,
				ai.WithModelName(a.config.ModelName),
				ai.WithSystem(SystemPrompt()),
				ai.WithMessages(messages...),
				ai.WithTools(a.tools...),
			)
			if err != nil {
				return nil, err
			}

			responseText := resp.Text()

			// Store in session
			sessions.append(input.SessionID, input.Message, responseText)

			return &SessionOutput{
				SessionID: input.SessionID,
				Response:  responseText,
			}, nil
		},
	)
}

// buildMessages converts history and current message into Genkit message format.
func buildMessages(history []Message, currentMessage string) []*ai.Message {
	messages := make([]*ai.Message, 0, len(history)+1)

	for _, msg := range history {
		switch msg.Role {
		case "user":
			messages = append(messages, ai.NewUserTextMessage(msg.Content))
		case "model":
			messages = append(messages, ai.NewModelTextMessage(msg.Content))
		}
	}

	// Add current user message
	messages = append(messages, ai.NewUserTextMessage(currentMessage))

	return messages
}

// appendToHistory adds the current exchange to history.
func appendToHistory(history []Message, userMessage, modelResponse string) []Message {
	newHistory := make([]Message, len(history), len(history)+2)
	copy(newHistory, history)
	newHistory = append(newHistory,
		Message{Role: "user", Content: userMessage},
		Message{Role: "model", Content: modelResponse},
	)
	return newHistory
}
