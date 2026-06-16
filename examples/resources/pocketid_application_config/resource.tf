# Manage the global application configuration of a Pocket-ID instance.
#
# This is a singleton resource: only one should exist per instance. Any
# attribute left unset inherits the current server-side value. Removing the
# resource from your configuration leaves the live configuration untouched.
resource "pocketid_application_config" "this" {
  app_name         = "My Company SSO"
  session_duration = "60"
  accent_color     = "#3b82f6"

  allow_user_signups = "disabled"
  require_user_email = "true"
  disable_animations = "false"
  emails_verified    = "false"
}

# Example: configure SMTP for outgoing email
resource "pocketid_application_config" "with_smtp" {
  app_name = "My Company SSO"

  smtp_host             = "smtp.example.com"
  smtp_port             = "587"
  smtp_from             = "no-reply@example.com"
  smtp_user             = "smtp-user"
  smtp_password         = var.smtp_password # mark sensitive in your variables
  smtp_tls              = "starttls"
  smtp_skip_cert_verify = "false"

  email_login_notification_enabled = "true"
  email_verification_enabled       = "true"
}
