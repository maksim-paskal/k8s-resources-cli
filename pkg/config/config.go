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
package config

import (
	"flag"
	"io/ioutil"
	"os"

	"github.com/maksim-paskal/k8s-resources-cli/pkg/types"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

type AppConfig struct {
	ConfigFile           *string
	KubeConfigFile       *string
	LogLevel             *string
	Namespace            *string
	Filter               *string
	ShowQoS              *bool
	ShowSafeToEvict      *bool
	NoCPURequest         *bool
	NoMemoryRequest      *bool
	PrometheusURL        *string
	PrometheusUser       *string
	PrometheusPassword   *string
	ShowDebugJSON        *bool
	PodLabelSelector     *string
	PrometheusGroupField *string
	PrometheusGroupValue *string
	PrometheusHorizont   *string
	LimitsStrategy       *string
}

func (c *AppConfig) String() string {
	out, err := yaml.Marshal(c)
	if err != nil {
		return errors.Wrap(err, "error marshal app config").Error()
	}

	return string(out)
}

//nolint:gochecknoglobals
var appConfig = &AppConfig{
	ConfigFile:           flag.String("config", "", "application config"),
	Namespace:            flag.String("namespace", "", "filter by namespace"),
	LogLevel:             flag.String("logLevel", "INFO", "log level"),
	KubeConfigFile:       flag.String("kubeconfig", os.Getenv("KUBECONFIG"), "kubeconfig path"),
	Filter:               flag.String("filter", "", "golang filter expression"),
	PodLabelSelector:     flag.String("podLabelSelector", "", "pod label selector"),
	ShowQoS:              flag.Bool("ShowQoS", false, "show QoS"),
	ShowSafeToEvict:      flag.Bool("ShowSafeToEvict", false, "show Autoscaler safeToEvict"),
	NoCPURequest:         flag.Bool("NoCPURequest", false, "show only when cpu request is not set"),
	NoMemoryRequest:      flag.Bool("NoMemoryRequest", false, "show only when memory request is not set"),
	PrometheusURL:        flag.String("prometheus.url", "", "prometheus url"),
	PrometheusUser:       flag.String("prometheus.user", "", "prometheus basic auth user"),
	PrometheusPassword:   flag.String("prometheus.password", "", "prometheus basic auth password"),
	PrometheusGroupField: flag.String("prometheus.group.field", "", "prometheus shared group field"),
	PrometheusGroupValue: flag.String("prometheus.group.value", "", "prometheus shared group value"),
	PrometheusHorizont:   flag.String("prometheus.horizont", "7d", "metrics period to process"),
	ShowDebugJSON:        flag.Bool("ShowDebugJSON", false, "show debug json"),
	LimitsStrategy:       flag.String("LimitsStrategy", "conservative", "limits strategy"),
}

func Load() error {
	configByte, err := ioutil.ReadFile(*appConfig.ConfigFile)
	if err != nil {
		log.Debug(err)

		return nil
	}

	err = yaml.Unmarshal(configByte, &appConfig)
	if err != nil {
		return errors.Wrap(err, "error unmarshal config")
	}

	return nil
}

func Check() error {
	_, err := types.ParseStrategyType(*appConfig.LimitsStrategy)
	if err != nil {
		return errors.Wrap(err, "error parse limits strategy")
	}

	return nil
}

func Get() *AppConfig {
	return appConfig
}

//nolint:gochecknoglobals
var gitVersion = "dev"

func GetVersion() string {
	return gitVersion
}
