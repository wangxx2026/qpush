package config

// Value contains all config info
type Value struct {
	Env       string
	Ak        string
	Sk        string
	GwHost    string
	RedisAddr string
	Servers   []string
	RabbitMQ  string
	PushQueue string
}

var (
	config Value
)

const (
	// ProdEnv is the production string for prod
	ProdEnv = "prod"
	// DevEnv is for dev
	DevEnv = "dev"
)

// Load init conf for environment
func Load(env string) (*Value, error) {
	err := DecodeTOMLFile("pkg/config/"+env+".toml", &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

// Get fetches the config value
func Get() *Value {
	return &config
}
