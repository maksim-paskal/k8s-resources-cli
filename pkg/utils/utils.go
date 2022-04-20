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
package utils

import (
	"fmt"
	"strings"
)

const (
	BytesUnit = 1000
)

func ByteCountSI(b int64) string {
	if b < BytesUnit {
		return fmt.Sprintf("%dm", b)
	}

	div, exp := int64(BytesUnit), 0
	for n := b / BytesUnit; n >= BytesUnit; n /= BytesUnit {
		div *= BytesUnit
		exp++
	}

	q := float64(b) / float64(div)

	result := fmt.Sprintf("%.2f%ci", q, "KMGTPE"[exp])

	return strings.ReplaceAll(result, ".00", "")
}
