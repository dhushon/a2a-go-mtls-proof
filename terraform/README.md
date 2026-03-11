# Okta Configuration for mTLS Proof-of-Concept

This document describes how to set up Okta as a custom OAuth2 authentication server for the `a2a-go-mtls-proof` project. The Terraform code in this directory will create the necessary Okta resources, but there are manual steps required in both the Okta UI and the Go code to complete the integration.

## 1. Prerequisites

*   An Okta account with administrative privileges.
*   [Terraform CLI](https://learn.hashicorp.com/tutorials/terraform/install-cli) installed.
*   The [Okta Terraform provider](https://registry.terraform.io/providers/okta/okta/latest/docs) configured with your Okta organization's URL (`OKTA_ORG_NAME`) and an API token (`OKTA_API_TOKEN`) as environment variables.

## 2. Terraform Setup

The Terraform code will provision the following in your Okta organization:
*   A custom authorization server.
*   A service application to represent the Go client.
*   A custom scope (`api://responder-agent/access`).
*   A custom `cnf` claim to hold the certificate thumbprint.
*   The necessary policies for the On-Behalf-Of (OBO) flow.

**To run the Terraform code:**

1.  Navigate to this directory (`./terraform`).
2.  Initialize Terraform:
    ```bash
    terraform init
    ```
3.  Apply the configuration:
    ```bash
    terraform apply
    ```
    After you approve the plan, Terraform will create the resources and output the following values. **Save these, as you will need them to configure the Go client.**
    *   `client_id`
    *   `client_secret`
    *   `authorization_server_metadata_url`

## 3. Okta Manual Configuration: Enforcing mTLS

The Terraform code cannot enforce mTLS on the token endpoint. You must do this manually in the Okta Admin UI.

1.  Log in to your Okta Admin Console.
2.  Navigate to **Security > API**.
3.  Select the **A2A mTLS Proof Auth Server** from the list of authorization servers.
4.  Click on the **Access Policies** tab.
5.  Edit the **API Access Policy**.
6.  Click on the **Add Rule** button.
7.  Give the rule a name (e.g., "Enforce mTLS").
8.  In the **IF** section, for **Client is**, select **Any client**.
9.  In the **THEN** section, for **Client authentication**, select **TLS (Transport Layer Security)**.
10. Click **Create Rule**.

This rule ensures that any client calling the token endpoint of this authorization server must successfully complete an mTLS handshake.

## 4. Go Application Configuration

You will need to update your Go client and server to use the new Okta configuration.

### Client (`client/main.go`)

The client needs to be updated to perform a real OBO token exchange with Okta. You will need to use a library like `golang.org/x/oauth2`.

```go
// In your client/main.go

import (
    "context"
    "golang.org/x/oauth2/clientcredentials"
)

// In your main function:
func main() {
    // ... (GetClientTLSConfig remains the same)

    // Configure the OAuth2 client with the credentials from Terraform
    oauthConfig := clientcredentials.Config{
        ClientID:     "your-okta-client-id", // From Terraform output
        ClientSecret: "your-okta-client-secret", // From Terraform output
        TokenURL:     "https://your-okta-domain.com/oauth2/ausexample/v1/token", // Construct from metadata URL
        Scopes:       []string{"api://responder-agent/access"},
    }

    // ...

    // The getOBOToken function needs to be rewritten
    // This is a simplified example
    func getOBOToken(ctx context.Context, userToken string, tlsConfig *tls.Config, oauthConfig clientcredentials.Config) (string, error) {
        mtlsClient := &http.Client{
            Transport: &http.Transport{
                TLSClientConfig: tlsConfig,
            },
        }
        ctx = context.WithValue(ctx, oauth2.HTTPClient, mtlsClient)

        // The specifics of the OBO flow may vary based on your library and Okta's expectations.
        // You may need to manually add the 'subject_token' to the request.
        // This is a conceptual example.
        token, err := oauthConfig.Token(ctx)
        if err != nil {
            return "", err
        }
        return token.AccessToken, nil
    }
}
```

### Server (`server/main.go`)

The server must be updated to validate the incoming JWT against Okta's public keys.

```go
// In your server/main.go

import (
    "github.com/coreos/go-oidc/v3/oidc"
    "golang.org/x/oauth2"
)

// In your main function or an init() function:
func main() {
    // Use the metadata URL from Terraform to discover the provider's configuration
    provider, err := oidc.NewProvider(context.Background(), "https://your-okta-domain.com/oauth2/ausexample") // From Terraform output
    if err != nil {
        log.Fatalf("Failed to discover OIDC provider: %v", err)
    }

    // Create a verifier for the tokens.
    // The ClientID here is the audience of the token.
    verifier := provider.Verifier(&oidc.Config{ClientID: "api://responder-agent/access"})
    
    // ... (rest of server setup)

    // Pass the verifier to your middleware
    mux.Handle("/task", MTLSBindingMiddleware(finalHandler, verifier))
}

// Update the MTLSBindingMiddleware to accept the verifier
func MTLSBindingMiddleware(next http.Handler, verifier *oidc.IDTokenVerifier) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // ... (extract tokenStr from header)

        // Verify the token's signature and standard claims
        idToken, err := verifier.Verify(r.Context(), tokenStr)
        if err != nil {
            http.Error(w, "Invalid token", http.StatusUnauthorized)
            return
        }
        
        var claims Claims
        if err := idToken.Claims(&claims); err != nil {
            // handle error
        }
        
        // ... (The rest of the 'cnf' claim check remains the same)
    })
}
```

By following these steps, you will have a fully functional end-to-end mTLS and OBO flow, with Okta acting as your certificate-aware OAuth2 server.
