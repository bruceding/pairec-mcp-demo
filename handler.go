package main

import (
	"encoding/json"
	"fmt"
)

type Implementation struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}
type InitializeResponse struct {
	ProtocolVersion string              `json:"protocolVersion"`
	Instructions    string              `json:"instructions"`
	Capabilities    *ServerCapabilities `json:"capabilities"`
	ServerInfo      Implementation      `json:"serverInfo"`
}

func initializeRequest(request JSONRPCRequest) (*InitializeResponse, *JSONRPCError) {
	var protocolVersion string
	if version, ok := request.Params["protocolVersion"]; ok {
		protocolVersion = version.(string)
	}
	response := InitializeResponse{
		ProtocolVersion: protocolVersion,
		Instructions:    "Hello, welcome to the PAI-Rec MCP server!",
		Capabilities: &ServerCapabilities{
			// 修改结构体定义以匹配 JSON 标签要求
			Tools: &struct {
				ListChanged bool `json:"listChanged,omitempty"`
			}{ListChanged: true},
		},
		ServerInfo: Implementation{
			Name:    "PAI-Rec MCP Server",
			Version: "1.0.0",
		},
	}

	return &response, nil
}

type ListToolsResult struct {
	Tools []Tool `json:"tools"`
}
type Tool struct {
	Description string      `json:"description"`
	InputSchema InputSchema `json:"inputSchema"`
	Name        string      `json:"name"`
}
type InputSchema struct {
	Properties map[string]any `json:"properties,omitempty"`
	Required   []string       `json:"required,omitempty"`
	Type       string         `json:"type"`
}

func listToolsRequest(request JSONRPCRequest) (*ListToolsResult, *JSONRPCError) {
	response := ListToolsResult{
		Tools: []Tool{
			{
				Description: "Verify pairec conf",
				Name:        "verify_pairec_conf",
				InputSchema: InputSchema{
					Type: "object",
					Properties: map[string]any{
						"pairec_conf": map[string]any{
							"description": "pairec config to verify",
							"type":        "string",
						},
					},
					Required: []string{"pairec_conf"},
				},
			},
		},
	}

	return &response, nil
}

type CallToolResult struct {
	Content []TextContent `json:"content,omitempty"`
}
type TextContent struct {
	Text string `json:"text"`
	Type string `json:"type"`
}

func callToolRequest(request JSONRPCRequest) (*CallToolResult, *JSONRPCError) {
	if toolName, ok := request.Params["name"]; ok {
		if toolName != "verify_pairec_conf" {
			return nil, &JSONRPCError{
				Code:    INVALID_PARAMS,
				Message: "Invalid tool name",
			}
		}
		arguments := request.Params["arguments"].(map[string]any)
		if config, ok := arguments["pairec_conf"].(string); !ok {
			return nil, &JSONRPCError{
				Code:    INVALID_PARAMS,
				Message: "Invalid pairec config",
			}
		} else {
			m := make(map[string]any)
			if err := json.Unmarshal([]byte(config), &m); err != nil {
				return nil, &JSONRPCError{
					Code:    INVALID_PARAMS,
					Message: fmt.Sprintf("Invalid pairec config, error: %v", err),
				}
			} else {
				return &CallToolResult{
					Content: []TextContent{
						{
							Text: "Verify pairec config successfully",
							Type: "text",
						},
					},
				}, nil
			}
		}

	} else {
		return nil, &JSONRPCError{
			Code:    INVALID_PARAMS,
			Message: "Invalid tool name",
		}
	}

}
