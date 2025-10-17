package types

// GitCommitRequest Git提交请求
type GitCommitRequest struct {
	BasePath   string `json:"base_path"`   // Git仓库基础路径，可选，为空时使用config.go中的basePath/funNovel
	CommitMsg  string `json:"commit_msg"`  // 提交信息，可选
	BranchName string `json:"branch_name"` // 分支名称，可选
	RemoteName string `json:"remote_name"` // 远程仓库名称，默认为origin
	TargetRef  string `json:"target_ref"`  // 目标引用，默认为HEAD:refs/for/uni/funNovel/devNew
}

// GitPullBranchRequest Git创建分支请求
type GitPullBranchRequest struct {
	RepositoryURL string `json:"repository_url"` // 代码库地址，如 git@code.funshion.com:somalia/funnovel.git
	BranchName    string `json:"branch_name"`    // 要创建的分支名称
	RemoteName    string `json:"remote_name"`    // 远程仓库名称，默认为origin
	DefaultBranch string `json:"default_branch"` // 默认分支，可选，为空时自动检测
}

// GitOperationResult Git操作结果
type GitOperationResult struct {
	Success bool                 `json:"success"`
	Message string               `json:"message"`
	Details []GitOperationDetail `json:"details"`
	Error   string               `json:"error,omitempty"`
}

// GitOperationDetail Git操作详情
type GitOperationDetail struct {
	Operation string `json:"operation"` // 操作名称
	Status    string `json:"status"`    // 操作状态：success, error, warning
	Message   string `json:"message"`   // 操作消息
	Output    string `json:"output"`    // 命令输出
	Duration  int64  `json:"duration"`  // 执行时长(毫秒)
}

// GitStatus Git状态信息
type GitStatus struct {
	Branch     string   `json:"branch"`      // 当前分支
	Status     string   `json:"status"`      // 工作区状态
	Staged     []string `json:"staged"`      // 已暂存文件
	Modified   []string `json:"modified"`    // 已修改文件
	Untracked  []string `json:"untracked"`   // 未跟踪文件
	StashCount int      `json:"stash_count"` // 暂存数量
	Ahead      int      `json:"ahead"`       // 领先远程分支的提交数
	Behind     int      `json:"behind"`      // 落后远程分支的提交数
}

// GitLogRequest Git日志请求
type GitLogRequest struct {
	Limit    int    `json:"limit"`     // 限制数量，默认15
	FilePath string `json:"file_path"` // 可选：特定文件的日志
}

// GitCommit Git提交信息
type GitCommit struct {
	CommitId string `json:"commit_id"` // 提交ID
	Author   string `json:"author"`    // 作者
	Email    string `json:"email"`     // 邮箱
	Date     string `json:"date"`      // 提交时间
	Message  string `json:"message"`   // 提交信息
}

// GitLogResponse Git日志响应
type GitLogResponse struct {
	Commits []GitCommit `json:"commits"`
}
