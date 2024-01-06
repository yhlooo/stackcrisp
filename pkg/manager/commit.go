package manager

import (
	"time"

	"github.com/yhlooo/stackcrisp/pkg/spaces/trees"
	"github.com/yhlooo/stackcrisp/pkg/utils/uid"
)

const (
	nodeAnnoCommitDate    = "commit-date"
	nodeAnnoCommitMessage = "commit-message"
)

// Commit 提交信息
type Commit struct {
	ID      uid.UID
	Date    *time.Time
	Message string
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
func GetCommitFromNode(node trees.Node) Commit {
	anno := node.Annotations()

	var date *time.Time
	dateStr := anno[nodeAnnoCommitDate]
	if dateStr != "" {
		d, err := time.Parse(time.RFC3339, dateStr)
		if err == nil {
			date = &d
		}
	}

	return Commit{
		ID:      node.ID(),
		Date:    date,
		Message: anno[nodeAnnoCommitMessage],
	}
}
