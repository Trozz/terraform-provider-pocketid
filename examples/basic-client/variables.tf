variable "base_url" {
  description = "The base URL of your Pocket-ID instance"
  type        = string
}

variable "api_token" {
  description = "API token for Pocket-ID authentication"
  type        = string
  sensitive   = true
}

variable "client_name" {
  description = "Name of the OIDC client"
  type        = string
  default     = "My Web Application"
}

variable "callback_urls" {
  description = "OAuth2 callback URLs for the client"
  type        = list(string)
  default = [
    "http://localhost:3000/auth/callback",
    "https://myapp.example.com/auth/callback"
  ]
}

variable "logout_callback_urls" {
  description = "Logout callback URLs for the client"
  type        = list(string)
  default = [
    "http://localhost:3000",
    "https://myapp.example.com"
  ]
}
