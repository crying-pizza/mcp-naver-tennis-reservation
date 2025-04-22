package main

import (
	"fmt"
	"github.com/mark3labs/mcp-go/server"
	"log"
	"naver/tools"
)

var VERSION string = "v0.0.1"

// func main2() {
// 	tools.GetAvailableTimeSlotHandler()
// }

func main() {
	s := server.NewMCPServer(
		"naver_reservation",
		VERSION,
	)

	s.AddTool(tools.GetAvailableTimeSlot())

	log.Print("Starting server...")
	if err := server.ServeStdio(s); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}

// https://github.com/abhirockzz/mcp_cosmosdb_go
