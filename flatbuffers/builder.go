package flatbuffers

// FlatBuffers 中，minalign（也称为对齐因子，表示内存对齐）用于指定表中字段的内存对齐方式。
// FlatBuffers 使用自定义二进制格式来表示数据，通过将数据结构组织成表，可以在不进行任何解析的情况下进行直接访问。
// minalign 用于确保字段在内存中的正确对齐，以提高访问效率。
//
// 由于计算机需要将数据从内存中读取到寄存器中进行处理，因此数据在内存中的存储位置对性能至关重要。
// 对齐方式可以确保多字节字段的地址始终位于其大小的倍数上，以便有效地读取和写入数据。
//
// minalign 指定了字段的最小对齐方式，以字节为单位。
// 例如，minalign=1 表示字段可以在任何字节边界上对齐，而 minalign=4 表示字段需要在4字节边界上对齐。
// 较小的对齐方式可以节省内存空间，但会增加访问成本，因为计算机需要执行额外的计算来访问不对齐的字段。
//
// 总之，minalign 用于控制 FlatBuffers 中字段的内存对齐方式，以平衡内存使用和访问效率。

// 例如，如果一个属性的大小为 3 字节，并且 minalign 设置为 4，那么该属性应该使用 4 字节的存储空间，并且要在前面添加一个额外的字节来完成对齐。
// 这样做可以提高访问速度和存储空间的使用效率。
//
// 设置较小的 minalign 可以减少内存消耗, 但可能导致性能下降。
// 设置较大的 minalign 可以提高访问速度, 但同时也会增加内存占用。
//
// 一般情况下，minalign 保持默认值 4 就可以，只有在内存很宝贵或者性能极为关键时，才需要考虑调整这个值，根据实际需求选择合适的值,平衡内存和性能的 tradeoff 。

// vtable 的元素都是 VOffsetT 类型，它是 uint16 。
// 第一个元素是 vtable 的大小（以字节为单位），包括自身。
// 第二个元素是对象的大小，以字节为单位（包括 vtable 偏移量）。这个大小可以用于流式传输，知道要读取多少字节才能访问对象的所有内联 inline 字段。
// 第三个元素开始是 N 个偏移量，其中 N 是编译构建此 buffer 的代码编译时（因此，表的大小为 N + 2）时在 schema 中声明的字段数量(包括 deprecated 字段)。每个以 SizeVOffsetT 字节为宽度。

// Builder is a state machine for creating FlatBuffer objects.
// Use a Builder to construct object(s) starting from leaf nodes.
//
// A Builder constructs byte buffers in a last-first manner for simplicity and
// performance.
type Builder struct {
	// `Bytes` gives raw access to the buffer. Most users will want to use
	// FinishedBytes() instead.
	Bytes []byte

	minalign  int
	vtable    []UOffsetT // 存储当前正在构建的对象的 VTable 。当使用 Builder 构建一个对象时，vtable 会被填充并最终添加到 vtables 中。这样，在序列化时，可以通过索引来引用正确的 VTable 。
	objectEnd UOffsetT
	vtables   []UOffsetT // 存储 FlatBuffers 对象中的所有 VTables 。每个 VTable 都表示一个对象的字段布局和访问信息。
	head      UOffsetT
	nested    bool
	finished  bool
}

const fileIdentifierLength = 4

// NewBuilder initializes a Builder of size `initial_size`.
// The internal buffer is grown as needed.
//
// 创建 `Builder` 实例, 其大小会根据需要自动增长，不必担心空间不够
func NewBuilder(initialSize int) *Builder {
	if initialSize <= 0 {
		initialSize = 0
	}

	b := &Builder{}
	b.Bytes = make([]byte, initialSize)
	b.head = UOffsetT(initialSize)
	b.minalign = 1
	b.vtables = make([]UOffsetT, 0, 16) // sensible default capacity

	return b
}

// Reset truncates the underlying Builder buffer, facilitating alloc-free
// reuse of a Builder. It also resets bookkeeping data.
func (b *Builder) Reset() {
	if b.Bytes != nil {
		b.Bytes = b.Bytes[:cap(b.Bytes)]
	}

	if b.vtables != nil {
		b.vtables = b.vtables[:0]
	}

	if b.vtable != nil {
		b.vtable = b.vtable[:0]
	}

	b.head = UOffsetT(len(b.Bytes))
	b.minalign = 1
	b.nested = false
	b.finished = false
}

// FinishedBytes returns a pointer to the written data in the byte buffer.
// Panics if the builder is not in a finished state (which is caused by calling
// `Finish()`).
func (b *Builder) FinishedBytes() []byte {
	b.assertFinished()
	return b.Bytes[b.Head():]
}

// StartObject 作用是初始化一个新对象的写入过程，包括设置嵌套状态、准备 vtable ，并记录对象的结束偏移量。
// StartObject 的参数 numfields 表示对象中的字段（fields）数量；
//
// vtable 是 FlatBuffers 中用于描述对象布局的虚拟表，用于支持对象的跨平台访问。
//
// 首先，判断当前 vtable 的容量是否小于 numfields 或者是否为 nil ：
//	如果是，则重新分配容量为 numfields 的 vtable；
//	如果否，则保留 vtable 的容量，但将长度限制为 numfields（即截断 vtable 的长度为 numfields ）；然后，将 vtable 中的每个元素设置为 0 ，表示当前字段的偏移量。
//
// 最后，记录当前对象的结束偏移量
//
// StartObject initializes bookkeeping for writing a new object.
func (b *Builder) StartObject(numfields int) {
	b.assertNotNested() // 确保当前没有嵌套写入操作
	b.nested = true     // 设置当前处于嵌套写入操作

	// use 32-bit offsets so that arithmetic doesn't overflow.
	if cap(b.vtable) < numfields || b.vtable == nil {
		b.vtable = make([]UOffsetT, numfields)
	} else {
		b.vtable = b.vtable[:numfields]
		for i := 0; i < len(b.vtable); i++ {
			b.vtable[i] = 0
		}
	}

	b.objectEnd = b.Offset()
}

//
//
//
//

// WriteVtable serializes the vtable for the current object, if applicable.
//
// Before writing out the vtable, this checks pre-existing vtables for equality
// to this one. If an equal vtable is found, point the object to the existing
// vtable and return.
//
// Because vtable values are sensitive to alignment of object data, not all
// logically-equal vtables will be deduplicated.
//
// A vtable has the following format:
//   <VOffsetT: size of the vtable in bytes, including this value>
//   <VOffsetT: size of the object in bytes, including the vtable offset>
//   <VOffsetT: offset for a field> * N, where N is the number of fields in
//	        the schema for this type. Includes deprecated fields.
// Thus, a vtable is made of 2 + N elements, each SizeVOffsetT bytes wide.
//
// An object has the following format:
//   <SOffsetT: offset to this object's vtable (may be negative)>
//   <byte: data>+
func (b *Builder) WriteVtable() (n UOffsetT) {
	// Prepend a zero scalar to the object. Later in this function we'll
	// write an offset here that points to the object's vtable:
	//
	// Object 的开头是 4B 的 SOffsetT 偏移量，指向关联的 vtable 。
	// 这里属于预写入(0值)，后面在写完 vtable 后会覆盖写入真正的偏移量。
	b.PrependSOffsetT(0)

	objectOffset := b.Offset() // 当前 object 的 offset ，
	existingVtable := UOffsetT(0)

	// Trim vtable of trailing zeroes.
	//
	// 去掉末尾 0
	i := len(b.vtable) - 1
	for ; i >= 0 && b.vtable[i] == 0; i-- {
	}
	b.vtable = b.vtable[:i+1]

	// Search backwards through existing vtables, because similar vtables
	// are likely to have been recently appended. See
	// BenchmarkVtableDeduplication for a case in which this heuristic
	// saves about 30% of the time used in writing objects with duplicate
	// tables.
	//
	// 从 vtables 中逆向搜索已经存储过的 vtable ，如果存在相同的且已经存储过的 vtable ，直接找到它，索引指向它即可；
	// 可以查看 BenchmarkVtableDeduplication 的测试结果，通过索引指向相同的 vtable，而不是新建一个，这种做法可以提高 30% 性能；
	for i := len(b.vtables) - 1; i >= 0; i-- {
		// Find the other vtable, which is associated with `i`:
		// 选定一个 vtable
		vt2Offset := b.vtables[i]
		vt2Start := len(b.Bytes) - int(vt2Offset)
		vt2Len := GetVOffsetT(b.Bytes[vt2Start:])

		metadata := VtableMetadataFields * SizeVOffsetT
		vt2End := vt2Start + int(vt2Len)
		vt2 := b.Bytes[vt2Start+metadata : vt2End]

		// Compare the other vtable to the one under consideration.
		// If they are equal, store the offset and break:
		if vtableEqual(b.vtable, objectOffset, vt2) {
			existingVtable = vt2Offset
			break
		}
	}

	if existingVtable == 0 {
		// Did not find a vtable, so write this one to the buffer.

		// Write out the current vtable in reverse , because
		// serialization occurs in last-first order:
		for i := len(b.vtable) - 1; i >= 0; i-- {
			var off UOffsetT
			if b.vtable[i] != 0 {
				// Forward reference to field;
				// use 32bit number to assert no overflow:
				off = objectOffset - b.vtable[i]
			}

			b.PrependVOffsetT(VOffsetT(off))
		}

		// The two metadata fields are written last.

		// First, store the object bytesize:
		objectSize := objectOffset - b.objectEnd
		b.PrependVOffsetT(VOffsetT(objectSize))

		// Second, store the vtable bytesize:
		vBytes := (len(b.vtable) + VtableMetadataFields) * SizeVOffsetT
		b.PrependVOffsetT(VOffsetT(vBytes))

		// Next, write the offset to the new vtable in the
		// already-allocated SOffsetT at the beginning of this object:
		objectStart := SOffsetT(len(b.Bytes)) - SOffsetT(objectOffset)
		WriteSOffsetT(b.Bytes[objectStart:], SOffsetT(b.Offset())-SOffsetT(objectOffset))

		// Finally, store this vtable in memory for future
		// deduplication:
		b.vtables = append(b.vtables, b.Offset())
	} else {
		// Found a duplicate vtable.

		objectStart := SOffsetT(len(b.Bytes)) - SOffsetT(objectOffset)
		b.head = UOffsetT(objectStart)

		// Write the offset to the found vtable in the
		// already-allocated SOffsetT at the beginning of this object:
		WriteSOffsetT(b.Bytes[b.head:], SOffsetT(existingVtable)-SOffsetT(objectOffset))
	}

	b.vtable = b.vtable[:0]
	return objectOffset
}

// EndObject writes data necessary to finish object construction.
func (b *Builder) EndObject() UOffsetT {
	b.assertNested()
	n := b.WriteVtable()
	b.nested = false
	return n
}

// Doubles the size of the byteslice, and copies the old data towards the
// end of the new byteslice (since we build the buffer backwards).
//
// 扩容到原来 2 倍的大小，旧数据会被 copy 到新扩容以后数组的末尾，因为 build buffer 是从后往前 build 的，旧数据在后边。
func (b *Builder) growByteBuffer() {
	if (int64(len(b.Bytes)) & int64(0xC0000000)) != 0 {
		panic("cannot grow buffer beyond 2 gigabytes")
	}
	newLen := len(b.Bytes) * 2
	if newLen == 0 {
		newLen = 1
	}

	if cap(b.Bytes) >= newLen {
		b.Bytes = b.Bytes[:newLen]
	} else {
		extension := make([]byte, newLen-len(b.Bytes))
		b.Bytes = append(b.Bytes, extension...)
	}

	middle := newLen / 2
	copy(b.Bytes[middle:], b.Bytes[:middle])
}

// Head gives the start of useful data in the underlying byte buffer.
// Note: unlike other functions, this value is interpreted as from the left.
//
// Head 返回底层 buffer 中有用数据的起始位置。
func (b *Builder) Head() UOffsetT {
	return b.head
}

// Offset relative to the end of the buffer.
// 反映当前 bytebuffer 中存储数据的长度，也能表明当前对象相对于 bytebuffer 结尾的偏移
func (b *Builder) Offset() UOffsetT {
	return UOffsetT(len(b.Bytes)) - b.head
}

// Pad places zeros at the current offset.
func (b *Builder) Pad(n int) {
	for i := 0; i < n; i++ {
		b.PlaceByte(0)
	}
}

// Prep 用于预留空间，该空间足以容纳 size + additionalBytes 且按照 size 对齐。
//
// `Prep` 该函数接受两个参数，`size`表示要写入的元素的大小，`additionalBytes` 表示已经写入的额外字节数量。
// 如果只需要进行对齐操作而不需要写入额外字节，`additionalBytes` 参数将为 0 。
//
// 函数执行的逻辑如下：
//	1. 首先，记录当前已对齐的最大大小，即若要对齐`size`，所需的大小。
//	2. 计算在已经写入了`additionalBytes`字节后，使`size`能够正常对齐所需的对齐大小。具体的计算方式是找到一个对齐大小，使得对齐大小加上已经写入的字节数后，加1取反再取与`size - 1`的与操作结果为0。这样可以确保对齐大小加上已经写入的字节数后，再向下取整能够被`size`整除。
//	3. 如果缓冲区的头指针`head`加上对齐大小、`size`和`additionalBytes`后的结果小于等于缓冲区的长度，则表示缓冲区的空间不足以写入新的数据，需要进行动态扩容操作。
//	4. 调用`growByteBuffer`函数进行缓冲区扩容，并更新`head`指针为扩容后的长度减去旧的缓冲区长度。
//	5. 调用`Pad`函数进行对齐操作，将缓冲区的头指针移动到对齐的位置。
//
// 这个函数的作用是为要写入的元素做准备，包括计算对齐大小、扩容缓冲区和进行对齐操作，以确保数据写入时能够满足对齐要求，并且缓冲区具有足够的空间来存储数据。
// 这样可以避免数据的错位和内存越界访问。

// 该段代码是 FlatBuffers 中的 `Builder` 结构的 `Prep` 方法。它用于在写入 `additionalBytes` 字节后，准备写入大小为 `size` 的元素。
//
// `Prep` 方法的功能是为要写入的元素做对齐准备。
// 当写入某个元素时，可能需要对齐以确保各个部分正确对齐。例如，如果要写入一个字符串，在字符串数据之前需要对齐整数长度字段的对齐方式（使用 `SizeInt32`）。
// 这里的 `additionalBytes` 参数表示已经写入的额外字节数。如果只需要对齐，则 `additionalBytes` 为 0。
//
// 该方法首先比较 `size` 和 `minalign`（用于追踪之前对齐的最大值）的大小，将较大的值赋给 `minalign`。
// 这样，`minalign` 就会记录下要写入元素的最大对齐值。
//
// 接下来，根据已写入的字节数和要写入的元素大小计算所需的对齐大小。
//`alignSize` 的计算方式是，取反后 `len(b.Bytes) - int(b.Head()) + additionalBytes`，然后加1，再与 `size - 1` 进行按位与操作。
//
// 然后，检查是否需要重新分配缓冲区以适应写入元素所需的空间。
// 通过比较 `b.head` 和 `alignSize + size + additionalBytes` 的大小，判断当前缓冲区的剩余空间是否足够。
// 如果不足，则调用 `growByteBuffer` 方法扩展缓冲区。
//
// 最后，调用 `Pad` 方法，根据对齐大小 `alignSize` 进行填充，确保写入位置合适的对齐。
//
// 总结来说，`Prep` 方法的功能是根据给定大小和已写入的额外字节数，在适当的位置进行对齐预处理，并根据需要对缓冲区进行重新分配以确保写入元素的空间足够。

// [重要]
// Prep() 函数的第一个入参是 size，这里的 size 是字节单位，有多少个字节大小，这里的 size 就是多少。
// 例如 SizeUint8 = 1、SizeUint16 = 2、SizeUint32 = 4、SizeUint64 = 8。其他类型以此类推。
// 比较特殊的 3 个 offset，大小也是固定的，SOffsetT int32，它的 size = 4；UOffsetT uint32，它的 size = 4；VOffsetT uint16，它的 size = 2。
//
// Prep() 函数能确保分配 additional_bytes 个字节之后的 offset 是 size 的整数倍，过程中可能需要对齐填充，如果有需要还会 Reallocate buffer 。

// Prep prepares to write an element of `size` after `additional_bytes`
// have been written, e.g. if you write a string, you need to align such
// the int length field is aligned to SizeInt32, and the string data follows it
// directly.
// If all you need to do is align, `additionalBytes` will be 0.
func (b *Builder) Prep(size, additionalBytes int) {
	// Track the biggest thing we've ever aligned to.
	if size > b.minalign {
		b.minalign = size
	}

	// Find the amount of alignment needed such that `size` is properly
	// aligned after `additionalBytes`:
	alignSize := (^(len(b.Bytes) - int(b.Head()) + additionalBytes)) + 1
	alignSize &= (size - 1)

	// Reallocate the buffer if needed:
	//
	// b.head 代表当前 buffer 的剩余空间，如果少于待添加的数据量，需要扩容；
	for int(b.head) <= alignSize+size+additionalBytes {
		oldBufSize := len(b.Bytes)
		b.growByteBuffer()
		// 扩容的 bytes size = len(b.Bytes) - oldBufSize ，把 b.head 执行向后移动，腾出空间
		b.head += UOffsetT(len(b.Bytes) - oldBufSize)
	}

	// 填入用于对齐的 0 字节
	b.Pad(alignSize)
}

// PrependSOffsetT prepends an SOffsetT, relative to where it will be written.
func (b *Builder) PrependSOffsetT(off SOffsetT) {
	b.Prep(SizeSOffsetT, 0) // Ensure alignment is already done.
	if !(UOffsetT(off) <= b.Offset()) {
		panic("unreachable: off <= b.Offset()")
	}
	off2 := SOffsetT(b.Offset()) - off + SOffsetT(SizeSOffsetT)
	b.PlaceSOffsetT(off2)
}

// 给定一个 offset1，通过当前 offset 间接定位到 offset1 的方式是，在 offset 处存储 |offset - offset1| 的偏移量。
// 注意，因为偏移量本身大小 4B ，所以需要额外加 4B 的相对偏移。

// PrependUOffsetT prepends an UOffsetT, relative to where it will be written.
func (b *Builder) PrependUOffsetT(off UOffsetT) {
	// 确保 buffer 足以容纳一个 UOffsetT 类型对象
	b.Prep(SizeUOffsetT, 0) // Ensure alignment is already done.
	// 合法性校验
	if !(off <= b.Offset()) {
		panic("unreachable: off <= b.Offset()")
	}
	// 计算相对偏移
	off2 := b.Offset() - off + UOffsetT(SizeUOffsetT)
	// 将相对偏移存入 b.Bytes[] 中
	b.PlaceUOffsetT(off2)
}

// StartVector initializes bookkeeping for writing a new vector.
//
// A vector has the following format:
//   <UOffsetT: number of elements in this vector>
//   <T: data>+, where T is the type of elements of this vector.
func (b *Builder) StartVector(elemSize, numElems, alignment int) UOffsetT {
	b.assertNotNested()
	b.nested = true

	b.Prep(SizeUint32, elemSize*numElems)
	b.Prep(alignment, elemSize*numElems) // Just in case alignment > int.
	return b.Offset()
}

// EndVector writes data necessary to finish vector construction.
// 结束 vector 的创建，同时会将此 vector 的长度信息写入，同时返回当前 vector 相对于 ByteBuffer 结尾的偏移
func (b *Builder) EndVector(vectorNumElems int) UOffsetT {
	b.assertNested()

	// we already made space for this, so write without PrependUint32
	b.PlaceUOffsetT(UOffsetT(vectorNumElems)) // 保存 vector 的成员个数，不是存储空间长度

	b.nested = false
	return b.Offset() // 表明当前对象相对于 bytebuffer 结尾的偏移
}

// FlatBuffers 在实现字符串写入的时候将字符串的编码数组当做了一维的 vector 来实现，
// startVector 函数是写入前的初始化，并且在写入编码数组之前我们又看到了先将 space 往前移动数组长度的距离，然后再写入；
// 写入完成后调用 endVector 进行收尾，endVector 再将 vector 的成员数量，在这里就是字符串数组的长度写入；
// 然后调用 offset 返回写入数据的起点。

// CreateString writes a null-terminated string as a vector.
// 创建字符串，从左到右依次是[字符串长度，字符串数据，结尾"0"]
func (b *Builder) CreateString(s string) UOffsetT {
	b.assertNotNested()
	b.nested = true

	b.Prep(int(SizeUOffsetT), (len(s)+1)*SizeByte)
	b.PlaceByte(0) // string 的末尾是 null 结束符，要加一个字节的 0

	l := UOffsetT(len(s)) // string 的字符长度

	b.head -= l                       // 移动 head ，腾出 l 个字符的长度
	copy(b.Bytes[b.head:b.head+l], s) // 把字符串 s 复制到相应的 offset 中

	return b.EndVector(len(s)) // 把字符串 s 的长度（不含末尾 0 ）写入 b.Bytes[b.Offset():] 中，返回 b.Offset() 。
}

// CreateByteString writes a byte slice as a string (null-terminated).
func (b *Builder) CreateByteString(s []byte) UOffsetT {
	b.assertNotNested()
	b.nested = true

	b.Prep(int(SizeUOffsetT), (len(s)+1)*SizeByte)
	b.PlaceByte(0)

	l := UOffsetT(len(s))

	b.head -= l
	copy(b.Bytes[b.head:b.head+l], s)

	return b.EndVector(len(s))
}

// CreateByteVector writes a ubyte vector
func (b *Builder) CreateByteVector(v []byte) UOffsetT {
	b.assertNotNested()
	b.nested = true

	b.Prep(int(SizeUOffsetT), len(v)*SizeByte)

	l := UOffsetT(len(v))

	b.head -= l
	copy(b.Bytes[b.head:b.head+l], v)

	return b.EndVector(len(v))
}

func (b *Builder) assertNested() {
	// If you get this assert, you're in an object while trying to write
	// data that belongs outside of an object.
	// To fix this, write non-inline data (like vectors) before creating
	// objects.
	if !b.nested {
		panic("Incorrect creation order: must be inside object.")
	}
}

func (b *Builder) assertNotNested() {
	// If you hit this, you're trying to construct a Table/Vector/String
	// during the construction of its parent table (between the MyTableBuilder
	// and builder.Finish()).
	// Move the creation of these sub-objects to above the MyTableBuilder to
	// not get this assert.
	// Ignoring this assert may appear to work in simple cases, but the reason
	// it is here is that storing objects in-line may cause vtable offsets
	// to not fit anymore. It also leads to vtable duplication.
	if b.nested {
		panic("Incorrect creation order: object must not be nested.")
	}
}

func (b *Builder) assertFinished() {
	// If you get this assert, you're attempting to get access a buffer
	// which hasn't been finished yet. Be sure to call builder.Finish()
	// with your root table.
	// If you really need to access an unfinished buffer, use the Bytes
	// buffer directly.
	if !b.finished {
		panic("Incorrect use of FinishedBytes(): must call 'Finish' first.")
	}
}

// PrependBoolSlot prepends a bool onto the object at vtable slot `o`.
// If value `x` equals default `d`, then the slot will be set to zero and no
// other data will be written.
func (b *Builder) PrependBoolSlot(o int, x, d bool) {
	val := byte(0)
	if x {
		val = 1
	}
	def := byte(0)
	if d {
		def = 1
	}
	b.PrependByteSlot(o, val, def)
}

// PrependByteSlot prepends a byte onto the object at vtable slot `o`.
// If value `x` equals default `d`, then the slot will be set to zero and no
// other data will be written.
func (b *Builder) PrependByteSlot(o int, x, d byte) {
	if x != d {
		b.PrependByte(x)
		b.Slot(o)
	}
}

// PrependUint8Slot prepends a uint8 onto the object at vtable slot `o`.
// If value `x` equals default `d`, then the slot will be set to zero and no
// other data will be written.
func (b *Builder) PrependUint8Slot(o int, x, d uint8) {
	if x != d {
		b.PrependUint8(x)
		b.Slot(o)
	}
}

// PrependUint16Slot prepends a uint16 onto the object at vtable slot `o`.
// If value `x` equals default `d`, then the slot will be set to zero and no
// other data will be written.
func (b *Builder) PrependUint16Slot(o int, x, d uint16) {
	if x != d {
		b.PrependUint16(x)
		b.Slot(o)
	}
}

// PrependUint32Slot prepends a uint32 onto the object at vtable slot `o`.
// If value `x` equals default `d`, then the slot will be set to zero and no
// other data will be written.
func (b *Builder) PrependUint32Slot(o int, x, d uint32) {
	if x != d {
		b.PrependUint32(x)
		b.Slot(o)
	}
}

// PrependUint64Slot prepends a uint64 onto the object at vtable slot `o`.
// If value `x` equals default `d`, then the slot will be set to zero and no
// other data will be written.
func (b *Builder) PrependUint64Slot(o int, x, d uint64) {
	if x != d {
		b.PrependUint64(x)
		b.Slot(o)
	}
}

// PrependInt8Slot prepends a int8 onto the object at vtable slot `o`.
// If value `x` equals default `d`, then the slot will be set to zero and no
// other data will be written.
func (b *Builder) PrependInt8Slot(o int, x, d int8) {
	if x != d {
		b.PrependInt8(x)
		b.Slot(o)
	}
}

// PrependInt16Slot 将一个 `int16` 类型的值添加到对象的 vtable 槽位 `o` 前面。
//
// PrependInt16Slot 接受三个参数：
//	- `o` 表示 vtable 的槽位
//  - `x` 表示要添加的值
//  - `d` 表示默认值
//
// 函数执行的逻辑如下：
//	1. 首先，判断 `x` 是否等于默认值 `d` 。
//	2. 如果 `x` 不等于 `d` ，则调用 `PrependInt16` 函数将 `x` 插入到对象数据缓冲区的前面。
//	3. 调用 `Slot(o)` 函数将槽位 `o` 设置为当前缓冲区中起始偏移量，从而记录这个字段在缓冲区中的位置。
//
// 这个函数的作用是将一个 `int16` 类型的字段添加到对象中，如果字段的值和默认值不相等，则将字段的值写入缓冲区，并设置对应的 vtable 槽位。
// 如果字段的值和默认值相等，则不写入数据，而是将对应的 vtable 槽位设置为0。这样可以节省空间，避免写入不必要的数据。

// PrependInt16Slot prepends a int16 onto the object at vtable slot `o`.
// If value `x` equals default `d`, then the slot will be set to zero and no
// other data will be written.
func (b *Builder) PrependInt16Slot(o int, x, d int16) {
	if x != d {
		// 把 x 插入到 b.Bytes[b.Offset():] 中；
		b.PrependInt16(x)
		// 把 b.Offset() 保存到 b.vtable[o] 上，即在 vtable 中保存第 o 字段的偏移量；
		b.Slot(o)
	}
}

// PrependInt32Slot prepends a int32 onto the object at vtable slot `o`.
// If value `x` equals default `d`, then the slot will be set to zero and no
// other data will be written.
func (b *Builder) PrependInt32Slot(o int, x, d int32) {
	if x != d {
		b.PrependInt32(x)
		b.Slot(o)
	}
}

// PrependInt64Slot prepends a int64 onto the object at vtable slot `o`.
// If value `x` equals default `d`, then the slot will be set to zero and no
// other data will be written.
func (b *Builder) PrependInt64Slot(o int, x, d int64) {
	if x != d {
		b.PrependInt64(x)
		b.Slot(o)
	}
}

// PrependFloat32Slot prepends a float32 onto the object at vtable slot `o`.
// If value `x` equals default `d`, then the slot will be set to zero and no
// other data will be written.
func (b *Builder) PrependFloat32Slot(o int, x, d float32) {
	if x != d {
		b.PrependFloat32(x)
		b.Slot(o)
	}
}

// PrependFloat64Slot prepends a float64 onto the object at vtable slot `o`.
// If value `x` equals default `d`, then the slot will be set to zero and no
// other data will be written.
func (b *Builder) PrependFloat64Slot(o int, x, d float64) {
	if x != d {
		b.PrependFloat64(x)
		b.Slot(o)
	}
}

// PrependUOffsetTSlot prepends an UOffsetT onto the object at vtable slot `o`.
// If value `x` equals default `d`, then the slot will be set to zero and no
// other data will be written.
func (b *Builder) PrependUOffsetTSlot(o int, x, d UOffsetT) {
	if x != d {
		b.PrependUOffsetT(x)
		b.Slot(o)
	}
}

// PrependStructSlot prepends a struct onto the object at vtable slot `o`.
// Structs are stored inline, so nothing additional is being added.
// In generated code, `d` is always 0.
func (b *Builder) PrependStructSlot(voffset int, x, d UOffsetT) {
	if x != d {
		b.assertNested()
		if x != b.Offset() {
			panic("inline data write outside of object")
		}
		b.Slot(voffset)
	}
}

// vtable 是用于存储和索引对象字段偏移量的表，使用 `Slot` 函数可以将字段的偏移量写入到 vtable 中，以便后续能够正确地访问和读取字段值。

// Slot sets the vtable key `voffset` to the current location in the buffer.
func (b *Builder) Slot(slotnum int) {
	b.vtable[slotnum] = UOffsetT(b.Offset())
}

// FinishWithFileIdentifier finalizes a buffer, pointing to the given `rootTable`.
// as well as applys a file identifier
func (b *Builder) FinishWithFileIdentifier(rootTable UOffsetT, fid []byte) {
	if fid == nil || len(fid) != fileIdentifierLength {
		panic("incorrect file identifier length")
	}
	// In order to add a file identifier to the flatbuffer message, we need
	// to prepare an alignment and file identifier length
	b.Prep(b.minalign, SizeInt32+fileIdentifierLength)
	for i := fileIdentifierLength - 1; i >= 0; i-- {
		// place the file identifier
		b.PlaceByte(fid[i])
	}
	// finish
	b.Finish(rootTable)
}

// Finish finalizes a buffer, pointing to the given `rootTable`.
func (b *Builder) Finish(rootTable UOffsetT) {
	b.assertNotNested()
	b.Prep(b.minalign, SizeUOffsetT)
	b.PrependUOffsetT(rootTable)
	b.finished = true
}

// 这段代码实现了一个函数 `vtableEqual`，用于比较一个未写入的 VTable 和一个已写入的 VTable 是否相等。
//
// 该函数的参数如下：
//	- `a` 是一个未写入的 VTable，类型为 `[]UOffsetT`。它是一个存储偏移量的整数切片，表示各个字段在对象中的布局和访问信息。
//	- `objectStart` 是对象的起始偏移量，类型为 `UOffsetT`。它表示未写入的 VTable 对应的对象的起始位置。
//	- `b` 是一个已写入的 VTable 对应的字节切片，类型为 `[]byte`。它包含了已写入的 VTable 内容。
//
// 函数首先会通过比较两个切片的长度来判断它们是否相等。如果两个切片长度不相等，说明它们不可能表示相同的 VTable，直接返回 `false`。
// 接下来，函数利用一个循环遍历未写入的 VTable 的每个元素。对于每个元素，它首先通过 `GetVOffsetT` 函数从已写入的 VTable 字节切片中获取相应的值。
// 然后，它会检查获取到的值是否为默认值（0），以及未写入的 VTable 对应的元素是否也为默认值（0）。如果两者都为默认值，该元素的比较被跳过。如果其中有一个不是默认值，或者两者都不是默认值但值不相等，说明两个 VTable 不相等，函数返回 `false`。
// 最后，如果所有元素的比较都相等，函数返回 `true`，表示未写入的 VTable 和已写入的 VTable 相等。
//
// 这个函数的作用是用于检查两个 VTable 是否一致，从而可以确定一个对象是否已经写入正确的 VTable，并且可以在需要时进行比较和验证。

// vtableEqual compares an unwritten vtable to a written vtable.
func vtableEqual(a []UOffsetT, objectStart UOffsetT, b []byte) bool {
	if len(a)*SizeVOffsetT != len(b) {
		return false
	}

	for i := 0; i < len(a); i++ {
		x := GetVOffsetT(b[i*SizeVOffsetT : (i+1)*SizeVOffsetT])

		// Skip vtable entries that indicate a default value.
		if x == 0 && a[i] == 0 {
			continue
		}

		y := SOffsetT(objectStart) - SOffsetT(a[i])
		if SOffsetT(x) != y {
			return false
		}
	}
	return true
}

// PrependBool prepends a bool to the Builder buffer.
// Aligns and checks for space.
func (b *Builder) PrependBool(x bool) {
	b.Prep(SizeBool, 0)
	b.PlaceBool(x)
}

// PrependUint8 prepends a uint8 to the Builder buffer.
// Aligns and checks for space.
func (b *Builder) PrependUint8(x uint8) {
	b.Prep(SizeUint8, 0)
	b.PlaceUint8(x)
}

// PrependUint16 prepends a uint16 to the Builder buffer.
// Aligns and checks for space.
func (b *Builder) PrependUint16(x uint16) {
	b.Prep(SizeUint16, 0)
	b.PlaceUint16(x)
}

// PrependUint32 prepends a uint32 to the Builder buffer.
// Aligns and checks for space.
func (b *Builder) PrependUint32(x uint32) {
	b.Prep(SizeUint32, 0)
	b.PlaceUint32(x)
}

// PrependUint64 prepends a uint64 to the Builder buffer.
// Aligns and checks for space.
func (b *Builder) PrependUint64(x uint64) {
	b.Prep(SizeUint64, 0)
	b.PlaceUint64(x)
}

// PrependInt8 prepends a int8 to the Builder buffer.
// Aligns and checks for space.
func (b *Builder) PrependInt8(x int8) {
	b.Prep(SizeInt8, 0)
	b.PlaceInt8(x)
}

// PrependInt16 功能是将一个 int16 类型的值插入到 Builder 缓冲区的前面。
// PrependInt16 接受一个参数 x 表示要插入的 int16 值。
//
// 函数执行的逻辑如下：
//	首先，调用 Prep 函数来为 int16 值预留空间。
//	Prep 函数接受两个参数，SizeInt16 表示int16值的大小（在FlatBuffers中为2字节），0 表示对齐方式（不进行对齐）。
//	Prep 函数会根据这两个参数来确定是否需要扩展缓冲区的大小，并调整缓冲区中的位置指针。
//
//	然后，调用 PlaceInt16 函数将 int16 值写入到缓冲区中。
//	PlaceInt16 函数接受一个 int16 值作为参数，将该值写入到缓冲区的当前位置，并根据对齐方式对缓冲区进行相应的调整。

// PrependInt16 prepends a int16 to the Builder buffer.
// Aligns and checks for space.
func (b *Builder) PrependInt16(x int16) {
	// 确保能容纳 16 Bytes
	b.Prep(SizeInt16, 0)
	// 存入 x
	b.PlaceInt16(x)
}

// PrependInt32 prepends a int32 to the Builder buffer.
// Aligns and checks for space.
func (b *Builder) PrependInt32(x int32) {
	b.Prep(SizeInt32, 0)
	b.PlaceInt32(x)
}

// PrependInt64 prepends a int64 to the Builder buffer.
// Aligns and checks for space.
func (b *Builder) PrependInt64(x int64) {
	b.Prep(SizeInt64, 0)
	b.PlaceInt64(x)
}

// PrependFloat32 prepends a float32 to the Builder buffer.
// Aligns and checks for space.
func (b *Builder) PrependFloat32(x float32) {
	b.Prep(SizeFloat32, 0)
	b.PlaceFloat32(x)
}

// PrependFloat64 prepends a float64 to the Builder buffer.
// Aligns and checks for space.
func (b *Builder) PrependFloat64(x float64) {
	b.Prep(SizeFloat64, 0)
	b.PlaceFloat64(x)
}

// PrependByte prepends a byte to the Builder buffer.
// Aligns and checks for space.
func (b *Builder) PrependByte(x byte) {
	b.Prep(SizeByte, 0)
	b.PlaceByte(x)
}

// PrependVOffsetT prepends a VOffsetT to the Builder buffer.
// Aligns and checks for space.
func (b *Builder) PrependVOffsetT(x VOffsetT) {
	b.Prep(SizeVOffsetT, 0)
	b.PlaceVOffsetT(x)
}

// PlaceBool prepends a bool to the Builder, without checking for space.
func (b *Builder) PlaceBool(x bool) {
	b.head -= UOffsetT(SizeBool)
	WriteBool(b.Bytes[b.head:], x)
}

// PlaceUint8 prepends a uint8 to the Builder, without checking for space.
func (b *Builder) PlaceUint8(x uint8) {
	b.head -= UOffsetT(SizeUint8)
	WriteUint8(b.Bytes[b.head:], x)
}

// PlaceUint16 prepends a uint16 to the Builder, without checking for space.
func (b *Builder) PlaceUint16(x uint16) {
	b.head -= UOffsetT(SizeUint16)
	WriteUint16(b.Bytes[b.head:], x)
}

// PlaceUint32 prepends a uint32 to the Builder, without checking for space.
func (b *Builder) PlaceUint32(x uint32) {
	b.head -= UOffsetT(SizeUint32)
	WriteUint32(b.Bytes[b.head:], x)
}

// PlaceUint64 prepends a uint64 to the Builder, without checking for space.
func (b *Builder) PlaceUint64(x uint64) {
	b.head -= UOffsetT(SizeUint64)
	WriteUint64(b.Bytes[b.head:], x)
}

// PlaceInt8 prepends a int8 to the Builder, without checking for space.
func (b *Builder) PlaceInt8(x int8) {
	b.head -= UOffsetT(SizeInt8)
	WriteInt8(b.Bytes[b.head:], x)
}

// PlaceInt16 prepends a int16 to the Builder, without checking for space.
func (b *Builder) PlaceInt16(x int16) {
	b.head -= UOffsetT(SizeInt16)
	WriteInt16(b.Bytes[b.head:], x)
}

// PlaceInt32 prepends a int32 to the Builder, without checking for space.
func (b *Builder) PlaceInt32(x int32) {
	b.head -= UOffsetT(SizeInt32)
	WriteInt32(b.Bytes[b.head:], x)
}

// PlaceInt64 prepends a int64 to the Builder, without checking for space.
func (b *Builder) PlaceInt64(x int64) {
	b.head -= UOffsetT(SizeInt64)
	WriteInt64(b.Bytes[b.head:], x)
}

// PlaceFloat32 prepends a float32 to the Builder, without checking for space.
func (b *Builder) PlaceFloat32(x float32) {
	b.head -= UOffsetT(SizeFloat32)
	WriteFloat32(b.Bytes[b.head:], x)
}

// PlaceFloat64 prepends a float64 to the Builder, without checking for space.
func (b *Builder) PlaceFloat64(x float64) {
	b.head -= UOffsetT(SizeFloat64)
	WriteFloat64(b.Bytes[b.head:], x)
}

// PlaceByte prepends a byte to the Builder, without checking for space.
func (b *Builder) PlaceByte(x byte) {
	b.head -= UOffsetT(SizeByte)   // 向前挪动 1 个位置，腾出一个 byte 的空间
	WriteByte(b.Bytes[b.head:], x) // 存入 1 byte 的 x
}

// PlaceVOffsetT prepends a VOffsetT to the Builder, without checking for space.
func (b *Builder) PlaceVOffsetT(x VOffsetT) {
	b.head -= UOffsetT(SizeVOffsetT)   // 向前挪动 2 个位置，腾出一个 VOffsetT 的空间
	WriteVOffsetT(b.Bytes[b.head:], x) // 存入 2 byte 的 x
}

// PlaceSOffsetT prepends a SOffsetT to the Builder, without checking for space.
func (b *Builder) PlaceSOffsetT(x SOffsetT) {
	b.head -= UOffsetT(SizeSOffsetT)   // 向前挪动 4 个位置，腾出一个 SOffsetT 的空间
	WriteSOffsetT(b.Bytes[b.head:], x) // 存入 4 byte 的 x
}

// PlaceUOffsetT prepends a UOffsetT to the Builder, without checking for space.
func (b *Builder) PlaceUOffsetT(x UOffsetT) {
	b.head -= UOffsetT(SizeUOffsetT)   // 向前挪动 4 个位置，腾出一个 UOffsetT 的空间
	WriteUOffsetT(b.Bytes[b.head:], x) // 存入 4 byte 的 x
}
