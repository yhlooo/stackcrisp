package workspaces

import (
	"time"

	"github.com/yhlooo/stackcrisp/pkg/spaces/trees"
	"github.com/yhlooo/stackcrisp/pkg/utils/uid"
)

func NewCommitInfo(msg string) CommitInfo {
	now := time.Now()
	return &defaultCommit{
		date:    &now,
		message: msg,
	}
}

// GetCommitFromNode 从节点获取提交信息
func GetCommitFromNode(ws Workspace, node trees.Node) Commit {
	anno := node.Annotations()

	// 获取提交日期
	var date *time.Time
	dateStr := anno[nodeAnnoCommitDate]
	if dateStr != "" {
		d, err := time.Parse(time.RFC3339, dateStr)
		if err == nil {
			date = &d
		}
	}

	nodeIDHex := node.ID().Hex()

	// 查找关联分支
	var branches []Branch
	for _, b := range ws.AllBranches() {
		if b.Head().Parent().ID().Hex() == nodeIDHex {
			branches = append(branches, b)
		}
	}

	// 查找关联标签
	var tags []string
	for t, n := range ws.Space().Tree().Tags() {
		if n.ID().Hex() == nodeIDHex {
			tags = append(tags, t)
		}
	}

	return &defaultCommit{
		id:       node.ID(),
		date:     date,
		message:  anno[nodeAnnoCommitMessage],
		branches: branches,
		tags:     tags,
	}
}

const (
	nodeAnnoCommitDate    = "commit-date"
	nodeAnnoCommitMessage = "commit-message"
)

// defaultCommit 是 Commit 的一个默认实现
type defaultCommit struct {
	id       uid.UID
	date     *time.Time
	message  string
	branches []Branch
	tags     []string
}

var _ Commit = &defaultCommit{}
var _ CommitInfo = &defaultCommit{}

// ID 返回提交 ID
func (commit *defaultCommit) ID() uid.UID {
	return commit.id
}

// Date 返回提交日期时间
func (commit *defaultCommit) Date() *time.Time {
	return commit.date
}

// Message 返回提交信息
func (commit *defaultCommit) Message() string {
	return commit.message
}

// Branches 返回提交对应分支头指针的分支
func (commit *defaultCommit) Branches() []Branch {
	return commit.branches
}

// Tags 返回提交对应标签
func (commit *defaultCommit) Tags() []string {
	return commit.tags
}

// SetToNode 设置提交信息到节点
func (commit *defaultCommit) SetToNode(node trees.Node) {
	if commit == nil {
		return
	}
	if commit.date != nil {
		node.AddAnnotation(nodeAnnoCommitDate, commit.date.Format(time.RFC3339))
	}
	node.AddAnnotation(nodeAnnoCommitMessage, commit.message)
}
