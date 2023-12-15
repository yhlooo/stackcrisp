package uid

// UID 唯一 ID
type UID interface {
	// Hex 返回十六进制表示
	Hex() string
	// Base32 返回 base32 编码表示
	Base32() string
	// String 返回字符串表示
	// 与 Base32 结果一致
	String() string
}
