package db2se

import (
	"context"
	"db2se/client"
	"errors"
	"fmt"
	"github.com/hashicorp/vault/sdk/helper/ldaputil"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
)

const (
	configStoragePath = "config"
	configPath            = "config"
)

// hashiCupsConfig includes the minimum configuration
// required to instantiate a new HashiCups client.
type db2Config struct {
	ConnectionString string `json:"connection_string"`
}

// pathConfig extends the Vault API with a `/config`
// endpoint for the backend. You can choose whether
// or not certain attributes should be displayed,
// required, and named. For example, password
// is marked as sensitive and will not be output
// when you read the configuration.
func (b *db2Backend) pathConfig() []*framework.Path {
	return []*framework.Path{
		{
			Pattern: configPath,
			Fields:  b.configFields(),
			Operations: map[logical.Operation]framework.OperationHandler{
				logical.CreateOperation: &framework.PathOperation{
					Callback: b.configCreateUpdateOperation,
				},
				logical.UpdateOperation: &framework.PathOperation{
					Callback: b.configCreateUpdateOperation,
				},
				logical.ReadOperation: &framework.PathOperation{
					Callback: b.configReadOperation,
				},
				//logical.DeleteOperation: &framework.PathOperation{
				//	Callback: b.configDeleteOperation,
				//},
			},
			HelpSynopsis: "Configure the OpenLDAP secret engine plugin.",
			HelpDescription: "This path configures the OpenLDAP secret engine plugin. See the documentation for the " +
				"plugin specified for a full list of accepted connection details.",
		},
	}
}

func (b *db2Backend) configFields() map[string]*framework.FieldSchema {
	fields := ldaputil.ConfigFields()
	fields["ttl"] = &framework.FieldSchema{
		Type:        framework.TypeDurationSecond,
		Description: "The default password time-to-live.",
	}
	fields["max_ttl"] = &framework.FieldSchema{
		Type:        framework.TypeDurationSecond,
		Description: "The maximum password time-to-live.",
	}
	fields["password_policy"] = &framework.FieldSchema{
		Type:        framework.TypeString,
		Description: "Password policy to use to generate passwords",
	}
	fields["connection_string"] = &framework.FieldSchema{
		Type:        framework.TypeString,
		Description: "DB2 Database Connection string",
	}

	return fields
}

func (b *db2Backend) configCreateUpdateOperation(ctx context.Context, req *logical.Request, fieldData *framework.FieldData) (*logical.Response, error) {

	connectionString := fieldData.Get("connection_string").(string)

	if connectionString == "" {
		return nil, errors.New("Connection String is required")
	}


	passPolicy := fieldData.Get("password_policy").(string)

	if passPolicy != "" {
		// If both a password policy and a password length are set, we can't figure out what to do
		return nil, fmt.Errorf("cannot set both 'password_policy' and 'length'")
	}

	config := config{
		DB2: &client.Config{
			ConnectionString: connectionString,
		},
		PasswordPolicy: passPolicy,
	}

	_ ,err := writeConfig(ctx, req.Storage, config)
	if err != nil {
		return nil, err
	}

	// Respond with a 204.
	return nil, nil
}

func readConfig(ctx context.Context, storage logical.Storage) (*config, error) {
	entry, err := storage.Get(ctx, configPath)
	if err != nil {
		return nil, err
	}
	if entry == nil {
		return nil, nil
	}
	config := &config{}
	if err := entry.DecodeJSON(config); err != nil {
		return nil, err
	}
	return config, nil
}

func writeConfig(ctx context.Context, storage logical.Storage, config config) (err error, err2 error) {
	entry, err := logical.StorageEntryJSON(configPath, config)
	if err != nil {
		return err, err2
	}
	err = storage.Put(ctx, entry)
	if err != nil {
		return err, err2
	}
	return nil, err2
}

func (b *db2Backend) configReadOperation(ctx context.Context, req *logical.Request, _ *framework.FieldData) (*logical.Response, error) {
	config, err := readConfig(ctx, req.Storage)
	if err != nil {
		return nil, err
	}
	if config == nil {
		return nil, nil
	}

	// "password" is intentionally not returned by this endpoint
	configMap := make(map[string]interface{})

	if config.PasswordPolicy != "" {
		configMap["password_policy"] = config.PasswordPolicy
	}

	if config.PasswordPolicy != "connection_string" {
		configMap["connection_string"] = config.DB2.ConnectionString
	}

	resp := &logical.Response{
		Data: configMap,
	}
	return resp, nil
}

// pathConfigExistenceCheck verifies if the configuration exists.
func (b *db2Backend) pathConfigExistenceCheck(ctx context.Context, req *logical.Request, data *framework.FieldData) (bool, error) {
	out, err := req.Storage.Get(ctx, req.Path)
	if err != nil {
		return false, fmt.Errorf("existence check failed: %w", err)
	}

	return out != nil, nil
}

func getConfig(ctx context.Context, s logical.Storage) (*db2Config, error) {
	entry, err := s.Get(ctx, configStoragePath)
	if err != nil {
		return nil, err
	}

	if entry == nil {
		return nil, nil
	}

	config := new(db2Config)
	if err := entry.DecodeJSON(&config); err != nil {
		return nil, fmt.Errorf("error reading root configuration: %w", err)
	}

	// return the config, we are done
	return config, nil
}

type config struct {
	DB2           *client.Config
	PasswordPolicy string `json:"password_policy,omitempty"`
}


// pathConfigHelpSynopsis summarizes the help text for the configuration
const pathConfigHelpSynopsis = `Configure the HashiCups backend.`

// pathConfigHelpDescription describes the help text for the configuration
const pathConfigHelpDescription = `
The HashiCups secret backend requires credentials for managing
JWTs issued to users working with the products API.

You must sign up with a username and password and
specify the HashiCups address for the products API
before using this secrets backend.
`
