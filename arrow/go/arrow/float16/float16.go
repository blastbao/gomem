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

package float16 // import "github.com/apache/arrow/go/arrow/float16"

import (
	"math"
	"strconv"
)

// 根据 IEEE 754 标准，不同的指数位和尾数位的组合方式可以表示不同的数值区间。
// 例如，
//	当指数位全为 0 时，即 exp == 0，表示的是非正规化数，此时尾数 fc 相当于小数部分，计算公式为 2^(-14) * fc。
// 	当指数位全为 1 时，即 exp == 0xff，表示特殊数或无穷数，此时尾数的值不重要。
//	当指数位在 1~30 范围内时，表示正常的浮点数，此时尾数 fc 相当于小数部分，计算公式为 1 + fc * 2^(-10)。
//
// 根据指数计算出的对应值 res 为 0 或 1~30 时，将符号位、指数和尾数按位拼接到一起，构成一个 16 位的半精度浮点数，存储在 Num 类型的 bits 字段中。
// 如果 res 超过了 30，表示溢出了半精度浮点数能够表示的最大值，此时将其置为 31，同时将尾数清零，得到的结果相当于无穷大。
// 如果 res 小于 1，表示半精度浮点数能够表示的最小非规格化值，此时将其置为 0，同时将尾数清零，得到的结果相当于 0。

// Num represents a half-precision floating point value (float16)
// stored on 16 bits.
//
// See https://en.wikipedia.org/wiki/Half-precision_floating-point_format for more informations.
type Num struct {
	bits uint16
}

// New creates a new half-precision floating point value from the provided
// float32 value.
func New(f float32) Num {
	b := math.Float32bits(f)      // float32 => uint32
	sn := uint16((b >> 31) & 0x1) // 符号位 sn
	exp := (b >> 23) & 0xff       // 指数 exp
	res := int16(exp) - 127 + 15
	fc := uint16(b>>13) & 0x3ff // 尾数 fc
	switch {
	case exp == 0:
		res = 0
	case exp == 0xff:
		res = 0x1f
	case res > 0x1e:
		res = 0x1f
		fc = 0
	case res < 0x01:
		res = 0
		fc = 0
	}
	return Num{bits: (sn << 15) | uint16(res<<10) | fc}
}

func (f Num) Float32() float32 {
	sn := uint32((f.bits >> 15) & 0x1)
	exp := (f.bits >> 10) & 0x1f
	res := uint32(exp) + 127 - 15
	fc := uint32(f.bits & 0x3ff)
	switch {
	case exp == 0:
		res = 0
	case exp == 0x1f:
		res = 0xff
	}
	return math.Float32frombits((sn << 31) | (res << 23) | (fc << 13))
}

func (f Num) Uint16() uint16 { return f.bits }
func (f Num) String() string { return strconv.FormatFloat(float64(f.Float32()), 'g', -1, 32) }
