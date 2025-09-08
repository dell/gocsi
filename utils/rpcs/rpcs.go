/*
 *
 * Copyright © 2021-2024 Dell Inc. or its subsidiaries. All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package rpcs

import (
	"fmt"
	"regexp"
	"strconv"
)

const parseMethodPatt = `^/csi\.v(\d+)\.([^/]+?)/(.+)$`

// ParseMethod parses a gRPC method and returns the CSI version, service
// to which the method belongs, and the method's name. An example value
// for the "fullMethod" argument is "/csi.v0.Identity/GetPluginInfo".
func ParseMethod(
	fullMethod string,
) (version int32, service, methodName string, err error) {
	rx := regexp.MustCompile(parseMethodPatt)
	m := rx.FindStringSubmatch(fullMethod)
	if len(m) == 0 {
		return 0, "", "", fmt.Errorf("ParseMethod: invalid: %s", fullMethod)
	}
	v, err := strconv.ParseInt(m[1], 10, 32)
	if err != nil {
		return 0, "", "", fmt.Errorf("ParseMethod: %v", err)
	}
	return int32(v), m[2], m[3], nil
}
