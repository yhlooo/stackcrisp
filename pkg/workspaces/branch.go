package workspaces

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/yhlooo/stackcrisp/pkg/utils/uid"
)

const (
	globalBranchPrefix      = "global/"
	globalBranchLocalPrefix = "origin/"
)

var (
	branchFullNameRegexp = regexp.MustCompile(`^(global|[A-Z0-9]{26})/(.*)$`)
)

// ParseBranchFullName 解析分支完整名，获取分支
func ParseBranchFullName(fullName string) (Branch, error) {
	groups := branchFullNameRegexp.FindStringSubmatch(fullName)
	if groups == nil {
		return nil, fmt.Errorf("invalid branch full name: %q (not match %q)", fullName, branchFullNameRegexp.String())
	}
	prefix := groups[1]
	name := groups[2]
	if prefix == "global" {
		return NewGlobalBranch(name), nil
	}

	// 解析 workspace ID
	workspaceID, err := uid.DecodeUID128FromBase32(prefix)
	if err != nil {
		return nil, fmt.Errorf("parse the workspace ID %q of the branch %q error: %w", prefix, fullName, err)
	}

	return NewLocalBranch(workspaceID, name), nil
}

// ParseBranchLocalName 解析分支本地名，获取分支
//
// Note: 可能得到多个可能的结果
func ParseBranchLocalName(workspaceID uid.UID, localName string) []Branch {
	var ret []Branch
	// 首先可能是个本地分支
	ret = append(ret, NewLocalBranch(workspaceID, localName))
	// 如果前缀匹配，也可能是个全局分支
	if strings.HasPrefix(localName, globalBranchLocalPrefix) {
		ret = append(ret, NewGlobalBranch(strings.TrimPrefix(localName, globalBranchLocalPrefix)))
	}
	return ret
}

// NewLocalBranch 创建一个本地分支
func NewLocalBranch(workspaceID uid.UID, name string) Branch {
	return &defaultBranch{
		name:        name,
		workspaceID: workspaceID,
	}
}

// NewGlobalBranch 创建一个远程分支
func NewGlobalBranch(name string) Branch {
	return &defaultBranch{
		name: name,
	}
}

// defaultBranch 是 Branch 的一个默认实现
type defaultBranch struct {
	name        string
	workspaceID uid.UID
}

var _ Branch = &defaultBranch{}

// Name 返回分支名
func (branch *defaultBranch) Name() string {
	return branch.name
}

// FullName 返回在 spaces.Space 中存储的完整名
func (branch *defaultBranch) FullName() string {
	if branch.workspaceID != nil {
		return fmt.Sprintf("%s/%s", branch.workspaceID.Base32(), branch.name)
	}
	return globalBranchPrefix + branch.name
}

// LocalName 返回在 Workspace 中的本地名
func (branch *defaultBranch) LocalName() string {
	if branch.workspaceID != nil {
		return branch.name
	}
	return globalBranchLocalPrefix + branch.name
}

// IsGlobal 返回该分支是否全局分支
func (branch *defaultBranch) IsGlobal() bool {
	return branch.workspaceID == nil
}

// IsLocal 返回该分支是否 Workspace 本地分支
func (branch *defaultBranch) IsLocal() bool {
	return branch.workspaceID != nil
}

// WorkspaceID 对于 Workspace 本地分支，返回所属 Workspace ID ，否则返回 nil
func (branch *defaultBranch) WorkspaceID() uid.UID {
	return branch.workspaceID
}
