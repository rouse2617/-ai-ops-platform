package permission

import "fmt"

type Checker struct {
	config *Config
}

type CheckResult struct {
	Action  Action
	Message string
}

func NewChecker(configPath string) (*Checker, error) {
	cfg, err := LoadConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("load permission config: %w", err)
	}

	return &Checker{config: cfg}, nil
}

func (c *Checker) Check(toolName string) CheckResult {
	for _, rule := range c.config.Rules {
		if Match(rule.Pattern, toolName) {
			return CheckResult{
				Action:  rule.Action,
				Message: rule.Message,
			}
		}
	}

	return CheckResult{
		Action: c.config.Default,
	}
}
