package resources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/Trozz/terraform-provider-pocketid/internal/client"
)

// applicationConfigID is the fixed identifier used for the singleton
// application configuration resource.
const applicationConfigID = "application-configuration"

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &applicationConfigResource{}
	_ resource.ResourceWithConfigure   = &applicationConfigResource{}
	_ resource.ResourceWithImportState = &applicationConfigResource{}
)

// NewApplicationConfigResource is a helper function to simplify the provider implementation.
func NewApplicationConfigResource() resource.Resource {
	return &applicationConfigResource{}
}

// applicationConfigResource is the resource implementation.
type applicationConfigResource struct {
	client *client.Client
}

// applicationConfigModel maps the application configuration schema data. It is
// shared in shape with the data source model.
type applicationConfigModel struct {
	ID types.String `tfsdk:"id"`

	// General
	AppName                   types.String `tfsdk:"app_name"`
	SessionDuration           types.String `tfsdk:"session_duration"`
	HomePageURL               types.String `tfsdk:"home_page_url"`
	EmailsVerified            types.String `tfsdk:"emails_verified"`
	DisableAnimations         types.String `tfsdk:"disable_animations"`
	AllowOwnAccountEdit       types.String `tfsdk:"allow_own_account_edit"`
	AllowUserSignups          types.String `tfsdk:"allow_user_signups"`
	SignupDefaultUserGroupIDs types.String `tfsdk:"signup_default_user_group_ids"`
	SignupDefaultCustomClaims types.String `tfsdk:"signup_default_custom_claims"`
	AccentColor               types.String `tfsdk:"accent_color"`
	RequireUserEmail          types.String `tfsdk:"require_user_email"`

	// Email / SMTP
	SmtpHost           types.String `tfsdk:"smtp_host"`
	SmtpPort           types.String `tfsdk:"smtp_port"`
	SmtpFrom           types.String `tfsdk:"smtp_from"`
	SmtpUser           types.String `tfsdk:"smtp_user"`
	SmtpPassword       types.String `tfsdk:"smtp_password"`
	SmtpTls            types.String `tfsdk:"smtp_tls"`
	SmtpSkipCertVerify types.String `tfsdk:"smtp_skip_cert_verify"`

	EmailOneTimeAccessAsAdminEnabled           types.String `tfsdk:"email_one_time_access_as_admin_enabled"`
	EmailOneTimeAccessAsUnauthenticatedEnabled types.String `tfsdk:"email_one_time_access_as_unauthenticated_enabled"`
	EmailLoginNotificationEnabled              types.String `tfsdk:"email_login_notification_enabled"`
	EmailApiKeyExpirationEnabled               types.String `tfsdk:"email_api_key_expiration_enabled"`
	EmailVerificationEnabled                   types.String `tfsdk:"email_verification_enabled"`

	// LDAP
	LdapEnabled                        types.String `tfsdk:"ldap_enabled"`
	LdapUrl                            types.String `tfsdk:"ldap_url"`
	LdapBindDn                         types.String `tfsdk:"ldap_bind_dn"`
	LdapBindPassword                   types.String `tfsdk:"ldap_bind_password"`
	LdapBase                           types.String `tfsdk:"ldap_base"`
	LdapUserSearchFilter               types.String `tfsdk:"ldap_user_search_filter"`
	LdapUserGroupSearchFilter          types.String `tfsdk:"ldap_user_group_search_filter"`
	LdapSkipCertVerify                 types.String `tfsdk:"ldap_skip_cert_verify"`
	LdapAttributeUserUniqueIdentifier  types.String `tfsdk:"ldap_attribute_user_unique_identifier"`
	LdapAttributeUserUsername          types.String `tfsdk:"ldap_attribute_user_username"`
	LdapAttributeUserEmail             types.String `tfsdk:"ldap_attribute_user_email"`
	LdapAttributeUserFirstName         types.String `tfsdk:"ldap_attribute_user_first_name"`
	LdapAttributeUserLastName          types.String `tfsdk:"ldap_attribute_user_last_name"`
	LdapAttributeUserDisplayName       types.String `tfsdk:"ldap_attribute_user_display_name"`
	LdapAttributeUserProfilePicture    types.String `tfsdk:"ldap_attribute_user_profile_picture"`
	LdapAttributeGroupMember           types.String `tfsdk:"ldap_attribute_group_member"`
	LdapAttributeGroupUniqueIdentifier types.String `tfsdk:"ldap_attribute_group_unique_identifier"`
	LdapAttributeGroupName             types.String `tfsdk:"ldap_attribute_group_name"`
	LdapAdminGroupName                 types.String `tfsdk:"ldap_admin_group_name"`
	LdapSoftDeleteUsers                types.String `tfsdk:"ldap_soft_delete_users"`
}

// applicationConfigToModel maps a client.ApplicationConfig onto the framework
// model, preserving the singleton ID.
func applicationConfigToModel(cfg *client.ApplicationConfig, m *applicationConfigModel) {
	m.ID = types.StringValue(applicationConfigID)

	m.AppName = types.StringValue(cfg.AppName)
	m.SessionDuration = types.StringValue(cfg.SessionDuration)
	m.HomePageURL = types.StringValue(cfg.HomePageURL)
	m.EmailsVerified = types.StringValue(cfg.EmailsVerified)
	m.DisableAnimations = types.StringValue(cfg.DisableAnimations)
	m.AllowOwnAccountEdit = types.StringValue(cfg.AllowOwnAccountEdit)
	m.AllowUserSignups = types.StringValue(cfg.AllowUserSignups)
	m.SignupDefaultUserGroupIDs = types.StringValue(cfg.SignupDefaultUserGroupIDs)
	m.SignupDefaultCustomClaims = types.StringValue(cfg.SignupDefaultCustomClaims)
	m.AccentColor = types.StringValue(cfg.AccentColor)
	m.RequireUserEmail = types.StringValue(cfg.RequireUserEmail)

	m.SmtpHost = types.StringValue(cfg.SmtpHost)
	m.SmtpPort = types.StringValue(cfg.SmtpPort)
	m.SmtpFrom = types.StringValue(cfg.SmtpFrom)
	m.SmtpUser = types.StringValue(cfg.SmtpUser)
	m.SmtpPassword = types.StringValue(cfg.SmtpPassword)
	m.SmtpTls = types.StringValue(cfg.SmtpTls)
	m.SmtpSkipCertVerify = types.StringValue(cfg.SmtpSkipCertVerify)

	m.EmailOneTimeAccessAsAdminEnabled = types.StringValue(cfg.EmailOneTimeAccessAsAdminEnabled)
	m.EmailOneTimeAccessAsUnauthenticatedEnabled = types.StringValue(cfg.EmailOneTimeAccessAsUnauthenticatedEnabled)
	m.EmailLoginNotificationEnabled = types.StringValue(cfg.EmailLoginNotificationEnabled)
	m.EmailApiKeyExpirationEnabled = types.StringValue(cfg.EmailApiKeyExpirationEnabled)
	m.EmailVerificationEnabled = types.StringValue(cfg.EmailVerificationEnabled)

	m.LdapEnabled = types.StringValue(cfg.LdapEnabled)
	m.LdapUrl = types.StringValue(cfg.LdapUrl)
	m.LdapBindDn = types.StringValue(cfg.LdapBindDn)
	m.LdapBindPassword = types.StringValue(cfg.LdapBindPassword)
	m.LdapBase = types.StringValue(cfg.LdapBase)
	m.LdapUserSearchFilter = types.StringValue(cfg.LdapUserSearchFilter)
	m.LdapUserGroupSearchFilter = types.StringValue(cfg.LdapUserGroupSearchFilter)
	m.LdapSkipCertVerify = types.StringValue(cfg.LdapSkipCertVerify)
	m.LdapAttributeUserUniqueIdentifier = types.StringValue(cfg.LdapAttributeUserUniqueIdentifier)
	m.LdapAttributeUserUsername = types.StringValue(cfg.LdapAttributeUserUsername)
	m.LdapAttributeUserEmail = types.StringValue(cfg.LdapAttributeUserEmail)
	m.LdapAttributeUserFirstName = types.StringValue(cfg.LdapAttributeUserFirstName)
	m.LdapAttributeUserLastName = types.StringValue(cfg.LdapAttributeUserLastName)
	m.LdapAttributeUserDisplayName = types.StringValue(cfg.LdapAttributeUserDisplayName)
	m.LdapAttributeUserProfilePicture = types.StringValue(cfg.LdapAttributeUserProfilePicture)
	m.LdapAttributeGroupMember = types.StringValue(cfg.LdapAttributeGroupMember)
	m.LdapAttributeGroupUniqueIdentifier = types.StringValue(cfg.LdapAttributeGroupUniqueIdentifier)
	m.LdapAttributeGroupName = types.StringValue(cfg.LdapAttributeGroupName)
	m.LdapAdminGroupName = types.StringValue(cfg.LdapAdminGroupName)
	m.LdapSoftDeleteUsers = types.StringValue(cfg.LdapSoftDeleteUsers)
}

// mergedString returns the planned value if it is set (known and non-null),
// otherwise it falls back to the current server-side value. This lets unset
// attributes inherit the existing configuration so that required server-side
// fields are never sent as empty values.
func mergedString(planned types.String, current string) string {
	if planned.IsNull() || planned.IsUnknown() {
		return current
	}
	return planned.ValueString()
}

// modelToApplicationConfig builds the client payload from the plan, merging in
// the current server values for any attribute that is not explicitly set.
func modelToApplicationConfig(plan *applicationConfigModel, current *client.ApplicationConfig) *client.ApplicationConfig {
	return &client.ApplicationConfig{
		AppName:                   mergedString(plan.AppName, current.AppName),
		SessionDuration:           mergedString(plan.SessionDuration, current.SessionDuration),
		HomePageURL:               mergedString(plan.HomePageURL, current.HomePageURL),
		EmailsVerified:            mergedString(plan.EmailsVerified, current.EmailsVerified),
		DisableAnimations:         mergedString(plan.DisableAnimations, current.DisableAnimations),
		AllowOwnAccountEdit:       mergedString(plan.AllowOwnAccountEdit, current.AllowOwnAccountEdit),
		AllowUserSignups:          mergedString(plan.AllowUserSignups, current.AllowUserSignups),
		SignupDefaultUserGroupIDs: mergedString(plan.SignupDefaultUserGroupIDs, current.SignupDefaultUserGroupIDs),
		SignupDefaultCustomClaims: mergedString(plan.SignupDefaultCustomClaims, current.SignupDefaultCustomClaims),
		AccentColor:               mergedString(plan.AccentColor, current.AccentColor),
		RequireUserEmail:          mergedString(plan.RequireUserEmail, current.RequireUserEmail),

		SmtpHost:           mergedString(plan.SmtpHost, current.SmtpHost),
		SmtpPort:           mergedString(plan.SmtpPort, current.SmtpPort),
		SmtpFrom:           mergedString(plan.SmtpFrom, current.SmtpFrom),
		SmtpUser:           mergedString(plan.SmtpUser, current.SmtpUser),
		SmtpPassword:       mergedString(plan.SmtpPassword, current.SmtpPassword),
		SmtpTls:            mergedString(plan.SmtpTls, current.SmtpTls),
		SmtpSkipCertVerify: mergedString(plan.SmtpSkipCertVerify, current.SmtpSkipCertVerify),

		EmailOneTimeAccessAsAdminEnabled:           mergedString(plan.EmailOneTimeAccessAsAdminEnabled, current.EmailOneTimeAccessAsAdminEnabled),
		EmailOneTimeAccessAsUnauthenticatedEnabled: mergedString(plan.EmailOneTimeAccessAsUnauthenticatedEnabled, current.EmailOneTimeAccessAsUnauthenticatedEnabled),
		EmailLoginNotificationEnabled:              mergedString(plan.EmailLoginNotificationEnabled, current.EmailLoginNotificationEnabled),
		EmailApiKeyExpirationEnabled:               mergedString(plan.EmailApiKeyExpirationEnabled, current.EmailApiKeyExpirationEnabled),
		EmailVerificationEnabled:                   mergedString(plan.EmailVerificationEnabled, current.EmailVerificationEnabled),

		LdapEnabled:                        mergedString(plan.LdapEnabled, current.LdapEnabled),
		LdapUrl:                            mergedString(plan.LdapUrl, current.LdapUrl),
		LdapBindDn:                         mergedString(plan.LdapBindDn, current.LdapBindDn),
		LdapBindPassword:                   mergedString(plan.LdapBindPassword, current.LdapBindPassword),
		LdapBase:                           mergedString(plan.LdapBase, current.LdapBase),
		LdapUserSearchFilter:               mergedString(plan.LdapUserSearchFilter, current.LdapUserSearchFilter),
		LdapUserGroupSearchFilter:          mergedString(plan.LdapUserGroupSearchFilter, current.LdapUserGroupSearchFilter),
		LdapSkipCertVerify:                 mergedString(plan.LdapSkipCertVerify, current.LdapSkipCertVerify),
		LdapAttributeUserUniqueIdentifier:  mergedString(plan.LdapAttributeUserUniqueIdentifier, current.LdapAttributeUserUniqueIdentifier),
		LdapAttributeUserUsername:          mergedString(plan.LdapAttributeUserUsername, current.LdapAttributeUserUsername),
		LdapAttributeUserEmail:             mergedString(plan.LdapAttributeUserEmail, current.LdapAttributeUserEmail),
		LdapAttributeUserFirstName:         mergedString(plan.LdapAttributeUserFirstName, current.LdapAttributeUserFirstName),
		LdapAttributeUserLastName:          mergedString(plan.LdapAttributeUserLastName, current.LdapAttributeUserLastName),
		LdapAttributeUserDisplayName:       mergedString(plan.LdapAttributeUserDisplayName, current.LdapAttributeUserDisplayName),
		LdapAttributeUserProfilePicture:    mergedString(plan.LdapAttributeUserProfilePicture, current.LdapAttributeUserProfilePicture),
		LdapAttributeGroupMember:           mergedString(plan.LdapAttributeGroupMember, current.LdapAttributeGroupMember),
		LdapAttributeGroupUniqueIdentifier: mergedString(plan.LdapAttributeGroupUniqueIdentifier, current.LdapAttributeGroupUniqueIdentifier),
		LdapAttributeGroupName:             mergedString(plan.LdapAttributeGroupName, current.LdapAttributeGroupName),
		LdapAdminGroupName:                 mergedString(plan.LdapAdminGroupName, current.LdapAdminGroupName),
		LdapSoftDeleteUsers:                mergedString(plan.LdapSoftDeleteUsers, current.LdapSoftDeleteUsers),
	}
}

func optionalComputedString(description string, sensitive bool) schema.StringAttribute {
	return schema.StringAttribute{
		Description: description,
		Optional:    true,
		Computed:    true,
		Sensitive:   sensitive,
	}
}

// Metadata returns the resource type name.
func (r *applicationConfigResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_application_config"
}

// Schema defines the schema for the resource.
func (r *applicationConfigResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Manages the global application configuration of a Pocket-ID instance.",
		MarkdownDescription: "Manages the global application configuration of a Pocket-ID instance. This is a singleton resource: only one should exist per instance. Any attribute left unset inherits the current server-side value, and removing the resource from configuration leaves the live configuration untouched.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Fixed identifier of the application configuration singleton.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},

			"app_name":                      optionalComputedString("The name of the application.", false),
			"session_duration":              optionalComputedString("Session duration in minutes.", false),
			"home_page_url":                 optionalComputedString("URL of the application home page.", false),
			"emails_verified":               optionalComputedString("Whether user emails are considered verified (\"true\" or \"false\").", false),
			"disable_animations":            optionalComputedString("Whether to disable UI animations (\"true\" or \"false\").", false),
			"allow_own_account_edit":        optionalComputedString("Whether users can edit their own account (\"true\" or \"false\").", false),
			"allow_user_signups":            optionalComputedString("User signup mode: \"disabled\", \"withToken\", or \"open\".", false),
			"signup_default_user_group_ids": optionalComputedString("JSON array of user group IDs assigned to users created via signup.", false),
			"signup_default_custom_claims":  optionalComputedString("JSON object of custom claims assigned to users created via signup.", false),
			"accent_color":                  optionalComputedString("Accent color used in the UI.", false),
			"require_user_email":            optionalComputedString("Whether a user email is required (\"true\" or \"false\").", false),

			"smtp_host":             optionalComputedString("SMTP server host.", false),
			"smtp_port":             optionalComputedString("SMTP server port.", false),
			"smtp_from":             optionalComputedString("Email address used as the sender.", false),
			"smtp_user":             optionalComputedString("SMTP authentication user.", false),
			"smtp_password":         optionalComputedString("SMTP authentication password.", true),
			"smtp_tls":              optionalComputedString("SMTP TLS mode: \"none\", \"starttls\", or \"tls\".", false),
			"smtp_skip_cert_verify": optionalComputedString("Whether to skip SMTP certificate verification (\"true\" or \"false\").", false),

			"email_one_time_access_as_admin_enabled":           optionalComputedString("Whether admins can use one-time access email links (\"true\" or \"false\").", false),
			"email_one_time_access_as_unauthenticated_enabled": optionalComputedString("Whether unauthenticated users can request one-time access email links (\"true\" or \"false\").", false),
			"email_login_notification_enabled":                 optionalComputedString("Whether login notification emails are enabled (\"true\" or \"false\").", false),
			"email_api_key_expiration_enabled":                 optionalComputedString("Whether API key expiration emails are enabled (\"true\" or \"false\").", false),
			"email_verification_enabled":                       optionalComputedString("Whether email verification is enabled (\"true\" or \"false\").", false),

			"ldap_enabled":                           optionalComputedString("Whether LDAP integration is enabled (\"true\" or \"false\").", false),
			"ldap_url":                               optionalComputedString("LDAP server URL.", false),
			"ldap_bind_dn":                           optionalComputedString("LDAP bind DN.", false),
			"ldap_bind_password":                     optionalComputedString("LDAP bind password.", true),
			"ldap_base":                              optionalComputedString("LDAP search base.", false),
			"ldap_user_search_filter":                optionalComputedString("LDAP user search filter.", false),
			"ldap_user_group_search_filter":          optionalComputedString("LDAP user group search filter.", false),
			"ldap_skip_cert_verify":                  optionalComputedString("Whether to skip LDAP certificate verification (\"true\" or \"false\").", false),
			"ldap_attribute_user_unique_identifier":  optionalComputedString("LDAP attribute for the user unique identifier.", false),
			"ldap_attribute_user_username":           optionalComputedString("LDAP attribute for the username.", false),
			"ldap_attribute_user_email":              optionalComputedString("LDAP attribute for the user email.", false),
			"ldap_attribute_user_first_name":         optionalComputedString("LDAP attribute for the user first name.", false),
			"ldap_attribute_user_last_name":          optionalComputedString("LDAP attribute for the user last name.", false),
			"ldap_attribute_user_display_name":       optionalComputedString("LDAP attribute for the user display name.", false),
			"ldap_attribute_user_profile_picture":    optionalComputedString("LDAP attribute for the user profile picture.", false),
			"ldap_attribute_group_member":            optionalComputedString("LDAP attribute for group membership.", false),
			"ldap_attribute_group_unique_identifier": optionalComputedString("LDAP attribute for the group unique identifier.", false),
			"ldap_attribute_group_name":              optionalComputedString("LDAP attribute for the group name.", false),
			"ldap_admin_group_name":                  optionalComputedString("LDAP group name granting admin privileges.", false),
			"ldap_soft_delete_users":                 optionalComputedString("Whether to soft-delete users removed from LDAP (\"true\" or \"false\").", false),
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *applicationConfigResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = c
}

// applyConfig merges the plan with the current server config, performs the PUT
// and writes the response back into the plan model.
func (r *applicationConfigResource) applyConfig(ctx context.Context, plan *applicationConfigModel, diags *diag.Diagnostics) {
	current, err := r.client.GetApplicationConfig()
	if err != nil {
		diags.AddError(
			"Error reading application configuration",
			"Could not read current application configuration: "+err.Error(),
		)
		return
	}

	payload := modelToApplicationConfig(plan, current)

	tflog.Debug(ctx, "Updating application configuration")

	updated, err := r.client.UpdateApplicationConfig(payload)
	if err != nil {
		diags.AddError(
			"Error updating application configuration",
			"Could not update application configuration: "+err.Error(),
		)
		return
	}

	applicationConfigToModel(updated, plan)
}

// Create creates the resource and sets the initial Terraform state.
func (r *applicationConfigResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan applicationConfigModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.applyConfig(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *applicationConfigResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state applicationConfigModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading application configuration")

	cfg, err := r.client.GetApplicationConfig()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading application configuration",
			"Could not read application configuration: "+err.Error(),
		)
		return
	}

	applicationConfigToModel(cfg, &state)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *applicationConfigResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan applicationConfigModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.applyConfig(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete removes the resource from state. The application configuration is a
// singleton that always exists, so the live configuration is left untouched.
func (r *applicationConfigResource) Delete(ctx context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
	tflog.Info(ctx, "Removing application configuration from state; the live Pocket-ID configuration is left unchanged")
}

// ImportState imports the singleton application configuration into Terraform.
func (r *applicationConfigResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
