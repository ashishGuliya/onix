package handler

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ashishGuliya/onix/pkg/log"
	"github.com/ashishGuliya/onix/pkg/model"
	"github.com/ashishGuliya/onix/pkg/plugin/definition"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

// 🔹 Sign Step
type signStep struct {
	signer definition.Signer
	km     definition.KeyManager
}

// newSignStep creates and returns the sign step after validation
func newSignStep(signer definition.Signer, km definition.KeyManager) (definition.Step, error) {
	if signer == nil {
		return nil, fmt.Errorf("invalid config: Signer plugin not configured")
	}
	if km == nil {
		return nil, fmt.Errorf("invalid config: KeyManager plugin not configured")
	}

	return &signStep{signer: signer, km: km}, nil
}

func (s *signStep) Run(ctx *model.StepContext) error {
	keyID, key, err := s.km.SigningPrivateKey(ctx, ctx.SubID)
	if err != nil {
		return fmt.Errorf("failed to get signing key: %w", err)
	}
	createdAt := time.Now().Unix()
	validTill := time.Now().Add(5 * time.Minute).Unix()
	sign, err := s.signer.Sign(ctx, ctx.Body, key, createdAt, validTill)
	if err != nil {
		return fmt.Errorf("failed to sign request: %w", err)
	}
	authHeader := fmt.Sprintf("Signature keyId=\"%s|%s|ed25519\",algorithm=\"ed25519\",created=\"%d\",expires=\"%d\",headers=\"(created) (expires) digest\",signature=\"%s\"", ctx.SubID, keyID, createdAt, validTill, sign)
	header := model.AuthHeaderSubscriber
	if ctx.Role == model.RoleGateway {
		header = model.AuthHeaderGateway
	}
	ctx.Request.Header.Set(header, authHeader)
	return nil
}

// 🔹 Validate Sign Step
type validateSignStep struct {
	validator definition.SignValidator
	km        definition.KeyManager
}

// newValidateSignStep creates and returns the validateSign step after validation
func newValidateSignStep(signValidator definition.SignValidator, km definition.KeyManager) (definition.Step, error) {
	if signValidator == nil {
		return nil, fmt.Errorf("invalid config: SignValidator plugin not configured")
	}
	if km == nil {
		return nil, fmt.Errorf("invalid config: KeyManager plugin not configured")
	}
	return &validateSignStep{validator: signValidator, km: km}, nil
}

func (s *validateSignStep) Run(ctx *model.StepContext) error {
	unauthHeader := fmt.Sprintf("Signature realm=\"%s\",headers=\"(created) (expires) digest\"", ctx.SubID)
	headerValue := ctx.Request.Header.Get(model.AuthHeaderGateway)
	if len(headerValue) != 0 {
		if err := s.validate(ctx, headerValue); err != nil {
			ctx.RespHeader.Set(model.UnaAuthorizedHeaderGateway, unauthHeader)
			return model.NewSignValidationErrf("failed to validate %s: %w", model.AuthHeaderGateway, err)
		}
	}
	headerValue = ctx.Request.Header.Get(model.AuthHeaderSubscriber)
	if len(headerValue) == 0 {
		ctx.RespHeader.Set(model.UnaAuthorizedHeaderSubscriber, unauthHeader)
		return model.NewSignValidationErrf("%s missing", model.UnaAuthorizedHeaderSubscriber)
	}
	if err := s.validate(ctx, headerValue); err != nil {
		ctx.RespHeader.Set(model.UnaAuthorizedHeaderSubscriber, unauthHeader)
		return model.NewSignValidationErrf("failed to validate %s: %w", model.AuthHeaderSubscriber, err)
	}
	return nil
}

func (s *validateSignStep) validate(ctx *model.StepContext, value string) error {
	headerParts := strings.Split(value, "|")
	ids := strings.Split(headerParts[0], "\"")
	if len(ids) < 2 || len(headerParts) < 3 {
		return fmt.Errorf("malformed sign header")
	}
	subID := ids[1]
	keyID := headerParts[1]
	key, err := s.km.SigningPublicKey(ctx, subID, keyID)
	if err != nil {
		return fmt.Errorf("failed to get validation key: %w", err)
	}
	if err := s.validator.Validate(ctx, ctx.Body, value, key); err != nil {
		return fmt.Errorf("sign validation failed: %w", err)
	}
	return nil
}

// 🔹 Validate Schema Step
type validateSchemaStep struct {
	validator definition.SchemaValidator
}

// newValidateSchemaStep creates and returns the validateSchema step after validation
func newValidateSchemaStep(schemaValidator definition.SchemaValidator) (definition.Step, error) {
	if schemaValidator == nil {
		return nil, fmt.Errorf("invalid config: SchemaValidator plugin not configured")
	}
	log.Debug(context.Background(), "adding schema validator")
	return &validateSchemaStep{validator: schemaValidator}, nil
}

func (s *validateSchemaStep) Run(ctx *model.StepContext) error {
	if err := s.validator.Validate(ctx, ctx.Request.URL, ctx.Body); err != nil {
		return fmt.Errorf("schema validation failed: %w", err)
	}
	return nil
}

// 🔹 Get Route Step
type addRouteStep struct {
	router definition.Router
}

// newRouteStep creates and returns the addRoute step after validation
func newRouteStep(router definition.Router) (definition.Step, error) {
	if router == nil {
		return nil, fmt.Errorf("invalid config: Router plugin not configured")
	}
	return &addRouteStep{router: router}, nil
}

func (s *addRouteStep) Run(ctx *model.StepContext) error {
	route, err := s.router.Route(ctx, ctx.Request.URL, ctx.Body)
	if err != nil {
		return fmt.Errorf("failed to determine route: %w", err)
	}
	log.Debugf(ctx, "Routing to %#v", route)
	ctx.Route = route

	log.Debugf(ctx, "ctx.Route to %#v", ctx.Route)
	return nil
}

// 🔹 Broadcast Step (Stub Implementation)
type broadcastStep struct{}

func (b *broadcastStep) Run(ctx *model.StepContext) error {
	// TODO: Implement broadcast logic if needed
	return nil
}

// 🔹 Subscribe Step (Stub Implementation)
type subscribeStep struct{}

func (s *subscribeStep) Run(ctx *model.StepContext) error {
	// TODO: Implement subscription logic if needed
	return nil
}

// tracingStep wraps a Step with OpenTelemetry tracing
type tracingStep struct {
	step definition.Step
	name string
}

// Run executes the step with tracing.
func (t *tracingStep) Run(ctx *model.StepContext) error {
	// Preserve original context
	orgCtx := ctx.Context

	// Start tracing
	tracer := otel.Tracer("step")
	newCtx, span := tracer.Start(orgCtx, t.name)
	defer span.End()

	// Set the new context safely.
	ctx.WithContext(newCtx)

	// Record execution time.
	startTime := time.Now()
	err := t.step.Run(ctx)
	duration := time.Since(startTime)

	// Set span attributes.
	span.SetAttributes(
		attribute.String("step.name", t.name),
		attribute.Float64("execution_time_ms", duration.Seconds()*1000),
	)

	// Capture errors.
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}

	// Restore original context to avoid unintended modifications.
	ctx.WithContext(orgCtx)

	return err
}

// traceWrapper wraps a Step with tracing
func traceWrapper(name string, step definition.Step) definition.Step {
	return &tracingStep{step: step, name: name}
}
