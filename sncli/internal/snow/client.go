	package snow

	import (
	    "bytes"
	    "encoding/json"
	    "fmt"
	    "io"
	    "net/http"
	    "path"
	)

	// UserInfo contains authenticated user details
	type UserInfo struct {
	    Name     string `json:"name"`
	    Email    string `json:"email"`
	    UserID   string `json:"user_id"`
	}

	// Client represents a ServiceNow API client
	type Client struct {
	    BaseURL     string
	    Username    string
	    Password    string
	    httpClient  *http.Client
	    UserInfo    *UserInfo
	}

	// Error types for the client
	type AuthError struct {
	    Message string
	    Status  int
	}

	func (e *AuthError) Error() string {
	    return fmt.Sprintf("authentication failed: %s (status: %d)", e.Message, e.Status)
	}

	// NewClient creates a new ServiceNow API client
	func NewClient(instanceName, username, password string) (*Client, error) {
	    if instanceName == "" || username == "" || password == "" {
	        return nil, fmt.Errorf("instance name, username and password are required")
	    }
	    baseURL := fmt.Sprintf("https://%s.service-now.com", instanceName)
	    return &Client{
	        BaseURL:     baseURL,
	        Username:    username,
	        Password:    password,
	        httpClient:  &http.Client{},
	    }, nil
	}

	// Authenticate verifies credentials and returns user information
	func (c *Client) Authenticate() (*UserInfo, error) {
	    endpoint := "/api/now/v1/table/sys_user?sysparm_query=user_name=" + c.Username
	    data, err := c.Request("GET", endpoint, nil)
	    if err != nil {
	        return nil, &AuthError{Message: err.Error(), Status: 401}
	    }

	    var response struct {
	        Result []UserInfo `json:"result"`
	    }
	    if err := json.Unmarshal(data, &response); err != nil {
	        return nil, fmt.Errorf("failed to parse user info: %w", err)
	    }

	    if len(response.Result) == 0 {
	        return nil, &AuthError{Message: "user not found", Status: 404}
	    }

	    c.UserInfo = &response.Result[0]
	    return c.UserInfo, nil
	}

	// SaveConfig saves the client configuration
	func (c *Client) SaveConfig() error {
	    cfg := &Config{
	        Instance: path.Base(c.BaseURL),
	        Username: c.Username,
	        Password: c.Password,
	    }
	    return SaveConfig(cfg)
	}

	// Request makes an HTTP request to the ServiceNow API
	func (c *Client) Request(method, endpoint string, data interface{}) ([]byte, error) {
	    var body io.Reader
	    if data != nil {
	        jsonData, err := json.Marshal(data)
	        if err != nil {
	            return nil, fmt.Errorf("failed to marshal request data: %w", err)
	        }
	        body = bytes.NewBuffer(jsonData)
	    }

	    req, err := http.NewRequest(method, c.BaseURL+endpoint, body)
	    if err != nil {
	        return nil, fmt.Errorf("failed to create request: %w", err)
	    }

	    req.SetBasicAuth(c.Username, c.Password)
	    req.Header.Set("Content-Type", "application/json")
	    req.Header.Set("Accept", "application/json")

	    resp, err := c.httpClient.Do(req)
	    if err != nil {
	        return nil, fmt.Errorf("failed to send request: %w", err)
	    }
	    defer resp.Body.Close()

	    responseBody, err := io.ReadAll(resp.Body)
	    if err != nil {
	        return nil, fmt.Errorf("failed to read response: %w", err)
	    }

	    if resp.StatusCode < 200 || resp.StatusCode > 299 {
	        return nil, &AuthError{
	            Message: string(responseBody),
	            Status:  resp.StatusCode,
	        }
	    }

	    return responseBody, nil
	      }

							// Table represents a ServiceNow table metadata
							type Table struct {
							    Name            string           `json:"name"`
							    Label           string           `json:"label"`
							    SysID          string           `json:"sys_id"`
							    Scope          string           `json:"scope"`
							    Description    string           `json:"description"`
							    SuperClass     string           `json:"super_class"`
							    AccessibleFrom string           `json:"accessible_from"`
							    Extendable     bool             `json:"extendable"`
							    NumberPrefix   string           `json:"number_prefix"`
							    Fields        []TableField      `json:"fields,omitempty"`
							}

							// TableField represents a field in a table
							type TableField struct {
							    Name          string `json:"name"`
							    Label         string `json:"label"`
							    Type         string `json:"type"`
							    Length       int    `json:"length"`
							    Reference    string `json:"reference"`
							    IsMandatory  bool   `json:"mandatory"`
							    IsUnique     bool   `json:"unique"`
							}

							// RelationshipInfo represents a relationship between tables
							type RelationshipInfo struct {
							    SourceTable     string `json:"source_table"`
							    TargetTable     string `json:"target_table"`
							    Field          string `json:"field"`
							    Type           string `json:"type"`
							    IsParentChild  bool   `json:"is_parent_child"`
							    Cardinality    string `json:"cardinality"`
							}

							// GetTables retrieves all tables from a specific scope
							func (c *Client) GetTables(scope string, detailed bool) ([]Table, error) {
							    query := "sys_scope.scope=" + scope
							    if scope == "global" {
							        query = "sys_scope=global"
							    }
							    
							    endpoint := "/api/now/table/sys_db_object?sysparm_query=" + query
							    if detailed {
							        endpoint += "&sysparm_display_value=true&sysparm_fields=name,label,sys_id,scope,description,super_class,accessible_from,extendable,number_prefix"
							    }

							    data, err := c.Request("GET", endpoint, nil)
							    if err != nil {
							        return nil, fmt.Errorf("failed to fetch tables: %w", err)
							    }

							    var response struct {
							        Result []Table `json:"result"`
							    }
							    if err := json.Unmarshal(data, &response); err != nil {
							        return nil, fmt.Errorf("failed to parse table data: %w", err)
							    }

							    if detailed {
							        for i, table := range response.Result {
							            fields, err := c.getTableFields(table.Name)
							            if err != nil {
							                return nil, fmt.Errorf("failed to fetch fields for table %s: %w", table.Name, err)
							            }
							            response.Result[i].Fields = fields
							        }
							    }

							    return response.Result, nil
							}

							// getTableFields retrieves all fields for a specific table
							func (c *Client) getTableFields(tableName string) ([]TableField, error) {
							    endpoint := fmt.Sprintf("/api/now/table/sys_dictionary?sysparm_query=name=%s&sysparm_fields=element,column_label,internal_type,max_length,reference,mandatory,unique", tableName)
							    
							    data, err := c.Request("GET", endpoint, nil)
							    if err != nil {
							        return nil, fmt.Errorf("failed to fetch fields: %w", err)
							    }

							    var response struct {
							        Result []TableField `json:"result"`
							    }
							    if err := json.Unmarshal(data, &response); err != nil {
							        return nil, fmt.Errorf("failed to parse field data: %w", err)
							    }

							    return response.Result, nil
							}

							// GetRelationships retrieves all relationships for the given tables
							func (c *Client) GetRelationships(tables []Table) ([]RelationshipInfo, error) {
							    var relationships []RelationshipInfo

							    for _, table := range tables {
							        // Get reference fields pointing to this table
							        endpoint := fmt.Sprintf("/api/now/table/sys_dictionary?sysparm_query=internal_type=reference^reference=%s&sysparm_fields=name,element,column_label,reference,table", table.Name)
							        data, err := c.Request("GET", endpoint, nil)
							        if err != nil {
							            return nil, fmt.Errorf("failed to fetch relationships for table %s: %w", table.Name, err)
							        }

							        var response struct {
							            Result []struct {
							                Name      string `json:"table"`
							                Field     string `json:"element"`
							                Label     string `json:"column_label"`
							                Reference string `json:"reference"`
							            } `json:"result"`
							        }
							        if err := json.Unmarshal(data, &response); err != nil {
							            return nil, fmt.Errorf("failed to parse relationship data: %w", err)
							        }

							        for _, rel := range response.Result {
							            relationships = append(relationships, RelationshipInfo{
							                SourceTable:    rel.Name,
							                TargetTable:    table.Name,
							                Field:         rel.Field,
							                Type:          "reference",
							                IsParentChild: false,
							                Cardinality:   "N:1",
							            })
							        }

							        // Check for parent-child relationships
							        if table.SuperClass != "" {
							            relationships = append(relationships, RelationshipInfo{
							                SourceTable:    table.Name,
							                TargetTable:    table.SuperClass,
							                Field:         "super_class",
							                Type:          "inheritance",
							                IsParentChild: true,
							                Cardinality:   "1:1",
							            })
							        }
							    }

							    return relationships, nil
							}
