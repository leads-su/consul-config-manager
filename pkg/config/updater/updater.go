package updater

// Updater describes structure for updater configuration
type Updater struct {
	Enabled     bool   `mapstructure:"enabled"`
	Type        string `mapstructure:"type"`
	Scheme      string `mapstructure:"scheme"`
	Host        string `mapstructure:"host"`
	Port        uint   `mapstructure:"port"`
	Owner       string `mapstructure:"owner"`
	Repository  string `mapstructure:"repository"`
	ProjectID   uint   `mapstructure:"project_id"`
	AccessToken string `mapstructure:"access_token"`
}

// InitializeDefaults create new updater config instance with default values
func InitializeDefaults() *Updater {
	return &Updater{
		Enabled:     false,
		Type:        "gitlab",
		Scheme:      "http",
		Host:        "gitlab.com",
		Port:        443,
		Owner:       "",
		Repository:  "",
		ProjectID:   0,
		AccessToken: "",
	}
}
