package config

import (
	"fmt"
	"time"

	"github.com/KimMachineGun/automemlimit/memlimit"
	"go.uber.org/zap"
)

func SetupMemLimit(c *Config) error {
	if !c.AutoMemoryLimit.Enabled {
		return nil
	}

	if c.AutoMemoryLimit.Ratio <= 0 || c.AutoMemoryLimit.Ratio > 1.0 {
		return fmt.Errorf("value %f is invalid: ratio must be greater than 0 and less than or equal to 1", c.AutoMemoryLimit.Ratio)
	}

	zap.L().Info("setup memlimit...", zap.Float64("ratio", c.AutoMemoryLimit.Ratio))
	limit, err := memlimit.SetGoMemLimitWithOpts(
		memlimit.WithRatio(c.AutoMemoryLimit.Ratio),
		memlimit.WithProvider(
			memlimit.ApplyFallback(
				memlimit.FromCgroup,
				memlimit.FromSystem,
			),
		),
		memlimit.WithRefreshInterval(5*time.Minute),
	)
	zap.L().Info("configured memlimit...", zap.Int64("limit", limit))

	return err
}
