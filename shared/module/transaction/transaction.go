package transaction

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"

	"onix/shared/log"
	"onix/shared/module/config"
	"onix/shared/module/middleware"
	"onix/shared/plugin"
	"onix/shared/plugin/definition"
)

// callerSOP handles the proxying of requests.
func callerSOP(ctx context.Context, router definition.Router, signer definition.Signer) func(http.Handler) http.Handler {
	// Create the HTTP handler.
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			// Read the request body.
			body, err := io.ReadAll(r.Body)
			if err != nil {
				log.Errorf(r.Context(), err, "failed to read message body")
				http.Error(w, "Internal error", http.StatusInternalServerError)
				return
			}
			// Close the request body after reading.
			defer r.Body.Close()

			// Route the request.
			target, err := router.Target(r.Context(), body)
			if err != nil {
				log.Errorf(r.Context(), err, "routing error")
				http.Error(w, "Routing Error", http.StatusBadRequest)
				return
			}

			// Update the request URL.
			r.URL = target
			r.Header.Set("X-Forwarded-Host", r.Header.Get("Host"))
			sign, err := signer.Sign(ctx, body, "key")
			if err != nil {
				log.Errorf(r.Context(), err, "failed to sign request body")
				http.Error(w, "Internal error", http.StatusInternalServerError)
				return
			}
			r.Header.Set("SignHeader", sign)

			// Restore the request body for the next handler.
			r.Body = io.NopCloser(bytes.NewBuffer(body))

			next.ServeHTTP(w, r)
		})
	}
}

type PreProxyHookFunc func(ctx context.Context, r *http.Request, target *url.URL, body *[]byte) error

func reverseProxy(ctx context.Context, router definition.Router, signer definition.Signer, preProxyHooks ...PreProxyHookFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close() // Close the body after function execution

		// Read the request body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			log.Errorf(r.Context(), err, "failed to read message body")
			http.Error(w, "Failed to read request body", http.StatusInternalServerError)
			return
		}

		// Route the request
		target, err := router.Target(r.Context(), body)
		if err != nil {
			log.Errorf(r.Context(), err, "routing error")
			http.Error(w, "Invalid routing target", http.StatusBadRequest)
			return
		}

		// Sign the request
		sign, err := signer.Sign(ctx, body, "key")
		if err != nil {
			log.Errorf(r.Context(), err, "failed to sign request body")
			http.Error(w, "Failed to sign request", http.StatusInternalServerError)
			return
		}
		r.Header.Set("SignHeader", sign)

		// Execute hooks
		for _, hook := range preProxyHooks {
			if err := hook(ctx, r, target, &body); err != nil {
				log.Errorf(r.Context(), err, "hook error")
				http.Error(w, "Hook processing failed", http.StatusInternalServerError)
				return
			}
		}

		// Update the request URL properly
		r.URL.Scheme = target.Scheme
		r.URL.Host = target.Host
		r.URL.Path = target.Path

		// Set forwarding headers
		r.Header.Set("X-Forwarded-Host", r.Host)

		// Restore the request body
		r.Body = io.NopCloser(bytes.NewBuffer(body))

		// Create and configure the reverse proxy
		proxy := httputil.NewSingleHostReverseProxy(&url.URL{Scheme: target.Scheme, Host: target.Host})
		log.Debugf(r.Context(), "Proxying request to: %s", target)

		// Serve the request through the proxy
		proxy.ServeHTTP(w, r)
	})
}

// Guidelines for Implementing Pre-Proxy Hooks
/*
1. **Purpose:**
   - Pre-Proxy Hooks allow modification of the request, headers, body, and target URL before proxying.
   - Use them to inject custom processing logic like authentication, logging, or transformation.

2. **Signature:**
   ```go
   type PreProxyHookFunc func(ctx context.Context, r *http.Request, target *url.URL, body *[]byte) error
   ```

3. **Modifying the Request:**
   - You can change the request method, add custom headers, or modify query parameters:
   ```go
   func AddCustomHeader(ctx context.Context, r *http.Request, target *url.URL, body *[]byte) error {
       r.Header.Set("X-Custom-Header", "SomeValue")
       return nil
   }
   ```

4. **Modifying the Request Body:**
   - You can modify or replace the request body as needed:
   ```go
   func ModifyRequestBody(ctx context.Context, r *http.Request, target *url.URL, body *[]byte) error {
       newBody := append(*body, []byte("\nExtra Data")...)
       *body = newBody
       return nil
   }
   ```

5. **Modifying the Target URL:**
   - Redirect or change the target dynamically:
   ```go
   func ModifyTargetURL(ctx context.Context, r *http.Request, target *url.URL, body *[]byte) error {
       target.Host = "new-target.example.com"
       return nil
   }
   ```

6. **Handling Errors:**
   - If a hook encounters an issue, return an error to stop request processing:
   ```go
   func ValidateRequest(ctx context.Context, r *http.Request, target *url.URL, body *[]byte) error {
       if len(*body) == 0 {
           return fmt.Errorf("empty request body")
       }
       return nil
   }
   ```

7. **Registering Hooks:**
   - Pass hooks to `reverseProxy` when initializing:
   ```go
   handler := reverseProxy(ctx, router, signer, AddCustomHeader, ModifyRequestBody)
   ```
*/
func process(ctx context.Context) {

}

// NetworkCaller handles the proxying of requests.
func recieverSOP(ctx context.Context, router definition.Router, signValidator definition.SignValidator, schemaValidator definition.SchemaValidator) func(http.Handler) http.Handler {
	// Create the HTTP handler.
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, err := io.ReadAll(r.Body)
			if err != nil {
				log.Errorf(r.Context(), err, "failed to read request body")
				http.Error(w, "Failed to read request body", http.StatusBadRequest)
				return
			}
			defer r.Body.Close()
			if valid, err := signValidator.Verify(r.Context(), body, r.Header.Get("HeaderString"), "key"); !valid || err != nil {
				log.Errorf(r.Context(), err, "Sign Validation Error.")
				http.Error(w, "Sign Validation Error", http.StatusBadRequest)
				return
			}
			target, err := router.Target(r.Context(), body)
			if err != nil {
				log.Errorf(r.Context(), err, "routing error")
				http.Error(w, "Routing Error", http.StatusBadRequest)
				return
			}

			// Update the request URL.
			r.URL.Host = target.Host
			r.URL.Scheme = target.Scheme
			r.Header.Set("X-Forwarded-Host", r.Header.Get("Host"))

			if err := schemaValidator.Validate(ctx, body); err != nil {
				log.Errorf(r.Context(), err, "Schema Validation Error.")
				http.Error(w, "Schema Validation Error", http.StatusBadRequest)
				return
			}

			// Restore the request body for the next handler.
			r.Body = io.NopCloser(bytes.NewBuffer(body))
			next.ServeHTTP(w, r)
		})
	}

}

func RegisterReciever(ctx context.Context, mgr *plugin.Manager, c *config.ModuleCfg) (http.Handler, error) {
	log.Debugf(ctx, "Intitalizing Reciver with cfg: %#v", c)
	signValidator, err := mgr.SignValidator(ctx, c.Plugins.SignValidator)
	if err != nil {
		return nil, fmt.Errorf("failed to get sign validator: %w", err)
	}
	schemaValidator, err := mgr.Validator(ctx, c.Plugins.SchemaValidator)
	if err != nil {
		return nil, fmt.Errorf("failed to get schema validator: %w", err)
	}
	router, err := mgr.Router(ctx, c.Plugins.Router)
	if err != nil {
		return nil, fmt.Errorf("failed to get router: %w", err)
	}
	var reciever http.Handler
	switch c.TargetType {
	case "msgQ":
		publisher, err := mgr.Publisher(ctx, c.Plugins.Publisher)
		if err != nil {
			return nil, fmt.Errorf("failed to get publisher: %w", err)
		}
		reciever = middleware.MsgPublisher(publisher)
	case "http":
		reciever = middleware.ReverseProxy()
	default:
		return nil, fmt.Errorf("Invalid module caller type: %s", c.TargetType)
	}
	h, err := middleware.Chain(ctx, mgr, reciever, c.Plugins.PostProcessors)
	if err != nil {
		return nil, fmt.Errorf("failed to add post processors: %w", err)

	}

	h = recieverSOP(ctx, router, signValidator, schemaValidator)(h)
	h, err = middleware.Chain(ctx, mgr, h, c.Plugins.PreProcessors)
	if err != nil {
		return nil, fmt.Errorf("failed to add pre processors: %w", err)

	}
	return h, nil
}

func RegisterCaller(ctx context.Context, mgr *plugin.Manager, c *config.ModuleCfg) (http.Handler, error) {

	router, err := mgr.Router(ctx, c.Plugins.Router)
	if err != nil {
		return nil, fmt.Errorf("failed to get router: %w", err)
	}
	signer, err := mgr.Signer(ctx, c.Plugins.Signer)
	if err != nil {
		return nil, fmt.Errorf("failed to get signer: %w", err)
	}
	h, err := middleware.Chain(ctx, mgr, middleware.ReverseProxy(), c.Plugins.PostProcessors)
	if err != nil {
		return nil, fmt.Errorf("failed to add post processors: %w", err)
	}
	h = callerSOP(ctx, router, signer)(h)
	h, err = middleware.Chain(ctx, mgr, h, c.Plugins.PreProcessors)
	if err != nil {
		return nil, fmt.Errorf("failed to add pre processors: %w", err)

	}
	return h, nil
}

// // OutgoingJWTMiddleware generates and adds a JWT token to outgoing requests
// func OutgoingJWTMiddleware() func(http.Handler) http.Handler {
// 	return func(next http.Handler) http.Handler {
// 		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 			token, err := generateJWT()
// 			if err != nil {
// 				http.Error(w, "Failed to generate JWT", http.StatusInternalServerError)
// 				return
// 			}

// 			r.Header.Set("Authorization", "Bearer "+token)
// 			next.ServeHTTP(w, r)
// 		})
// 	}
// }

// // OutgoingGCPAuthMiddleware adds a Google IAM Identity Token to outgoing requests
// func OutgoingGCPAuthMiddleware(audience string) func(http.Handler) http.Handler {
// 	return func(next http.Handler) http.Handler {
// 		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 			ctx := context.Background()
// 			creds, err := credentials.NewIamCredentialsClient(ctx)
// 			if err != nil {
// 				http.Error(w, "Failed to get IAM credentials", http.StatusInternalServerError)
// 				return
// 			}
// 			defer creds.Close()

// 			tokenResp, err := creds.GenerateIDToken(ctx, &credentials.GenerateIDTokenRequest{
// 				Name:         "projects/-/serviceAccounts/default",
// 				Audience:     audience,
// 				IncludeEmail: true,
// 			})
// 			if err != nil {
// 				http.Error(w, "Failed to generate identity token", http.StatusInternalServerError)
// 				return
// 			}

// 			r.Header.Set("Authorization", "Bearer "+tokenResp.Token)
// 			next.ServeHTTP(w, r)
// 		})
// 	}
// }
