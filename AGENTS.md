# Orihuelacosta.ai — Agent Guidelines

## 1. Project Overview

- **Name:** Orihuelacosta.ai
- **Concept:** A Vertical AI Agent & Managed Marketplace for expats on Costa Blanca.
- **Core Value Prop:** The "Habeno for living." A digital bridge between Spanish bureaucracy/service providers and Northern European homeowners.
- **Key Differentiator:** Automating cross-border tax compliance (Swedish ROT/RUT, German Handwerkerleistungen, UK CGT compliance) and acting as a "Managed Marketplace" with escrow and quality guarantee.

## 2. Technical Stack

- **Language:** Go (Golang) — Preferred for performance, type safety, and concurrency.
- **Orchestration:** Firebase Genkit (Go SDK).
- **Architectural Pattern:** Agentic Workflows (Multi-agent system), RAG (Retrieval-Augmented Generation).
- **Interface:** Model Context Protocol (MCP) for real-time tool use (weather, currency, local govt APIs).
- **Database:** Vector Database (e.g., Supabase/pgvector or ChromaDB).
- **Models:** Claude Opus 4.5 / Gemini 1.5 Pro for reasoning; Gemini 1.5 Flash for ingestion/summarization.

## 3. Developer Persona & Context

- **Lead Developer:** Swedish male, 20+ years of software development experience.
- **Local Context:** Resident in Orihuela Costa, Spain. Deep understanding of local expat pain points (cars, renovation, bureaucracy).
- **Expertise Level:** Senior. Expect concise, high-quality idiomatic Go code. No need for basic explanations, focus on architectural patterns and complex logic.

## 4. Domain Logic (The "Vertical" in Vertical AI)

- **Tax Engine:** Must handle logic for Swedish ROT/RUT (30-50% deduction), German tax laws (§ 35a EStG), and Spanish invoicing requirements (Factura Legal).
- **Managed Marketplace:** Logic for Escrow payments, milestone verification (via vision/LLM), and contractor vetting.
- **Ingestion Pipeline:** Automated scraping of BOE (Spanish state bulletin), local news, and municipal data.

## 5. Interaction Guidelines for Claude

- **Role:** Senior Software Architect & Business Strategist.
- **Style:** Empathetic but intellectually honest. Focus on "Managed Marketplace" logic and "Vertical AI" advantages.
- **Priority:** Always consider the "Hot Lead" aspect and transactional security (Escrow/Contractor verification).
- **Language:** Discussion in Swedish is fine, but code, comments, and documentation should be in English unless specified.

## 6. Project Structure (Conventions)

```
/cmd              # Application entrypoints
/internal
  /agent          # Agent definitions and orchestration
  /logic          # Domain logic (tax, marketplace, etc.)
  /tools          # MCP tools (deadline calculators, API wrappers)
  /flows          # Genkit flows
/pkg              # Shared utilities
/data             # Vector DB seeds, RAG documents
```

## 7. Architectural Philosophy & Data Strategy

To ensure scalability and minimize manual maintenance, the system follows a three-layer intelligence model:

- **Layer 1:** Generalist Intelligence (Zero-Touch): Use LLM base knowledge + Google Search/Tavily for non-critical expat questions (e.g., "Where is the nearest pharmacy?"). No custom Go logic required.

- **Layer 2:** Autonomous Scout (Agentic RAG): An automated Go-based pipeline that monitors a curated list of "Master Sources" (Official bulletins, local news).
  - **Logic:** Periodically scrape, summarize with Gemini Flash, and upsert to Vector DB.
  - **Goal:** Handle 80% of local info updates without developer intervention.

- **Layer 3:** Vertical Business Flows (Deep Logic): Hard-coded, high-precision Go modules for high-value transactions (Car Import, RUT/ROT, Escrow).
  - **Logic:** Detailed state machines, complex tax calculations, and partner integrations.
  - **Goal:** Monetization and high-stakes reliability.

Claude's Instruction: When suggesting features or code, always categorize the task into one of these layers. Avoid over-engineering Layer 1/2 tasks and focus maximum architectural rigor on Layer 3.

## 8. Code Quality Expectations

- **Testing:** Unit tests for all domain logic; integration tests for flows.
- **Error Handling:** Use explicit error returns (Go style). Wrap errors with context.
- **Commits:** Conventional commits (`feat:`, `fix:`, `docs:`, `refactor:`).

## 9. Testing Standards

Testing is critical. All code changes must be covered by proper tests with thorough assertions. We follow these principles:

### Test Framework & Assertions

- **Use `github.com/stretchr/testify/suite`** for integration tests requiring shared setup/teardown.
- **ALWAYS use `s.Require()` (not `s.Assert()`)**. Fail-fast on first error — no point continuing if preconditions fail.
- Example: `s.Require().NoError(err)`, `s.Require().Len(items, 3)`, `s.Require().Equal(expected, actual)`

### Test Isolation & Cleanup

- **Defer cleanup immediately after insert**: Insert test data, then `defer s.deleteXxx(id)` on the next line.
- **Never leave test data behind**: Each test cleans up its own data, so tests remain independent.
- **Add "cleanup verification" tests**: Explicitly verify that previous tests' data was removed.

### Test Suite Pattern (Integration Tests)

```go
type MySuite struct {
    suite.Suite
    store *store.Store
    ctx   context.Context
}

func (s *MySuite) SetupSuite() {
    // Start testcontainer, create store, run migrations ONCE
}

func (s *MySuite) TearDownSuite() {
    // Close connections, terminate container
}

func TestMySuite(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test suite")
    }
    suite.Run(t, new(MySuite))
}
```

### Test Structure

1. **Arrange**: Set up test data with helper methods (`s.insertTestDocument(...)`)
2. **Defer cleanup**: `defer s.deleteDocument(id)` — immediately after insert
3. **Act**: Call the method under test
4. **Assert**: Thorough assertions with `s.Require()` — check all relevant fields

### Integration Tests with testcontainers

- Use `testcontainers-go` with `pgvector/pgvector:pg16` for database tests.
- Use `testcontainers-go` with `ollama/ollama:latest` for embedding tests.
- Reusable containers are in `internal/testcontainers/` — use `NewPostgres()` and `NewOllama()`.
- Container starts once in `SetupSuite`, terminates in `TearDownSuite`.
- Run with `-short` flag to skip integration tests in CI fast-feedback loops.

### Running Tests

**CRITICAL: Always run tests sequentially with `-p=1` flag!**

```bash
go test ./... -count=1 -p=1
```

Why? Multiple packages run in parallel by default, causing Docker resource contention when multiple Ollama/Postgres containers start simultaneously. This leads to flaky 404 errors and race conditions.

- `-p=1`: Run one package at a time (sequential)
- `-count=1`: Disable test caching
- `-short`: Skip integration tests (for fast feedback)
- `-v`: Verbose output (for debugging)

### What to Assert

- **Lengths**: `s.Require().Len(results, expected)` — exact count matters
- **Empty checks**: `s.Require().Empty(results)` for no-data scenarios
- **Field values**: Check all business-critical fields, not just existence
- **Ordering**: If sorted, verify order with loop: `s.Require().LessOrEqual(prev, curr)`
- **Edge cases**: Empty inputs, non-existent IDs, boundary conditions

