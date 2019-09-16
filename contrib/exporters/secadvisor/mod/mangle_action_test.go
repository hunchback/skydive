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

const logStatusConfig = []byte(`
pipeline:
  mangle:
    type: action
`)

func getMangleAction(t *testing.T) core.Mangler {
	cfg := ConfigFromJSON(logStatusConfig)
	mangler, err := NewMangleAction(cfg)
	if err != nil {
		t.Fatalf("Mangle creation returned unexpected error: %s", err)
	}
	return mangler.(core.Mangler)
}

func TestMangleActionReject(t *testing.T) {
	mangler := getMangleAction(t)

	in := []interface{}{
		&SecurityAdvisorFlow{
			UUID: "A",
			l3TrackingID: 1,
			linkID: 1,
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

	if flow.Action != awsflowlogs.ActionReject {
		t.Fatalf("Expected flow.Action '%s' but got: '%s'", awsflowlogs.ActionReject, flow.Action)
	}
}

func TestMangleActionAccepot(t *testing.T) {
	mangler := getMangleAction(t)

	in := []interface{}{
		&SecurityAdvisorFlow{
			UUID: "A",
			l3TrackingID: 1,
			linkID: 1,
		},
		&SecurityAdvisorFlow{
			UUID: "B",
			l3TrackingID: 1,
			linkID: 2,
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

	if flowOk.Action != awsflowlogs.ActionAccept {
		t.Fatalf("Expected flow.Action '%s' but got: '%s'", awsflowlogs.ActionAccept, flow.Action)
	}
}
