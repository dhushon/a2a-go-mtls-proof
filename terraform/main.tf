terraform {
  required_providers {
    okta = {
      source  = "okta/okta"
      version = "~> 4.7"
    }
  }
}

# This assumes you have your Okta provider configured via environment variables
# (OKTA_ORG_NAME and OKTA_API_TOKEN)
provider "okta" {}

# 1. Create the Service Application for the Client
resource "okta_app_oauth" "client_app" {
  label                      = "A2A mTLS Proof Client"
  type                       = "service"
  grant_types                = ["client_credentials", "urn:ietf:params:oauth:grant-type:token-exchange"]
  token_endpoint_auth_method = "client_secret_basic"
  response_types             = ["token"]
}

# 2. Create the Custom Authorization Server
resource "okta_auth_server" "auth_server" {
  name      = "A2A mTLS Proof Auth Server"
  audiences = ["api://responder-agent/access"]
}

# 3. Create a Custom Scope for the API
resource "okta_auth_server_scope" "api_scope" {
  auth_server_id = okta_auth_server.auth_server.id
  name           = "api://responder-agent/access"
  display_name   = "Access the Responder Agent"
  description    = "Allows the client to access the responder agent's API."
  consent        = "IMPLICIT"
}

# 4. Create the 'cnf' Claim
# This claim will contain the SHA-256 thumbprint of the client certificate.
# Okta's Expression Language provides access to the certificate from the mTLS handshake.
resource "okta_auth_server_claim" "cnf_claim" {
  auth_server_id          = okta_auth_server.auth_server.id
  name                    = "cnf"
  claim_type              = "RESOURCE"
  value_type              = "EXPRESSION"
  value                   = "'{'\"x5t#S256\"': tls.clientCertificate.hash_sha256}'"
  scopes                  = [okta_auth_server_scope.api_scope.name]
  always_include_in_token = true
}

# 5. Create a Policy for the Authorization Server
resource "okta_auth_server_policy" "api_policy" {
  auth_server_id = okta_auth_server.auth_server.id
  name           = "API Access Policy"
  description    = "Policy for clients accessing the API."
  priority       = 1
  client_whitelist = [
    okta_app_oauth.client_app.id
  ]
}

# 6. Create a Policy Rule to Allow the OBO Flow
resource "okta_auth_server_policy_rule" "obo_rule" {
  auth_server_id = okta_auth_server.auth_server.id
  policy_id      = okta_auth_server_policy.api_policy.id
  name           = "Allow OBO Flow"
  priority       = 1
  status         = "ACTIVE"

  grant_type_whitelist = ["urn:ietf:params:oauth:grant-type:token-exchange"]
  scope_whitelist      = [okta_auth_server_scope.api_scope.id]
}

# Output the client credentials and the authorization server's metadata URL
output "client_id" {
  value = okta_app_oauth.client_app.client_id
}

output "client_secret" {
  value     = okta_app_oauth.client_app.client_secret
  sensitive = true
}

output "authorization_server_metadata_url" {
  value = okta_auth_server.auth_server.issuer
}
