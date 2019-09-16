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

type mangleAction struct {
}

func (m *mangleAction) makeReject(flow *SecurityAdvisorFlow) *SecurityAdvisorFlow {
	newFlow = *flow
	newFlow.Action = awsflowlogs.ActionReject
	return &newFlow
}

func (m *mangleAction) makeAccept(flow1, flow2 *SecurityAdvisorFlow) *SecurityAdvisorFlow {
	newFlow := *flow1
	if flow1.Start > flow2.Start {
		newFlow = *flow2
	}
	newFlow.Action = awsflowlogs.ActionAccept
	return &newFlow
}

// Mangle action
func (m *mangleAction) Mangle(in []interface{}) (out []interface{}) {
	flowsToLinkIDs := make(map[string][string]*SecurityAdvisorFlow)
	for _, flow := range in {
		flow := flow.(*SecurityAdvisorFlow)
		flowsToLinkIDs[flow.l3TrackingID][flow.linkID] = flow
	}

	for trackingID, byLinkID := range flowsToLinkIDs {
		switch len(byLinkID) {
		case 1:
			out := append(out, m.makeReject(byLinkIDs[0]))
		case 2:
			out := append(out, m.makeAccept(byLinkIDs[0], byLinkIDs[1])
		default:
			logging.GetLogger().Warningf("Flow %s captured in more than two points: %v", byLinkID[0].UUID, byLinkIDs)
		}
	}
}

// NewMangleAction create a new mangle
func NewMangleAction(cfg *viper.Viper) (interface{}, error) {
	return &mangleAction{}, nil
}
