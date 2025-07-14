terraform {
  required_providers {
    pocketid = {
      source  = "Trozz/pocketid"
      version = "~> 0.1.0"
    }
  }
}

provider "pocketid" {
  base_url  = var.pocketid_base_url
  api_token = var.pocketid_api_token
}

# Variables
variable "pocketid_base_url" {
  description = "Base URL for PocketID instance"
  type        = string
  default     = "http://localhost:1411"
}

variable "pocketid_api_token" {
  description = "API token for PocketID authentication"
  type        = string
  sensitive   = true
}

# Resource IDs from the resource creation test
# These would typically be passed in via variables or read from outputs
variable "admin_user_id" {
  description = "ID of the admin user created in resource test"
  type        = string
}

variable "dev_user_id" {
  description = "ID of the developer user created in resource test"
  type        = string
}

variable "admin_group_id" {
  description = "ID of the admin group created in resource test"
  type        = string
}

variable "developers_group_id" {
  description = "ID of the developers group created in resource test"
  type        = string
}

variable "web_app_client_id" {
  description = "ID of the web app client created in resource test"
  type        = string
}

variable "service_account_client_id" {
  description = "ID of the service account client created in resource test"
  type        = string
}

# Individual resource lookups by ID
data "pocketid_user" "lookup_admin" {
  id = var.admin_user_id
}

data "pocketid_user" "lookup_dev" {
  id = var.dev_user_id
}

data "pocketid_group" "lookup_admin_group" {
  id = var.admin_group_id
}

data "pocketid_group" "lookup_dev_group" {
  id = var.developers_group_id
}

data "pocketid_client" "lookup_web_client" {
  id = var.web_app_client_id
}

data "pocketid_client" "lookup_service_account" {
  id = var.service_account_client_id
}

# List all resources
data "pocketid_users" "all_users" {}

data "pocketid_groups" "all_groups" {}

data "pocketid_clients" "all_clients" {}

# Filtered lookups
data "pocketid_users" "enabled_users" {
  filter {
    name   = "enabled"
    values = ["true"]
  }
}

data "pocketid_users" "disabled_users" {
  filter {
    name   = "enabled"
    values = ["false"]
  }
}
