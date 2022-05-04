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
	"fmt"

	"github.com/pkg/errors"
)

// Recommend for container resources.
type Recomendations struct {
	MemoryRequest string
	MemoryLimit   string
	CPURequest    string
	CPULimit      string
	OOMKilled     bool
}

// Pod results.
type PodResources struct {
	PodName        string
	PodTemplate    string
	ContainerName  string
	NodeName       string
	Namespace      string
	MemoryRequest  string
	MemoryLimit    string
	CPURequest     string
	CPULimit       string
	QoS            string
	SafeToEvict    bool
	recomendations *Recomendations
}

func (r *PodResources) SetRecomendation(recomendations *Recomendations) {
	r.recomendations = recomendations
}

func (r *PodResources) GetPodNamespaceName() string {
	return fmt.Sprintf("%s/%s", r.Namespace, r.PodName)
}

func (r *PodResources) GetFormattedResources() *PodResources {
	if r.recomendations == nil {
		return r
	}

	result := PodResources{}

	if len(r.recomendations.MemoryRequest) > 0 {
		result.MemoryRequest = fmt.Sprintf("%s / %s", r.MemoryRequest, r.recomendations.MemoryRequest)
	} else {
		result.MemoryRequest = r.MemoryRequest
	}

	if len(r.recomendations.MemoryLimit) > 0 {
		result.MemoryLimit = fmt.Sprintf("%s / %s", r.MemoryLimit, r.recomendations.MemoryLimit)
	} else {
		result.MemoryLimit = r.MemoryLimit
	}

	if len(r.recomendations.CPURequest) > 0 {
		result.CPURequest = fmt.Sprintf("%s / %s", r.CPURequest, r.recomendations.CPURequest)
	} else {
		result.CPURequest = r.CPURequest
	}

	if len(r.recomendations.CPULimit) > 0 {
		result.CPULimit = fmt.Sprintf("%s / %s", r.CPULimit, r.recomendations.CPULimit)
	} else {
		result.CPULimit = r.CPULimit
	}

	if r.recomendations.OOMKilled {
		result.MemoryLimit = fmt.Sprintf("%s OOMKilled", result.MemoryLimit)
	} else {
		result.MemoryLimit = r.MemoryLimit
	}

	return &result
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

// Grouping metrics key.
type GroupBy string

const (
	GroupByContainer   = GroupBy("container")
	GroupByPod         = GroupBy("pod")
	GroupByPodTemplate = GroupBy("podtemplate")
)

func ParseGroupBy(groupBy string) (GroupBy, error) {
	switch groupBy {
	case "container":
		return GroupByContainer, nil
	case "pod":
		return GroupByPod, nil
	case "podtemplate":
		return GroupByPodTemplate, nil
	default:
		return "", errors.Errorf("unknown collector type %s", groupBy)
	}
}
