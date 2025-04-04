package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/sashabaranov/go-openai"
)

func connectToServer(serverPath string) (*client.StdioMCPClient, error) {
	mcpClient, err := client.NewStdioMCPClient(serverPath, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to server: %v", err)
	}
	_, err = mcpClient.Initialize(context.Background(), mcp.InitializeRequest{
		Params: struct {
			ProtocolVersion string                 "json:\"protocolVersion\""
			Capabilities    mcp.ClientCapabilities "json:\"capabilities\""
			ClientInfo      mcp.Implementation     "json:\"clientInfo\""
		}{
			ProtocolVersion: "2024-11-05",
			Capabilities:    mcp.ClientCapabilities{},
			ClientInfo: mcp.Implementation{
				Name:    "mcp-go",
				Version: "0.1.0",
			},
		},
	})

	if err != nil {
		return nil, fmt.Errorf("failed to initialize client: %v", err)
	}
	return mcpClient, nil
}
func main() {
	dir, err := os.Getwd()
	if err != nil {
		fmt.Println("get pwd dir fail:", err)
		return
	}
	serverPath := dir + "/pairec-mcp-demo"
	mcpClient, err := connectToServer(serverPath)
	if err != nil {
		fmt.Println("create mcp client fail:", err)
		return
	}
	defer mcpClient.Close()
	toolsResult, err := mcpClient.ListTools(context.Background(), mcp.ListToolsRequest{})
	if err != nil {
		fmt.Println("list tools fail:", err)
		return
	}
	fmt.Println("list tools result:", toolsResult)
	config := openai.DefaultConfig(os.Getenv("API_KEY"))
	config.BaseURL = "https://dashscope.aliyuncs.com/compatible-mode/v1"
	// 初始化客户端
	client := openai.NewClientWithConfig(config)

	var tools []openai.Tool
	for _, tool := range toolsResult.Tools {
		tools = append(tools, openai.Tool{
			Type: "function",
			Function: &openai.FunctionDefinition{
				Name:        tool.Name,
				Description: tool.Description,
				Parameters:  tool.InputSchema.Properties,
			},
		})
	}
	query := "校验下下面的pairec 配置是否正确 '{\"Listen\":{\"Http\":8080}}'"
	// 创建聊天请求
	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: "qwen-plus",
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: query,
				},
			},
			Tools:      tools,
			ToolChoice: "auto",
		},
	)

	if err != nil {
		fmt.Printf("ChatCompletion error: %v\n", err)
		return
	}

	// 输出结果

	if len(resp.Choices[0].Message.ToolCalls) == 0 {
		fmt.Println("no tool use")
		fmt.Println("最终答案：")
		fmt.Println(resp.Choices[0].Message.Content)
		return
	}
	toolCall := resp.Choices[0].Message.ToolCalls[0].Function
	callTooRequest := mcp.CallToolRequest{}
	callTooRequest.Params.Name = toolCall.Name
	callTooRequest.Params.Arguments = make(map[string]interface{})
	if err := json.Unmarshal([]byte(toolCall.Arguments), &callTooRequest.Params.Arguments); err != nil {
		fmt.Println("json unmarshal fail:", err)
		return
	}
	toolCallResult, err := mcpClient.CallTool(context.Background(), callTooRequest)
	if err != nil {
		fmt.Println("call tool fail:", err)
		return
	}
	fmt.Println("call tool result:", toolCallResult)
	// 修改ChatCompletion请求，确保消息顺序正确
	messages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleUser,
			Content: query,
		},
		{
			Role:      openai.ChatMessageRoleAssistant,
			Content:   "",
			ToolCalls: resp.Choices[0].Message.ToolCalls,
		},
		{
			Role:       openai.ChatMessageRoleTool,
			ToolCallID: resp.Choices[0].Message.ToolCalls[0].ID,
			Content:    toolCallResult.Content[0].(mcp.TextContent).Text,
		},
	}

	resp, err = client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:    "qwen-plus",
			Messages: messages,
		},
	)
	if err != nil {
		fmt.Printf("ChatCompletion error: %v\n", err)
		return
	}
	fmt.Println("final answer：")
	fmt.Println(resp.Choices[0].Message.Content)
}
