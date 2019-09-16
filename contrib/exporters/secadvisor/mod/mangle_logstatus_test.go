/*
 * Copyright (C) 2019 IBM, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy ofthe License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specificlanguage governing permissions and
 * limitations under the License.
 *
 */

package mod

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/spf13/viper"

	"github.com/skydive-project/skydive/contrib/exporters/core"
)

func areEqualJSON(buf1, buf2 []byte) (bool, error) {
	var o1 interface{}
	var o2 interface{}

	var err error
	err = json.Unmarshal(buf1, &o1)
	if err != nil {
		return false, err
	}
	err = json.Unmarshal(buf2, &o2)
	if err != nil {
		return false, err
	}
	return reflect.DeepEqual(o1, o2), nil
}

const logStatusConfig = []byte(`
pipeline:
  mangle:
    type: logstatus
`)

func ConfigFromJSON(json string) *viper.Viper {
	cfg := viper.New()
	viper.SetConfigType("json")
	r := bytes.NewReader(json)
	unmarshalReader(r, cfg)
	return cfg
}

func getMangleLogStatus(t *testing.T) core.Mangler {
	cfg := ConfigFromJSON(logStatusConfig)
	mangler, err := NewMangleLogStatus(cfg)
	if err != nil {
		t.Fatalf("Mangle creation returned unexpected error: %s", err)
	}
	return mangler.(core.Mangler)
}

func TestMangleLogStatusOK(t *testing.T) {
	mangler := getMangleLogStatus(t)

	in := []interface{}{
		&SecurityAdvisorFlow{
			UUID: "A",
		},
	}

	out := mangler.Mangle(in)

	if len(out) != 1 {
		t.Fatalf("Expected 1 flow entry but got: %d", len(out))
	}

	flow := out[0]
	if flow.UUID != "A" {
		t.Fatalf("Expected flow.UUID 'A' but got: '%s'", flow.UUID)
	}

	if flow.LogStatus != awsflowlogs.LogStatusOk {
		t.Fatalf("Expected flow.LogStatus '%s' but got: '%s'", awsflowlogs.LogStatusOk, flow.UUID)
	}
}

func TestMangleLogStatusNoData(t *testing.T) {
	mangler := getMangleLogStatus(t)

	in1 := []interface{}{
		&SecurityAdvisorFlow{
			L3TrackingID: 1,
			UUID: "A",
		},
		&SecurityAdvisorFlow{
			L3TrackingID: 1,
			UUID: "B",
		},
	}

	in2 := []interface{}{
		&SecurityAdvisorFlow{
			L3TrackingID: 1,
			UUID: "A",
		},
	}

	mangler.Mangle(in1)
	out := mangler.Mangle(in2)

	if len(out) != 2 {
		t.Fatalf("Expected 2 flow entry but got: %d", len(out))
	}

	flowOk := out[0]

	if flowOk.UUID != "A" {
		t.Fatalf("Expected flow.UUID 'A' but got: '%s'", flowOk.UUID)
	}

	if flowOk.LogStatus != awsflowlogs.LogStatusOk {
		t.Fatalf("Expected flow.LogStatus '%s' but got: '%s'", awsflowlogs.LogStatusOk, flowOk.LogStatus)
	}

	flowNoData = out[1]

	if flowNoData.UUID != "B" {
		t.Fatalf("Expected flow.UUID 'B' but got: '%s'", flowNoData.UUID)
	}

	if flowNoDatsa.LogStatus != awsflowlogs.LogStatusOk {
		t.Fatalf("Expected flow.LogStatus '%s' but got: '%s'", awsflowlogs.LogStatusOk, flowNoData.LogStatus)
	}
}

func TestMangleLogStatusSkipData(t *testing.T) {
	mangler := getMangleLogStatus(t)

	in1 := []interface{}{
		&SecurityAdvisorFlow{
			L3TrackingID: 1,
			UUID: "A",
			Metadata: sameA,
		},
		&SecurityAdvisorFlow{
			L3TrackingID: 1,
			UUID: "B",
			Metadata: sameB,
		},
	}

	in2 := []interface{}{
		&SecurityAdvisorFlow{
			L3TrackingID: 1,
			UUID: "A",
			OldUpdateMetadata: sameA,
		},
		&SecurityAdvisorFlow{
			L3TrackingID: 1,
			UUID: "B",
			OldUpdateMetadata: diffB,
		},
	}

	mangler.Mangle(in1)
	out := mangler.Mangle(in2)
}
