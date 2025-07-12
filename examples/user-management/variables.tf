variable "base_url" {
  description = "The base URL of your Pocket-ID instance"
  type        = string
}

variable "api_token" {
  description = "API token for Pocket-ID authentication"
  type        = string
  sensitive   = true
}

variable "admin_username" {
  description = "Username for the main admin user"
  type        = string
  default     = "terraform.admin"
}

variable "admin_email" {
  description = "Email for the main admin user"
  type        = string
  default     = "admin@example.com"
}
