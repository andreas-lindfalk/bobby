# MCP Integration

Bobby exposes its car import tools via [Model Context Protocol (MCP)](https://modelcontextprotocol.io), allowing Claude Desktop, VS Code with Copilot, and other MCP clients to use the tools directly.

## Available Tools

| Tool | Description |
|------|-------------|
| `calculate_deadlines` | Calculate car matriculation deadlines based on arrival date and padrón registration |
| `estimate_tax` | Estimate Spanish car registration tax based on car value and CO2 emissions |

## Setup

### 1. Build the MCP server

```bash
make build-mcp
```

This creates the `bobby-mcp` binary.

### 2. Configure Claude Desktop

Add to your Claude Desktop config (`~/Library/Application Support/Claude/claude_desktop_config.json` on macOS):

```json
{
  "mcpServers": {
    "bobby": {
      "command": "/path/to/bobby/bobby-mcp"
    }
  }
}
```

Replace `/path/to/bobby` with the actual path to your bobby directory.

### 3. Configure VS Code (with GitHub Copilot)

Add to your VS Code settings (`.vscode/mcp.json` in your workspace or global settings):

```json
{
  "servers": {
    "bobby": {
      "type": "stdio",
      "command": "/path/to/bobby/bobby-mcp"
    }
  }
}
```

## Testing

You can test the MCP server with the [MCP Inspector](https://github.com/modelcontextprotocol/inspector):

```bash
npx @modelcontextprotocol/inspector ./bobby-mcp
```

## Example Usage in Claude

Once configured, you can ask Claude things like:

- "I arrived in Spain on 2025-09-15. What are my car import deadlines?"
- "Estimate the tax for importing a car worth €25,000 with 120 g/km CO2 emissions"
- "I have a Tesla, is it exempt from registration tax?"

Claude will automatically use Bobby's tools to calculate accurate answers.
