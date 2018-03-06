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
	"strconv"
)

// ValueString a value used within query constructs
type ValueString string

// newValueStringFromArgument via inferance creates a correct ValueString
func NewValueStringFromArgument(v interface{}) ValueString {
	switch v := v.(type) {
	case ValueString:
		return v
	case string:
		return Quote(v)
	case int:
		return ValueString(strconv.Itoa(v))
	default:
		panic(fmt.Sprintf("argument %v: type %T not supported", v, v))
	}
}

// String converts value to string
func (v ValueString) String() string {
	return string(v)
}

// DESC const definition
const DESC = ValueString("DESC")

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

func newValueString(name string, list ...interface{}) ValueString {
	s := fmt.Sprintf("%s(", name)
	first := true
	for _, v := range list {
		if !first {
			s = s + ", "
		}
		first = false
		s = s + NewValueStringFromArgument(v).String()
	}
	return ValueString(s + ")")
}

// Between append a Between() operation to query
func Between(list ...interface{}) ValueString {
	return newValueString("Between", list...)
}

// Contains append a Contains() operation to query
func Contains(v interface{}) ValueString {
	return newValueString("Contains", v)
}

// Gt append a Gt() operation to query
func Gt(v interface{}) ValueString {
	return newValueString("Gt", v)
}

// Gte append a Gte() operation to query
func Gte(v interface{}) ValueString {
	return newValueString("Gte", v)
}

// Ipv4Range append a Ipv4Range() operation to query
func Ipv4Range(list ...interface{}) ValueString {
	return newValueString("Ipv4Range", list...)
}

// Inside append a Inside() operation to query
func Inside(list ...interface{}) ValueString {
	return newValueString("Inside", list...)
}

// Lt append a Lt() operation to query
func Lt(v interface{}) ValueString {
	return newValueString("Lt", v)
}

// Lte append a Lte() operation to query
func Lte(v interface{}) ValueString {
	return newValueString("Lte", v)
}

// Metadata append a Metadata() operation to query
func Metadata(list ...interface{}) ValueString {
	return newValueString("Metadata", list...)
}

// Within append a Within() operation to query
func Within(list ...interface{}) ValueString {
	return newValueString("Within", list...)
}
