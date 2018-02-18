/*
 * Copyright (C) 2018 IBM, Inc.
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
func (q QueryString) Append(s string) QueryString {
	return QueryString(q.String() + s)
}

// Append a Context() operation to query
func (q QueryString) Context(list ...interface{}) QueryString {
	newQ := q.Append(".Context(")
	first := true
	for _, v := range list {
		if !first {
			newQ = newQ.Append(", ")
		}
		switch v := v.(type) {
		case time.Time:
			if v.IsZero() {
				return q
			}
			newQ = newQ.Append(fmt.Sprintf("%d", common.UnixMillis(v)))
		case int:
			newQ = newQ.Append(fmt.Sprintf("%d", v))
		}
	}
	return newQ.Append(")")
}

// V append a V() operation to query
func (q QueryString) V() QueryString {
	return q.Append(".V()")
}

// Has append a Has() operation to query
func (q QueryString) Has(list ...interface{}) QueryString {
	q = q.Append(".Has(")
	first := true
	for _, v := range list {
		if !first {
			q = q.Append(", ")
		}
		first = false
		q = q.Append(NewValueStringFromArgument(v).String())
	}
	return q.Append(")")
}

// Flows append a Flows() operation to query
func (q QueryString) Flows() QueryString {
	return q.Append(".Flows()")
}

// Sort append a Sort() operation to query
func (q QueryString) Sort() QueryString {
	return q.Append(".Sort()")
}

// ValueString a value used within query constructs
type ValueString string

// NewValueStringFromArgument via inferance creates a correct ValueString
func NewValueStringFromArgument(v interface{}) ValueString {
	switch v := v.(type) {
	case ValueString:
		return v
	case string:
		return Quote(v)
	default:
		panic("argument type not supported")
	}
}

// String converts value to string
func (v ValueString) String() string {
	return string(v)
}

// Quote used to quote string values as needed by query
func (v ValueString) Quote() ValueString {
	return ValueString(fmt.Sprintf(`"%s"`, v))
}

// Regex used for constructing a regexp expression string
func (v ValueString) Regex() ValueString {
	return ValueString(fmt.Sprintf("Regex(%s)", v.Quote()))
}

// StartsWith construct a regexp representing all that start with string
func (v ValueString) StartsWith() ValueString {
	return ValueString(fmt.Sprintf("%s.*", v)).Regex()
}

// EndsWith construct a regexp representing all that end with string
func (v ValueString) EndsWith() ValueString {
	return ValueString(fmt.Sprintf(".*%s", v)).Regex()
}

// Quote used to quote string values as needed by query
func Quote(s string) ValueString {
	return ValueString(s).Quote()
}

// Regex used for constructing a regexp expression string
func Regex(s string) ValueString {
	return ValueString(s).Regex()
}

// StartsWith construct a regexp representing all that start with string
func StartsWith(s string) ValueString {
	return ValueString(s).StartsWith()
}

// EndsWith construct a regexp representing all that end with string
func EndsWith(s string) ValueString {
	return ValueString(s).EndsWith()
}
