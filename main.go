package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/golang/glog" // 添加glog导入
)

func main() {

	flag.Parse()
	// 在main函数结束时添加
	defer glog.Flush()

	stdin := os.Stdin
	stdout := os.Stdout

	reader := bufio.NewReader(stdin)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()

		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				glog.Error("read from client error", err)
				return
			}
			glog.Info("read from client\t", line, err)
			request := JSONRPCRequest{}
			if err := json.Unmarshal([]byte(line), &request); err != nil {
				glog.Error(err) // 替换fmt为glog
				continue
			}
			switch request.Method {
			case "initialize":
				if response, err := initializeRequest(request); err == nil {
					jsonrpcResponse := JSONRPCResponse{
						JSONRPC: request.JSONRPC,
						ID:      request.ID,
						Result:  response,
					}

					writeResponse(stdout, jsonrpcResponse)

				}
			case "tools/list":
				if response, err := listToolsRequest(request); err == nil {
					jsonrpcResponse := JSONRPCResponse{
						JSONRPC: request.JSONRPC,
						ID:      request.ID,
						Result:  response,
					}

					writeResponse(stdout, jsonrpcResponse)
				}
			case "tools/call":
				if response, err := callToolRequest(request); err == nil {
					jsonrpcResponse := JSONRPCResponse{
						JSONRPC: request.JSONRPC,
						ID:      request.ID,
						Result:  response,
					}
					writeResponse(stdout, jsonrpcResponse)
				} else {
					jsonrpcResponse := JSONRPCResponse{
						JSONRPC: request.JSONRPC,
						ID:      request.ID,
						Error:   err,
					}
					writeResponse(stdout, jsonrpcResponse)
				}
			case "notifications/initialized":
				// do nothing
			default:
				jsonrpcResponse := JSONRPCResponse{
					JSONRPC: request.JSONRPC,
					ID:      request.ID,
					Error: &JSONRPCError{
						Code:    METHOD_NOT_FOUND,
						Message: "not support",
					},
				}
				writeResponse(stdout, jsonrpcResponse)
			}
		}

	}()

	wg.Wait()
}

func writeResponse(writer io.Writer, response JSONRPCResponse) error {
	bytes, err := json.Marshal(response)
	if err != nil {
		return err
	}
	glog.Info("write response\t", string(bytes)) // 替换fmt为glog
	if _, err := fmt.Fprintf(writer, "%s\n", bytes); err != nil {
		glog.Error(fmt.Sprintf("write error, error:%v", err)) // 替换fmt为glog
		return err
	}
	return nil
}
