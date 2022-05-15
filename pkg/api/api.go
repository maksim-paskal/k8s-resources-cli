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
package api

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"strings"

	"github.com/cheggaaa/pb"
	"github.com/maksim-paskal/k8s-resources-cli/pkg/config"
	"github.com/maksim-paskal/k8s-resources-cli/pkg/recomender"
	"github.com/maksim-paskal/k8s-resources-cli/pkg/types"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// nolint: gochecknoglobals
var clientset *kubernetes.Clientset

func Init() error {
	var (
		kubeconfig *rest.Config
		err        error
	)

	if len(*config.Get().KubeConfigFile) > 0 {
		kubeconfig, err = clientcmd.BuildConfigFromFlags("", *config.Get().KubeConfigFile)
		if err != nil {
			return errors.Wrap(err, "clientcmd.BuildConfigFromFlags")
		}
	} else {
		kubeconfig, err = rest.InClusterConfig()
		if err != nil {
			return errors.Wrap(err, "rest.InClusterConfig")
		}
	}

	clientset, err = kubernetes.NewForConfig(kubeconfig)
	if err != nil {
		return errors.Wrap(err, "kubernetes.NewForConfig")
	}

	return nil
}

func GetPodResources() ([]*types.PodResources, error) { //nolint: funlen,cyclop,gocognit
	ctx := context.Background()

	pods, err := clientset.CoreV1().Pods(*config.Get().Namespace).List(ctx, metav1.ListOptions{
		LabelSelector: *config.Get().PodLabelSelector,
	})
	if err != nil {
		return nil, errors.Wrap(err, "error get pods")
	}

	if len(pods.Items) == 0 {
		return nil, errors.New("no pods found")
	}

	results := make([]*types.PodResources, 0)

	for _, pod := range pods.Items {
		for _, container := range pod.Spec.Containers {
			item := types.PodResources{
				PodName:       pod.Name,
				PodTemplate:   pod.GenerateName,
				ContainerName: container.Name,
				Namespace:     pod.Namespace,
				NodeName:      pod.Spec.NodeName,
				MemoryRequest: container.Resources.Requests.Memory().String(),
				MemoryLimit:   container.Resources.Limits.Memory().String(),
				CPURequest:    container.Resources.Requests.Cpu().String(),
				CPULimit:      container.Resources.Limits.Cpu().String(),
				QoS:           string(pod.Status.QOSClass),
				SafeToEvict:   false,
			}

			podTemplateHash := pod.Labels["pod-template-hash"]

			if len(podTemplateHash) > 0 {
				podTemplateHash = fmt.Sprintf("%s-", podTemplateHash)
				item.PodTemplate = strings.TrimSuffix(item.PodTemplate, podTemplateHash)
			}

			if pod.Annotations["cluster-autoscaler.kubernetes.io/safe-to-evict"] == "false" {
				item.SafeToEvict = true
			}

			if isContainerTerminatedReason(pod, container.Name, "OOMKilled") {
				item.OOMKilled = true
			}

			showResult := false

			if len(*config.Get().Filter) > 0 {
				showResult, err = filterResult(item)
				if err != nil {
					return nil, errors.Wrap(err, "error filtering result")
				}
			}

			if *config.Get().NoMemoryRequest && container.Resources.Requests.Memory().IsZero() {
				showResult = true
			}

			if *config.Get().NoCPURequest && container.Resources.Requests.Cpu().IsZero() {
				showResult = true
			}

			if *config.Get().OOMKilled && item.OOMKilled {
				showResult = true
			}

			if !*config.Get().NoMemoryRequest && !*config.Get().NoCPURequest && !*config.Get().OOMKilled && len(*config.Get().Filter) == 0 { //nolint:lll
				showResult = true
			}

			if showResult {
				results = append(results, &item)
			}
		}
	}

	if err := calculateRecomendations(results); err != nil {
		return nil, errors.Wrap(err, "error adding recommendations")
	}

	return results, nil
}

func calculateRecomendations(results []*types.PodResources) error {
	if len(*config.Get().PrometheusURL) == 0 {
		return nil
	}

	bar := pb.New(len(results))

	showBar := log.GetLevel() < log.DebugLevel

	if showBar {
		bar.Start()
	}

	for i, result := range results {
		recommend, err := recomender.Get(result)
		if err != nil {
			return errors.Wrap(err, "error get metrics")
		}

		results[i].SetRecomendation(recommend)

		bar.Increment()
	}

	if showBar {
		bar.Finish()
	}

	return nil
}

const conditionParts = 2

func filterResult(item types.PodResources) (bool, error) {
	conditions := strings.Split(*config.Get().Filter, ",")
	for _, condition := range conditions {
		eq := strings.Split(condition, "==")
		if len(eq) != conditionParts {
			return false, errors.Errorf("invalid filter condition: %s, it must be .Field==value", condition)
		}

		value, err := templateItem(eq[0], item)
		if err != nil {
			return false, errors.Wrap(err, "error template item")
		}

		if value == eq[1] {
			return true, nil
		}
	}

	return false, nil
}

func templateItem(value string, item types.PodResources) (string, error) {
	tmpl, err := template.New("test").Parse(fmt.Sprintf("{{ %s }}", value))
	if err != nil {
		return "", errors.Wrap(err, "error parsing filter")
	}

	var tpl bytes.Buffer

	err = tmpl.Execute(&tpl, item)
	if err != nil {
		return "", errors.Wrap(err, "error executing filter")
	}

	return tpl.String(), nil
}

func isContainerTerminatedReason(pod corev1.Pod, containerName string, reason string) bool {
	if len(pod.Status.ContainerStatuses) == 0 {
		return false
	}

	for _, containerStatus := range pod.Status.ContainerStatuses {
		if containerStatus.Name == containerName {
			return containerStatus.LastTerminationState.Terminated != nil && containerStatus.LastTerminationState.Terminated.Reason == reason //nolint:lll
		}
	}

	return false
}
