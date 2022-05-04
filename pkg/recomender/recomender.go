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
	"github.com/maksim-paskal/k8s-resources-cli/pkg/utils"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	promConfig "github.com/prometheus/common/config"
	"github.com/prometheus/common/model"
	log "github.com/sirupsen/logrus"
)

// nolint:gochecknoglobals
var recomendationCache = make(map[string]*types.Recomendations)

func Get(pod *types.PodResources) (*types.Recomendations, error) { //nolint:funlen,cyclop
	limitsStrategy, err := types.ParseStrategyType(*config.Get().Strategy)
	if err != nil {
		return nil, errors.Wrap(err, "error parsing strategy")
	}

	groupBy, err := types.ParseGroupBy(*config.Get().GroupBy)
	if err != nil {
		return nil, errors.Wrap(err, "error parsing collector type")
	}

	cacheKey := fmt.Sprintf("%s:%s", pod.ContainerName, pod.Namespace)

	metricsExtra := ""

	// search by pod template name
	if groupBy == types.GroupByPodTemplate {
		// if pod is created without replica set, pod template name will be empty
		if len(pod.PodTemplate) == 0 {
			log.Warnf("no pod template value %s/%s, use pod group", pod.Namespace, pod.PodName)

			// use pod group by for pod without template
			groupBy = types.GroupByPod
		} else {
			cacheKey = fmt.Sprintf("%s:%s:%s", pod.PodTemplate, pod.ContainerName, pod.Namespace)
			metricsExtra += fmt.Sprintf(`,pod=~"%s.+"`, pod.PodTemplate)
		}
	}

	// extra fields
	if len(*config.Get().PrometheusGroupField) > 0 {
		metricsExtra += fmt.Sprintf(`,%s=~"%s"`, *config.Get().PrometheusGroupField, *config.Get().PrometheusGroupValue)
	}

	// search by pod name
	if groupBy == types.GroupByPod {
		cacheKey = fmt.Sprintf("%s:%s:%s", pod.PodName, pod.ContainerName, pod.Namespace)
		metricsExtra += fmt.Sprintf(`,pod="%s"`, pod.PodName)
	}

	// check for recomendation in cache
	if _, ok := recomendationCache[cacheKey]; ok {
		log.Debugf("recomendation found in cache key=%s", cacheKey)

		return recomendationCache[cacheKey], nil
	}

	// requests = 50 percentile of resource usage
	// limits (conservate) = max resource usage
	// limits (aggressive) = 99 percentile of resource usage
	memoryRequestQuery := fmt.Sprintf(`max(quantile_over_time(0.50,container_memory_working_set_bytes{container="%s",namespace="%s"%s}[%s]))`, pod.ContainerName, pod.Namespace, metricsExtra, *config.Get().PrometheusRetention)                 //nolint:lll
	memoryLimitQueryAggresive := fmt.Sprintf(`max(quantile_over_time(0.99,container_memory_working_set_bytes{container="%s",namespace="%s"%s}[%s]))`, pod.ContainerName, pod.Namespace, metricsExtra, *config.Get().PrometheusRetention)          //nolint:lll
	memoryLimitQueryConservate := fmt.Sprintf(`max(max_over_time(container_memory_working_set_bytes{container="%s",namespace="%s"%s}[%s]))`, pod.ContainerName, pod.Namespace, metricsExtra, *config.Get().PrometheusRetention)                   //nolint:lll
	cpuRequestQuery := fmt.Sprintf(`max(quantile_over_time(0.50,rate(container_cpu_usage_seconds_total{container="%s",namespace="%s"%s}[1m])[%s:1m]))`, pod.ContainerName, pod.Namespace, metricsExtra, *config.Get().PrometheusRetention)        //nolint:lll
	cpuLimitQueryAggresive := fmt.Sprintf(`max(quantile_over_time(0.99,rate(container_cpu_usage_seconds_total{container="%s",namespace="%s"%s}[1m])[%s:1m]))`, pod.ContainerName, pod.Namespace, metricsExtra, *config.Get().PrometheusRetention) //nolint:lll
	cpuLimitQueryConservate := fmt.Sprintf(`max(max_over_time(rate(container_cpu_usage_seconds_total{container="%s",namespace="%s"%s}[1m])[%s:1m]))`, pod.ContainerName, pod.Namespace, metricsExtra, *config.Get().PrometheusRetention)          //nolint:lll
	oomkilled := fmt.Sprintf(`sum_over_time(kube_pod_container_status_last_terminated_reason{reason="OOMKilled",container="%s",namespace="%s"%s}[%s])`, pod.ContainerName, pod.Namespace, metricsExtra, *config.Get().PrometheusRetention)        //nolint:lll

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

	containerOOMKilled, err := getMetrics(oomkilled)
	if err != nil {
		return nil, errors.Wrap(err, "error getting OOMKilled")
	}

	result := types.Recomendations{}

	if len(memoryRequest) == 1 {
		result.MemoryRequest = utils.ByteCountSI(int64(memoryRequest[0].Value))
	}

	if len(memoryLimit) == 1 {
		result.MemoryLimit = utils.ByteCountSI(int64(memoryLimit[0].Value))
	}

	if len(cpuRequest) == 1 {
		b := fmt.Sprintf("%.0fm", cpuRequest[0].Value*utils.BytesUnit)

		result.CPURequest = strings.ReplaceAll(b, ".00", "")
	}

	if len(cpuLimit) == 1 {
		b := fmt.Sprintf("%.0fm", cpuLimit[0].Value*utils.BytesUnit)

		result.CPULimit = strings.ReplaceAll(b, ".00", "")
	}

	if len(containerOOMKilled) == 1 {
		if containerOOMKilled[0].Value > 0 {
			result.OOMKilled = true
		}
	}

	// add result in cache
	recomendationCache[cacheKey] = &result

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
