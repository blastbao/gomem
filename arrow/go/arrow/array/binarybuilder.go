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

package array

import (
	"math"
	"sync/atomic"

	"github.com/apache/arrow/go/arrow"
	"github.com/apache/arrow/go/arrow/internal/debug"
	"github.com/apache/arrow/go/arrow/memory"
)

const (
	binaryArrayMaximumCapacity = math.MaxInt32
)

// A BinaryBuilder is used to build a Binary array using the Append methods.
type BinaryBuilder struct {
	builder // []bit ，存储第 i 个 value 是否为 null ，底层是 []byte

	dtype   arrow.BinaryDataType // 数据类型
	offsets *int32BufferBuilder  // []int32 ，存储第 i 个 value 的偏移量
	values  *byteBufferBuilder   // []byte ，以铺平的方式存储 values
}

func NewBinaryBuilder(mem memory.Allocator, dtype arrow.BinaryDataType) *BinaryBuilder {
	b := &BinaryBuilder{
		builder: builder{refCount: 1, mem: mem},
		dtype:   dtype,
		offsets: newInt32BufferBuilder(mem),
		values:  newByteBufferBuilder(mem),
	}
	return b
}

// Release decreases the reference count by 1.
// When the reference count goes to zero, the memory is freed.
// Release may be called simultaneously from multiple goroutines.
func (b *BinaryBuilder) Release() {
	debug.Assert(atomic.LoadInt64(&b.refCount) > 0, "too many releases")
	if atomic.AddInt64(&b.refCount, -1) == 0 {
		if b.nullBitmap != nil {
			b.nullBitmap.Release()
			b.nullBitmap = nil
		}
		if b.offsets != nil {
			b.offsets.Release()
			b.offsets = nil
		}
		if b.values != nil {
			b.values.Release()
			b.values = nil
		}
	}
}

func (b *BinaryBuilder) Append(v []byte) {
	b.Reserve(1)
	// 添加到 `offsets` ，保存当前 v 的 offset 到 offsets 中
	b.appendNextOffset()
	// 添加到 `values` ，保存当前 v
	b.values.Append(v)
	// 添加到 `nullBitmap`
	b.UnsafeAppendBoolToBitmap(true)
}

func (b *BinaryBuilder) AppendString(v string) {
	b.Append([]byte(v))
}

func (b *BinaryBuilder) AppendNull() {
	b.Reserve(1)
	// 添加到 `offsets` ，值得注意的是，即使是 null 元素也要为其保存一个无效的 offset ，但是 value 是不需要的。
	b.appendNextOffset()
	// 添加到 `nullBitmap`
	b.UnsafeAppendBoolToBitmap(false)
}

// AppendValues will append the values in the v slice. The valid slice determines which values
// in v are valid (not null). The valid slice must either be empty or be equal in length to v. If empty,
// all values in v are appended and considered valid.
func (b *BinaryBuilder) AppendValues(v [][]byte, valid []bool) {
	if len(v) != len(valid) && len(valid) != 0 {
		panic("len(v) != len(valid) && len(valid) != 0")
	}

	if len(v) == 0 {
		return
	}

	b.Reserve(len(v))
	for _, vv := range v {
		b.appendNextOffset()
		b.values.Append(vv)
	}

	b.builder.unsafeAppendBoolsToBitmap(valid, len(v))
}

// AppendStringValues will append the values in the v slice. The valid slice determines which values
// in v are valid (not null). The valid slice must either be empty or be equal in length to v. If empty,
// all values in v are appended and considered valid.
func (b *BinaryBuilder) AppendStringValues(v []string, valid []bool) {
	if len(v) != len(valid) && len(valid) != 0 {
		panic("len(v) != len(valid) && len(valid) != 0")
	}
	if len(v) == 0 {
		return
	}
	b.Reserve(len(v))
	for _, vv := range v {
		b.appendNextOffset()
		b.values.Append([]byte(vv))
	}
	b.builder.unsafeAppendBoolsToBitmap(valid, len(v))
}

func (b *BinaryBuilder) Value(i int) []byte {
	// 取第 i 个 value 的 offset
	offsets := b.offsets.Values()
	start := int(offsets[i])
	// 取第 i + 1 个 value 的 offset
	var end int
	if i == (b.length - 1) {
		end = b.values.Len()
	} else {
		end = int(offsets[i+1])
	}
	// 返回 [off(i), off(i+1)) 之间的 []bytes
	return b.values.Bytes()[start:end]
}

func (b *BinaryBuilder) init(capacity int) {
	b.builder.init(capacity)
	b.offsets.resize((capacity + 1) * arrow.Int32SizeBytes)
}

// DataLen returns the number of bytes in the data array.
func (b *BinaryBuilder) DataLen() int { return b.values.length }

// DataCap returns the total number of bytes that can be stored
// without allocating additional memory.
func (b *BinaryBuilder) DataCap() int { return b.values.capacity }

// Reserve ensures there is enough space for appending n elements
// by checking the capacity and calling Resize if necessary.
func (b *BinaryBuilder) Reserve(n int) {
	b.builder.reserve(n, b.Resize)
}

// ReserveData ensures there is enough space for appending n bytes
// by checking the capacity and resizing the data buffer if necessary.
func (b *BinaryBuilder) ReserveData(n int) {
	if b.values.capacity < b.values.length+n {
		b.values.resize(b.values.Len() + n)
	}
}

// Resize adjusts the space allocated by b to n elements. If n is greater than b.Cap(),
// additional memory will be allocated. If n is smaller, the allocated memory may be reduced.
func (b *BinaryBuilder) Resize(n int) {
	b.offsets.resize((n + 1) * arrow.Int32SizeBytes)
	b.builder.resize(n, b.init)
}

// NewArray creates a Binary array from the memory buffers used by the builder and resets the BinaryBuilder
// so it can be used to build a new array.
func (b *BinaryBuilder) NewArray() Interface {
	return b.NewBinaryArray()
}

// NewBinaryArray creates a Binary array from the memory buffers used by the builder and resets the BinaryBuilder
// so it can be used to build a new array.
func (b *BinaryBuilder) NewBinaryArray() (a *Binary) {
	data := b.newData()
	a = NewBinaryData(data)
	data.Release()
	return
}

func (b *BinaryBuilder) newData() (data *Data) {
	b.appendNextOffset()

	offsets := b.offsets.Finish() // 取底层数组
	values := b.values.Finish()   // 取底层数组

	data = NewData(
		b.dtype,
		b.length,
		[]*memory.Buffer{b.nullBitmap, offsets, values},
		nil,
		b.nulls,
		0,
	)

	if offsets != nil {
		offsets.Release()
	}
	if values != nil {
		values.Release()
	}
	b.builder.reset()
	return
}

func (b *BinaryBuilder) appendNextOffset() {
	// 取当前 values 的字节总数，作为新 value 的起始 offset
	numBytes := b.values.Len()
	// TODO(sgc): check binaryArrayMaximumCapacity?
	// 把当前 offset 存入 offsets 中
	b.offsets.AppendValue(int32(numBytes))
}

var (
	_ Builder = (*BinaryBuilder)(nil)
)
