package codec

// Data 约束序列化结果的载体类型。
type Data interface {
	string | []byte
}
