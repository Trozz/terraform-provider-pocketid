package client_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Trozz/terraform-provider-pocketid/internal/client"
)

func TestClient_GetUser(t *testing.T) {
	expectedUser := &client.User{
		ID:        "test-user-id",
		Username:  "testuser",
		Email:     "test@example.com",
		FirstName: "Test",
		LastName:  "User",
		IsAdmin:   false,
		Disabled:  false,
		Locale:    stringPtr("en"),
		UserGroups: []client.UserGroup{
			{ID: "group1", Name: "Group 1"},
			{ID: "group2", Name: "Group 2"},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/users/test-user-id", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(expectedUser); err != nil {
			t.Fatalf("Failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	c, err := client.NewClient(server.URL, "test-token", false, 30)
	require.NoError(t, err)

	result, err := c.GetUser("test-user-id")
	assert.NoError(t, err)
	assert.Equal(t, expectedUser, result)
}

func TestClient_GetUser_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		if _, err := fmt.Fprint(w, `{"error": "User not found"}`); err != nil {
			t.Fatalf("Failed to write response: %v", err)
		}
	}))
	defer server.Close()

	c, err := client.NewClient(server.URL, "test-token", false, 30)
	require.NoError(t, err)

	result, err := c.GetUser("nonexistent-id")
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "HTTP 404")
}

func TestClient_UpdateUser(t *testing.T) {
	updateReq := &client.UserCreateRequest{
		Username:  "updateduser",
		Email:     "updated@example.com",
		FirstName: "Updated",
		LastName:  "User",
		IsAdmin:   true,
		Disabled:  false,
		Locale:    stringPtr("fr"),
	}

	expectedUser := &client.User{
		ID:        "test-user-id",
		Username:  updateReq.Username,
		Email:     updateReq.Email,
		FirstName: updateReq.FirstName,
		LastName:  updateReq.LastName,
		IsAdmin:   updateReq.IsAdmin,
		Disabled:  updateReq.Disabled,
		Locale:    updateReq.Locale,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		assert.Equal(t, "/api/users/test-user-id", r.URL.Path)

		var req client.UserCreateRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)
		assert.Equal(t, updateReq, &req)

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(expectedUser); err != nil {
			t.Fatalf("Failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	c, err := client.NewClient(server.URL, "test-token", false, 30)
	require.NoError(t, err)

	result, err := c.UpdateUser("test-user-id", updateReq)
	assert.NoError(t, err)
	assert.Equal(t, expectedUser, result)
}

func TestClient_DeleteUser(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)
		assert.Equal(t, "/api/users/test-user-id", r.URL.Path)

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c, err := client.NewClient(server.URL, "test-token", false, 30)
	require.NoError(t, err)

	err = c.DeleteUser("test-user-id")
	assert.NoError(t, err)
}

func TestClient_DeleteUser_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		if _, err := fmt.Fprint(w, `{"error": "Insufficient permissions"}`); err != nil {
			t.Fatalf("Failed to write response: %v", err)
		}
	}))
	defer server.Close()

	c, err := client.NewClient(server.URL, "test-token", false, 30)
	require.NoError(t, err)

	err = c.DeleteUser("test-user-id")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "HTTP 403")
}

func TestClient_ListUsers(t *testing.T) {
	expectedUsers := []client.User{
		{
			ID:        "user1",
			Username:  "testuser1",
			Email:     "test1@example.com",
			FirstName: "Test",
			LastName:  "User1",
			IsAdmin:   false,
			Disabled:  false,
		},
		{
			ID:        "user2",
			Username:  "testuser2",
			Email:     "test2@example.com",
			FirstName: "Test",
			LastName:  "User2",
			IsAdmin:   true,
			Disabled:  false,
			Locale:    stringPtr("en"),
		},
	}

	expectedResponse := &client.PaginatedResponse[client.User]{
		Data: expectedUsers,
		Pagination: client.PaginationInfo{
			TotalItems:   2,
			CurrentPage:  1,
			ItemsPerPage: 10,
			TotalPages:   1,
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/users", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(expectedResponse); err != nil {
			t.Fatalf("Failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	c, err := client.NewClient(server.URL, "test-token", false, 30)
	require.NoError(t, err)

	result, err := c.ListUsers()
	assert.NoError(t, err)
	assert.Equal(t, expectedResponse, result)
	assert.Equal(t, expectedUsers, result.Data)
}

func TestClient_ListUsers_Empty(t *testing.T) {
	expectedResponse := &client.PaginatedResponse[client.User]{
		Data: []client.User{},
		Pagination: client.PaginationInfo{
			TotalItems:   0,
			CurrentPage:  1,
			ItemsPerPage: 10,
			TotalPages:   0,
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/users", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(expectedResponse); err != nil {
			t.Fatalf("Failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	c, err := client.NewClient(server.URL, "test-token", false, 30)
	require.NoError(t, err)

	result, err := c.ListUsers()
	assert.NoError(t, err)
	assert.Equal(t, expectedResponse, result)
	assert.Empty(t, result.Data)
}

func TestClient_GetUserGroup(t *testing.T) {
	expectedGroup := &client.UserGroup{
		ID:           "test-group-id",
		Name:         "test-group",
		FriendlyName: "Test Group",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/user-groups/test-group-id", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(expectedGroup); err != nil {
			t.Fatalf("Failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	c, err := client.NewClient(server.URL, "test-token", false, 30)
	require.NoError(t, err)

	result, err := c.GetUserGroup("test-group-id")
	assert.NoError(t, err)
	assert.Equal(t, expectedGroup, result)
}

func TestClient_GetUserGroup_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		if _, err := fmt.Fprint(w, `{"error": "Group not found"}`); err != nil {
			t.Fatalf("Failed to write response: %v", err)
		}
	}))
	defer server.Close()

	c, err := client.NewClient(server.URL, "test-token", false, 30)
	require.NoError(t, err)

	result, err := c.GetUserGroup("nonexistent-id")
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "HTTP 404")
}

func TestClient_UpdateUserGroup(t *testing.T) {
	updateReq := &client.UserGroupCreateRequest{
		Name:         "updated-group",
		FriendlyName: "Updated Group",
	}

	expectedGroup := &client.UserGroup{
		ID:           "test-group-id",
		Name:         updateReq.Name,
		FriendlyName: updateReq.FriendlyName,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		assert.Equal(t, "/api/user-groups/test-group-id", r.URL.Path)

		var req client.UserGroupCreateRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)
		assert.Equal(t, updateReq, &req)

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(expectedGroup); err != nil {
			t.Fatalf("Failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	c, err := client.NewClient(server.URL, "test-token", false, 30)
	require.NoError(t, err)

	result, err := c.UpdateUserGroup("test-group-id", updateReq)
	assert.NoError(t, err)
	assert.Equal(t, expectedGroup, result)
}

func TestClient_DeleteUserGroup(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)
		assert.Equal(t, "/api/user-groups/test-group-id", r.URL.Path)

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c, err := client.NewClient(server.URL, "test-token", false, 30)
	require.NoError(t, err)

	err = c.DeleteUserGroup("test-group-id")
	assert.NoError(t, err)
}

func TestClient_DeleteUserGroup_InUse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusConflict)
		if _, err := fmt.Fprint(w, `{"error": "Group is in use"}`); err != nil {
			t.Fatalf("Failed to write response: %v", err)
		}
	}))
	defer server.Close()

	c, err := client.NewClient(server.URL, "test-token", false, 30)
	require.NoError(t, err)

	err = c.DeleteUserGroup("test-group-id")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "HTTP 409")
}

func TestClient_ListUserGroups(t *testing.T) {
	expectedGroups := []client.UserGroup{
		{
			ID:           "group1",
			Name:         "admins",
			FriendlyName: "Administrators",
		},
		{
			ID:           "group2",
			Name:         "users",
			FriendlyName: "Regular Users",
		},
		{
			ID:           "group3",
			Name:         "developers",
			FriendlyName: "Developers",
		},
	}

	expectedResponse := &client.PaginatedResponse[client.UserGroup]{
		Data: expectedGroups,
		Pagination: client.PaginationInfo{
			TotalItems:   3,
			CurrentPage:  1,
			ItemsPerPage: 10,
			TotalPages:   1,
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/user-groups", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(expectedResponse); err != nil {
			t.Fatalf("Failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	c, err := client.NewClient(server.URL, "test-token", false, 30)
	require.NoError(t, err)

	result, err := c.ListUserGroups()
	assert.NoError(t, err)
	assert.Equal(t, expectedResponse, result)
	assert.Equal(t, expectedGroups, result.Data)
}

func TestClient_ListUserGroups_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		if _, err := fmt.Fprint(w, `{"error": "Internal server error"}`); err != nil {
			t.Fatalf("Failed to write response: %v", err)
		}
	}))
	defer server.Close()

	c, err := client.NewClient(server.URL, "test-token", false, 30)
	require.NoError(t, err)

	result, err := c.ListUserGroups()
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "HTTP 500")
}

// Helper function to create a string pointer
func stringPtr(s string) *string {
	return &s
}
