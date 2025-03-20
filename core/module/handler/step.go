package handler

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ashishGuliya/onix/pkg/log"
	"github.com/ashishGuliya/onix/pkg/plugin/definition"
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

func (s *signStep) Run(ctx *definition.StepContext) error {
	subID, ok := ctx.Value("subscriber_id").(string)
	if !ok {
		return fmt.Errorf("failed to sign request: Subscriber  id not set")
	}
	keyID, key, err := s.km.SigningPrivateKey(ctx, subID)
	if err != nil {
		return fmt.Errorf("failed to get signing key: %w", err)
	}
	createdAt := time.Now().Unix()
	validTill := time.Now().Add(5 * time.Minute).Unix()
	sign, err := s.signer.Sign(ctx, ctx.Body, key, createdAt, validTill)
	if err != nil {
		return fmt.Errorf("failed to sign request: %w", err)
	}
	authHeader := fmt.Sprintf("Signature keyId=\"%s|%s|ed25519\",algorithm=\"ed25519\",created=\"%d\",expires=\"%d\",headers=\"(created) (expires) digest\",signature=\"%s\"", subID, keyID, createdAt, validTill, sign)
	ctx.Request.Header.Set("Authorization", authHeader)
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

func (s *validateSignStep) Run(ctx *definition.StepContext) error {
	headerValue := ctx.Request.Header.Get("Authorization")
	headerParts := strings.Split(headerValue, "|")
	subID := strings.Split(headerParts[0], "\"")[1]
	keyID := headerParts[1]
	key, err := s.km.SigningPublicKey(ctx, subID, keyID)
	if err != nil {
		return fmt.Errorf("failed to get validation key: %w", err)
	}
	if err := s.validator.Validate(ctx, ctx.Body, headerValue, key); err != nil {
		return fmt.Errorf("sign validation failed: %w", err)
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
	log.Debugf(ctx, "Routing to %#v", route)
	ctx.Route = route

	log.Debugf(ctx, "ctx.Route to %#v", ctx.Route)
	return nil
}

// newValidateSchemaStep creates and returns the validateSchema step after validation
func newValidateSchemaStep(schemaValidator definition.SchemaValidator) (definition.Step, error) {
	if schemaValidator == nil {
		return nil, fmt.Errorf("invalid config: SchemaValidator plugin not configured")
	}
	log.Debug(context.Background(), "adding schema validator")
	return &validateSchemaStep{validator: schemaValidator}, nil
}

// newRouteStep creates and returns the addRoute step after validation
func newRouteStep(router definition.Router) (definition.Step, error) {
	if router == nil {
		return nil, fmt.Errorf("invalid config: Router plugin not configured")
	}
	return &addRouteStep{router: router}, nil
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
