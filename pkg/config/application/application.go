package application

import "github.com/leads-su/version"

type Application struct {
	Version   string
	CommitSha string
	BuildDate string
	BuiltBy   string
}

// InitializeDefaults create new application config instance with default values
func InitializeDefaults() *Application {
	return &Application{
		Version:   version.GetVersion(),
		CommitSha: version.GetCommit(),
		BuildDate: version.GetBuildDate(),
		BuiltBy:   version.GetBuiltBy(),
	}
}
