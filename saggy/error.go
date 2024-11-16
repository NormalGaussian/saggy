package saggy

import (
	"encoding/json"
	"fmt"
	"runtime"
	"strings"
)

type SilentError struct {
	Err error
	ExitCode int
}

func (e *SilentError) Error() string {
	return ""
}

func (e *SilentError) Unwrap() error {
	return e.Err
}

func NewSilentError(err error, exitCode int) error {
	return &SilentError{Err: err, ExitCode: exitCode}
}

type SaggyError struct {
	Message string
	Err     error
	File    string
	Line    int
	Meta    interface{}
}

func jsonifyStruct(s interface{}) string {
	if s == nil {
		return ""
	}
	b, _ := json.MarshalIndent(s, "\t", "\t")
	return string(b)
}

func indent(s string) string {
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		lines[i] = "\t" + line
	}
	return strings.Join(lines, "\n")
}

func (e *SaggyError) Error() string {
	if e == nil {
		return "SaggyError is nil"
	}
	result := "SaggyError"

	location := ""
	location += e.File
	if e.File != "" && e.Line != 0 {
		location += ":"
	}
	location += fmt.Sprint(e.Line)

	if location != "" {
		result += "@" + location
	}

	if e.Message != "" {
		result += ": " + e.Message
	}

	if e.Err != nil {
		result += "\n\t" + e.Err.Error()
	}

	if e.Meta != nil {
		result += "\n\t" + indent(jsonifyStruct(e.Meta))
	}

	return result
}

func (e *SaggyError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

func NewSaggyError_skipFrames(message string, err error, meta interface{}, skip int) error {
	_, file, line, _ := runtime.Caller(skip)
	error := &SaggyError{Message: message, Err: err, Meta: meta, File: file, Line: line}
	return error
}

func NewSaggyErrorWithMeta(message string, err error, meta interface{}) error {
	return NewSaggyError_skipFrames(message, err, meta, 0)
}

func NewSaggyError(message string, err error) error {
	return NewSaggyError_skipFrames(message, err, nil, 0)
}

func NewExecutionError(message string, output string, status int, command string, args []string, dir string) error {
	meta := struct {
		Status  int
		Output  string
		Command string
		Args    []string
		Dir     string
	}{Status: status, Output: output, Command: command, Args: args, Dir: dir}
	return NewSaggyError_skipFrames(message, nil, meta, 2)
}
