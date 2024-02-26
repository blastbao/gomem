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
	"sync/atomic"

	"github.com/apache/arrow/go/arrow"
	"github.com/apache/arrow/go/arrow/bitutil"
	"github.com/apache/arrow/go/arrow/internal/debug"
	"github.com/apache/arrow/go/arrow/memory"
)

type BooleanBuilder struct {
	builder

	data    *memory.Buffer
	rawData []byte
}

func NewBooleanBuilder(mem memory.Allocator) *BooleanBuilder {
	return &BooleanBuilder{
		builder: builder{
			refCount: 1,
			mem:      mem,
		},
	}
}

// Release decreases the reference count by 1.
// When the reference count goes to zero, the memory is freed.
// Release may be called simultaneously from multiple goroutines.
func (b *BooleanBuilder) Release() {
	debug.Assert(atomic.LoadInt64(&b.refCount) > 0, "too many releases")

	if atomic.AddInt64(&b.refCount, -1) == 0 {
		if b.nullBitmap != nil {
			b.nullBitmap.Release()
			b.nullBitmap = nil
		}
		if b.data != nil {
			b.data.Release()
			b.data = nil
			b.rawData = nil
		}
	}
}

func (b *BooleanBuilder) Append(v bool) {
	b.Reserve(1)
	b.UnsafeAppend(v)
}

func (b *BooleanBuilder) AppendByte(v byte) {
	b.Reserve(1)
	b.UnsafeAppend(v != 0)
}

func (b *BooleanBuilder) AppendNull() {
	b.Reserve(1)
	b.UnsafeAppendBoolToBitmap(false)
}

func (b *BooleanBuilder) UnsafeAppend(v bool) {
	// 更新 `nullBitmap` 中第 b.length 个 bit 为 1 ，标识其非空
	bitutil.SetBit(b.nullBitmap.Bytes(), b.length)
	// 更新 data buffer
	if v {
		// 设置 `b.rawData` 中第 b.length 个 bit 为 1 ，标识其为 true
		bitutil.SetBit(b.rawData, b.length)
	} else {
		// 设置 `b.rawData` 中第 b.length 个 bit 为 0 ，标识其为 false
		bitutil.ClearBit(b.rawData, b.length)
	}
	// 更新元素总数
	b.length++
}

func (b *BooleanBuilder) AppendValues(v []bool, valid []bool) {
	if len(v) != len(valid) && len(valid) != 0 {
		panic("len(v) != len(valid) && len(valid) != 0")
	}

	if len(v) == 0 {
		return
	}

	b.Reserve(len(v))
	for i, vv := range v {
		bitutil.SetBitTo(b.rawData, b.length+i, vv)
	}
	b.builder.unsafeAppendBoolsToBitmap(valid, len(v))
}

func (b *BooleanBuilder) init(capacity int) {
	// 初始化底层 builder ，用于管理 nullBitmap 。
	b.builder.init(capacity)
	// 创建 data buffer ，用于存储数据
	b.data = memory.NewResizableBuffer(b.mem)
	// 计算 n 个 boolean 需要占用多少个 bytes
	bytesN := arrow.BooleanTraits.BytesRequired(capacity)
	// 调整 data buffer 的容量，使之能容纳 N 个 bytes
	b.data.Resize(bytesN)
	// 引用底层的 []byte ，加速访问
	b.rawData = b.data.Bytes()
}

// Reserve ensures there is enough space for appending n elements
// by checking the capacity and calling Resize if necessary.
func (b *BooleanBuilder) Reserve(n int) {
	b.builder.reserve(n, b.Resize)
}

// Resize adjusts the space allocated by b to n elements. If n is greater than b.Cap(),
// additional memory will be allocated. If n is smaller, the allocated memory may reduced.
//
// 使足以容纳 n 个元素。
func (b *BooleanBuilder) Resize(n int) {
	if n < minBuilderCapacity {
		n = minBuilderCapacity
	}
	if b.capacity == 0 {
		b.init(n)
	} else {
		// resize `nullBitmap` builder
		b.builder.resize(n, b.init)
		// resize data buffer
		b.data.Resize(arrow.BooleanTraits.BytesRequired(n))
		// 更新引用，因为 resize 操作可能会新建底层 []byte
		b.rawData = b.data.Bytes()
	}
}

// NewArray creates a Boolean array from the memory buffers used by the builder and resets the BooleanBuilder
// so it can be used to build a new array.
func (b *BooleanBuilder) NewArray() Interface {
	return b.NewBooleanArray()
}

// NewBooleanArray creates a Boolean array from the memory buffers used by the builder and resets the BooleanBuilder
// so it can be used to build a new array.
func (b *BooleanBuilder) NewBooleanArray() (a *Boolean) {
	data := b.newData()
	a = NewBooleanData(data)
	data.Release()
	return
}

func (b *BooleanBuilder) newData() *Data {
	// 计算 n 个 boolean 需要占用多少个 bytes
	bytesRequired := arrow.BooleanTraits.BytesRequired(b.length)
	// 缩减 data buffer
	if bytesRequired > 0 && bytesRequired < b.data.Len() {
		// trim buffers
		b.data.Resize(bytesRequired)
	}

	// 基于当前的 b 构造一个 *Data
	res := NewData(
		arrow.FixedWidthTypes.Boolean,
		b.length,
		[]*memory.Buffer{b.nullBitmap, b.data},
		nil,
		b.nulls,
		0,
	)

	// reset `nullBitmap`
	b.reset()
	// reset data buffer
	if b.data != nil {
		b.data.Release()
		b.data = nil
		b.rawData = nil
	}

	return res
}

var (
	_ Builder = (*BooleanBuilder)(nil)
)
