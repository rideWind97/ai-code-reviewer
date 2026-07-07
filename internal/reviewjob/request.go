package reviewjob

// Request 是写入队列的审查任务 payload。
//
// GitHub/GitLab webhook payload 字段差异很大；进入队列前先转换成这个统一结构，
// worker 就不需要理解各个 provider 的原始 webhook JSON。
type Request struct {
	Provider string `json:"provider"` // provider 名称，例如 github/gitlab。
	Repo     string `json:"repo"`     // 仓库全名。
	Number   int    `json:"number"`   // PR/MR 编号。
}
