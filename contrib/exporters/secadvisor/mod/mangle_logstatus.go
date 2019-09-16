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
	"time"

	cache "github.com/pmylund/go-cache"
	"github.com/spf13/viper"

	"github.com/skydive-project/skydive/common"
	awsflowlogs "github.com/skydive-project/skydive/contrib/exporters/awsflowlogs/mod"
	"github.com/skydive-project/skydive/flow"
)

type mangleLogStatus struct {
	linkIDs map[string]true
	flows   map[string]*SecurityAdvisorFlow
}

// Mangle log-status book keeping
func (m *mangleLogStatus) Mangle(in []interface{}) (out []interface{}) {
	now := common.UnixMillis(time.Now())

	// OK: all flows
	for _, flow := range in {
		okFlow := flow.(*SecurityAdvisorFlow)
		okFlow.LogStatus = awsflowlogs.LogStatusOK
		out = append(out, &okFlow)
	}

	// NODATA: when nothing was seen per intrerface in measurement window
	linkIDs := make(map[string]true)
	for _, flow := range in {
		linkIDs[flow.Link.ID] = true
	}

	for key := range linkIDs {
		if !m.linkIDs[key] {
			noFlow := &SecurityAdvisorFlow{
				Link: &SecurityAdvisorFlowLayer{
					ID: key,
				},
				Start:     now,
				Last:      now,
				LogStatus: awsflowlogs.LogStatusNoData,
			}
			out = append(out, noFlow)
		}

	}

	for _, flow := range in {
		m.linkIDs[flow.Link.ID] = true
	}

	// SKIPDATA: flows were lost due to error or resource constraint
	for _, flow := range in {
		flow := flow.(*SecurityAdvisorFlow)
		key := flow.TrackingID + "," + flow.NodeTID
		if last, ok := m.flows[key]; ok {
			if flow.LastUpdateMetadata != m.flows[key].Metadata {
				skipFlow := &SecurityAdvisorFlow{
					Link:           flow.Link,
					Network:        flow.Network,
					TransportLayer: flow.TransportLayer,
					ICMPLayer:      flow.ICMPLayer,
					Start:          now,
					Last:           now,
					TrackingID:     flow.TrackingID,
					L3TrackingID:   flow.L3TrackingID,
					NodeTID:        flow.NodeTID,
					LogStatus:      awsflowlogs.LogStatusSkipData,
				}
				out = append(out, skipFlow)
			}
		}
	}

	return
}

// NewMangleLogStatus create a new mangle
func NewMangleLogStatus(cfg *viper.Viper) (interface{}, error) {
	return &mangleLogStatus{
		linkIDs: make(map[string]bool),
		flows:   make(map[string]*SecurityAdvisorFlow),
	}, nil
}
