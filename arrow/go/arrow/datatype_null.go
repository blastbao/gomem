// Licensed to the Apache Software Foundation (ASF) under one
// or more contributor license agreements.  See the NOTICE file
// distributed with this work for additional information
// regarding copyright ownership.  The ASF licenses this file
// to you under the Apache License, Version 2.0 (the
// "License"); you may not use this file except in compliance
// with the License.  You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package arrow

// Null type 并非 null ，它是一种无需真正分配内存的 logical type 。
// struct{} 不占用任何真实内存空间，NullType 则“继承”了这点 。
//
// NullType describes a degenerate array, with zero physical storage.
type NullType struct{}

func (*NullType) ID() Type       { return NULL }
func (*NullType) Name() string   { return "null" }
func (*NullType) String() string { return "null" }

var (
	Null *NullType
	_    DataType = Null
)
