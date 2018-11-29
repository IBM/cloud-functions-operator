/*
 * Copyright 2017-2018 IBM Corporation
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package common

import (
	"errors"
	"fmt"
	"strings"

	"github.com/apache/incubator-openwhisk-client-go/whisk"
	"k8s.io/apimachinery/pkg/runtime"

	context "github.com/ibm/cloud-operators/pkg/context"
	kv "github.com/ibm/cloud-operators/pkg/types/apis/keyvalue/v1"
)

// ConvertKeyValues convert key value array to whisk key values
func ConvertKeyValues(ctx context.Context, obj runtime.Object, params []kv.KeyValue, what string) (whisk.KeyValueArr, bool, error) {
	keyValArr, err := ToKeyValueArr(ctx, params)
	if err != nil {
		if strings.Contains(err.Error(), "Missing") {
			return nil, true, fmt.Errorf("%v (Retrying)", err)
		}
		return nil, false, fmt.Errorf("Error converting %s: %v", what, err)
	}
	return keyValArr, false, nil
}

// ToKeyValueArr converts a list of key-value pairs to Whisk format
func ToKeyValueArr(ctx context.Context, vars []kv.KeyValue) (whisk.KeyValueArr, error) {
	keyValueArr := make(whisk.KeyValueArr, 0)
	for _, v := range vars {
		var keyVal whisk.KeyValue
		keyVal.Key = v.Name

		value, err := v.ToJSON(ctx)
		if err != nil {
			return nil, err
		}

		if value != nil {
			keyVal.Value = value
			keyValueArr = append(keyValueArr, keyVal)
		}
	}

	return keyValueArr, nil
}

// ToKeyValueArrFromMap converts raw JSON to whisk param format
func ToKeyValueArrFromMap(m interface{}) (whisk.KeyValueArr, error) {
	obj, ok := m.(map[string]interface{})
	if !ok {
		return nil, errors.New("error: JSON value is not an object")
	}
	keyValueArr := make(whisk.KeyValueArr, 0)

	for key := range obj {
		var keyVal whisk.KeyValue
		keyVal.Key = key
		keyVal.Value = obj[key]
		keyValueArr = append(keyValueArr, keyVal)
	}

	return keyValueArr, nil
}

// GetValueString gets the string value for the key
func GetValueString(keyValueArr whisk.KeyValueArr, key string) (string, error) {
	value := keyValueArr.GetValue(key)
	if str, ok := value.(string); ok {
		return str, nil
	}
	return "", fmt.Errorf("missing string value '%v' for key '%s'", value, key)
}
