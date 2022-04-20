/*
Copyright paskal.maksim@gmail.com
Licensed under the Apache License, Version 2.0 (the "License")
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package types

import (
	"github.com/pkg/errors"
)

// Recommend for container resources.
type Recomendations struct {
	MemoryRequest string
	MemoryLimit   string
	CPURequest    string
	CPULimit      string
}

// Pod results.
type PodResources struct {
	PodName       string
	PodTemplate   string
	ContainerName string
	Namespace     string
	MemoryRequest string
	MemoryLimit   string
	CPURequest    string
	CPULimit      string
	QoS           string
	SafeToEvict   bool
	Recommend     *Recomendations
}

// strategy to calculate resources.
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

// Type of information to collect.
type CollectorType string

const (
	CollectorTypeContainer   = CollectorType("container")
	CollectorTypePod         = CollectorType("pod")
	CollectorTypePodTemplate = CollectorType("podtemplate")
)

func ParseCollectorType(collectorType string) (CollectorType, error) {
	switch collectorType {
	case "container":
		return CollectorTypeContainer, nil
	case "pod":
		return CollectorTypePod, nil
	case "podtemplate":
		return CollectorTypePodTemplate, nil
	default:
		return "", errors.Errorf("unknown collector type %s", collectorType)
	}
}
