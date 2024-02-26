package flatbuffers

// 定义表的结构：
//	table MyTable {
//	 field1: int;    // 偏移 0
//	 field2: string; // 偏移 4
//	 field3: int;    // 偏移 6
//	}
//
// 假设：
//	field1 值为 42。
//	field2 值为 "hello"。
//	field3 未设置（默认值）。
//
// 二进制数据布局：
// 1. vtable 存储
//	[0x0A] [0x10] [0x04] [0x08] [0x00]
// 其中：
//	0x0A (vtable_size=10)：vtable 大小。
//	0x10 (object_size=16)：MyTable 对象大小。
//	0x04：field1 相对于表起始地址的偏移。
//	0x08：field2 相对于表起始地址的偏移。
//	0x00：field3 未设置。
//
// 2. 表数据存储
//	[0x04 0x00] [0x2A 0x00 0x00 0x00] [0x08 0x00] ["hello"] ...
// 其中：
//	0x04 0x00：指向 vtable 的偏移。
//	0x2A (42)：field1 数据。
//	0x08 0x00：field2 的偏移指针，指向 "hello" 字符串数据
//
// 字段访问：
//	访问 field1
//		定位 vtable：根据表头的 vtable_offset (0x04 0x00) 找到 vtable。
//		在 vtable 中找到 field1_offset (0x04)。
//		使用 field1_offset 定位到 field1 数据地址：0x0A + 0x04 = 0x0C。
//		读取数据：42。
//	访问 field2
//		定位 vtable：同上。
//		在 vtable 中找到 field2_offset (0x08)。
//		使用 field2_offset 定位到 field2 数据指针地址：0x0A + 0x08 = 0x10。
//		读取数据指针：指向字符串数据 0x12。
//		读取字符串 "hello"。
//	访问 field3
//		定位 vtable：同上。
//		在 vtable 中找到 field3_offset (0x00)。
//		偏移为 0，表示字段未设置，使用默认值。
//
// 示例代码：
//
//	// 假设 `buf` 是 FlatBuffer 的字节数组
//	tableOffset := flatbuffers.GetUOffsetT(buf[0:4]) // 获取表起始偏移
//	vtableOffset := tableOffset - flatbuffers.GetInt32(buf[tableOffset:tableOffset+4])
//
//	// 获取字段 1 偏移
//	field1Offset := vtableOffset + 4
//	// 获取字段 1 数据
//	field1Data := flatbuffers.GetInt32(buf[tableOffset+field1Offset:])
//
//	field2Offset := vtableOffset + 8
//	field2Pointer := flatbuffers.GetUOffsetT(buf[tableOffset+field2Offset:])
//	field2String := flatbuffers.String(buf[field2Pointer:])

// 原理一
//
// vtable 中的字段偏移量存储位置由字段的顺序（schema 定义顺序）决定，具体为：
//	第 1 个字段：存储在 vtable 偏移 4。
//	第 2 个字段：存储在 vtable 偏移 6。
//	第 3 个字段：存储在 vtable 偏移 8。
//
// 因此，字段在 vtable 中的存储位置为：
//	vtable_field_offset = vtable_address + 4 + (field_index * 2)
//
// 其中：
//  4 是 2B 的 vtable 大小和 2B 的数据体大小
//	field_index 是字段的索引（从 0 开始）。
//	2 是每个字段偏移存储的大小（int16）。
//
// vtable 的设计使得字段的读取是基于固定偏移计算的，无需线性扫描数据，直接通过索引定位：
//	- 通过 vtable_offset 找到 vtable。
//	- 按照字段的索引找到字段偏移位置。
//	- 从偏移值读取字段数据，或确定字段是否存在（offset = 0 时字段未设置）。
// 可以高效、快速地访问表中的数据，同时确保字段是稀疏存储且向后兼容的。

// 原理二
//
// 在 FlatBuffers 中，根对象通常指的是序列化数据的起始位置，也就是表头。
// 它包含一个指向 vtable 的偏移量，以便在读取时快速定位 vtable 并访问字段数据。
//
// 代码示例：
//	int32_t root_offset = *(int32_t*)buffer_address;    	// 读取根对象偏移量，buffer_address 是文件起始地址
//	int32_t table_offset = buffer_address + root_offset; 	// 计算表头地址
//	int32_t vtable_offset = *(int32_t*)table_offset;    	// 从表头读取 vtable_offset
//	int32_t vtable_address = table_offset - vtable_offset; 	// 根据表头计算 vtable 地址
//	printf("表头地址: 0x%X\n", table_offset);
//	printf("vtable 地址: 0x%X\n", vtable_address);

// Q&A
// 网络传输过程中，flatbuffer 是不是每条消息都要存储 vtable , 这不是导致数据量增加吗？
// fb 在每条消息中都会包含 vtable 来描述其字段布局，这确实会增加一些数据量，但这是为了实现高效解析和动态扩展性所做的权衡。
//
// 每个 vtable 需要存储以下信息：
//	- vtable 本身的大小（2 字节）
//	- ...（2 字节）
//	- 每个字段的偏移（每个字段 2 字节）
//	- 数据区相对于 vtable 的偏移（2 字节）
//
// 示例：
//	如果一个表有 5 个字段，则 vtable 的大小为：
//		2+2（大小）+ 2（数据区偏移）+ 2 × 5（字段偏移） = 16 字节
// 对于一个表中只有少量字段的情况下（如数十字节），vtable 的相对开销较大。
// 对于复杂表结构（如嵌套表、数组等），vtable 的开销可以忽略不计。
//
// 优化：
// 	压缩数据：FlatBuffers 消息可以使用压缩算法（如 Gzip）进行压缩，由于 vtable 占用较少空间，压缩效果通常很好。
//	共享 vtable：在某些情况下（如批量传输类似结构的表），可以复用相同的 vtable（需要手动设计，不是 FlatBuffers 的默认行为）。
//	裁剪消息：使用自定义 schema 或字段移除工具，仅保留需要传输的字段，减少冗余数据。

// Q&A 与 Protobuf 的不同
//
//	Protobuf 不需要 vtable： 每个字段都通过字段标签（field tag）编码，这些标签描述了字段的类型和顺序。
//	优缺点：
//	 - Protobuf 消息中没有类似 vtable 的结构，因此数据更加紧凑。
//	 - 解析时需要扫描字段标签并解码，效率比 FlatBuffers 稍低。

// Q&A FlatBuffers 是自描述的吗？
//
// “自描述”意味着数据结构中包含足够的信息来解释其内容和布局，无需依赖外部的 schema。例如：
//	- JSON 和 XML 是完全自描述的，因为它们直接包含字段名和值。
//	- Protobuf 不是完全自描述的，因为它的二进制数据需要 schema（.proto 文件）来解析字段含义。
//
// FlatBuffers 不是完全自描述：
//	- 无法独立解码：二进制数据不包含字段名或类型信息，如果缺少 schema，尽管可以读取字段值，但无法知道字段的实际含义。
//	- 依赖外部工具生成解析器：FlatBuffers 生成的解析代码使用 schema 来定义数据结构。

// Q&A FlatBuffers 如何支持嵌套表？
//
//
// 定义模式文件
//
// 	``` schema.fbs
//	table Address {
//		city: string;
//		street: string;
//	}
//
//	table Person {
//		name: string;
//		age: int;
//		address: Address;
//	}
//
//	root_type Person;
// 	```
//
// 生成代码
// 	flatc --go schema.fbs
//
//
// 构建嵌套表
//
//	``` 构造
//	builder := flatbuffers.NewBuilder(0)
//
//	// 创建 Address 表
//	city := builder.CreateString("New York")
//	street := builder.CreateString("5th Avenue")
//	code.AddressStart(builder)
//	code.AddressAddCity(builder, city)
//	code.AddressAddStreet(builder, street)
//	address := code.AddressEnd(builder)
//
//	// 创建 Person 表，包含 Address
//	name := builder.CreateString("John Doe")
//	code.PersonStart(builder)
//	code.PersonAddName(builder, name)
//	code.PersonAddAge(builder, 30)
//	code.PersonAddAddress(builder, address)
//	person := code.PersonEnd(builder)
//
//	// 完成缓冲区
//	builder.Finish(person)
//	// 获取数据
//	buf := builder.FinishedBytes()
//	```
//
//	``` 解析
//	// 解析 Person 表
//	personObj := code.GetRootAsPerson(buf, 0)
//	// 获取基础字段
//	fmt.Println("Name:", string(personObj.Name()))
//	fmt.Println("Age:", personObj.Age())
//	// 获取 address 字段
//	addressObj := new(code.Address)
//	personObj.Address(addressObj)
//	// 解析 Address 表
//	fmt.Println("City:", string(addressObj.City()))
//	fmt.Println("Street:", string(addressObj.Street()))
//	```
