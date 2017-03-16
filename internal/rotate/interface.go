//go:generate stringer -type=Status
package rotate

import (
	"context"
	"time"
)

type Result struct {
	Name   string
	Err    error
	Status Status
}

type Status int

const (
	SUCCESS Status = iota
	FAIL
	SKIP
)

type LogInfo struct {
	Path     string
	Size     uint64
	ChangeAt time.Time
}

type Callback func()

func (this Callback) callback() {
	this()
}

type RotateJob interface {
	Run(ctx context.Context) Result
	Test(ctx context.Context) Result
	Rotate(ctx context.Context) error
	CanRotate(info LogInfo, ctx context.Context) (bool, error)
	LogState(ctx context.Context) (LogInfo, error)
	Callback(ctx context.Context)
}
