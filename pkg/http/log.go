package http

import (
	"bufio"
	"encoding/json"
	"fmt"
	netHttp "net/http"
	"os"

	cfg "github.com/leads-su/consul-config-manager/pkg/config"
	"github.com/leads-su/storage"
)

// LogServer describes structure of log server handler
type LogServer struct {
	storage *storage.Storage
}

// LogStructure describes log file structure
type LogStructure struct {
	Level   string      `json:"level"`
	Hash    string      `json:"commit"`
	Source  string      `json:"source"`
	Time    string      `json:"time"`
	Version string      `json:"version"`
	Message interface{} `json:"msg"`
}

// NewLogServer creates new instance of log server handler
func NewLogServer(config *cfg.Config) *LogServer {
	return &LogServer{
		storage: storage.NewStorage(storage.Options{
			WorkingDirectory: config.Log.WriteTo,
		}),
	}
}

// RegisterRoutes registers list of routes supported by the log server
func (logServer *LogServer) RegisterRoutes() {
	netHttp.HandleFunc("/logs/info", logServer.handleInfoLogRequest)
	netHttp.HandleFunc("/logs/error", logServer.handleErrorLogRequest)
}

// handleInfoLogRequest handles incoming request for info log
func (logServer *LogServer) handleInfoLogRequest(response netHttp.ResponseWriter, request *netHttp.Request) {
	logServer.handleLogRequest(response, request, "info")
}

// handleErrorLogRequest handles incoming request for error log
func (logServer *LogServer) handleErrorLogRequest(response netHttp.ResponseWriter, request *netHttp.Request) {
	logServer.handleLogRequest(response, request, "error")
}

// handleLogRequest handles internal processing of logs
func (logServer *LogServer) handleLogRequest(response netHttp.ResponseWriter, request *netHttp.Request, logName string) {
	response.Header().Set("Content-Type", "application/json")
	entries := make([]LogStructure, 0)

	logPath := logServer.storage.AbsolutePath(fmt.Sprintf("%s.log", logName))

	if !logServer.storage.Exists(logPath) {
		response.WriteHeader(200)
		json.NewEncoder(response).Encode(ResponseStructure{
			Success: true,
			Status:  200,
			Message: fmt.Sprintf("Successfully retrieved %s logs", logName),
			Data:    entries,
		})
		return
	}

	entries, err := logServer.readLogToStructure(logPath)

	if err != nil {
		response.WriteHeader(500)
		json.NewEncoder(response).Encode(ResponseStructure{
			Success: false,
			Status:  500,
			Message: fmt.Sprintf("Failed to read %s log", logName),
			Data:    err.Error(),
		})
		return
	}
	response.WriteHeader(200)
	json.NewEncoder(response).Encode(ResponseStructure{
		Success: true,
		Status:  200,
		Message: fmt.Sprintf("Successfully retrieved %s logs", logName),
		Data:    entries,
	})
}

// readLogToStructure reads log and returns it as a structure
func (logServer *LogServer) readLogToStructure(path string) ([]LogStructure, error) {
	var entries []LogStructure
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var entry LogStructure
		err := json.Unmarshal(scanner.Bytes(), &entry)
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return entries, nil
}
