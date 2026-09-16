package config

import "fmt"

// ValidateInitialReports rejects modes that cannot acknowledge local persistence.
func (c *Config) ValidateInitialReports() error {
	if !c.REST.WaitForInitialReports {
		return nil
	}
	if !c.REST.Enabled {
		return fmt.Errorf("rest.waitForInitialReports requires rest.enabled")
	}
	if c.Database.Type != "" && c.Database.Type != "sqlite" {
		return fmt.Errorf("rest.waitForInitialReports requires SQLite; database type %q is unsupported", c.Database.Type)
	}
	return nil
}
