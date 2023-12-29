package uid

import (
	"encoding/base32"
	"encoding/hex"
	"fmt"
	"hash"
	"math/rand"
	"time"
)

// uid128 是 UID 的一个实现，长度 128 位
type uid128 [16]byte

var _ UID = uid128{}

// Hex 返回十六进制表示
// 32 个字符
func (uid uid128) Hex() string {
	return hex.EncodeToString(uid[:])
}

// Base32 返回 base32 编码表示
// 26 个字符
func (uid uid128) Base32() string {
	return base32Encoding.EncodeToString(uid[:])
}

// String 返回字符串表示
// 与 Base32 结果一致
func (uid uid128) String() string {
	return fmt.Sprintf("%s(%s)", uid.Hex(), uid.Base32())
}

// randObj 随机数生成器
var randObj *rand.Rand

// base32Encoding base32 编码器
var base32Encoding = base32.StdEncoding.WithPadding(base32.NoPadding)

func init() {
	randObj = rand.New(rand.NewSource(time.Now().UnixNano()))
}

// NewUID128 生成一个随机的 128 位 UID
func NewUID128() UID {
	var ret uid128
	for i := 0; i < 2; i++ {
		v := randObj.Uint64()
		for j := 0; j < 8; j++ {
			ret[i*8+j] = uint8(v & uint64(255))
			v = v >> 8
		}
	}
	return ret
}

// DecodeUID128FromHex 从十六进制形式解码 128 位 UID
// 输入 32 个字符
func DecodeUID128FromHex(in string) (UID, error) {
	if len(in) != 32 {
		return nil, fmt.Errorf("wrong length in bytes: %d (expected: 32)", len(in))
	}
	ret, err := hex.DecodeString(in)
	if err != nil {
		return nil, err
	}
	return uid128(ret), err
}

// DecodeUID128FromBase32 从 base32 形式解码 128 位 UID
// 输入 26 个字符
func DecodeUID128FromBase32(in string) (UID, error) {
	if len(in) != 26 {
		return nil, fmt.Errorf("wrong length in bytes: %d (expected: 26)", len(in))
	}
	ret, err := base32Encoding.DecodeString(in)
	if err != nil {
		return nil, err
	}
	return uid128(ret), err
}

// NewUID128FromHash 通过指定哈希算法生成一个 128 位 UID
// 如果哈希算法输出大于 128 位，则取高 128 位，不足 128 位则以 0 补足
func NewUID128FromHash(data []byte, algorithm func() hash.Hash) UID {
	h := algorithm()
	_, _ = h.Write(data)
	return uid128(h.Sum(nil))
}
