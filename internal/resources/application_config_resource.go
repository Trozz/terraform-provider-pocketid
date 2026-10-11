package resources

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
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

	// Images
	Logo                  types.Object `tfsdk:"logo"`
	DarkLogo              types.Object `tfsdk:"dark_logo"`
	EmailLogo             types.Object `tfsdk:"email_logo"`
	BackgroundImage       types.Object `tfsdk:"background_image"`
	Favicon               types.Object `tfsdk:"favicon"`
	DefaultProfilePicture types.Object `tfsdk:"default_profile_picture"`
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
//
// The payload starts as a copy of the current server configuration because the
// update endpoint replaces the configuration in full: any field left at its
// zero value is reset server-side. Copying first keeps settings the provider
// does not expose, such as the WebAuthn options and the CIMD URL allowlist,
// from being wiped on every apply.
func modelToApplicationConfig(plan *applicationConfigModel, current *client.ApplicationConfig) *client.ApplicationConfig {
	cfg := *current
	cfg.AppName = mergedString(plan.AppName, current.AppName)
	cfg.SessionDuration = mergedString(plan.SessionDuration, current.SessionDuration)
	cfg.HomePageURL = mergedString(plan.HomePageURL, current.HomePageURL)
	cfg.EmailsVerified = mergedString(plan.EmailsVerified, current.EmailsVerified)
	cfg.DisableAnimations = mergedString(plan.DisableAnimations, current.DisableAnimations)
	cfg.AllowOwnAccountEdit = mergedString(plan.AllowOwnAccountEdit, current.AllowOwnAccountEdit)
	cfg.AllowUserSignups = mergedString(plan.AllowUserSignups, current.AllowUserSignups)
	cfg.SignupDefaultUserGroupIDs = mergedString(plan.SignupDefaultUserGroupIDs, current.SignupDefaultUserGroupIDs)
	cfg.SignupDefaultCustomClaims = mergedString(plan.SignupDefaultCustomClaims, current.SignupDefaultCustomClaims)
	cfg.AccentColor = mergedString(plan.AccentColor, current.AccentColor)
	cfg.RequireUserEmail = mergedString(plan.RequireUserEmail, current.RequireUserEmail)

	cfg.SmtpHost = mergedString(plan.SmtpHost, current.SmtpHost)
	cfg.SmtpPort = mergedString(plan.SmtpPort, current.SmtpPort)
	cfg.SmtpFrom = mergedString(plan.SmtpFrom, current.SmtpFrom)
	cfg.SmtpUser = mergedString(plan.SmtpUser, current.SmtpUser)
	cfg.SmtpPassword = mergedString(plan.SmtpPassword, current.SmtpPassword)
	cfg.SmtpTls = mergedString(plan.SmtpTls, current.SmtpTls)
	cfg.SmtpSkipCertVerify = mergedString(plan.SmtpSkipCertVerify, current.SmtpSkipCertVerify)

	cfg.EmailOneTimeAccessAsAdminEnabled = mergedString(plan.EmailOneTimeAccessAsAdminEnabled, current.EmailOneTimeAccessAsAdminEnabled)
	cfg.EmailOneTimeAccessAsUnauthenticatedEnabled = mergedString(plan.EmailOneTimeAccessAsUnauthenticatedEnabled, current.EmailOneTimeAccessAsUnauthenticatedEnabled)
	cfg.EmailLoginNotificationEnabled = mergedString(plan.EmailLoginNotificationEnabled, current.EmailLoginNotificationEnabled)
	cfg.EmailApiKeyExpirationEnabled = mergedString(plan.EmailApiKeyExpirationEnabled, current.EmailApiKeyExpirationEnabled)
	cfg.EmailVerificationEnabled = mergedString(plan.EmailVerificationEnabled, current.EmailVerificationEnabled)

	cfg.LdapEnabled = mergedString(plan.LdapEnabled, current.LdapEnabled)
	cfg.LdapUrl = mergedString(plan.LdapUrl, current.LdapUrl)
	cfg.LdapBindDn = mergedString(plan.LdapBindDn, current.LdapBindDn)
	cfg.LdapBindPassword = mergedString(plan.LdapBindPassword, current.LdapBindPassword)
	cfg.LdapBase = mergedString(plan.LdapBase, current.LdapBase)
	cfg.LdapUserSearchFilter = mergedString(plan.LdapUserSearchFilter, current.LdapUserSearchFilter)
	cfg.LdapUserGroupSearchFilter = mergedString(plan.LdapUserGroupSearchFilter, current.LdapUserGroupSearchFilter)
	cfg.LdapSkipCertVerify = mergedString(plan.LdapSkipCertVerify, current.LdapSkipCertVerify)
	cfg.LdapAttributeUserUniqueIdentifier = mergedString(plan.LdapAttributeUserUniqueIdentifier, current.LdapAttributeUserUniqueIdentifier)
	cfg.LdapAttributeUserUsername = mergedString(plan.LdapAttributeUserUsername, current.LdapAttributeUserUsername)
	cfg.LdapAttributeUserEmail = mergedString(plan.LdapAttributeUserEmail, current.LdapAttributeUserEmail)
	cfg.LdapAttributeUserFirstName = mergedString(plan.LdapAttributeUserFirstName, current.LdapAttributeUserFirstName)
	cfg.LdapAttributeUserLastName = mergedString(plan.LdapAttributeUserLastName, current.LdapAttributeUserLastName)
	cfg.LdapAttributeUserDisplayName = mergedString(plan.LdapAttributeUserDisplayName, current.LdapAttributeUserDisplayName)
	cfg.LdapAttributeUserProfilePicture = mergedString(plan.LdapAttributeUserProfilePicture, current.LdapAttributeUserProfilePicture)
	cfg.LdapAttributeGroupMember = mergedString(plan.LdapAttributeGroupMember, current.LdapAttributeGroupMember)
	cfg.LdapAttributeGroupUniqueIdentifier = mergedString(plan.LdapAttributeGroupUniqueIdentifier, current.LdapAttributeGroupUniqueIdentifier)
	cfg.LdapAttributeGroupName = mergedString(plan.LdapAttributeGroupName, current.LdapAttributeGroupName)
	cfg.LdapAdminGroupName = mergedString(plan.LdapAdminGroupName, current.LdapAdminGroupName)
	cfg.LdapSoftDeleteUsers = mergedString(plan.LdapSoftDeleteUsers, current.LdapSoftDeleteUsers)

	return &cfg
}

func optionalComputedString(description string, sensitive bool) schema.StringAttribute {
	return schema.StringAttribute{
		Description: description,
		Optional:    true,
		Computed:    true,
		Sensitive:   sensitive,
	}
}

func applicationImageAttribute(description string) schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: description,
		Optional:    true,
		Attributes: map[string]schema.Attribute{
			"source": schema.StringAttribute{
				Description: "Path to the image file.",
				Required:    true,
			},
			"sha256": schema.StringAttribute{
				Description:   "SHA-256 of the image file, computed at plan time.",
				Computed:      true,
				PlanModifiers: []planmodifier.String{sourceFileSHA256PlanModifier{}},
			},
		},
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

			"logo":                    applicationImageAttribute("The application logo."),
			"dark_logo":               applicationImageAttribute("The application logo in dark mode."),
			"email_logo":              applicationImageAttribute("The logo used in emails."),
			"background_image":        applicationImageAttribute("The sign-in page background image."),
			"favicon":                 applicationImageAttribute("The favicon."),
			"default_profile_picture": applicationImageAttribute("The profile picture of users without their own."),
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

	// Upload the configured images
	none := types.ObjectNull(sourceFileAttrTypes)
	r.applyApplicationImages(ctx, &plan, applicationConfigModel{
		Logo: none, DarkLogo: none, EmailLogo: none, BackgroundImage: none, Favicon: none, DefaultProfilePicture: none,
	}, &resp.Diagnostics)
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

	r.refreshApplicationImages(&state, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *applicationConfigResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan applicationConfigModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state applicationConfigModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.applyConfig(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Upload, replace or reset images
	r.applyApplicationImages(ctx, &plan, state, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// applyApplicationImages makes the server's images match the plan and writes
// the values to store in state back into the plan model.
func (r *applicationConfigResource) applyApplicationImages(ctx context.Context, plan *applicationConfigModel, prior applicationConfigModel, diags *diag.Diagnostics) {
	var d diag.Diagnostics
	plan.Logo, d = r.applyApplicationImage(ctx, client.ApplicationImageLogo, "logo", true, plan.Logo, prior.Logo)
	diags.Append(d...)
	plan.DarkLogo, d = r.applyApplicationImage(ctx, client.ApplicationImageDarkLogo, "dark_logo", true, plan.DarkLogo, prior.DarkLogo)
	diags.Append(d...)
	plan.EmailLogo, d = r.applyApplicationImage(ctx, client.ApplicationImageEmailLogo, "email_logo", false, plan.EmailLogo, prior.EmailLogo)
	diags.Append(d...)
	plan.BackgroundImage, d = r.applyApplicationImage(ctx, client.ApplicationImageBackground, "background_image", true, plan.BackgroundImage, prior.BackgroundImage)
	diags.Append(d...)
	plan.Favicon, d = r.applyApplicationImage(ctx, client.ApplicationImageFavicon, "favicon", false, plan.Favicon, prior.Favicon)
	diags.Append(d...)
	plan.DefaultProfilePicture, d = r.applyApplicationImage(ctx, client.ApplicationImageDefaultProfilePicture, "default_profile_picture", true, plan.DefaultProfilePicture, prior.DefaultProfilePicture)
	diags.Append(d...)
}

// refreshApplicationImages drops managed images that are gone from the server
// (deleted in the UI) from state, so the next plan uploads them again.
func (r *applicationConfigResource) refreshApplicationImages(state *applicationConfigModel, diags *diag.Diagnostics) {
	state.Logo = r.refreshApplicationImage(client.ApplicationImageLogo, state.Logo, diags)
	state.DarkLogo = r.refreshApplicationImage(client.ApplicationImageDarkLogo, state.DarkLogo, diags)
	state.EmailLogo = r.refreshApplicationImage(client.ApplicationImageEmailLogo, state.EmailLogo, diags)
	state.BackgroundImage = r.refreshApplicationImage(client.ApplicationImageBackground, state.BackgroundImage, diags)
	state.Favicon = r.refreshApplicationImage(client.ApplicationImageFavicon, state.Favicon, diags)
	state.DefaultProfilePicture = r.refreshApplicationImage(client.ApplicationImageDefaultProfilePicture, state.DefaultProfilePicture, diags)
}

// refreshApplicationImage returns the managed image value to keep in state:
// null when Pocket ID no longer has a custom version of it.
func (r *applicationConfigResource) refreshApplicationImage(image client.ApplicationImage, managed types.Object, diags *diag.Diagnostics) types.Object {
	if managed.IsNull() {
		return managed
	}
	set, err := r.client.ApplicationImageSet(image)
	if err != nil {
		diags.AddError("Error reading application image", "Could not read application image: "+err.Error())
		return managed
	}
	if !set {
		return types.ObjectNull(sourceFileAttrTypes)
	}
	return managed
}

// applyApplicationImage uploads, resets or stops managing one image, and
// returns the value to store in state. Pocket ID can only reset an image that
// is deletable; the others stay on the server when removed from the
// configuration.
func (r *applicationConfigResource) applyApplicationImage(ctx context.Context, image client.ApplicationImage, attribute string, deletable bool, planned, prior types.Object) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics
	if planned.Equal(prior) {
		return planned, diags
	}

	if planned.IsNull() {
		if !deletable {
			diags.AddAttributeWarning(path.Root(attribute), "Image not removed from Pocket ID",
				fmt.Sprintf("Pocket ID cannot remove the %s. It is no longer managed by Terraform and stays until replaced.", attribute))
			return planned, diags
		}
		if err := r.client.DeleteApplicationImage(image); err != nil {
			diags.AddAttributeError(path.Root(attribute), "Error resetting application image", err.Error())
		}
		return planned, diags
	}

	var file sourceFileModel
	diags.Append(planned.As(ctx, &file, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return planned, diags
	}

	sourcePath := path.Root(attribute).AtName("source")
	data, sum, err := readSourceFile(file.Source.ValueString())
	if err != nil {
		diags.AddAttributeError(sourcePath, "Error setting application image", err.Error())
		return planned, diags
	}

	filename := filepath.Base(file.Source.ValueString())
	if err := r.client.UpdateApplicationImage(image, filename, data); err != nil {
		diags.AddAttributeError(sourcePath, "Error setting application image",
			fmt.Sprintf("Pocket ID rejected %q: %s", filename, err))
		return planned, diags
	}

	file.SHA256 = types.StringValue(sum)
	obj, d := types.ObjectValueFrom(ctx, sourceFileAttrTypes, file)
	diags.Append(d...)
	return obj, diags
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
