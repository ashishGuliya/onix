package gcpAuthMdw

import (
	"context"
	"fmt"
	"net/http"

	"github.com/ashishGuliya/onix/pkg/log"

	credentials "cloud.google.com/go/iam/credentials/apiv1"
	"cloud.google.com/go/iam/credentials/apiv1/credentialspb"
)

func New(ctx context.Context, cfg map[string]string) func(http.Handler) http.Handler {
	audience, ok := cfg["audience"]
	if !ok {
		panic("missing required 'audience' config")
	}

	creds, err := credentials.NewIamCredentialsClient(ctx)
	if err != nil {
		panic(fmt.Sprintf("failed to create IAM credentials client: %v", err))
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenResp, err := creds.GenerateIdToken(r.Context(), &credentialspb.GenerateIdTokenRequest{ // <-- Fixed package reference
				Name:         cfg["serviceAccount"],
				Audience:     audience,
				IncludeEmail: true,
			})

			if err != nil {
				log.Errorf(r.Context(), err, "Failed to generate identity token")
				http.Error(w, "Failed to generate identity token", http.StatusInternalServerError)
				return
			}

			r.Header.Set("Authorization", "Bearer "+tokenResp.Token)
			next.ServeHTTP(w, r)
		})
	}
}
