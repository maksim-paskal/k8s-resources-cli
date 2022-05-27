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
	"encoding/json"
	"fmt"
	"math"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/api/resource"
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
	OOMKilled      bool
	Evicted        bool
	recomendations *Recomendations
}

func (r *PodResources) String() string {
	result, err := json.Marshal(r)
	if err != nil {
		return err.Error()
	}

	return string(result)
}

func (r *PodResources) SetRecomendation(recomendations *Recomendations) {
	r.recomendations = recomendations
}

func (r *PodResources) GetPodNamespaceName() string {
	return fmt.Sprintf("%s/%s", r.Namespace, r.PodName)
}

func (r *PodResources) GetFormattedResources() *PodResources { //nolint:cyclop
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

	if r.recomendations.OOMKilled || r.OOMKilled {
		result.OOMKilled = true
	}

	const scoreFormat = "%s OK"

	if !result.OOMKilled {
		if score := scoreResourcePlaning(MemoryResourcePlaningType, r.MemoryRequest, r.recomendations.MemoryRequest); score >= GoodResourcePlaningResult { //nolint:lll
			result.MemoryRequest = fmt.Sprintf(scoreFormat, result.MemoryRequest)
		}
	}

	if score := scoreResourcePlaning(CPUResourcePlaningType, r.CPURequest, r.recomendations.CPURequest); score >= GoodResourcePlaningResult { //nolint:lll
		result.CPURequest = fmt.Sprintf(scoreFormat, result.CPURequest)
	}

	return &result
}

type ResourcePlaningType string

const (
	MemoryResourcePlaningType ResourcePlaningType = "memory"
	CPUResourcePlaningType    ResourcePlaningType = "cpu"
)

type ResourcePlaningResult int

func (p *ResourcePlaningResult) String() string {
	result := ""

	for i := 0; i < int(*p); i++ {
		result += "*"
	}

	return result
}

const (
	UnknownResourcePlaningResult ResourcePlaningResult = -1
	BadResourcePlaningResult     ResourcePlaningResult = 0
	GoodResourcePlaningResult    ResourcePlaningResult = 1
	PerfectResourcePlaningResult ResourcePlaningResult = 2
	GeniousResourcePlaningResult ResourcePlaningResult = 3
	GodResourcePlaningResult     ResourcePlaningResult = 4
)

func scoreResourcePlaning(planingType ResourcePlaningType, req, reqrecomend string) ResourcePlaningResult {
	if len(req) == 0 || len(reqrecomend) == 0 {
		return UnknownResourcePlaningResult
	}

	resReq := resource.MustParse(req)
	resReqRecomend := resource.MustParse(reqrecomend)

	f := resReqRecomend.AsApproximateFloat64() / resReq.AsApproximateFloat64()
	if f == 1 {
		return GodResourcePlaningResult
	}

	if f >= 0.8 && f <= 1.2 {
		return GeniousResourcePlaningResult
	}

	okDiffPerfect := resource.MustParse("10Mi")
	okDiffGood := resource.MustParse("100Mi")

	if planingType == CPUResourcePlaningType {
		okDiffPerfect = resource.MustParse("10m")
		okDiffGood = resource.MustParse("20m")
	}

	planDiff := math.Abs(resReqRecomend.AsApproximateFloat64() - resReq.AsApproximateFloat64())

	if planDiff < okDiffPerfect.AsApproximateFloat64() {
		return PerfectResourcePlaningResult
	}

	if planDiff < okDiffGood.AsApproximateFloat64() {
		return GoodResourcePlaningResult
	}

	return BadResourcePlaningResult
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
