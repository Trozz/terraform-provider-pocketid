package client

import (
	"encoding/json"
	"fmt"
)

// GetLDAPConfig retrieves the current LDAP configuration
func (c *Client) GetLDAPConfig() (*LDAPConfig, error) {
	body, err := c.doRequest("GET", "/api/application-configuration/all", nil)
	if err != nil {
		return nil, err
	}

	// Parse the array of config variables
	var configVars []AppConfigVariable
	if err := json.Unmarshal(body, &configVars); err != nil {
		return nil, fmt.Errorf("error unmarshaling config response: %w", err)
	}

	// Map config variables to LDAPConfig struct
	config := &LDAPConfig{}
	for _, v := range configVars {
		switch v.Key {
		case "ldapEnabled":
			if b, ok := v.Value.(bool); ok {
				config.Enabled = b
			}
		case "ldapUrl":
			if s, ok := v.Value.(string); ok {
				config.URL = s
			}
		case "ldapBindDn":
			if s, ok := v.Value.(string); ok {
				config.BindDN = s
			}
		case "ldapBindPassword":
			if s, ok := v.Value.(string); ok {
				config.BindPassword = s
			}
		case "ldapBase":
			if s, ok := v.Value.(string); ok {
				config.BaseDN = s
			}
		case "ldapSkipCertVerify":
			if b, ok := v.Value.(bool); ok {
				config.SkipCertVerify = b
			}
		case "ldapUserSearchFilter":
			if s, ok := v.Value.(string); ok {
				config.UserSearchFilter = s
			}
		case "ldapGroupSearchFilter":
			if s, ok := v.Value.(string); ok {
				config.UserGroupSearchFilter = s
			}
		case "ldapAttributeUserUniqueIdentifier":
			if s, ok := v.Value.(string); ok {
				config.UserUniqueAttribute = s
			}
		case "ldapAttributeUserUsername":
			if s, ok := v.Value.(string); ok {
				config.UserUsernameAttribute = s
			}
		case "ldapAttributeUserEmail":
			if s, ok := v.Value.(string); ok {
				config.UserEmailAttribute = s
			}
		case "ldapAttributeUserFirstName":
			if s, ok := v.Value.(string); ok {
				config.UserFirstNameAttribute = s
			}
		case "ldapAttributeUserLastName":
			if s, ok := v.Value.(string); ok {
				config.UserLastNameAttribute = s
			}
		case "ldapAttributeGroupMember":
			if s, ok := v.Value.(string); ok {
				config.GroupMemberAttribute = s
			}
		case "ldapAttributeGroupUniqueIdentifier":
			if s, ok := v.Value.(string); ok {
				config.GroupUniqueAttribute = s
			}
		case "ldapAttributeGroupName":
			if s, ok := v.Value.(string); ok {
				config.GroupNameAttribute = s
			}
		case "ldapAttributeAdminGroup":
			if s, ok := v.Value.(string); ok {
				config.AdminGroupName = s
			}
		case "ldapSoftDeleteUsers":
			if b, ok := v.Value.(bool); ok {
				config.SoftDeleteUsers = b
			}
		}
	}

	return config, nil
}

// UpdateLDAPConfig updates the LDAP configuration
func (c *Client) UpdateLDAPConfig(req *LDAPConfigUpdateRequest) error {
	_, err := c.doRequest("PUT", "/api/application-configuration", req)
	return err
}

// SyncLDAP triggers an LDAP synchronization
func (c *Client) SyncLDAP() error {
	_, err := c.doRequest("POST", "/api/application-configuration/sync-ldap", nil)
	return err
}
