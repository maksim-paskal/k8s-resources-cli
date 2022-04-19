package types

import "github.com/pkg/errors"

type StrategyType string

const (
	StrategyTypeAggressive   = StrategyType("aggressive")
	StrategyTypeConservative = StrategyType("conservative")
)

func ParseStrategyType(strategyType string) (StrategyType, error) {
	switch strategyType {
	case "aggressive":
		return StrategyTypeAggressive, nil
	case "conservative":
		return StrategyTypeConservative, nil
	default:
		return "", errors.Errorf("unknown strategy type %s", strategyType)
	}
}
