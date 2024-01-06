# StackCrisp

使用类 [git]((https://git-scm.com/)) 命令操作 [overlayfs](https://docs.kernel.org/filesystems/overlayfs.html) 挂载

---

> **该项目还处于非常早期阶段，没有文档、特性不稳定、缺陷数不清。请勿生产使用，有什么想法或者建议欢迎提 [Issue](https://github.com/yhlooo/stackcrisp/issues/new) 。**

## 这个项目从何而来

因为一些巧合我接触到了 [Overlay 文件系统](https://docs.kernel.org/filesystems/overlayfs.html) （或者说 overlayfs ），在了解它的过程中，我有一种模糊的感觉，它和我理解的 [git]((https://git-scm.com/)) 有某些相似性：

overlayfs 是一种分层的文件系统，其中每一层存储的是基于下层的“变更”，所有层叠加起来是文件系统存储的完整内容。十分类似的是， git 中每个 commit 存储的是基于之前内容的“变更”，从仓库建立到 HEAD 指针之间所有 commit 叠加起来就是当前仓库中存储的文件内容。 **那么能否使用 overlayfs 实现一些类似 git 的版本管理操作呢？**

我尝试将 git 中的某些操作概念套到 overlayfs 中，发现基本都能找到对应：

- 如果将当前 overlayfs 挂载中所有 lower 和 upper 组合成为新的 lower ，加上一个新的 upper 进行挂载，就相当于把原本 upper 中缓存的变更提交了（类似于 `git commit ...` ）
- 如果去掉 overlayfs 顶上的一些层重新挂载，就能得到这个文件系统中存储内容的一个比较早期的版本（类似于 `git reset ...` ）
- 如果基于某个 overlayfs 挂载的 lower ，加上一个新的 upper 进行挂载，就相当于从原来挂载中分出了一个分支（类似于 `git checkout -b <branch> ...`）
- ...

我震惊于我的“发现”。它看起来非常可行，但是我尝试了各种关键字通过 Google 和 GitHub 搜索，都没有找到类似的东西。 `docker commit ...` 算是个类似的东西，但是它太简单了，没有回退、分支等概念。

所以，经过一段时间的思考之后，有了这个项目。虽然我还没有想好它能做什么，不知道它能解决什么现实世界的问题，但是它不一定需要有什么用，我觉得它很有趣，这就足够了。

如果它启发了你，欢迎提 [Issue](https://github.com/yhlooo/stackcrisp/issues/new) 分享下你的想法。

## 入门

执行以下命令安装：

```shell
go install github.com/yhlooo/stackcrisp/cmd/stackcrisp@latest
```

使用方式参考 `stackcrisp --help`

（它跟 [`git`](https://git-scm.com/) 非常类似，你可以假装它就是 `git` ）

## 进度

项目看起来像个半成品，它确实是。目前它仅包含我觉得足够演示它核心能力的最少实现，包括以下类 git 命令：

- `init` 初始化一个目录
- `clone` 克隆一个已有目录
- `commit` 提交变更
- `checkout` 切换到指定 commit
- `log` 查看提交历史

已知问题：

- `commit` `checkout` 时不能在目标目录内操作，因为这会导致至少有一个 shell 进程工作目录在目标目录内使 overlay 无法卸载，无法完成 overlay 挂载的切换。需要在其它目录通过 `-C <path>` 参数指定要操作的目标目录。

以下是规划中的能力：（按我认为的优先级由高到低排序）

1. 分支概念。包括 `reset` `branch` `switch` 等命令，以及在 `checkout` 命令中添加分支相关操作
2. 更完善的状态和历史查询功能。包括 `log` `status` `diff` 等命令
3. tag 概念。包括 `tag` 命令，以及其它命令中 tag 相关操作
4. （可选）编辑提交树。包括 `rebase` `cherry-pick` `merge` 等命令

## 为什么是 `StackCrisp`

关于 `StackCrisp` 名字的由来，见 [为什么是 StackCrisp](docs/name.md)
