// Copyright 2020 the Exposure Notifications Verification Server authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package config

import (
	"context"
	"fmt"
	"time"

	"github.com/google/exposure-notifications-verification-server/pkg/cache"
	"github.com/google/exposure-notifications-verification-server/pkg/database"
	"github.com/google/exposure-notifications-verification-server/pkg/ratelimit"

	"github.com/google/exposure-notifications-server/pkg/observability"

	"github.com/sethvargo/go-envconfig"
)

var _ IssueAPIConfig = (*AdminAPIServerConfig)(nil)

// AdminAPIServerConfig represents the environment based config for the Admin API Server.
type AdminAPIServerConfig struct {
	Database      database.Config
	Observability observability.Config
	Cache         cache.Config
	Features      FeatureConfig

	// SMSSigning defines the SMS signing configuration.
	SMSSigning SMSSigningConfig

	// DevMode produces additional debugging information. Do not enable in
	// production environments.
	DevMode bool `env:"DEV_MODE"`

	// If MaintenanceMode is true, the server is temporarily read-only and will not issue codes.
	MaintenanceMode bool `env:"MAINTENANCE_MODE"`

	// Rate limiting configuration
	RateLimit ratelimit.Config

	Port                string        `env:"PORT,default=8080"`
	APIKeyCacheDuration time.Duration `env:"API_KEY_CACHE_DURATION,default=5m"`

	Issue IssueAPIVars
}

// NewAdminAPIServerConfig returns the environment config for the Admin API server.
// Only needs to be called once per instance, but may be called multiple times.
func NewAdminAPIServerConfig(ctx context.Context) (*AdminAPIServerConfig, error) {
	var config AdminAPIServerConfig
	if err := ProcessWith(ctx, &config, envconfig.OsLookuper()); err != nil {
		return nil, err
	}
	return &config, nil
}

func (c *AdminAPIServerConfig) Validate() error {
	fields := []struct {
		Var  time.Duration
		Name string
	}{
		{c.APIKeyCacheDuration, "API_KEY_CACHE_DURATION"},
	}

	for _, f := range fields {
		if err := checkPositiveDuration(f.Var, f.Name); err != nil {
			return err
		}
	}

	if err := c.Issue.Validate(); err != nil {
		return fmt.Errorf("failed to validate issue API configuration: %w", err)
	}

	return nil
}

func (c *AdminAPIServerConfig) IssueConfig() *IssueAPIVars {
	return &c.Issue
}

func (c *AdminAPIServerConfig) GetRateLimitConfig() *ratelimit.Config {
	return &c.RateLimit
}

func (c *AdminAPIServerConfig) GetFeatureConfig() *FeatureConfig {
	return &c.Features
}

func (c *AdminAPIServerConfig) ObservabilityExporterConfig() *observability.Config {
	return &c.Observability
}

func (c *AdminAPIServerConfig) IsMaintenanceMode() bool {
	return c.MaintenanceMode
}

func (c *AdminAPIServerConfig) GetAuthenticatedSMSFailClosed() bool {
	return c.SMSSigning.FailClosed
}
