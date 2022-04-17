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
	"context"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/maksim-paskal/k8s-resources-cli/pkg/config"
	"github.com/pkg/errors"
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

func PrintResources() error { //nolint: funlen,cyclop
	ctx := context.Background()

	pods, err := clientset.CoreV1().Pods(*config.Get().Namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return errors.Wrap(err, "error get pods")
	}

	if len(pods.Items) == 0 {
		return errors.New("no pods found")
	}

	podsFound := 0
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
	// defer w.Flush()

	header := []string{
		"Pod",
		"Container",
		"MemoryRequest",
		"MemoryLimit",
		"CPURequest",
		"CPULimit",
	}

	if *config.Get().ShowQoS {
		header = append(header, "QoS")
	}

	fmt.Fprintln(w, strings.Join(header, "\t"))

	const separatorString = "------"

	separator := make([]string, len(header))
	for i := range header {
		separator[i] = "------"
	}

	for _, pod := range pods.Items {
		for _, container := range pod.Spec.Containers {
			podName := pod.Name

			if len(*config.Get().Namespace) == 0 {
				podName = fmt.Sprintf("%s/%s", pod.Namespace, pod.Name)
			}

			showResult := false
			result := make([]string, len(header))

			result[0] = podName
			result[1] = container.Name
			result[2] = container.Resources.Requests.Memory().String()
			result[3] = container.Resources.Limits.Memory().String()
			result[4] = container.Resources.Requests.Cpu().String()
			result[5] = container.Resources.Limits.Cpu().String()

			if *config.Get().ShowQoS {
				result[6] = string(pod.Status.QOSClass)
			}

			if *config.Get().NoMemoryRequest && container.Resources.Requests.Memory().IsZero() {
				showResult = true
			}

			if *config.Get().NoCPURequest && container.Resources.Requests.Cpu().IsZero() {
				showResult = true
			}

			if !*config.Get().NoMemoryRequest && !*config.Get().NoCPURequest {
				showResult = true
			}

			if showResult {
				fmt.Fprintln(w, strings.Join(result, "\t"))
				podsFound++
			}
		}
	}

	if podsFound == 0 {
		return errors.New("no pods found")
	}

	w.Flush()

	return nil
}
