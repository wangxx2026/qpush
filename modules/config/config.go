package config

// Value contains all config info
type Value struct {
	Env string
}

var (
	config Value
)

const (
	// ProductionEnv is the production string for prod
	ProductionEnv = "prod"
)

// Load init conf for environment
func Load(env string) (*Value, error) {
	err := DecodeTOMLFile("modules/config/"+env+".toml", &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

// Get fetches the config value
func Get() *Value {
	return &config
}
