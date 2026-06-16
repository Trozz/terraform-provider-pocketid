package datasources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/Trozz/terraform-provider-pocketid/internal/client"
)

const applicationConfigID = "application-configuration"

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &applicationConfigDataSource{}
	_ datasource.DataSourceWithConfigure = &applicationConfigDataSource{}
)

// NewApplicationConfigDataSource creates a new application configuration data source.
func NewApplicationConfigDataSource() datasource.DataSource {
	return &applicationConfigDataSource{}
}

// applicationConfigDataSource is the data source implementation.
type applicationConfigDataSource struct {
	client *client.Client
}

// applicationConfigDataSourceModel describes the data source data model.
type applicationConfigDataSourceModel struct {
	ID types.String `tfsdk:"id"`

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

func computedString(description string, sensitive bool) schema.StringAttribute {
	return schema.StringAttribute{
		Description: description,
		Computed:    true,
		Sensitive:   sensitive,
	}
}

// Metadata returns the data source type name.
func (d *applicationConfigDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_application_config"
}

// Schema defines the schema for the data source.
func (d *applicationConfigDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves the global application configuration of a Pocket-ID instance.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Fixed identifier of the application configuration singleton.",
				Computed:    true,
			},

			"app_name":                      computedString("The name of the application.", false),
			"session_duration":              computedString("Session duration in minutes.", false),
			"home_page_url":                 computedString("URL of the application home page.", false),
			"emails_verified":               computedString("Whether user emails are considered verified.", false),
			"disable_animations":            computedString("Whether UI animations are disabled.", false),
			"allow_own_account_edit":        computedString("Whether users can edit their own account.", false),
			"allow_user_signups":            computedString("User signup mode.", false),
			"signup_default_user_group_ids": computedString("JSON array of user group IDs assigned to users created via signup.", false),
			"signup_default_custom_claims":  computedString("JSON object of custom claims assigned to users created via signup.", false),
			"accent_color":                  computedString("Accent color used in the UI.", false),
			"require_user_email":            computedString("Whether a user email is required.", false),

			"smtp_host":             computedString("SMTP server host.", false),
			"smtp_port":             computedString("SMTP server port.", false),
			"smtp_from":             computedString("Email address used as the sender.", false),
			"smtp_user":             computedString("SMTP authentication user.", false),
			"smtp_password":         computedString("SMTP authentication password.", true),
			"smtp_tls":              computedString("SMTP TLS mode.", false),
			"smtp_skip_cert_verify": computedString("Whether SMTP certificate verification is skipped.", false),

			"email_one_time_access_as_admin_enabled":           computedString("Whether admins can use one-time access email links.", false),
			"email_one_time_access_as_unauthenticated_enabled": computedString("Whether unauthenticated users can request one-time access email links.", false),
			"email_login_notification_enabled":                 computedString("Whether login notification emails are enabled.", false),
			"email_api_key_expiration_enabled":                 computedString("Whether API key expiration emails are enabled.", false),
			"email_verification_enabled":                       computedString("Whether email verification is enabled.", false),

			"ldap_enabled":                           computedString("Whether LDAP integration is enabled.", false),
			"ldap_url":                               computedString("LDAP server URL.", false),
			"ldap_bind_dn":                           computedString("LDAP bind DN.", false),
			"ldap_bind_password":                     computedString("LDAP bind password.", true),
			"ldap_base":                              computedString("LDAP search base.", false),
			"ldap_user_search_filter":                computedString("LDAP user search filter.", false),
			"ldap_user_group_search_filter":          computedString("LDAP user group search filter.", false),
			"ldap_skip_cert_verify":                  computedString("Whether LDAP certificate verification is skipped.", false),
			"ldap_attribute_user_unique_identifier":  computedString("LDAP attribute for the user unique identifier.", false),
			"ldap_attribute_user_username":           computedString("LDAP attribute for the username.", false),
			"ldap_attribute_user_email":              computedString("LDAP attribute for the user email.", false),
			"ldap_attribute_user_first_name":         computedString("LDAP attribute for the user first name.", false),
			"ldap_attribute_user_last_name":          computedString("LDAP attribute for the user last name.", false),
			"ldap_attribute_user_display_name":       computedString("LDAP attribute for the user display name.", false),
			"ldap_attribute_user_profile_picture":    computedString("LDAP attribute for the user profile picture.", false),
			"ldap_attribute_group_member":            computedString("LDAP attribute for group membership.", false),
			"ldap_attribute_group_unique_identifier": computedString("LDAP attribute for the group unique identifier.", false),
			"ldap_attribute_group_name":              computedString("LDAP attribute for the group name.", false),
			"ldap_admin_group_name":                  computedString("LDAP group name granting admin privileges.", false),
			"ldap_soft_delete_users":                 computedString("Whether users removed from LDAP are soft-deleted.", false),
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *applicationConfigDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = c
}

// Read refreshes the Terraform state with the latest data.
func (d *applicationConfigDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	tflog.Debug(ctx, "Reading application configuration")

	cfg, err := d.client.GetApplicationConfig()
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Application Configuration",
			err.Error(),
		)
		return
	}

	data := applicationConfigDataSourceModel{
		ID: types.StringValue(applicationConfigID),

		AppName:                   types.StringValue(cfg.AppName),
		SessionDuration:           types.StringValue(cfg.SessionDuration),
		HomePageURL:               types.StringValue(cfg.HomePageURL),
		EmailsVerified:            types.StringValue(cfg.EmailsVerified),
		DisableAnimations:         types.StringValue(cfg.DisableAnimations),
		AllowOwnAccountEdit:       types.StringValue(cfg.AllowOwnAccountEdit),
		AllowUserSignups:          types.StringValue(cfg.AllowUserSignups),
		SignupDefaultUserGroupIDs: types.StringValue(cfg.SignupDefaultUserGroupIDs),
		SignupDefaultCustomClaims: types.StringValue(cfg.SignupDefaultCustomClaims),
		AccentColor:               types.StringValue(cfg.AccentColor),
		RequireUserEmail:          types.StringValue(cfg.RequireUserEmail),

		SmtpHost:           types.StringValue(cfg.SmtpHost),
		SmtpPort:           types.StringValue(cfg.SmtpPort),
		SmtpFrom:           types.StringValue(cfg.SmtpFrom),
		SmtpUser:           types.StringValue(cfg.SmtpUser),
		SmtpPassword:       types.StringValue(cfg.SmtpPassword),
		SmtpTls:            types.StringValue(cfg.SmtpTls),
		SmtpSkipCertVerify: types.StringValue(cfg.SmtpSkipCertVerify),

		EmailOneTimeAccessAsAdminEnabled:           types.StringValue(cfg.EmailOneTimeAccessAsAdminEnabled),
		EmailOneTimeAccessAsUnauthenticatedEnabled: types.StringValue(cfg.EmailOneTimeAccessAsUnauthenticatedEnabled),
		EmailLoginNotificationEnabled:              types.StringValue(cfg.EmailLoginNotificationEnabled),
		EmailApiKeyExpirationEnabled:               types.StringValue(cfg.EmailApiKeyExpirationEnabled),
		EmailVerificationEnabled:                   types.StringValue(cfg.EmailVerificationEnabled),

		LdapEnabled:                        types.StringValue(cfg.LdapEnabled),
		LdapUrl:                            types.StringValue(cfg.LdapUrl),
		LdapBindDn:                         types.StringValue(cfg.LdapBindDn),
		LdapBindPassword:                   types.StringValue(cfg.LdapBindPassword),
		LdapBase:                           types.StringValue(cfg.LdapBase),
		LdapUserSearchFilter:               types.StringValue(cfg.LdapUserSearchFilter),
		LdapUserGroupSearchFilter:          types.StringValue(cfg.LdapUserGroupSearchFilter),
		LdapSkipCertVerify:                 types.StringValue(cfg.LdapSkipCertVerify),
		LdapAttributeUserUniqueIdentifier:  types.StringValue(cfg.LdapAttributeUserUniqueIdentifier),
		LdapAttributeUserUsername:          types.StringValue(cfg.LdapAttributeUserUsername),
		LdapAttributeUserEmail:             types.StringValue(cfg.LdapAttributeUserEmail),
		LdapAttributeUserFirstName:         types.StringValue(cfg.LdapAttributeUserFirstName),
		LdapAttributeUserLastName:          types.StringValue(cfg.LdapAttributeUserLastName),
		LdapAttributeUserDisplayName:       types.StringValue(cfg.LdapAttributeUserDisplayName),
		LdapAttributeUserProfilePicture:    types.StringValue(cfg.LdapAttributeUserProfilePicture),
		LdapAttributeGroupMember:           types.StringValue(cfg.LdapAttributeGroupMember),
		LdapAttributeGroupUniqueIdentifier: types.StringValue(cfg.LdapAttributeGroupUniqueIdentifier),
		LdapAttributeGroupName:             types.StringValue(cfg.LdapAttributeGroupName),
		LdapAdminGroupName:                 types.StringValue(cfg.LdapAdminGroupName),
		LdapSoftDeleteUsers:                types.StringValue(cfg.LdapSoftDeleteUsers),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
