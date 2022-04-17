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
	"os"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

type AppConfig struct {
	KubeConfigFile  *string
	LogLevel        *string
	Namespace       *string
	ShowQoS         *bool
	NoCPURequest    *bool
	NoMemoryRequest *bool
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
	Namespace:       flag.String("namespace", "", "filter by namespace"),
	LogLevel:        flag.String("logLevel", "INFO", "log level"),
	KubeConfigFile:  flag.String("kubeconfig", os.Getenv("KUBECONFIG"), "kubeconfig path"),
	ShowQoS:         flag.Bool("ShowQoS", false, "show QoS"),
	NoCPURequest:    flag.Bool("NoCPURequest", false, "show only when cpu request is not set"),
	NoMemoryRequest: flag.Bool("NoMemoryRequest", false, "show only when memory request is not set"),
}

func Get() *AppConfig {
	return appConfig
}

//nolint:gochecknoglobals
var gitVersion = "dev"

func GetVersion() string {
	return gitVersion
}
