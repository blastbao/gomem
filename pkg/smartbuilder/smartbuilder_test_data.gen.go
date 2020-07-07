// Code generated by pkg/smartbuilder/smartbuilder_test_data.gen.go.tmpl. DO NOT EDIT.

// Copyright 2019 Nick Poorman
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package smartbuilder

import (
	"strconv"

	"github.com/apache/arrow/go/arrow"
	"github.com/apache/arrow/go/arrow/decimal128"
	"github.com/apache/arrow/go/arrow/float16"
	"github.com/gomem/gomem/pkg/object"
)

type SmartBuilderTestCase struct {
	Values []interface{}
	Dtype  arrow.DataType
	Want   string
}

// TODO: Add boolean, null
func GenerateSmartBuilderTestCases() []SmartBuilderTestCase {
	return []SmartBuilderTestCase{
		{
			Values: BooleanGen(),
			Dtype:  arrow.FixedWidthTypes.Boolean,
			Want:   `col[0][col-bool]: [true false true false true false true false true (null)]`,
		},
		{
			Values: Date32Gen(),
			Dtype:  arrow.FixedWidthTypes.Date32,
			Want:   `col[0][col-date32]: [0 1 2 3 4 5 6 7 8 (null)]`,
		},
		{
			Values: Date64Gen(),
			Dtype:  arrow.FixedWidthTypes.Date64,
			Want:   `col[0][col-date64]: [0 1 2 3 4 5 6 7 8 (null)]`,
		},
		{
			Values: DayTimeIntervalGen(),
			Dtype:  arrow.FixedWidthTypes.DayTimeInterval,
			Want:   `col[0][col-day_time_interval]: [{0 0} {1 2} {2 4} {3 6} {4 8} {5 10} {6 12} {7 14} {8 16} (null)]`,
		},
		{
			Values: Decimal128Gen(),
			Dtype:  &arrow.Decimal128Type{Precision: 1, Scale: 10},
			Want:   `col[0][col-decimal]: [{0 0} {1 1} {2 2} {3 3} {4 4} {5 5} {6 6} {7 7} {8 8} (null)]`,
		},
		{
			Values: Duration_sGen(),
			Dtype:  arrow.FixedWidthTypes.Duration_s,
			Want:   `col[0][col-duration]: [0 1 2 3 4 5 6 7 8 (null)]`,
		},
		{
			Values: Duration_msGen(),
			Dtype:  arrow.FixedWidthTypes.Duration_ms,
			Want:   `col[0][col-duration]: [0 1 2 3 4 5 6 7 8 (null)]`,
		},
		{
			Values: Duration_usGen(),
			Dtype:  arrow.FixedWidthTypes.Duration_us,
			Want:   `col[0][col-duration]: [0 1 2 3 4 5 6 7 8 (null)]`,
		},
		{
			Values: Duration_nsGen(),
			Dtype:  arrow.FixedWidthTypes.Duration_ns,
			Want:   `col[0][col-duration]: [0 1 2 3 4 5 6 7 8 (null)]`,
		},
		{
			Values: Float16Gen(),
			Dtype:  arrow.FixedWidthTypes.Float16,
			Want:   `col[0][col-float16]: [0 1 2 3 4 5 6 7 8 (null)]`,
		},
		{
			Values: Float32Gen(),
			Dtype:  arrow.PrimitiveTypes.Float32,
			Want:   `col[0][col-float32]: [0 1 2 3 4 5 6 7 8 (null)]`,
		},
		{
			Values: Float64Gen(),
			Dtype:  arrow.PrimitiveTypes.Float64,
			Want:   `col[0][col-float64]: [0 1 2 3 4 5 6 7 8 (null)]`,
		},
		{
			Values: Int16Gen(),
			Dtype:  arrow.PrimitiveTypes.Int16,
			Want:   `col[0][col-int16]: [0 1 2 3 4 5 6 7 8 (null)]`,
		},
		{
			Values: Int32Gen(),
			Dtype:  arrow.PrimitiveTypes.Int32,
			Want:   `col[0][col-int32]: [0 1 2 3 4 5 6 7 8 (null)]`,
		},
		{
			Values: Int64Gen(),
			Dtype:  arrow.PrimitiveTypes.Int64,
			Want:   `col[0][col-int64]: [0 1 2 3 4 5 6 7 8 (null)]`,
		},
		{
			Values: Int8Gen(),
			Dtype:  arrow.PrimitiveTypes.Int8,
			Want:   `col[0][col-int8]: [0 1 2 3 4 5 6 7 8 (null)]`,
		},
		{
			Values: MonthIntervalGen(),
			Dtype:  arrow.FixedWidthTypes.MonthInterval,
			Want:   `col[0][col-month_interval]: [0 1 2 3 4 5 6 7 8 (null)]`,
		},
		{
			Values: StringGen(),
			Dtype:  arrow.BinaryTypes.String,
			Want:   `col[0][col-utf8]: ["0" "1" "2" "3" "4" "5" "6" "7" "8" (null)]`,
		},
		{
			Values: Time32sGen(),
			Dtype:  arrow.FixedWidthTypes.Time32s,
			Want:   `col[0][col-time32]: [0 1 2 3 4 5 6 7 8 (null)]`,
		},
		{
			Values: Time32msGen(),
			Dtype:  arrow.FixedWidthTypes.Time32ms,
			Want:   `col[0][col-time32]: [0 1 2 3 4 5 6 7 8 (null)]`,
		},
		{
			Values: Time64usGen(),
			Dtype:  arrow.FixedWidthTypes.Time64us,
			Want:   `col[0][col-time64]: [0 1 2 3 4 5 6 7 8 (null)]`,
		},
		{
			Values: Time64nsGen(),
			Dtype:  arrow.FixedWidthTypes.Time64ns,
			Want:   `col[0][col-time64]: [0 1 2 3 4 5 6 7 8 (null)]`,
		},
		{
			Values: Timestamp_sGen(),
			Dtype:  arrow.FixedWidthTypes.Timestamp_s,
			Want:   `col[0][col-timestamp]: [0 1 2 3 4 5 6 7 8 (null)]`,
		},
		{
			Values: Timestamp_msGen(),
			Dtype:  arrow.FixedWidthTypes.Timestamp_ms,
			Want:   `col[0][col-timestamp]: [0 1 2 3 4 5 6 7 8 (null)]`,
		},
		{
			Values: Timestamp_usGen(),
			Dtype:  arrow.FixedWidthTypes.Timestamp_us,
			Want:   `col[0][col-timestamp]: [0 1 2 3 4 5 6 7 8 (null)]`,
		},
		{
			Values: Timestamp_nsGen(),
			Dtype:  arrow.FixedWidthTypes.Timestamp_ns,
			Want:   `col[0][col-timestamp]: [0 1 2 3 4 5 6 7 8 (null)]`,
		},
		{
			Values: Uint16Gen(),
			Dtype:  arrow.PrimitiveTypes.Uint16,
			Want:   `col[0][col-uint16]: [0 1 2 3 4 5 6 7 8 (null)]`,
		},
		{
			Values: Uint32Gen(),
			Dtype:  arrow.PrimitiveTypes.Uint32,
			Want:   `col[0][col-uint32]: [0 1 2 3 4 5 6 7 8 (null)]`,
		},
		{
			Values: Uint64Gen(),
			Dtype:  arrow.PrimitiveTypes.Uint64,
			Want:   `col[0][col-uint64]: [0 1 2 3 4 5 6 7 8 (null)]`,
		},
		{
			Values: Uint8Gen(),
			Dtype:  arrow.PrimitiveTypes.Uint8,
			Want:   `col[0][col-uint8]: [0 1 2 3 4 5 6 7 8 (null)]`,
		},
	}
}

func BooleanGen() []interface{} {
	vals := make([]interface{}, 9)
	for i := range vals {
		vals[i] = object.NewBoolean(i%2 == 0)
	}
	return vals
}

func Date32Gen() []interface{} {
	vals := make([]interface{}, 9)
	for i := range vals {
		vals[i] = object.NewDate32(arrow.Date32(int32(i)))
	}
	return vals
}

func Date64Gen() []interface{} {
	vals := make([]interface{}, 9)
	for i := range vals {
		vals[i] = object.NewDate64(arrow.Date64(int64(i)))
	}
	return vals
}

func DayTimeIntervalGen() []interface{} {
	vals := make([]interface{}, 9)
	for i := range vals {
		vals[i] = object.NewDayTimeInterval(arrow.DayTimeInterval{Days: int32(i), Milliseconds: int32(i * 2)})
	}
	return vals
}

func Decimal128Gen() []interface{} {
	vals := make([]interface{}, 9)
	for i := range vals {
		vals[i] = object.NewDecimal128(decimal128.New(int64(i), uint64(i)))
	}
	return vals
}

func Duration_sGen() []interface{} {
	vals := make([]interface{}, 9)
	for i := range vals {
		vals[i] = object.NewDuration(arrow.Duration(int64(i)))
	}
	return vals
}

func Duration_msGen() []interface{} {
	vals := make([]interface{}, 9)
	for i := range vals {
		vals[i] = object.NewDuration(arrow.Duration(int64(i)))
	}
	return vals
}

func Duration_usGen() []interface{} {
	vals := make([]interface{}, 9)
	for i := range vals {
		vals[i] = object.NewDuration(arrow.Duration(int64(i)))
	}
	return vals
}

func Duration_nsGen() []interface{} {
	vals := make([]interface{}, 9)
	for i := range vals {
		vals[i] = object.NewDuration(arrow.Duration(int64(i)))
	}
	return vals
}

func Float16Gen() []interface{} {
	vals := make([]interface{}, 9)
	for i := range vals {
		vals[i] = object.NewFloat16(float16.New(float32(i)))
	}
	return vals
}

func Float32Gen() []interface{} {
	vals := make([]interface{}, 9)
	for i := range vals {
		vals[i] = object.NewFloat32(float32(i))
	}
	return vals
}

func Float64Gen() []interface{} {
	vals := make([]interface{}, 9)
	for i := range vals {
		vals[i] = object.NewFloat64(float64(i))
	}
	return vals
}

func Int16Gen() []interface{} {
	vals := make([]interface{}, 9)
	for i := range vals {
		vals[i] = object.NewInt16(int16(i))
	}
	return vals
}

func Int32Gen() []interface{} {
	vals := make([]interface{}, 9)
	for i := range vals {
		vals[i] = object.NewInt32(int32(i))
	}
	return vals
}

func Int64Gen() []interface{} {
	vals := make([]interface{}, 9)
	for i := range vals {
		vals[i] = object.NewInt64(int64(i))
	}
	return vals
}

func Int8Gen() []interface{} {
	vals := make([]interface{}, 9)
	for i := range vals {
		vals[i] = object.NewInt8(int8(i))
	}
	return vals
}

func MonthIntervalGen() []interface{} {
	vals := make([]interface{}, 9)
	for i := range vals {
		vals[i] = object.NewMonthInterval(arrow.MonthInterval(i))
	}
	return vals
}

func StringGen() []interface{} {
	vals := make([]interface{}, 9)
	for i := range vals {
		vals[i] = object.NewString(strconv.Itoa(i))
	}
	return vals
}

func Time32sGen() []interface{} {
	vals := make([]interface{}, 9)
	for i := range vals {
		vals[i] = object.NewTime32(arrow.Time32(int32(i)))
	}
	return vals
}

func Time32msGen() []interface{} {
	vals := make([]interface{}, 9)
	for i := range vals {
		vals[i] = object.NewTime32(arrow.Time32(int32(i)))
	}
	return vals
}

func Time64usGen() []interface{} {
	vals := make([]interface{}, 9)
	for i := range vals {
		vals[i] = object.NewTime64(arrow.Time64(int64(i)))
	}
	return vals
}

func Time64nsGen() []interface{} {
	vals := make([]interface{}, 9)
	for i := range vals {
		vals[i] = object.NewTime64(arrow.Time64(int64(i)))
	}
	return vals
}

func Timestamp_sGen() []interface{} {
	vals := make([]interface{}, 9)
	for i := range vals {
		vals[i] = object.NewTimestamp(arrow.Timestamp(int64(i)))
	}
	return vals
}

func Timestamp_msGen() []interface{} {
	vals := make([]interface{}, 9)
	for i := range vals {
		vals[i] = object.NewTimestamp(arrow.Timestamp(int64(i)))
	}
	return vals
}

func Timestamp_usGen() []interface{} {
	vals := make([]interface{}, 9)
	for i := range vals {
		vals[i] = object.NewTimestamp(arrow.Timestamp(int64(i)))
	}
	return vals
}

func Timestamp_nsGen() []interface{} {
	vals := make([]interface{}, 9)
	for i := range vals {
		vals[i] = object.NewTimestamp(arrow.Timestamp(int64(i)))
	}
	return vals
}

func Uint16Gen() []interface{} {
	vals := make([]interface{}, 9)
	for i := range vals {
		vals[i] = object.NewUint16(uint16(i))
	}
	return vals
}

func Uint32Gen() []interface{} {
	vals := make([]interface{}, 9)
	for i := range vals {
		vals[i] = object.NewUint32(uint32(i))
	}
	return vals
}

func Uint64Gen() []interface{} {
	vals := make([]interface{}, 9)
	for i := range vals {
		vals[i] = object.NewUint64(uint64(i))
	}
	return vals
}

func Uint8Gen() []interface{} {
	vals := make([]interface{}, 9)
	for i := range vals {
		vals[i] = object.NewUint8(uint8(i))
	}
	return vals
}
