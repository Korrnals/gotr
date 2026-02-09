// internal/client/configs.go
package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/Korrnals/gotr/internal/models/data"
)

// GetConfigs получает список конфигураций проекта
// https://support.testrail.com/hc/en-us/articles/7077719410580-Configurations#getconfigs
func (c *HTTPClient) GetConfigs(projectID int64) (data.GetConfigsResponse, error) {
	endpoint := fmt.Sprintf("get_configs/%d", projectID)
	resp, err := c.Get(endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting configs for project %d: %w", projectID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned %s for project %d: %s", resp.Status, projectID, string(body))
	}

	var configs data.GetConfigsResponse
	if err := json.NewDecoder(resp.Body).Decode(&configs); err != nil {
		return nil, fmt.Errorf("error decoding configs: %w", err)
	}
	return configs, nil
}

// AddConfigGroup создает новую группу конфигураций
// https://support.testrail.com/hc/en-us/articles/7077719410580-Configurations#addconfiggroup
func (c *HTTPClient) AddConfigGroup(projectID int64, req *data.AddConfigGroupRequest) (*data.ConfigGroup, error) {
	endpoint := fmt.Sprintf("add_config_group/%d", projectID)

	jsonBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %w", err)
	}

	resp, err := c.Post(endpoint, bytes.NewReader(jsonBody), nil)
	if err != nil {
		return nil, fmt.Errorf("error creating config group: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned %s: %s", resp.Status, string(body))
	}

	var group data.ConfigGroup
	if err := json.NewDecoder(resp.Body).Decode(&group); err != nil {
		return nil, fmt.Errorf("error decoding created config group: %w", err)
	}
	return &group, nil
}

// AddConfig создает новую конфигурацию в группе
// https://support.testrail.com/hc/en-us/articles/7077719410580-Configurations#addconfig
func (c *HTTPClient) AddConfig(groupID int64, req *data.AddConfigRequest) (*data.Config, error) {
	endpoint := fmt.Sprintf("add_config/%d", groupID)

	jsonBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %w", err)
	}

	resp, err := c.Post(endpoint, bytes.NewReader(jsonBody), nil)
	if err != nil {
		return nil, fmt.Errorf("error creating config: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned %s: %s", resp.Status, string(body))
	}

	var config data.Config
	if err := json.NewDecoder(resp.Body).Decode(&config); err != nil {
		return nil, fmt.Errorf("error decoding created config: %w", err)
	}
	return &config, nil
}

// UpdateConfigGroup обновляет группу конфигураций
// https://support.testrail.com/hc/en-us/articles/7077719410580-Configurations#updateconfiggroup
func (c *HTTPClient) UpdateConfigGroup(groupID int64, req *data.UpdateConfigGroupRequest) (*data.ConfigGroup, error) {
	endpoint := fmt.Sprintf("update_config_group/%d", groupID)

	jsonBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %w", err)
	}

	resp, err := c.Post(endpoint, bytes.NewReader(jsonBody), nil)
	if err != nil {
		return nil, fmt.Errorf("error updating config group %d: %w", groupID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned %s for group %d: %s", resp.Status, groupID, string(body))
	}

	var group data.ConfigGroup
	if err := json.NewDecoder(resp.Body).Decode(&group); err != nil {
		return nil, fmt.Errorf("error decoding updated config group: %w", err)
	}
	return &group, nil
}

// UpdateConfig обновляет конфигурацию
// https://support.testrail.com/hc/en-us/articles/7077719410580-Configurations#updateconfig
func (c *HTTPClient) UpdateConfig(configID int64, req *data.UpdateConfigRequest) (*data.Config, error) {
	endpoint := fmt.Sprintf("update_config/%d", configID)

	jsonBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %w", err)
	}

	resp, err := c.Post(endpoint, bytes.NewReader(jsonBody), nil)
	if err != nil {
		return nil, fmt.Errorf("error updating config %d: %w", configID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned %s for config %d: %s", resp.Status, configID, string(body))
	}

	var config data.Config
	if err := json.NewDecoder(resp.Body).Decode(&config); err != nil {
		return nil, fmt.Errorf("error decoding updated config: %w", err)
	}
	return &config, nil
}

// DeleteConfigGroup удаляет группу конфигураций
// https://support.testrail.com/hc/en-us/articles/7077719410580-Configurations#deleteconfiggroup
func (c *HTTPClient) DeleteConfigGroup(groupID int64) error {
	endpoint := fmt.Sprintf("delete_config_group/%d", groupID)

	resp, err := c.Post(endpoint, bytes.NewReader([]byte("{}")), nil)
	if err != nil {
		return fmt.Errorf("error deleting config group %d: %w", groupID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API returned %s for group %d: %s", resp.Status, groupID, string(body))
	}
	return nil
}

// DeleteConfig удаляет конфигурацию
// https://support.testrail.com/hc/en-us/articles/7077719410580-Configurations#deleteconfig
func (c *HTTPClient) DeleteConfig(configID int64) error {
	endpoint := fmt.Sprintf("delete_config/%d", configID)

	resp, err := c.Post(endpoint, bytes.NewReader([]byte("{}")), nil)
	if err != nil {
		return fmt.Errorf("error deleting config %d: %w", configID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API returned %s for config %d: %s", resp.Status, configID, string(body))
	}
	return nil
}
