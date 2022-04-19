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
package recomender

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/maksim-paskal/k8s-resources-cli/pkg/config"
	"github.com/maksim-paskal/k8s-resources-cli/pkg/types"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	promConfig "github.com/prometheus/common/config"
	"github.com/prometheus/common/model"
	log "github.com/sirupsen/logrus"
)

type Requests struct {
	MemoryRequest string
	MemoryLimit   string
	CPURequest    string
	CPULimit      string
}

const (
	bytesUnit = 1000
)

func Get(container, namespace string) (*Requests, error) { //nolint:funlen,cyclop
	limitsStrategy, err := types.ParseStrategyType(*config.Get().LimitsStrategy)
	if err != nil {
		return nil, errors.Wrap(err, "error parsing strategy")
	}

	metricsExtra := ""

	if len(*config.Get().PrometheusGroupField) > 0 {
		metricsExtra = fmt.Sprintf(`,%s=~"%s"`, *config.Get().PrometheusGroupField, *config.Get().PrometheusGroupValue)
	}

	// requests = 50 percentile of resource usage
	// limits (conservate) = max resource usage
	// limits (aggressive) = 99 percentile of resource usage
	memoryRequestQuery := fmt.Sprintf(`max(quantile_over_time(0.50,container_memory_working_set_bytes{container="%s",namespace="%s"%s}[%s]))`, container, namespace, metricsExtra, *config.Get().PrometheusHorizont)                 //nolint:lll
	memoryLimitQueryAggresive := fmt.Sprintf(`max(quantile_over_time(0.99,container_memory_working_set_bytes{container="%s",namespace="%s"%s}[%s]))`, container, namespace, metricsExtra, *config.Get().PrometheusHorizont)          //nolint:lll
	memoryLimitQueryConservate := fmt.Sprintf(`max(max_over_time(container_memory_working_set_bytes{container="%s",namespace="%s"%s}[%s]))`, container, namespace, metricsExtra, *config.Get().PrometheusHorizont)                   //nolint:lll
	cpuRequestQuery := fmt.Sprintf(`max(quantile_over_time(0.50,rate(container_cpu_usage_seconds_total{container="%s",namespace="%s"%s}[1m])[%s:1m]))`, container, namespace, metricsExtra, *config.Get().PrometheusHorizont)        //nolint:lll
	cpuLimitQueryAggresive := fmt.Sprintf(`max(quantile_over_time(0.99,rate(container_cpu_usage_seconds_total{container="%s",namespace="%s"%s}[1m])[%s:1m]))`, container, namespace, metricsExtra, *config.Get().PrometheusHorizont) //nolint:lll
	cpuLimitQueryConservate := fmt.Sprintf(`max(max_over_time(rate(container_cpu_usage_seconds_total{container="%s",namespace="%s"%s}[1m])[%s:1m]))`, container, namespace, metricsExtra, *config.Get().PrometheusHorizont)          //nolint:lll

	memoryRequest, err := getMetrics(memoryRequestQuery)
	if err != nil {
		return nil, errors.Wrap(err, "error getting memory request")
	}

	memoryLimitQuery := memoryLimitQueryConservate
	if limitsStrategy == types.StrategyTypeAggressive {
		memoryLimitQuery = memoryLimitQueryAggresive
	}

	memoryLimit, err := getMetrics(memoryLimitQuery)
	if err != nil {
		return nil, errors.Wrap(err, "error getting memory limits")
	}

	cpuRequest, err := getMetrics(cpuRequestQuery)
	if err != nil {
		return nil, errors.Wrap(err, "error getting cpu request")
	}

	cpuLimitQuery := cpuLimitQueryConservate
	if limitsStrategy == types.StrategyTypeAggressive {
		cpuLimitQuery = cpuLimitQueryAggresive
	}

	cpuLimit, err := getMetrics(cpuLimitQuery)
	if err != nil {
		return nil, errors.Wrap(err, "error getting cpu limits")
	}

	result := Requests{}

	if len(memoryRequest) == 1 {
		result.MemoryRequest = ByteCountSI(int64(memoryRequest[0].Value))
	}

	if len(memoryLimit) == 1 {
		result.MemoryLimit = ByteCountSI(int64(memoryLimit[0].Value))
	}

	if len(cpuRequest) == 1 {
		b := fmt.Sprintf("%.0fm", cpuRequest[0].Value*bytesUnit)

		result.CPURequest = strings.ReplaceAll(b, ".00", "")
	}

	if len(cpuLimit) == 1 {
		b := fmt.Sprintf("%.0fm", cpuLimit[0].Value*bytesUnit)

		result.CPULimit = strings.ReplaceAll(b, ".00", "")
	}

	return &result, nil
}

func getMetrics(query string) (model.Vector, error) {
	log.Debugf("query: %s", query)

	prometheusConfig := api.Config{
		Address: *config.Get().PrometheusURL,
	}

	if len(*config.Get().PrometheusUser) > 0 {
		prometheusConfig.RoundTripper = promConfig.NewBasicAuthRoundTripper(
			*config.Get().PrometheusUser,
			promConfig.Secret(*config.Get().PrometheusPassword),
			"",
			api.DefaultRoundTripper,
		)
	}

	client, err := api.NewClient(prometheusConfig)
	if err != nil {
		return nil, errors.Wrap(err, "error creating client")
	}

	v1api := v1.NewAPI(client)

	result, warnings, err := v1api.Query(context.Background(), query, time.Now())
	if err != nil {
		return nil, errors.Wrap(err, "error creating client")
	}

	if len(warnings) > 0 {
		log.Warn(warnings)
	}

	v, ok := result.(model.Vector)
	if !ok {
		return nil, errors.New("assertion error")
	}

	return v, nil
}

func ByteCountSI(b int64) string {
	if b < bytesUnit {
		return fmt.Sprintf("%dm", b)
	}

	div, exp := int64(bytesUnit), 0
	for n := b / bytesUnit; n >= bytesUnit; n /= bytesUnit {
		div *= bytesUnit
		exp++
	}

	q := float64(b) / float64(div)

	result := fmt.Sprintf("%.2f%ci", q, "KMGTPE"[exp])

	return strings.ReplaceAll(result, ".00", "")
}
