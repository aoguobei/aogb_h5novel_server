package utils

import "errors"

var (
	// ErrTooManyTasks 任务数量过多错误
	ErrTooManyTasks = errors.New("too many concurrent tasks")
	// ErrTaskNotFound 任务未找到错误
	ErrTaskNotFound = errors.New("task not found")
	// ErrTaskAlreadyRunning 任务已在运行错误
	ErrTaskAlreadyRunning = errors.New("task is already running")
)
