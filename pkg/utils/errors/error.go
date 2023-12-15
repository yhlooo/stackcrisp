package errors

import (
	"errors"
	"fmt"
)

// ReasonInternalError 内部错误
const ReasonInternalError = "InternalError"

// Status 执行某些过程的异常结果（错误）
type Status interface {
	error
	// Reason 返回错误原因
	// 一个大驼峰格式可枚举的值
	Reason() string
	// Code 返回错误码
	// 可选， 0 表示无错误码
	Code() uint32
	// Message 返回人类可读的错误描述
	Message() string
}

// defaultStatus 是 Status 的一个默认实现
type defaultStatus struct {
	code    uint32
	reason  string
	message string
}

var _ Status = defaultStatus{}

// Error 返回错误的字符串表示
func (status defaultStatus) Error() string {
	if status.code == 0 {
		return fmt.Sprintf("%s: %s", status.reason, status.message)
	}
	return fmt.Sprintf("%s(%d): %s", status.reason, status.code, status.message)
}

// Code 返回错误码
func (status defaultStatus) Code() uint32 {
	return status.code
}

// Reason 返回错误原因
// 一个大驼峰格式可枚举的值
func (status defaultStatus) Reason() string {
	return status.reason
}

// Message 返回人类可读的错误描述
func (status defaultStatus) Message() string {
	return status.message
}

// FromError 从 err 获取 Status
func FromError(err error) Status {
	var status Status
	if errors.As(err, &status) {
		return status
	}
	return New(ReasonInternalError, err.Error())
}

// New 创建一个 Status
func New(reason, message string) Status {
	return defaultStatus{
		reason:  reason,
		message: message,
	}
}

// NewWithCode 创建一个带错误码的 Status
func NewWithCode(code uint32, reason, message string) Status {
	return defaultStatus{
		code:    code,
		reason:  reason,
		message: message,
	}
}
