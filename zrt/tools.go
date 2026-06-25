package zrt

import "context"

// FunctionToolInfo describes a tool the LLM may call.
//
// ParametersSchema is a JSON Schema object (e.g.
// {"type":"object","properties":{...},"required":[...]}).
type FunctionToolInfo struct {
	Name             string
	Description      string
	ParametersSchema map[string]any
	Filler           string
}

// ToolHandler executes a tool call. args is the decoded arguments object. The
// returned value becomes the tool result: a string is used verbatim, anything
// else is JSON-encoded.
type ToolHandler func(ctx context.Context, args map[string]any) (any, error)

// FunctionTool is a callable tool exposed to the LLM.
//
// The tool's name, description and JSON schema are provided explicitly when the
// tool is constructed, and its handler is invoked when the LLM calls the tool.
type FunctionTool struct {
	Info    FunctionToolInfo
	Handler ToolHandler
}

// ToolOption configures optional FunctionTool metadata (see WithFiller).
type ToolOption func(*FunctionToolInfo)

// WithFiller makes the tool speak filler when it is called — a short line that
// covers the tool's latency. Pass it to NewFunctionTool.
func WithFiller(filler string) ToolOption {
	return func(i *FunctionToolInfo) { i.Filler = filler }
}

// NewFunctionTool builds a FunctionTool. schema may be nil (treated as an empty
// object schema). Optional behavior is set via ToolOptions, e.g.:
//
//	zrt.NewFunctionTool(name, desc, schema, handler, zrt.WithFiller("One moment..."))
func NewFunctionTool(name, description string, schema map[string]any, handler ToolHandler, opts ...ToolOption) *FunctionTool {
	info := FunctionToolInfo{Name: name, Description: description, ParametersSchema: schema}
	for _, opt := range opts {
		opt(&info)
	}
	return &FunctionTool{Info: info, Handler: handler}
}

// ToolInfo returns the tool metadata.
func (t *FunctionTool) ToolInfo() FunctionToolInfo { return t.Info }

// IsFunctionTool reports whether v is a non-nil *FunctionTool with a name.
func IsFunctionTool(v any) bool {
	t, ok := v.(*FunctionTool)
	return ok && t != nil && t.Info.Name != ""
}

// GetToolInfo returns a tool's info.
func GetToolInfo(t *FunctionTool) FunctionToolInfo { return t.Info }

// BuildOpenAISchema renders the tool as an OpenAI function-tool schema.
func BuildOpenAISchema(t *FunctionTool) map[string]any {
	params := t.Info.ParametersSchema
	if params == nil {
		params = map[string]any{"type": "object", "properties": map[string]any{}}
	}
	return map[string]any{
		"type": "function",
		"function": map[string]any{
			"name":        t.Info.Name,
			"description": t.Info.Description,
			"parameters":  params,
		},
	}
}

// BuildGeminiSchema renders the tool as a Gemini function declaration.
func BuildGeminiSchema(t *FunctionTool) map[string]any {
	schema := t.Info.ParametersSchema
	if schema == nil {
		schema = map[string]any{"type": "object", "properties": map[string]any{}}
	}
	return map[string]any{
		"name":        t.Info.Name,
		"description": t.Info.Description,
		"parameters":  simplifySchema(schema),
	}
}

// BuildNovaSonicSchema renders the tool as a Nova Sonic schema (= OpenAI form).
func BuildNovaSonicSchema(t *FunctionTool) map[string]any { return BuildOpenAISchema(t) }

// simplifySchema strips keys Gemini rejects.
func simplifySchema(schema map[string]any) map[string]any {
	skip := map[string]bool{"additionalProperties": true, "title": true, "$defs": true, "definitions": true}
	cleaned := make(map[string]any, len(schema))
	for k, v := range schema {
		if skip[k] {
			continue
		}
		switch vv := v.(type) {
		case map[string]any:
			cleaned[k] = simplifySchema(vv)
		case []any:
			out := make([]any, 0, len(vv))
			for _, item := range vv {
				if m, ok := item.(map[string]any); ok {
					out = append(out, simplifySchema(m))
				} else {
					out = append(out, item)
				}
			}
			cleaned[k] = out
		default:
			cleaned[k] = v
		}
	}
	return cleaned
}
