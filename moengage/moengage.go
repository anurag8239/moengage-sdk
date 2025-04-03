package moengage

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

type Moengage struct {
	BaseURL string
	AppID   string
	APIKey  string
	Client  *http.Client
}

func NewMoengage(baseURL, appID, apiKey string) *Moengage {
	return &Moengage{
		BaseURL: baseURL,
		AppID:   appID,
		APIKey:  apiKey,
		Client:  &http.Client{},
	}
}

func (m *Moengage) CreateOrUpdateUser(authID, name, phoneNumber, email string, attributes map[string]interface{}) error {
	data := map[string]interface{}{
		"type":        "customer",
		"customer_id": authID,
		"attributes":  attributes,
	}
	if name != "" {
		data["attributes"].(map[string]interface{})["name"] = name
	}
	if phoneNumber != "" {
		data["attributes"].(map[string]interface{})["mobile"] = phoneNumber
	}
	if email != "" {
		data["attributes"].(map[string]interface{})["email"] = email
	}

	url := fmt.Sprintf("%s/v1/customer/%s", m.BaseURL, m.AppID)
	return m.makeRequest(url, data)
}

func (m *Moengage) PublishEvent(authID, eventName string, attributes map[string]interface{}) error {
	data := map[string]interface{}{
		"type":        "event",
		"customer_id": authID,
		"actions": []map[string]interface{}{
			{
				"action":               eventName,
				"attributes":           attributes,
				"user_timezone_offset": 19800,
			},
		},
	}

	url := fmt.Sprintf("%s/v1/event/%s", m.BaseURL, m.AppID)
	return m.makeRequest(url, data)
}

func (m *Moengage) makeRequest(url string, data map[string]interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	encodedAuth := base64.StdEncoding.EncodeToString([]byte(m.AppID + ":" + m.APIKey))
	req.Header.Set("Authorization", "Basic "+encodedAuth)
	req.Header.Set("MOE-APPKEY", m.AppID)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := m.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var responseJSON map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&responseJSON); err != nil {
		return err
	}

	if status, exists := responseJSON["status"].(string); exists && status == "fail" {
		if errorMsg, ok := responseJSON["error"].(map[string]interface{})["message"].(string); ok {
			return errors.New("Moengage API error: " + errorMsg)
		}
		return errors.New("Moengage API returned failure status")
	}

	return nil
}
