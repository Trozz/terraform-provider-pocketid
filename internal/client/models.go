package client

// ErrorResponse represents an error response from the Pocket-ID API
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// PaginationInfo represents pagination metadata
type PaginationInfo struct {
	TotalPages   int `json:"totalPages"`
	TotalItems   int `json:"totalItems"`
	CurrentPage  int `json:"currentPage"`
	ItemsPerPage int `json:"itemsPerPage"`
}

// PaginatedResponse represents a paginated API response
type PaginatedResponse[T any] struct {
	Data       []T            `json:"data"`
	Pagination PaginationInfo `json:"pagination"`
}

// OIDCClient represents an OIDC client in Pocket-ID
type OIDCClient struct {
	ID                       string   `json:"id,omitempty"`
	Name                     string   `json:"name"`
	HasLogo                  bool     `json:"hasLogo,omitempty"`
	CallbackURLs             []string `json:"callbackURLs"`
	LogoutCallbackURLs       []string `json:"logoutCallbackURLs,omitempty"`
	IsPublic                 bool     `json:"isPublic"`
	RequiresReauthentication bool     `json:"requiresReauthentication,omitempty"`
	// Pointer so an absent field (Pocket-ID <= v2.8.0, which has no PAR support)
	// is distinguishable from an explicit false.
	RequiresPushedAuthorizationRequests *bool                 `json:"requiresPushedAuthorizationRequests,omitempty"`
	LaunchURL                           string                `json:"launchURL,omitempty"`
	PkceEnabled                         bool                  `json:"pkceEnabled"`
	IsGroupRestricted                   bool                  `json:"isGroupRestricted"`
	Credentials                         OIDCClientCredentials `json:"credentials"`
	AllowedUserGroups                   []UserGroup           `json:"allowedUserGroups,omitempty"`
	AllowedUserGroupsCount              int64                 `json:"allowedUserGroupsCount,omitempty"`
}

// OIDCClientCredentials represents federated identity credentials for an OIDC client
type OIDCClientCredentials struct {
	FederatedIdentities []OIDCClientFederatedIdentity `json:"federatedIdentities,omitempty"`
}

// OIDCClientFederatedIdentity represents a federated identity configuration
type OIDCClientFederatedIdentity struct {
	Issuer   string `json:"issuer"`
	Subject  string `json:"subject,omitempty"`
	Audience string `json:"audience,omitempty"`
	JWKS     string `json:"jwks,omitempty"`
}

// OIDCClientCreateRequest represents a request to create or update an OIDC client
type OIDCClientCreateRequest struct {
	Name                                string                `json:"name"`
	ClientID                            *string               `json:"id,omitempty"`
	CallbackURLs                        []string              `json:"callbackURLs"`
	LogoutCallbackURLs                  []string              `json:"logoutCallbackURLs,omitempty"`
	IsPublic                            bool                  `json:"isPublic"`
	RequiresReauthentication            bool                  `json:"requiresReauthentication,omitempty"`
	RequiresPushedAuthorizationRequests bool                  `json:"requiresPushedAuthorizationRequests"`
	LaunchURL                           *string               `json:"launchURL,omitempty"`
	PkceEnabled                         bool                  `json:"pkceEnabled"`
	IsGroupRestricted                   bool                  `json:"isGroupRestricted"`
	Credentials                         OIDCClientCredentials `json:"credentials"`
}

// ClientSecretResponse represents the response when generating a client secret
type ClientSecretResponse struct {
	Secret string `json:"secret"`
}

// UpdateAllowedUserGroupsRequest represents a request to update allowed user groups for a client
type UpdateAllowedUserGroupsRequest struct {
	UserGroupIDs []string `json:"userGroupIds"`
}

// User represents a user in Pocket-ID
type User struct {
	ID            string        `json:"id,omitempty"`
	Username      string        `json:"username"`
	Email         string        `json:"email"`
	FirstName     string        `json:"firstName,omitempty"`
	LastName      string        `json:"lastName,omitempty"`
	DisplayName   string        `json:"displayName,omitempty"`
	EmailVerified bool          `json:"emailVerified"`
	IsAdmin       bool          `json:"isAdmin"`
	Locale        *string       `json:"locale,omitempty"`
	Disabled      bool          `json:"disabled"`
	UserGroups    []UserGroup   `json:"userGroups,omitempty"`
	CustomClaims  []CustomClaim `json:"customClaims,omitempty"`
	LdapID        *string       `json:"ldapId,omitempty"`
}

// UserCreateRequest represents a request to create or update a user
type UserCreateRequest struct {
	Username      string  `json:"username"`
	Email         string  `json:"email"`
	FirstName     string  `json:"firstName,omitempty"`
	LastName      string  `json:"lastName,omitempty"`
	DisplayName   string  `json:"displayName,omitempty"`
	EmailVerified bool    `json:"emailVerified"`
	IsAdmin       bool    `json:"isAdmin"`
	Locale        *string `json:"locale,omitempty"`
	Disabled      bool    `json:"disabled"`
}

// UpdateUserGroupsRequest represents a request to update a user's groups
type UpdateUserGroupsRequest struct {
	UserGroupIDs []string `json:"userGroupIds"`
}

// UserGroup represents a user group in Pocket-ID
type UserGroup struct {
	ID           string        `json:"id,omitempty"`
	Name         string        `json:"name"`
	FriendlyName string        `json:"friendlyName"`
	Users        []User        `json:"users,omitempty"`
	UserCount    int           `json:"userCount,omitempty"`
	CustomClaims []CustomClaim `json:"customClaims,omitempty"`
	LdapID       *string       `json:"ldapId,omitempty"`
	CreatedAt    string        `json:"createdAt,omitempty"`
}

// UserGroupCreateRequest represents a request to create or update a user group
type UserGroupCreateRequest struct {
	Name         string `json:"name"`
	FriendlyName string `json:"friendlyName"`
}

// CustomClaim represents a custom claim for users or groups
type CustomClaim struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// ApplicationConfig represents the writable application configuration of a
// Pocket-ID instance. Every value is stored as a string by the API. The JSON
// tags match the keys returned by GET /api/application-configuration/all and
// the body expected by PUT /api/application-configuration.
type ApplicationConfig struct {
	// General
	AppName                   string `json:"appName"`
	SessionDuration           string `json:"sessionDuration"`
	HomePageURL               string `json:"homePageUrl"`
	EmailsVerified            string `json:"emailsVerified"`
	DisableAnimations         string `json:"disableAnimations"`
	AllowOwnAccountEdit       string `json:"allowOwnAccountEdit"`
	AllowUserSignups          string `json:"allowUserSignups"`
	SignupDefaultUserGroupIDs string `json:"signupDefaultUserGroupIDs"`
	SignupDefaultCustomClaims string `json:"signupDefaultCustomClaims"`
	AccentColor               string `json:"accentColor"`
	RequireUserEmail          string `json:"requireUserEmail"`

	// Email / SMTP
	SmtpHost           string `json:"smtpHost"`
	SmtpPort           string `json:"smtpPort"`
	SmtpFrom           string `json:"smtpFrom"`
	SmtpUser           string `json:"smtpUser"`
	SmtpPassword       string `json:"smtpPassword"`
	SmtpTls            string `json:"smtpTls"`
	SmtpSkipCertVerify string `json:"smtpSkipCertVerify"`

	EmailOneTimeAccessAsAdminEnabled           string `json:"emailOneTimeAccessAsAdminEnabled"`
	EmailOneTimeAccessAsUnauthenticatedEnabled string `json:"emailOneTimeAccessAsUnauthenticatedEnabled"`
	EmailLoginNotificationEnabled              string `json:"emailLoginNotificationEnabled"`
	EmailApiKeyExpirationEnabled               string `json:"emailApiKeyExpirationEnabled"`
	EmailVerificationEnabled                   string `json:"emailVerificationEnabled"`

	// LDAP
	LdapEnabled                        string `json:"ldapEnabled"`
	LdapUrl                            string `json:"ldapUrl"`
	LdapBindDn                         string `json:"ldapBindDn"`
	LdapBindPassword                   string `json:"ldapBindPassword"`
	LdapBase                           string `json:"ldapBase"`
	LdapUserSearchFilter               string `json:"ldapUserSearchFilter"`
	LdapUserGroupSearchFilter          string `json:"ldapUserGroupSearchFilter"`
	LdapSkipCertVerify                 string `json:"ldapSkipCertVerify"`
	LdapAttributeUserUniqueIdentifier  string `json:"ldapAttributeUserUniqueIdentifier"`
	LdapAttributeUserUsername          string `json:"ldapAttributeUserUsername"`
	LdapAttributeUserEmail             string `json:"ldapAttributeUserEmail"`
	LdapAttributeUserFirstName         string `json:"ldapAttributeUserFirstName"`
	LdapAttributeUserLastName          string `json:"ldapAttributeUserLastName"`
	LdapAttributeUserDisplayName       string `json:"ldapAttributeUserDisplayName"`
	LdapAttributeUserProfilePicture    string `json:"ldapAttributeUserProfilePicture"`
	LdapAttributeGroupMember           string `json:"ldapAttributeGroupMember"`
	LdapAttributeGroupUniqueIdentifier string `json:"ldapAttributeGroupUniqueIdentifier"`
	LdapAttributeGroupName             string `json:"ldapAttributeGroupName"`
	LdapAdminGroupName                 string `json:"ldapAdminGroupName"`
	LdapSoftDeleteUsers                string `json:"ldapSoftDeleteUsers"`
}

// AppConfigVariable represents a single key/value entry as returned by the
// application configuration endpoints.
type AppConfigVariable struct {
	Key   string `json:"key"`
	Type  string `json:"type"`
	Value string `json:"value"`
}

// APIKey represents an API key
type APIKey struct {
	ID                  string `json:"id,omitempty"`
	Name                string `json:"name"`
	Key                 string `json:"key,omitempty"`
	Description         string `json:"description,omitempty"`
	ExpiresAt           string `json:"expiresAt"`
	LastUsedAt          string `json:"lastUsedAt,omitempty"`
	CreatedAt           string `json:"createdAt,omitempty"`
	ExpirationEmailSent bool   `json:"expirationEmailSent"`
}

// ScimServiceProvider represents a SCIM service provider configuration attached
// to an OIDC client in Pocket-ID. The token is stored encrypted server-side but
// is returned (decrypted) on read.
type ScimServiceProvider struct {
	ID           string              `json:"id,omitempty"`
	Endpoint     string              `json:"endpoint"`
	Token        string              `json:"token,omitempty"`
	LastSyncedAt *string             `json:"lastSyncedAt,omitempty"`
	OidcClient   *OIDCClientMetadata `json:"oidcClient,omitempty"`
	CreatedAt    string              `json:"createdAt,omitempty"`
}

// OIDCClientMetadata represents the OIDC client metadata embedded in a SCIM
// service provider response.
type OIDCClientMetadata struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

// ScimServiceProviderCreateRequest represents a request to create or update a
// SCIM service provider configuration for an OIDC client.
type ScimServiceProviderCreateRequest struct {
	Endpoint     string `json:"endpoint"`
	Token        string `json:"token,omitempty"`
	OidcClientID string `json:"oidcClientId"`
}
