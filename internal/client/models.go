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
	ID                     string                `json:"id,omitempty"`
	Name                   string                `json:"name"`
	HasLogo                bool                  `json:"hasLogo,omitempty"`
	CallbackURLs           []string              `json:"callbackURLs"`
	LogoutCallbackURLs     []string              `json:"logoutCallbackURLs,omitempty"`
	IsPublic               bool                  `json:"isPublic"`
	PkceEnabled            bool                  `json:"pkceEnabled"`
	Credentials            OIDCClientCredentials `json:"credentials"`
	AllowedUserGroups      []UserGroup           `json:"allowedUserGroups,omitempty"`
	AllowedUserGroupsCount int64                 `json:"allowedUserGroupsCount,omitempty"`
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
	Name               string                `json:"name"`
	CallbackURLs       []string              `json:"callbackURLs"`
	LogoutCallbackURLs []string              `json:"logoutCallbackURLs,omitempty"`
	IsPublic           bool                  `json:"isPublic"`
	PkceEnabled        bool                  `json:"pkceEnabled"`
	Credentials        OIDCClientCredentials `json:"credentials"`
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
	ID           string        `json:"id,omitempty"`
	Username     string        `json:"username"`
	Email        string        `json:"email"`
	FirstName    string        `json:"firstName,omitempty"`
	LastName     string        `json:"lastName,omitempty"`
	IsAdmin      bool          `json:"isAdmin"`
	Locale       *string       `json:"locale,omitempty"`
	Disabled     bool          `json:"disabled"`
	UserGroups   []UserGroup   `json:"userGroups,omitempty"`
	CustomClaims []CustomClaim `json:"customClaims,omitempty"`
	LdapID       *string       `json:"ldapId,omitempty"`
	CreatedAt    string        `json:"createdAt,omitempty"`
	UpdatedAt    string        `json:"updatedAt,omitempty"`
}

// UserCreateRequest represents a request to create or update a user
type UserCreateRequest struct {
	Username  string  `json:"username"`
	Email     string  `json:"email"`
	FirstName string  `json:"firstName,omitempty"`
	LastName  string  `json:"lastName,omitempty"`
	IsAdmin   bool    `json:"isAdmin"`
	Locale    *string `json:"locale,omitempty"`
	Disabled  bool    `json:"disabled"`
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
	UpdatedAt    string        `json:"updatedAt,omitempty"`
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

// ApplicationConfiguration represents the complete application configuration
type ApplicationConfiguration struct {
	LDAP *LDAPConfiguration `json:"ldap,omitempty"`
	// Other configuration sections can be added here as needed
}

// LDAPConfiguration represents LDAP configuration settings
type LDAPConfiguration struct {
	Enabled              string                    `json:"enabled"`                        // "true" or "false"
	URL                  string                    `json:"url,omitempty"`                  // LDAP server URL
	BindDN               string                    `json:"bindDN,omitempty"`               // Bind DN for authentication
	BindPassword         string                    `json:"bindPassword,omitempty"`         // Bind password (sensitive)
	BaseDN               string                    `json:"baseDN,omitempty"`               // Base DN for searches
	SkipCertVerify       string                    `json:"skipCertVerify,omitempty"`       // "true" or "false"
	UserSearchFilter     string                    `json:"userSearchFilter,omitempty"`     // LDAP filter for users
	UserGroupSearchFilter string                   `json:"userGroupSearchFilter,omitempty"` // LDAP filter for groups
	UserAttributes       *LDAPUserAttributes       `json:"userAttributes,omitempty"`       // User attribute mappings
	GroupAttributes      *LDAPGroupAttributes      `json:"groupAttributes,omitempty"`      // Group attribute mappings
	SoftDeleteUsers      string                    `json:"softDeleteUsers,omitempty"`      // "true" or "false"
}

// LDAPUserAttributes represents LDAP user attribute mappings
type LDAPUserAttributes struct {
	UniqueIdentifier string `json:"uniqueIdentifier,omitempty"` // LDAP attribute for unique ID
	Username         string `json:"username,omitempty"`         // LDAP attribute for username
	Email            string `json:"email,omitempty"`            // LDAP attribute for email
	FirstName        string `json:"firstName,omitempty"`        // LDAP attribute for first name
	LastName         string `json:"lastName,omitempty"`         // LDAP attribute for last name
	ProfilePicture   string `json:"profilePicture,omitempty"`   // LDAP attribute for profile picture
}

// LDAPGroupAttributes represents LDAP group attribute mappings
type LDAPGroupAttributes struct {
	Member            string `json:"member,omitempty"`            // LDAP attribute for group members
	UniqueIdentifier  string `json:"uniqueIdentifier,omitempty"`  // LDAP attribute for unique group ID
	Name              string `json:"name,omitempty"`              // LDAP attribute for group name
	AdminGroupName    string `json:"adminGroupName,omitempty"`    // Name of admin group
}

// ApplicationConfigurationUpdateRequest represents a request to update application configuration
type ApplicationConfigurationUpdateRequest struct {
	LDAP *LDAPConfiguration `json:"ldap,omitempty"`
}

// LDAPSyncResponse represents the response from triggering an LDAP sync
type LDAPSyncResponse struct {
	Status    string `json:"status"`              // "success", "failed", "in_progress"
	Message   string `json:"message,omitempty"`   // Success or error message
	Timestamp string `json:"timestamp,omitempty"` // When sync was triggered
}
