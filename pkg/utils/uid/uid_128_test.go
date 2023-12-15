package uid

import (
	"encoding/base32"
	"encoding/hex"
	"reflect"
	"regexp"
	"strings"
	"testing"
)

var (
	uid128HexRegexp    = regexp.MustCompile(`^[0-9a-f]{32}$`)
	uid128Base32Regexp = regexp.MustCompile(`^[A-Z0-7]{26}$`)
)

// TestNewUID128 测试 NewUID128 方法
func TestNewUID128(t *testing.T) {
	// 随便试 1000 个
	lastOne := NewUID128()
	for i := 0; i < 1000; i++ {
		uid := NewUID128()

		// 检查是否与上一个结果重复
		// 对于这个量级的测试而言，理论上极小概率出现重复
		if uid.Hex() == lastOne.Hex() {
			t.Errorf("repeat with previous result: %q", uid)
		}
		// 检查是否有非预期格式的输出
		if hex := uid.Hex(); !uid128HexRegexp.MatchString(hex) {
			t.Errorf("unexpected hex result: %q (not match %q)", hex, uid128HexRegexp.String())
		}
		if b32 := uid.Base32(); !uid128Base32Regexp.MatchString(b32) {
			t.Errorf("unexpected base32 result: %q (not match %q)", b32, uid128Base32Regexp.String())
		}

		lastOne = uid
	}
}

// TestUid128_Hex 测试 uid128.Hex 方法
func TestUid128_Hex(t *testing.T) {
	in := uid128{123, 45, 67, 8, 90, 98, 76, 54, 32, 1, 233, 66, 99, 6, 71, 34}
	expected := "7b2d43085a624c362001e94263064722"
	ret := in.Hex()
	if ret != expected {
		t.Errorf("unexpected result: %q (expected %q)", ret, expected)
	}
}

// TestUid128_Base32 测试 uid128.Base32 方法
func TestUid128_Base32(t *testing.T) {
	in := uid128{123, 45, 67, 8, 90, 98, 76, 54, 32, 1, 233, 66, 99, 6, 71, 34}
	expected := "PMWUGCC2MJGDMIAB5FBGGBSHEI"
	ret := in.Base32()
	if ret != expected {
		t.Errorf("unexpected result: %q (expected %q)", ret, expected)
	}
}

// TestUid128_String 测试 uid128.String 方法
func TestUid128_String(t *testing.T) {
	in := uid128{123, 45, 67, 8, 90, 98, 76, 54, 32, 1, 233, 66, 99, 6, 71, 34}
	expected := in.Base32()
	ret := in.String()
	if ret != expected {
		t.Errorf("unexpected result: %q (expected %q)", ret, expected)
	}
}

// TestDecodeUID128FromHex 测试 DecodeUID128FromHex 方法
func TestDecodeUID128FromHex(t *testing.T) {
	in := "7b2d43085a624c362001e94263064722"
	expected := uid128{123, 45, 67, 8, 90, 98, 76, 54, 32, 1, 233, 66, 99, 6, 71, 34}
	ret, err := DecodeUID128FromHex(in)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}
	if !reflect.DeepEqual(ret, expected) {
		t.Errorf("unexpected result: %q (expected %q)", ret, expected)
	}

	if _, err := DecodeUID128FromHex(in + "12"); err == nil || !strings.Contains(err.Error(), "wrong length") {
		t.Errorf("expected \"wrong length ...\" error, but: %v", err)
	}
	if _, err := DecodeUID128FromHex(in[:20]); err == nil || !strings.Contains(err.Error(), "wrong length") {
		t.Errorf("expected \"wrong length ...\" error, but: %v", err)
	}
	if _, err := DecodeUID128FromHex(in[:30] + "xx"); err == nil || err != hex.InvalidByteError('x') {
		t.Errorf("expected InvalidByteError error, but: %v", err)
	}
}

// TestDecodeUID128FromString 测试 DecodeUID128FromString 方法
func TestDecodeUID128FromString(t *testing.T) {
	in := "PMWUGCC2MJGDMIAB5FBGGBSHEI"
	expected := uid128{123, 45, 67, 8, 90, 98, 76, 54, 32, 1, 233, 66, 99, 6, 71, 34}
	ret, err := DecodeUID128FromBase32(in)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}
	if !reflect.DeepEqual(ret, expected) {
		t.Errorf("unexpected result: %q (expected %q)", ret, expected)
	}

	if _, err := DecodeUID128FromBase32(in + "12"); err == nil || !strings.Contains(err.Error(), "wrong length") {
		t.Errorf("expected \"wrong length ...\" error, but: %v", err)
	}
	if _, err := DecodeUID128FromBase32(in[:20]); err == nil || !strings.Contains(err.Error(), "wrong length") {
		t.Errorf("expected \"wrong length ...\" error, but: %v", err)
	}
	if _, err := DecodeUID128FromBase32(in[:24] + "xx"); err == nil || err != base32.CorruptInputError(24) {
		t.Errorf("expected CorruptInputError error, but: %v", err)
	}
}
