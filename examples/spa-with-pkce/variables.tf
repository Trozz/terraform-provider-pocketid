variable "base_url" {
  description = "The base URL of your Pocket-ID instance"
  type        = string
}

variable "api_token" {
  description = "API token for Pocket-ID authentication"
  type        = string
  sensitive   = true
}

variable "app_name" {
  description = "Name of your Single Page Application"
  type        = string
  default     = "My React App"
}

variable "app_urls" {
  description = "URLs where your SPA is hosted (without trailing slash)"
  type        = list(string)
  default = [
    "http://localhost:3000",
    "https://app.example.com"
  ]
}

variable "allowed_groups" {
  description = "Optional: User groups allowed to access this application"
  type        = list(string)
  default     = []
}
