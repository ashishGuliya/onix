package handler

import (
	"fmt"

	"github.com/ashishGuliya/onix/pkg/plugin/definition"
)

// 🔹 Sign Step
type signStep struct {
	signer definition.Signer
}

func (s *signStep) Run(ctx *definition.StepContext) error {
	sign, err := s.signer.Sign(ctx, ctx.Body, ctx.SigningKey)
	if err != nil {
		return fmt.Errorf("failed to sign request: %w", err)
	}
	ctx.Request.Header.Set("SignHeader", sign)
	return nil
}

// 🔹 Validate Sign Step
type validateSignStep struct {
	validator definition.SignValidator
}

func (s *validateSignStep) Run(ctx *definition.StepContext) error {
	headerValue := ctx.Request.Header.Get("HeaderString")
	valid, err := s.validator.Verify(ctx, ctx.Body, headerValue, "key")
	if err != nil {
		return fmt.Errorf("sign validation failed: %w", err)
	}
	if !valid {
		return fmt.Errorf("sign validation failed: signature is invalid")
	}
	return nil
}

// 🔹 Validate Schema Step
type validateSchemaStep struct {
	validator definition.SchemaValidator
}

func (s *validateSchemaStep) Run(ctx *definition.StepContext) error {
	if err := s.validator.Validate(ctx, ctx.Request.URL, ctx.Body); err != nil {
		return fmt.Errorf("schema validation failed: %w", err)
	}
	return nil
}

// 🔹 Get Route Step
type addRouteStep struct {
	router definition.Router
}

func (s *addRouteStep) Run(ctx *definition.StepContext) error {
	route, err := s.router.Route(ctx, ctx.Request.URL, ctx.Body)
	if err != nil {
		return fmt.Errorf("failed to determine route: %w", err)
	}
	ctx.Route = route
	return nil
}

// 🔹 Broadcast Step (Stub Implementation)
type broadcastStep struct{}

func (b *broadcastStep) Run(ctx *definition.StepContext) error {
	// TODO: Implement broadcast logic if needed
	return nil
}

// 🔹 Subscribe Step (Stub Implementation)
type subscribeStep struct{}

func (s *subscribeStep) Run(ctx *definition.StepContext) error {
	// TODO: Implement subscription logic if needed
	return nil
}
