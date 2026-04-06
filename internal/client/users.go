// internal/client/users.go
package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/Korrnals/gotr/internal/models/data"
)

// GetUsers fetches the list of all users.
// https://support.testrail.com/hc/en-us/articles/7077807509812-Users#getusers
func (c *HTTPClient) GetUsers(ctx context.Context) (data.GetUsersResponse, error) {
	resp, err := c.Get(ctx, "get_users", nil)
	if err != nil {
		return nil, fmt.Errorf("error getting users: %w", err)
	}
	defer resp.Body.Close()

	var users data.GetUsersResponse
	if err := json.NewDecoder(resp.Body).Decode(&users); err != nil {
		return nil, fmt.Errorf("error decoding users: %w", err)
	}
	return users, nil
}

// GetUsersByProject fetches the user list for a project.
// https://support.testrail.com/hc/en-us/articles/7077807509812-Users#getusers
func (c *HTTPClient) GetUsersByProject(ctx context.Context, projectID int64) (data.GetUsersResponse, error) {
	endpoint := fmt.Sprintf("get_users/%d", projectID)
	resp, err := c.Get(ctx, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting users for project %d: %w", projectID, err)
	}
	defer resp.Body.Close()

	var users data.GetUsersResponse
	if err := json.NewDecoder(resp.Body).Decode(&users); err != nil {
		return nil, fmt.Errorf("error decoding users: %w", err)
	}
	return users, nil
}

// GetUser fetches a user by ID.
// https://support.testrail.com/hc/en-us/articles/7077807509812-Users#getuser
func (c *HTTPClient) GetUser(ctx context.Context, userID int64) (*data.User, error) {
	endpoint := fmt.Sprintf("get_user/%d", userID)
	resp, err := c.Get(ctx, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting user %d: %w", userID, err)
	}
	defer resp.Body.Close()

	var user data.User
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, fmt.Errorf("error decoding user: %w", err)
	}
	return &user, nil
}

// GetUserByEmail fetches a user by email.
// https://support.testrail.com/hc/en-us/articles/7077807509812-Users#getuserbyemail
func (c *HTTPClient) GetUserByEmail(ctx context.Context, email string) (*data.User, error) {
	resp, err := c.Get(ctx, "get_user_by_email", map[string]string{"email": email})
	if err != nil {
		return nil, fmt.Errorf("error getting user by email %s: %w", email, err)
	}
	defer resp.Body.Close()

	var user data.User
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, fmt.Errorf("error decoding user: %w", err)
	}
	return &user, nil
}

// AddUser creates a new user.
// https://support.testrail.com/hc/en-us/articles/7077807509812-Users#adduser
func (c *HTTPClient) AddUser(ctx context.Context, req data.AddUserRequest) (*data.User, error) {
	bodyBytes, _ := json.Marshal(req)
	resp, err := c.Post(ctx, "add_user", bytes.NewReader(bodyBytes), nil)
	if err != nil {
		return nil, fmt.Errorf("error adding user: %w", err)
	}
	defer resp.Body.Close()

	var user data.User
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, fmt.Errorf("error decoding user: %w", err)
	}
	return &user, nil
}

// UpdateUser updates an existing user.
// https://support.testrail.com/hc/en-us/articles/7077807509812-Users#updateuser
func (c *HTTPClient) UpdateUser(ctx context.Context, userID int64, req data.UpdateUserRequest) (*data.User, error) {
	endpoint := fmt.Sprintf("update_user/%d", userID)
	bodyBytes, _ := json.Marshal(req)
	resp, err := c.Post(ctx, endpoint, bytes.NewReader(bodyBytes), nil)
	if err != nil {
		return nil, fmt.Errorf("error updating user %d: %w", userID, err)
	}
	defer resp.Body.Close()

	var user data.User
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, fmt.Errorf("error decoding user: %w", err)
	}
	return &user, nil
}

// GetPriorities fetches the list of priorities.
// https://support.testrail.com/hc/en-us/articles/7077701636116-Priorities#getpriorities
func (c *HTTPClient) GetPriorities(ctx context.Context) (data.GetPrioritiesResponse, error) {
	resp, err := c.Get(ctx, "get_priorities", nil)
	if err != nil {
		return nil, fmt.Errorf("error getting priorities: %w", err)
	}
	defer resp.Body.Close()

	var priorities data.GetPrioritiesResponse
	if err := json.NewDecoder(resp.Body).Decode(&priorities); err != nil {
		return nil, fmt.Errorf("error decoding priorities: %w", err)
	}
	return priorities, nil
}

// GetStatuses fetches the list of statuses.
// https://support.testrail.com/hc/en-us/articles/7077812750372-Statuses#getstatuses
func (c *HTTPClient) GetStatuses(ctx context.Context) (data.GetStatusesResponse, error) {
	resp, err := c.Get(ctx, "get_statuses", nil)
	if err != nil {
		return nil, fmt.Errorf("error getting statuses: %w", err)
	}
	defer resp.Body.Close()

	var statuses data.GetStatusesResponse
	if err := json.NewDecoder(resp.Body).Decode(&statuses); err != nil {
		return nil, fmt.Errorf("error decoding statuses: %w", err)
	}
	return statuses, nil
}

// GetTemplates fetches the template list for a project.
// https://support.testrail.com/hc/en-us/articles/7077792420884-Templates#gettemplates
func (c *HTTPClient) GetTemplates(ctx context.Context, projectID int64) (data.GetTemplatesResponse, error) {
	endpoint := fmt.Sprintf("get_templates/%d", projectID)
	resp, err := c.Get(ctx, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting templates for project %d: %w", projectID, err)
	}
	defer resp.Body.Close()

	var templates data.GetTemplatesResponse
	if err := json.NewDecoder(resp.Body).Decode(&templates); err != nil {
		return nil, fmt.Errorf("error decoding templates: %w", err)
	}
	return templates, nil
}
