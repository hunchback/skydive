/*
 * Copyright (C) 2018 Skydive-Project Authors.
 *
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 *  http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 *
 */

package gremlin

import (
	"fmt"
	"time"

	"github.com/skydive-project/skydive/common"
)

// QueryString used to construct string representation of query
type QueryString string

// NewValueStringFromArgument via inferance creates a correct ValueString
func NewQueryStringFromArgument(v interface{}) QueryString {
	switch v := v.(type) {
	case QueryString:
		return v
	case string:
		return QueryString(v)
	default:
		panic("argument type not supported")
	}
}

// String converts value to string
func (v QueryString) String() string {
	return string(v)
}

// G represents the base graph token
const G = QueryString("G")

// Append appends string value to query
func (q QueryString) append(s string) QueryString {
	return QueryString(q.String() + s)
}

func (q QueryString) command(name string, list ...interface{}) QueryString {
	q = q.append(fmt.Sprintf(".%s(", name))
	first := true
	for _, v := range list {
		if !first {
			q = q.append(", ")
		}
		first = false
		q = q.append(NewValueStringFromArgument(v).String())
	}
	return q.append(")")
}

// Aggregates append a Aggregates() operation to query
func (q QueryString) Aggregates() QueryString {
	return q.command("Aggregates")
}

// At append a At() operation to query
func (q QueryString) At(list ...interface{}) QueryString {
	return q.command("At", list...)
}

// Both append a Both() operation to query
func (q QueryString) Both() QueryString {
	return q.command("Both")
}

// BPF append a PBF() operation to query
func (q QueryString) BPF(list ...interface{}) QueryString {
	return q.command("BPF", list)
}

// CaptureNode append a CaptureNode() operation to query
func (q QueryString) CaptureNode() QueryString {
	return q.command("CaptureNode")
}

// Append a Context() operation to query
func (q QueryString) Context(list ...interface{}) QueryString {
	newQ := q.append(".Context(")
	first := true
	for _, v := range list {
		if !first {
			newQ = newQ.append(", ")
		}
		switch v := v.(type) {
		case time.Time:
			if v.IsZero() {
				return q
			}
			newQ = newQ.append(fmt.Sprintf("%d", common.UnixMillis(v)))
		case int:
			newQ = newQ.append(fmt.Sprintf("%d", v))
		}
	}
	return newQ.append(")")
}

// Count append a Count() operation to query
func (q QueryString) Count() QueryString {
	return q.command("Count")
}

// Dedup append a Dedup() operation to query
func (q QueryString) Dedup() QueryString {
	return q.command("Dedup")
}

// Flows append a Flows() operation to query
func (q QueryString) Flows(list ...interface{}) QueryString {
	return q.command("Flows", list...)
}

// Has append a Has() operation to query
func (q QueryString) Has(list ...interface{}) QueryString {
	return q.command("Has", list...)
}

// HasKey append a HasKey() operation to query
func (q QueryString) HasKey(v interface{}) QueryString {
	return q.command("HasKey", v)
}

// Hops append a Hops() operation to query
func (q QueryString) Hops() QueryString {
	return q.command("Hops")
}

// In append a In() operation to query
func (q QueryString) In() QueryString {
	return q.command("In")
}

// Metrics append a Metrics() operation to query
func (q QueryString) Metrics() QueryString {
	return q.command("Metrics")
}

// RawPackets append a RawPackets() operation to query
func (q QueryString) RawPackets() QueryString {
	return q.command("RawPackets")
}

// ShortestPathTo append a ShortestPathTo() operation to query
func (q QueryString) ShortestPathTo(list ...interface{}) QueryString {
	return q.command("ShortestPathTo", list)
}

// Sort append a Sort() operation to query
func (q QueryString) Sort(list ...interface{}) QueryString {
	return q.command("Sort", list)
}

// Sum append a Sum() operation to query
func (q QueryString) Sum(list ...interface{}) QueryString {
	return q.command("Sum", list)
}

// Nodes append a Nodes() operation to query
func (q QueryString) Nodes() QueryString {
	return q.command("Nodes")
}

// Out append a Out() operation to query
func (q QueryString) Out() QueryString {
	return q.command("Out")
}

// V append a V() operation to query
func (q QueryString) V(list ...interface{}) QueryString {
	return q.command("V", list)
}
