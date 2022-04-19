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
package recomender_test

import (
	"testing"

	"github.com/maksim-paskal/k8s-resources-cli/pkg/recomender"
	"k8s.io/apimachinery/pkg/api/resource"
)

func TestMbToString(t *testing.T) {
	t.Parallel()

	tests := make(map[int64]string)

	tests[1] = "1m"
	tests[100] = "100m"
	tests[1000] = "1Ki"
	tests[10000] = "10Ki"
	tests[100000] = "100Ki"
	tests[1000000] = "1Mi"
	tests[1200000] = "1.20Mi"
	tests[13000000] = "13Mi"
	tests[140000000] = "140Mi"

	for in, want := range tests {
		got := recomender.ByteCountSI(in)

		// test that kubertnetes can parse this string
		_, err := resource.ParseQuantity(got)
		if err != nil {
			t.Fatal(err)
		}

		if got != want {
			t.Fatalf("want %s, got %s", want, got)
		}
	}
}
