// internal/client/configs.go
package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/Korrnals/gotr/internal/models/data"
)

// GetConfigs fetches the configuration list for a project.
// https://support.testrail.com/hc/en-us/articles/7077719410580-Configurations#getconfigs
func (c *HTTPClient) GetConfigs(ctx context.Context, projectID int64) (data.GetConfigsResponse, error) {
	endpoint := fmt.Sprintf("get_configs/%d", projectID)
	resp, err := c.Get(ctx, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting configs for project %d: %w", projectID, err)
	}
	defer resp.Body.Close()

	var configs data.GetConfigsResponse
	if err := json.NewDecoder(resp.Body).Decode(&configs); err != nil {
		return nil, fmt.Errorf("error decoding configs: %w", err)
	}
	return configs, nil
}

// AddConfigGroup creates a new configuration group.
// https://support.testrail.com/hc/en-us/articles/7077719410580-Configurations#addconfiggroup
func (c *HTTPClient) AddConfigGroup(ctx context.Context, projectID int64, req *data.AddConfigGroupRequest) (*data.ConfigGroup, error) {
	endpoint := fmt.Sprintf("add_config_group/%d", projectID)

	jsonBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := c.Post(ctx, endpoint, bytes.NewReader(jsonBody), nil)
	if err != nil {
		return nil, fmt.Errorf("error creating config group: %w", err)
	}
	defer resp.Body.Close()

	var group data.ConfigGroup
	if err := json.NewDecoder(resp.Body).Decode(&group); err != nil {
		return nil, fmt.Errorf("error decoding created config group: %w", err)
	}
	return &group, nil
}

// AddConfig creates a new configuration in a group.
// https://support.testrail.com/hc/en-us/articles/7077719410580-Configurations#addconfig
func (c *HTTPClient) AddConfig(ctx context.Context, groupID int64, req *data.AddConfigRequest) (*data.Config, error) {
	endpoint := fmt.Sprintf("add_config/%d", groupID)

	jsonBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := c.Post(ctx, endpoint, bytes.NewReader(jsonBody), nil)
	if err != nil {
		return nil, fmt.Errorf("error creating config: %w", err)
	}
	defer resp.Body.Close()

	var config data.Config
	if err := json.NewDecoder(resp.Body).Decode(&config); err != nil {
		return nil, fmt.Errorf("error decoding created config: %w", err)
	}
	return &config, nil
}

// UpdateConfigGroup updates a configuration group.
// https://support.testrail.com/hc/en-us/articles/7077719410580-Configurations#updateconfiggroup
func (c *HTTPClient) UpdateConfigGroup(ctx context.Context, groupID int64, req *data.UpdateConfigGroupRequest) (*data.ConfigGroup, error) {
	endpoint := fmt.Sprintf("update_config_group/%d", groupID)

	jsonBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := c.Post(ctx, endpoint, bytes.NewReader(jsonBody), nil)
	if err != nil {
		return nil, fmt.Errorf("error updating config group %d: %w", groupID, err)
	}
	defer resp.Body.Close()

	var group data.ConfigGroup
	if err := json.NewDecoder(resp.Body).Decode(&group); err != nil {
		return nil, fmt.Errorf("error decoding updated config group: %w", err)
	}
	return &group, nil
}

// UpdateConfig updates a configuration.
// https://support.testrail.com/hc/en-us/articles/7077719410580-Configurations#updateconfig
func (c *HTTPClient) UpdateConfig(ctx context.Context, configID int64, req *data.UpdateConfigRequest) (*data.Config, error) {
	endpoint := fmt.Sprintf("update_config/%d", configID)

	jsonBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := c.Post(ctx, endpoint, bytes.NewReader(jsonBody), nil)
	if err != nil {
		return nil, fmt.Errorf("error updating config %d: %w", configID, err)
	}
	defer resp.Body.Close()

	var config data.Config
	if err := json.NewDecoder(resp.Body).Decode(&config); err != nil {
		return nil, fmt.Errorf("error decoding updated config: %w", err)
	}
	return &config, nil
}

// DeleteConfigGroup deletes a configuration group.
// https://support.testrail.com/hc/en-us/articles/7077719410580-Configurations#deleteconfiggroup
func (c *HTTPClient) DeleteConfigGroup(ctx context.Context, groupID int64) error {
	endpoint := fmt.Sprintf("delete_config_group/%d", groupID)

	resp, err := c.Post(ctx, endpoint, bytes.NewReader([]byte("{}")), nil)
	if err != nil {
		return fmt.Errorf("error deleting config group %d: %w", groupID, err)
	}
	defer resp.Body.Close()

	return nil
}

// DeleteConfig deletes a configuration.
// https://support.testrail.com/hc/en-us/articles/7077719410580-Configurations#deleteconfig
func (c *HTTPClient) DeleteConfig(ctx context.Context, configID int64) error {
	endpoint := fmt.Sprintf("delete_config/%d", configID)

	resp, err := c.Post(ctx, endpoint, bytes.NewReader([]byte("{}")), nil)
	if err != nil {
		return fmt.Errorf("error deleting config %d: %w", configID, err)
	}
	defer resp.Body.Close()

	return nil
}
