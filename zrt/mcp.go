package zrt

// MCPServer describes a Model Context Protocol server to connect to.
type MCPServer interface {
	mcpType() string
	mcpStdio() (command string, args []string, env map[string]string)
	mcpHTTP() (url string, headers map[string]string)
}

// MCPServerStdio is a stdio-transport MCP server.
type MCPServerStdio struct {
	Command string
	Args    []string
	Env     map[string]string
}

func (m *MCPServerStdio) mcpType() string { return "stdio" }
func (m *MCPServerStdio) mcpStdio() (string, []string, map[string]string) {
	return m.Command, m.Args, m.Env
}
func (m *MCPServerStdio) mcpHTTP() (string, map[string]string) { return "", nil }

// MCPServerHTTP is an HTTP-transport MCP server.
type MCPServerHTTP struct {
	URL     string
	Headers map[string]string
}

func (m *MCPServerHTTP) mcpType() string                                 { return "http" }
func (m *MCPServerHTTP) mcpStdio() (string, []string, map[string]string) { return "", nil, nil }
func (m *MCPServerHTTP) mcpHTTP() (string, map[string]string)            { return m.URL, m.Headers }
