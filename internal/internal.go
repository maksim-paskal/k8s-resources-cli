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
package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/maksim-paskal/k8s-resources-cli/pkg/api"
	"github.com/maksim-paskal/k8s-resources-cli/pkg/config"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

func Run() error { //nolint:funlen,cyclop
	pods, err := api.GetPodResources()
	if err != nil {
		return err //nolint:wrapcheck
	}

	var b bytes.Buffer
	w := tabwriter.NewWriter(&b, 0, 0, 1, ' ', tabwriter.Debug)

	header := []string{
		"PodName",
		"ContainerName",
		"MemoryRequest",
		"MemoryLimit",
		"CPURequest",
		"CPULimit",
	}

	if *config.Get().ShowQoS {
		header = append(header, "QoS")
	}

	if *config.Get().ShowSafeToEvict {
		header = append(header, "SafeToEvict")
	}

	if *config.Get().ShowDebugJSON {
		header = append(header, "Debug")
	}

	fmt.Fprintln(w, strings.Join(header, "\t"))

	const separatorString = "------"

	separator := make([]string, len(header))
	for i := range header {
		separator[i] = "------"
	}

	fmt.Fprintln(w, strings.Join(separator, "\t"))

	if len(pods) == 0 {
		return errors.New("no pods found")
	}

	sort.Slice(pods, func(i, j int) bool {
		return pods[i].PodName < pods[j].PodName
	})

	for _, result := range pods {
		item := make([]string, 0)

		if result.Recommend != nil {
			result.MemoryRequest = fmt.Sprintf("%s / %s", result.MemoryRequest, result.Recommend.MemoryRequest)
			result.MemoryLimit = fmt.Sprintf("%s / %s", result.MemoryLimit, result.Recommend.MemoryLimit)
			result.CPURequest = fmt.Sprintf("%s / %s", result.CPURequest, result.Recommend.CPURequest)
			result.CPULimit = fmt.Sprintf("%s / %s", result.CPULimit, result.Recommend.CPULimit)
		}

		item = append(item, result.PodName)
		item = append(item, result.ContainerName)
		item = append(item, result.MemoryRequest)
		item = append(item, result.MemoryLimit)
		item = append(item, result.CPURequest)
		item = append(item, result.CPULimit)

		if *config.Get().ShowQoS {
			item = append(item, result.QoS)
		}

		if *config.Get().ShowSafeToEvict {
			item = append(item, strconv.FormatBool(result.SafeToEvict))
		}

		if *config.Get().ShowDebugJSON {
			info, err := json.Marshal(result)
			if err != nil {
				return errors.Wrap(err, "error marshalling result")
			}

			item = append(item, string(info))
		}

		fmt.Fprintln(w, strings.Join(item, "\t"))
	}

	w.Flush()

	fmt.Println(b.String()) //nolint:forbidigo

	const filePermission = 0o755

	err = ioutil.WriteFile("result.txt", b.Bytes(), os.FileMode(filePermission))
	if err != nil {
		log.WithError(err).Error("error writing result to file")
	}

	return nil
}
