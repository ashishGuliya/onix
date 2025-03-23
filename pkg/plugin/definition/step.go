package definition

import (
	"context"

	"github.com/ashishGuliya/onix/pkg/model"
)

type Step interface {
	Run(ctx *model.StepContext) error
}

type StepProvider interface {
	New(context.Context, map[string]string) (Step, func(), error)
}
