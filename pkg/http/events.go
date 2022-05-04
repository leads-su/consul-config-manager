package http

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	netHttp "net/http"
	"os"
	"text/template"
	"time"

	cfg "github.com/leads-su/consul-config-manager/pkg/config"
	"github.com/leads-su/consul-config-manager/pkg/config/agent"
	"github.com/leads-su/consul-config-manager/pkg/config/application"
	"github.com/leads-su/consul-config-manager/pkg/config/notifier"
	"github.com/leads-su/consul-config-manager/pkg/tasks"
	"github.com/leads-su/logger"
	notifierPackage "github.com/leads-su/notifier"
	"github.com/leads-su/runner"
	"github.com/leads-su/storage"
	"github.com/r3labs/sse/v2"
)

// EventServer describes structure of event server
type EventServer struct {
	application *application.Application
	agent       *agent.Agent
	server      *sse.Server
	notifier    *notifier.Notifier
	storage     *storage.Storage
}

// NewEventServer creates new instance of event server
func NewEventServer(config *cfg.Config) *EventServer {
	serverInstance := sse.New()
	serverInstance.Headers = map[string]string{
		"Access-Control-Allow-Origin":      "*",
		"Access-Control-Allow-Credentials": "true",
	}

	return &EventServer{
		application: config.Application,
		agent:       config.Agent,
		server:      serverInstance,
		notifier:    config.Notifier,
		storage: storage.NewStorage(storage.Options{
			WorkingDirectory: config.Sse.WriteTo,
		}),
	}
}

// RegisterRoutes registers list of routes supported by the events server
func (eventServer *EventServer) RegisterRoutes() *EventServer {
	netHttp.HandleFunc("/events/watch", eventServer.Server().ServeHTTP)
	netHttp.HandleFunc("/events/load", eventServer.streamLoaderHandler)
	netHttp.HandleFunc("/tasks/create", eventServer.taskCreatorHandler)
	return eventServer
}

// StartStream starts new stream
func (eventServer *EventServer) StartStream(streamID string) *sse.Stream {
	return eventServer.Server().CreateStream(streamID)
}

// StopStream stops existing stream
func (eventServer *EventServer) StopStream(streamID string) {
	eventServer.Server().RemoveStream(streamID)
}

// Publish publishes new event to event server at given stream id
func (eventServer *EventServer) Publish(streamID string, event *sse.Event) {
	eventServer.Server().Publish(streamID, event)
}

// Server returns instance of SSE server
func (eventServer *EventServer) Server() *sse.Server {
	return eventServer.server
}

// Storage returns instance of storage specific for sse
func (eventServer *EventServer) Storage() *storage.Storage {
	return eventServer.storage
}

// streamLoaderHandler handles loading of past streams for reading
func (eventServer *EventServer) streamLoaderHandler(response netHttp.ResponseWriter, request *netHttp.Request) {
	response.Header().Add("Content-Type", "application/json")
	streamID := request.URL.Query().Get("stream")
	if streamID == "" {
		eventServer.loaderBadRequestResponse(response, "Please specify stream ID!")
		return
	}

	eventPath := eventServer.Storage().AbsolutePath(fmt.Sprintf("%s.log", streamID))
	if !eventServer.Storage().Exists(eventPath) {
		response.WriteHeader(netHttp.StatusNotFound)
		json.NewEncoder(response).Encode(ResponseStructure{
			Success: false,
			Status:  netHttp.StatusNotFound,
			Message: fmt.Sprintf("Unable to find stream with ID - `%s`", streamID),
		})
		return
	}

	entries, err := eventServer.readEventToStructure(eventPath)
	if err != nil {
		response.WriteHeader(netHttp.StatusInternalServerError)
		json.NewEncoder(response).Encode(ResponseStructure{
			Success: false,
			Status:  netHttp.StatusInternalServerError,
			Message: fmt.Sprintf("Failed to read stream with ID - `%s`", streamID),
			Data:    err.Error(),
		})
		return
	}

	response.WriteHeader(netHttp.StatusOK)
	json.NewEncoder(response).Encode(ResponseStructure{
		Success: true,
		Status:  netHttp.StatusOK,
		Message: fmt.Sprintf("Successfully retrieved stream information for stream with id - `%s`", streamID),
		Data:    entries,
	})
}

func (eventServer *EventServer) taskCreatorHandler(response netHttp.ResponseWriter, request *netHttp.Request) {
	task, err := tasks.NewTask(request.Body)
	if err != nil {
		response.WriteHeader(netHttp.StatusBadRequest)
		json.NewEncoder(response).Encode(ResponseStructure{
			Success: false,
			Status:  netHttp.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	eventServer.StartStream(task.StreamID())
	logger.Infof("tasks:manager", "created new stream - `%s`", task.StreamID())

	runnerTask, err := task.ToRunnerTask(func(outputType int, outputLine string, outputSource int) {
		eventResponse := tasks.EventStructure{
			Source:    outputSource,
			Type:      outputType,
			Line:      outputLine,
			Timestamp: time.Now().Unix(),
		}
		data, _ := json.Marshal(eventResponse)

		eventServer.Publish(task.StreamID(), &sse.Event{
			Data: data,
		})

		go func() {
			newLine := []byte("\n")
			data = append(data, newLine...)
			err := eventServer.Storage().AppendBytesArrayToFile(eventServer.Storage().AbsolutePath(task.SteamIDLogFile()), data, 0755)
			if err != nil {
				logger.Errorf("task:realtime", "failed to write data to file - %s", err.Error())
			}
		}()
	})
	if err != nil {
		response.WriteHeader(netHttp.StatusBadRequest)
		json.NewEncoder(response).Encode(ResponseStructure{
			Success: false,
			Status:  netHttp.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	go runner.NewRunner(&runner.RunnerOptions{
		Task: runnerTask,
		OnError: func() {
			if eventServer.notifier.NotifyOn.Error {
				eventServer.sendNotification(
					notifierPackage.Error,
					task.ExecutionID,
					task.PipelineID,
					"Pipeline failed with error, please check UI for more information.",
				)
			}
			eventServer.StopStream(task.StreamID())
		},
		OnSuccess: func() {
			if eventServer.notifier.NotifyOn.Success {
				eventServer.sendNotification(
					notifierPackage.Success,
					task.ExecutionID,
					task.PipelineID,
					"Pipeline succeeded without errors.",
				)
			}
			eventServer.StopStream(task.StreamID())
		},
	}).Run()

	json.NewEncoder(response).Encode(task)
}

// readEventToStructure reads event file to structure
func (eventServer *EventServer) readEventToStructure(path string) ([]tasks.EventStructure, error) {
	var entries []tasks.EventStructure
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var entry tasks.EventStructure
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

// loaderBadRequestResponse handles `BAD REQUEST` error
func (eventServer *EventServer) loaderBadRequestResponse(response netHttp.ResponseWriter, message string) {
	response.WriteHeader(netHttp.StatusBadRequest)
	json.NewEncoder(response).Encode(ResponseStructure{
		Success: false,
		Status:  netHttp.StatusBadRequest,
		Message: message,
	})
}

// sendNotification send notification using configured notifiers
func (eventServer *EventServer) sendNotification(notificationType int, executionIdentifier, pipelineIdentifier, notificationMessage string) {
	if eventServer.notifier.IsEnabled() {
		tpl, err := template.New("").Parse(`
*Server:*      {{ .Address }}
*Version:*     {{ .Version }}-{{ .ShaHash }}
*Execution:*   {{ .Execution }}
*Pipeline:*    {{ .Pipeline }}
`)
		if err == nil {
			templateValues := struct {
				Version   string
				ShaHash   string
				Address   string
				Execution string
				Pipeline  string
			}{
				Version:   eventServer.application.Version,
				ShaHash:   eventServer.application.CommitSha,
				Address:   eventServer.agent.Address(),
				Execution: executionIdentifier,
				Pipeline:  pipelineIdentifier,
			}

			var templateBuffer bytes.Buffer
			err = tpl.Execute(&templateBuffer, templateValues)
			if err == nil {
				eventServer.notifier.DeliverNotification(notifier.TELEGRAM, notifierPackage.NewNotification(notifierPackage.NotificationOptions{
					Type: notificationType,
					Title: fmt.Sprintf(
						"%s - %s",
						eventServer.agent.Network.Hostname(),
						notificationMessage,
					),
					Message: templateBuffer.String(),
				}))
			}
		}
	}
}
