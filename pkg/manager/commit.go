package manager

import (
	"time"

	"github.com/yhlooo/stackcrisp/pkg/spaces/trees"
	"github.com/yhlooo/stackcrisp/pkg/utils/uid"
	"github.com/yhlooo/stackcrisp/pkg/workspaces"
)

const (
	nodeAnnoCommitDate    = "commit-date"
	nodeAnnoCommitMessage = "commit-message"
)

// Commit 提交信息
type Commit struct {
	ID       uid.UID
	Date     *time.Time
	Message  string
	Branches []workspaces.Branch
	Tags     []string
}

// SetToNode 设置提交信息到节点
func (commit *Commit) SetToNode(node trees.Node) {
	if commit == nil {
		return
	}
	if commit.Date != nil {
		node.AddAnnotation(nodeAnnoCommitDate, commit.Date.Format(time.RFC3339))
	}
	node.AddAnnotation(nodeAnnoCommitMessage, commit.Message)
}

// GetCommitFromNode 从节点获取提交信息
func GetCommitFromNode(ws workspaces.Workspace, node trees.Node) Commit {
	tree := ws.Space().Tree()
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

	// 查找分支
	var branches []workspaces.Branch
	for fullName, n := range tree.Branches() {
		b, err := workspaces.ParseBranchFullName(fullName)
		if err != nil {
			continue
		}
		if b.IsLocal() && b.WorkspaceID().Base32() != ws.ID().Base32() {
			// 其它 workspace 的本地分支
			continue
		}
		if n.Parent().ID().Hex() == node.ID().Hex() {
			branches = append(branches, b)
		}
	}

	// 查找关联标签
	var tags []string
	for t, n := range tree.Tags() {
		if n.ID().Hex() == node.ID().Hex() {
			tags = append(tags, t)
		}
	}

	return Commit{
		ID:       node.ID(),
		Date:     date,
		Message:  anno[nodeAnnoCommitMessage],
		Branches: branches,
		Tags:     tags,
	}
}
