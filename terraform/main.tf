terraform {
  required_providers {
    okta = {
      source  = "okta/okta"
      version = "~> 4.7"
    }
  }
}

# 1. Create the Service Application for the Client (Requester Agent)
resource "okta_app_oauth" "client_app" {
  label                      = "A2A Zero-Trust Agent"
  type                       = "service"
  grant_types                = ["client_credentials", "urn:ietf:params:oauth:grant-type:token-exchange"]
  token_endpoint_auth_method = "private_key_jwt"
  response_types             = ["token"]

  jwks {
    kty = "RSA"
    kid = "tgr8kWQ3CTDlFE8hCMe9EHfMVRE9BvWVuTWrPIpK9sA"
    e   = "AQAB"
    n   = "qFOqQ6Maw9TgC9m-32pG8kiSIuAfzYJ8Xay9vOri4npPwIEnqDScdzKfqwV_RJ2mQAthfx0nHUBukxJqac8v-F9Jp97aWUlY__fV47ov5jq3XbCPCe0tHbJ5C0jVGtTClUvyE-AlVwA6dXI1QXsqHCEimTm0pF0d93O3BpiBL4EDy3okebR-RZfqZBCBbN6gNAnVCfiszSAGLkiMw2r77mxEmG02p3dvmPKHOHQKAHAEa9mFYNYz9VTEI20ZBmpBgeLek85KIKuXGDrgCt9Dyhj4U3ss2c5kwnEkz8-S514GcD_-ROWXY1ELLWmCG9dR6H41SCTAVkyoPWOdcnErOw"
  }
}

# 2. Create the Custom Authorization Server
resource "okta_auth_server" "auth_server" {
  name      = "A2A Zero-Trust Auth Server"
  audiences = ["api://a2a-agents"]
}

# 3. Create Custom Scopes for the Agents
resource "okta_auth_server_scope" "responder_scope" {
  auth_server_id = okta_auth_server.auth_server.id
  name           = "api://responder-agent/access"
  display_name   = "Access Responder Agent"
  consent        = "IMPLICIT"
}

resource "okta_auth_server_scope" "weather_scope" {
  auth_server_id = okta_auth_server.auth_server.id
  name           = "api://weather-agent/access"
  display_name   = "Access Weather Agent"
  consent        = "IMPLICIT"
}

# 4. Create the binding claim
# Renamed to avoid forbidden symbol '#'
resource "okta_auth_server_claim" "x5t_claim" {
  auth_server_id          = okta_auth_server.auth_server.id
  name                    = "x5t_S256"
  claim_type              = "RESOURCE"
  value_type              = "EXPRESSION"
  value                   = "tls.clientCertificate.hash_sha256"
  scopes                  = [
    okta_auth_server_scope.responder_scope.name,
    okta_auth_server_scope.weather_scope.name
  ]
  always_include_in_token = true
}

# 5. Create a Policy for the Authorization Server
resource "okta_auth_server_policy" "api_policy" {
  auth_server_id = okta_auth_server.auth_server.id
  name           = "Agent Access Policy"
  description    = "Allows Token Exchange for A2A agents."
  priority       = 1
  client_whitelist = [okta_app_oauth.client_app.id]
}

data "okta_group" "everyone" {
  name = "Everyone"
}

# 6. Create a Policy Rule to Allow the OBO (Token Exchange) Flow
resource "okta_auth_server_policy_rule" "obo_rule" {
  auth_server_id = okta_auth_server.auth_server.id
  policy_id      = okta_auth_server_policy.api_policy.id
  name           = "Allow Token Exchange"
  priority       = 1
  status         = "ACTIVE"

  grant_type_whitelist = ["urn:ietf:params:oauth:grant-type:token-exchange"]
  scope_whitelist      = [
    okta_auth_server_scope.responder_scope.name,
    okta_auth_server_scope.weather_scope.name
  ]
  group_whitelist = [data.okta_group.everyone.id]
}

# Outputs
output "client_id" {
  value = okta_app_oauth.client_app.client_id
}

output "auth_server_issuer" {
  value = okta_auth_server.auth_server.issuer
}
