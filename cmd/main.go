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
package main

import (
	"flag"
	"os"

	"github.com/maksim-paskal/k8s-resources-cli/internal"
	"github.com/maksim-paskal/k8s-resources-cli/pkg/api"
	"github.com/maksim-paskal/k8s-resources-cli/pkg/config"
	log "github.com/sirupsen/logrus"
)

var version = flag.Bool("version", false, "show version")

func main() {
	flag.Parse()

	logLevel, err := log.ParseLevel(*config.Get().LogLevel)
	if err != nil {
		log.WithError(err).Fatal("error parse level")
	}

	log.SetLevel(logLevel)

	if *version {
		os.Stdout.WriteString(config.GetVersion())
		os.Exit(0)
	}

	if err := config.Load(); err != nil {
		log.WithError(err).Fatal("error loading config")
	}

	if err := config.Check(); err != nil {
		log.WithError(err).Fatal("error in config")
	}

	log.Debugf("using config:\n%s", config.Get().String())

	if err := api.Init(); err != nil {
		log.WithError(err).Fatal("error connecting kubernetes")
	}

	if err := internal.Run(); err != nil {
		log.WithError(err).Fatal()
	}
}
