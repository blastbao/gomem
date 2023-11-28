// Package flatbuffers provides facilities to read and write flatbuffers
// objects.
package flatbuffers


// 简单来说 FlatBuffers 就是把对象数据，保存在一个一维的数组中，将数据都缓存在一个 ByteBuffer 中，每个对象在数组中被分为两部分。
// 	元数据部分：负责存放索引。
// 	真实数据部分：存放实际的值。
//
// 然而 FlatBuffers 与大多数内存中的数据结构不同，它使用严格的对齐规则和字节顺序来确保 buffer 是跨平台的。
// 此外，对于 table 对象，FlatBuffers 提供前向/后向兼容性和 optional 字段，以支持大多数格式的演变。
// 除了解析效率以外，二进制格式还带来了另一个优势，数据的二进制表示通常更具有效率。
// 我们可以使用 4 字节的 UInt 而不是 10 个字符来存储 10 位数字的整数。
//
// FlatBuffers 对序列化基本使用原则：
//	小端模式。FlatBuffers 对各种基本数据的存储都是按照小端模式来进行的，因为这种模式目前和大部分处理器的存储模式是一致的，可以加快数据读写的数据。
//	写入数据方向和读取数据方向不同。


// FlatBuffers 向 ByteBuffer 中写入数据的顺序是从 ByteBuffer 的尾部向头部填充，由于这种增长方向和 ByteBuffer 默认的增长方向不同，
// 因此 FlatBuffers 在向 ByteBuffer 中写入数据的时候就不能依赖 ByteBuffer 的 position 来标记有效数据位置，
// 而是自己维护了一个 space 变量来指明有效数据的位置，在分析 FlatBuffersBuilder 的时候要特别注意这个变量的增长特点。
//
// 但是，和数据的写入方向不同的是，FlatBuffers 从 ByteBuffer 中解析数据的时候又是按照 ByteBuffer 正常的顺序来进行的。
// FlatBuffers 这样组织数据存储的好处是，在从左到右解析数据的时候，能够保证最先读取到的就是整个 ByteBuffer 的概要信息
//（例如 Table 类型的 vtable 字段），方便解析。


// table 是 FlatBuffers 的基石，为了解决数据结构变更的问题，table 通过 vtable 间接访问字段。
// 每个 table 都带有一个 vtable（可以在具有相同布局的多个 table 之间共享），并且包含存储此特定类型 vtable 实例的字段的信息。
// vtable 还可能表明该字段不存在（因为此 FlatBuffers 是使用旧版本的代码编写的，仅仅因为信息对于此实例不是必需的，或者被视为已弃用），
// 在这种情况下会返回默认值。
//
// table 的内存开销很小（因为 vtables 很小并且共享）访问成本也很小（间接访问），但是提供了很大的灵活性。
// table 在特殊情况下可能比等价的 struct 花费更少的内存，因为字段在等于默认值时不需要存储在 buffer 中。
// 这样的结构决定了一些复杂类型的成员都是使用相对寻址进行数据访问的，即先从Table 中取到成员常量的偏移，
// 然后根据这个偏移再去常量真正存储的地址去取真实数据。
//
// 单就结构来讲：首先可以将 Table 分为两个部分，
// 第一部分是存储 Table 中各个成员变量的概要，这里命名为 vtable，
// 第二部分是 Table 的数据部分，存储 Table 中各个成员的值，这里命名为 table_data 。
// 注意 Table 中的成员如果是简单类型或者 Struct 类型，那么这个成员的具体数值就直接存储在 table_data 中；
// 如果成员是复杂类型，那么 table_data 中存储的只是这个成员数据相对于写入地址的偏移，
// 也就是说要获得这个成员的真正数据还要取出 table_data 中的数据进行一次相对寻址。

