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

package memory

type GoAllocator struct{}

func NewGoAllocator() *GoAllocator { return &GoAllocator{} }


// Allocate 方法用于分配指定大小的内存，并确保内存地址是 64 字节对齐的。
// 如果分配的内存地址不是 64 字节对齐的，会在内存前面添加一些填充以实现对齐。
func (a *GoAllocator) Allocate(size int) []byte {
	buf := make([]byte, size+alignment) // padding for 64-byte alignment
	addr := int(addressOf(buf))
	next := roundUpToMultipleOf64(addr)
	if addr != next {
		shift := next - addr
		return buf[shift : size+shift : size+shift]
	}
	return buf[:size:size]
}

// Reallocate 方法用于重新分配内存，如果新的大小与原来的大小相同，则直接返回原来的内存。
// 如果不同，会重新分配内存并将原来的数据拷贝到新的内存中。
func (a *GoAllocator) Reallocate(size int, b []byte) []byte {
	if size == len(b) {
		return b
	}

	newBuf := a.Allocate(size)
	copy(newBuf, b)
	return newBuf
}

// Free 方法用于释放内存。
func (a *GoAllocator) Free(b []byte) {}

var (
	_ Allocator = (*GoAllocator)(nil)
)
