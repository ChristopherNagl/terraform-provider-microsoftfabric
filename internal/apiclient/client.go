package apiclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	
	"net/http"
	"net/url"
	"os"
	"time"
	"io"
)

// APIClient struct holds the information needed to authenticate and make requests
type APIClient struct {
	ClientID      string
	ClientSecret  string
	TenantID      string
	Token         string
	TokenExpiry   time.Time
	TokenFilePath string
}

// NewAPIClient initializes a new APIClient
func NewAPIClient(clientID, clientSecret, tenantID, tokenFilePath string) *APIClient {
	client := &APIClient{
		ClientID:      clientID,
		ClientSecret:  clientSecret,
		TenantID:      tenantID,
		TokenFilePath: tokenFilePath,
	}

	// If a token file is provided, try to read the token from the file
	if tokenFilePath != "" {
		client.readTokenFromFile()
	}

	return client
}

// readTokenFromFile reads the token from a JSON file if it exists
func (c *APIClient) readTokenFromFile() error {
	file, err := os.Open(c.TokenFilePath)
	if err != nil {
		// If the file does not exist, just return
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer file.Close()

	var tokenData struct {
		TokenType    string `json:"token_type"`
		Scope        string `json:"scope"`
		ExpiresIn    int    `json:"expires_in"`
		ExtExpiresIn int    `json:"ext_expires_in"`
		AccessToken  string `json:"access_token"`
	}

	if err := json.NewDecoder(file).Decode(&tokenData); err != nil {
		return err
	}

	c.Token = tokenData.AccessToken
	c.TokenExpiry = time.Now().Add(time.Duration(tokenData.ExpiresIn) * time.Second)

	return nil
}

// saveTokenToFile saves the token to a JSON file
func (c *APIClient) saveTokenToFile(tokenData map[string]interface{}) error {
	file, err := os.Create(c.TokenFilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	return json.NewEncoder(file).Encode(tokenData)
}

// GetAccessToken retrieves an access token from Azure AD
func (c *APIClient) GetAccessToken() error {
	// Check if the token is still valid
	if c.Token != "" && time.Now().Before(c.TokenExpiry) {
		return nil
	}

	authorityURL := "https://login.microsoftonline.com/" + c.TenantID + "/oauth2/v2.0/token"
	form := url.Values{}
	form.Set("grant_type", "client_credentials")
	form.Set("client_id", c.ClientID)
	form.Set("client_secret", c.ClientSecret)
	form.Set("scope", "https://analysis.windows.net/powerbi/api/.default")

	resp, err := http.PostForm(authorityURL, form)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	if token, ok := result["access_token"].(string); ok {
		c.Token = token
		c.TokenExpiry = time.Now().Add(time.Duration(result["expires_in"].(float64)) * time.Second)

		// Save the token to file if a token file path is provided
		if c.TokenFilePath != "" {
			c.saveTokenToFile(result)
		}

		return nil
	}

	return fmt.Errorf("failed to get access token")
}

func (c *APIClient) Get(url string) (map[string]interface{}, error) {
	// Ensure we have a valid token
	if err := c.GetAccessToken(); err != nil {
		return nil, fmt.Errorf("failed to acquire token: %v", err)
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.Token))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("resource not found: %v", resp.Status)
	}

	var respBody map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		return nil, err
	}

	return respBody, nil
}

// Post makes a POST request to the specified URL with the given body
func (c *APIClient) Post(url string, body map[string]string) (map[string]interface{}, error) {
    // Ensure we have a valid token
    if err := c.GetAccessToken(); err != nil {
        return nil, fmt.Errorf("failed to acquire token: %v", err)
    }

    bodyBytes, err := json.Marshal(body) // Use regular assignment here
    if err != nil {
        return nil, fmt.Errorf("failed to marshal request body: %v", err)
    }

    req, err := http.NewRequest("POST", url, bytes.NewBuffer(bodyBytes))
    if err != nil {
        return nil, err
    }
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.Token))

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    // Log the status code for debugging
    fmt.Printf("HTTP Status Code: %d\n", resp.StatusCode)

    // Read the response body
    bodyBytes, err = io.ReadAll(resp.Body) // Reassign using = here
    if err != nil {
        return nil, fmt.Errorf("failed to read response body: %v", err)
    }
    
    fmt.Printf("Response Body: %s\n", string(bodyBytes))

    // Handle non-success status codes
    if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
        errorMessage := string(bodyBytes)
        return nil, fmt.Errorf("request failed with status code %d: %s", resp.StatusCode, errorMessage)
    }

    // If the response body is empty, return an empty map
    if len(bodyBytes) == 0 {
        return make(map[string]interface{}), nil
    }

    // Parse the response body
    var responseBody map[string]interface{}
    if err := json.Unmarshal(bodyBytes, &responseBody); err != nil {
        return nil, fmt.Errorf("failed to parse response body: %v", err)
    }

    return responseBody, nil
}



// Delete makes a DELETE request to the specified URL
func (c *APIClient) Delete(url string) error {
	// Ensure we have a valid token
	if err := c.GetAccessToken(); err != nil {
		return fmt.Errorf("failed to acquire token: %v", err)
	}

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.Token))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check the HTTP status code
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

func (c *APIClient) Patch(url string, body map[string]string) (map[string]interface{}, error) {
	// Ensure we have a valid token
	if err := c.GetAccessToken(); err != nil {
		return nil, fmt.Errorf("failed to acquire token: %v", err)
	}

	bodyBytes, _ := json.Marshal(body)
	req, err := http.NewRequest("PATCH", url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.Token))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Check the status code
	switch resp.StatusCode {
	case http.StatusOK:
		// 200 OK: Return success, even if the body is empty
		var respBody map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
			if err == io.EOF {
				// Empty body, return an empty map as success
				return map[string]interface{}{}, nil
			}
			return nil, err
		}
		return respBody, nil

	case http.StatusBadRequest:
		// 400 Bad Request: Return an error
		return nil, fmt.Errorf("request failed with status code 400")

	case http.StatusNotFound:
		// 404 Not Found: Return an error
		return nil, fmt.Errorf("request failed with status code 404")

	default:
		// Handle other status codes if needed
		return nil, fmt.Errorf("request failed with status code %d", resp.StatusCode)
	}
}

// Put makes a PUT request to the specified URL with the given body
func (c *APIClient) Put(url string, body map[string]string) (map[string]interface{}, error) {
	// Ensure we have a valid token
	if err := c.GetAccessToken(); err != nil {
		return nil, fmt.Errorf("failed to acquire token: %v", err)
	}

	bodyBytes, _ := json.Marshal(body)
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.Token))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Check the status code
	switch resp.StatusCode {
	case http.StatusOK, http.StatusCreated:
		// 200 OK or 201 Created: Return success, even if the body is empty
		var respBody map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
			if err == io.EOF {
				// Empty body, return an empty map as success
				return map[string]interface{}{}, nil
			}
			return nil, err
		}
		return respBody, nil

	case http.StatusBadRequest:
		// 400 Bad Request: Return an error
		return nil, fmt.Errorf("request failed with status code 400")

	case http.StatusNotFound:
		// 404 Not Found: Return an error
		return nil, fmt.Errorf("request failed with status code 404")

	default:
		// Handle other status codes if needed
		return nil, fmt.Errorf("request failed with status code %d", resp.StatusCode)
	}
}

// PostBytes makes a POST request to the specified URL with the given body as bytes
// PostBytes makes a POST request to the specified URL with the given body as bytes
// PostBytes makes a POST request to the specified URL with the given body as bytes
// PostBytes makes a POST request to the specified URL with the given body as bytes
func (c *APIClient) PostBytes(url string, bodyBytes []byte) (map[string]interface{}, error) {
    // Ensure we have a valid token
    if err := c.GetAccessToken(); err != nil {
        return nil, fmt.Errorf("failed to acquire token: %v", err)
    }

    req, err := http.NewRequest("POST", url, bytes.NewBuffer(bodyBytes))
    if err != nil {
        return nil, err
    }
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.Token))

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    // Log the status code for debugging
    fmt.Printf("HTTP Status Code: %d\n", resp.StatusCode)

    // Read the response body for debugging
    responseBodyBytes, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, fmt.Errorf("failed to read response body: %v", err)
    }
    fmt.Printf("Response Body: %s\n", string(responseBodyBytes))

    // Handle non-success status codes
    if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
        return nil, fmt.Errorf("request failed with status code %d: %s", resp.StatusCode, string(responseBodyBytes))
    }

    // Handle empty response body
    if len(responseBodyBytes) == 0 {
        // Return nil for the response body if it's empty
        return nil, nil
    }

    // Parse the response body
    var responseBody map[string]interface{}
    if err := json.Unmarshal(responseBodyBytes, &responseBody); err != nil {
        return nil, fmt.Errorf("failed to parse response body: %v", err)
    }

    return responseBody, nil
}


// PostWithOperationCheck makes a POST request and checks the status of the long-running operation
func (c *APIClient) PostWithOperationCheck(url string, body map[string]string) (map[string]interface{}, error) {
    // Ensure we have a valid token
    if err := c.GetAccessToken(); err != nil {
        return nil, fmt.Errorf("failed to acquire token: %v", err)
    }

    bodyBytes, err := json.Marshal(body)
    if err != nil {
        return nil, fmt.Errorf("failed to marshal request body: %v", err)
    }

    req, err := http.NewRequest("POST", url, bytes.NewBuffer(bodyBytes))
    if err != nil {
        return nil, err
    }
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.Token))

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    // Log the status code for debugging
    fmt.Printf("HTTP Status Code: %d\n", resp.StatusCode)

    // Read the response body for debugging
    responseBodyBytes, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, fmt.Errorf("failed to read response body: %v", err)
    }
    fmt.Printf("Response Body: %s\n", string(responseBodyBytes))

    // Check if the response contains an operation ID
    operationID := resp.Header.Get("x-ms-operation-id")
    if operationID == "" {
        return nil, fmt.Errorf("no operation ID found in response")
    }

    // Poll the operation result
    return c.pollOperationResult(operationID)
}

// pollOperationResult polls the operation result until completion
func (c *APIClient) pollOperationResult(operationID string) (map[string]interface{}, error) {
    client := &http.Client{}
    for {
        url := fmt.Sprintf("https://api.fabric.microsoft.com/v1/operations/%s/result", operationID)
        req, err := http.NewRequest("GET", url, nil)
        if err != nil {
            return nil, err
        }
        req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.Token))

        resp, err := client.Do(req)
        if err != nil {
            return nil, err
        }
        defer resp.Body.Close()

        // Read the response body
        responseBodyBytes, err := io.ReadAll(resp.Body)
        if err != nil {
            return nil, fmt.Errorf("failed to read response body: %v", err)
        }
        fmt.Printf("Operation Status Response Body: %s\n", string(responseBodyBytes))

        var responseBody map[string]interface{}
        if err := json.Unmarshal(responseBodyBytes, &responseBody); err != nil {
            return nil, fmt.Errorf("failed to parse response body: %v", err)
        }

        // Check if the operation is complete
        if id, ok := responseBody["id"].(string); ok && id != "" {
            return responseBody, nil
        }

        // Sleep for a while before polling again
        time.Sleep(2 * time.Second)
    }
}





