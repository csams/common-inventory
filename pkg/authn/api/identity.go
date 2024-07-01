package api

// Identity is the identity of the requester
type Identity struct {
	Tenant string `yaml:"tenant"`

	Principcal string   `yaml:"principal"`
	Groups     []string `yaml:"groups"`

	// TODO: If we explicitly represent reporters in the database, do we need to distinguish them when they authenticate?
	IsReporter bool `yaml:"is_reporter"`
	IsGuest    bool `yaml:"is_guest"`
}
