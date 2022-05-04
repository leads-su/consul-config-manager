package tasks

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/leads-su/logger"
	"github.com/leads-su/runner"
	"io"
	"io/ioutil"
	"strings"
)

// Task describes structure for incoming task request
type Task struct {
	ExecutionID string   `json:"execution_id"`
	PipelineID  string   `json:"pipeline_id"`
	TaskID      string   `json:"task_id"`
	ActionID    string   `json:"action_id"`
	ServerID    string   `json:"server_id"`
	Command     string   `json:"command"`
	Arguments   []string `json:"arguments"`
	RunAs       string   `json:"run_as"`
	UseSudo     bool     `json:"use_sudo"`
	FailOnError bool     `json:"fail_on_error"`
	WorkingDir  string   `json:"working_dir"`
}

// NewTask creates new instance of task from received request
func NewTask(requestBodyReader io.ReadCloser) (*Task, error) {
	requestBody, err := ioutil.ReadAll(requestBodyReader)
	if err != nil {
		logger.Errorf("tasks:new", "failed to process request body - %s", err.Error())
		return nil, err
	}
	var task *Task
	err = json.Unmarshal(requestBody, &task)

	if err != nil {
		logger.Errorf("tasks:new", "failed to decode received request body - %s", err.Error())
		return nil, err
	}
	return task.validateTask()
}

// StreamID generates and returns stream id for current task
func (task *Task) StreamID() string {
	identifier := fmt.Sprintf(
		"%s_%s_%s_%s_%s",
		task.ExecutionID,
		task.PipelineID,
		task.TaskID,
		task.ActionID,
		task.ServerID,
	)
	encryptor := sha256.New()
	encryptor.Write([]byte(identifier))
	return hex.EncodeToString(encryptor.Sum(nil))
}

// SteamIDLogFile generates and returns stream id log file name
func (task *Task) SteamIDLogFile() string {
	return fmt.Sprintf("%s.log", task.StreamID())
}

// ToRunnerTask convert request task to runner task
func (task *Task) ToRunnerTask(outputHandler func(int, string, int)) (*runner.Task, error) {
	var err error

	runnerTask := runner.
		NewTask(task.Command).
		WithArguments(task.Arguments).
		WithWorkingDir(task.WorkingDir).
		WithRealtimeOutput(outputHandler)

	if task.RunAs != "" {
		runnerTask, err = runnerTask.RunAs(task.RunAs)
		if err != nil {
			return nil, err
		}
	}

	if task.UseSudo {
		runnerTask, err = runnerTask.RunAsSudo()
		if err != nil {
			return nil, err
		}
	}

	if !task.FailOnError {
		runnerTask = runnerTask.WithSkipError()
	}

	return runnerTask, nil
}

// validateTask validates that passed body is a valid task object
func (task *Task) validateTask() (*Task, error) {
	formattedExecutionID := strings.TrimSpace(task.ExecutionID)
	formattedPipelineID := strings.TrimSpace(task.PipelineID)
	formattedTaskID := strings.TrimSpace(task.TaskID)
	formattedActionID := strings.TrimSpace(task.ActionID)
	formattedServerID := strings.TrimSpace(task.ServerID)

	if formattedExecutionID == "" {
		return nil, fmt.Errorf("`execution_id` cannot be empty")
	}
	task.ExecutionID = formattedExecutionID

	if formattedPipelineID == "" {
		return nil, fmt.Errorf("`pipeline_id` cannot be empty")
	}
	task.PipelineID = formattedPipelineID

	if formattedTaskID == "" {
		return nil, fmt.Errorf("`task_id` cannot be empty")
	}
	task.TaskID = formattedTaskID

	if formattedActionID == "" {
		return nil, fmt.Errorf("`action_id` cannot be empty")
	}
	task.ActionID = formattedActionID

	if formattedServerID == "" {
		return nil, fmt.Errorf("`server_id` cannot be empty")
	}
	task.ServerID = formattedServerID

	formattedCommand := strings.TrimSpace(task.Command)

	if formattedCommand == "" {
		return nil, fmt.Errorf("`command` cannot be empty")
	}
	task.Command = formattedCommand

	if task.Arguments == nil || len(task.Arguments) == 0 {
		return nil, fmt.Errorf("`arguments` must have at least 1 argument passed to task")
	}
	for index, value := range task.Arguments {
		task.Arguments[index] = strings.TrimSpace(value)
	}

	task.RunAs = strings.TrimSpace(task.RunAs)
	task.WorkingDir = strings.TrimSpace(task.WorkingDir)

	return task, nil
}
