// Package aria2 提供 aria2 JSON-RPC HTTP 客户端和安装管理功能。
// 包括完整的 RPC 操作、安装/卸载/升级流程编排、配置模板生成。
package aria2

// RpcRequest JSON-RPC 2.0 请求结构体
type RpcRequest struct {
	Jsonrpc string        `json:"jsonrpc"`
	Method  string        `json:"method"`
	Params  []any `json:"params"`
	Id      string        `json:"id"`
}

// RpcResponse JSON-RPC 2.0 响应结构体
type RpcResponse struct {
	Jsonrpc string      `json:"jsonrpc"`
	Result  any `json:"result,omitempty"`
	Error   *RpcError   `json:"error,omitempty"`
	Id      string      `json:"id"`
}

// RpcError JSON-RPC 错误信息
type RpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// StatusInfo aria2 任务状态信息，覆盖 tellStatus/tellActive/tellWaiting/tellStopped 返回字段
type StatusInfo struct {
	Gid             string       `json:"gid"`
	Status          string       `json:"status"` // active / waiting / paused / error / complete / removed
	TotalLength     string       `json:"totalLength"`
	CompletedLength string       `json:"completedLength"`
	UploadLength    string       `json:"uploadLength"`
	DownloadSpeed   string       `json:"downloadSpeed"`
	UploadSpeed     string       `json:"uploadSpeed"`
	InfoHash        string       `json:"infoHash"`
	NumSeeders      string       `json:"numSeeders"`
	Connections     string       `json:"connections"`
	ErrorCode       string       `json:"errorCode"`
	ErrorMessage    string       `json:"errorMessage"`
	Files           []FileInfo   `json:"files"`
	Dir             string       `json:"dir"`
	Bitfield        string       `json:"bitfield"`
	PieceLength     string       `json:"pieceLength"`
	NumPieces       string       `json:"numPieces"`
	Following       string       `json:"following"`
	BelongsTo       string       `json:"belongsTo"`
	VerifiedLength  string       `json:"verifiedLength"`
	VerifyIntegrity string       `json:"verifyIntegrityPending"`
}

// FileInfo aria2 文件信息
type FileInfo struct {
	Index           string     `json:"index"`
	Path            string     `json:"path"`
	Length          string     `json:"length"`
	CompletedLength string     `json:"completedLength"`
	Selected        string     `json:"selected"`
	Uris            []FileUri   `json:"uris"`
}

// FileUri 文件下载 URI
type FileUri struct {
	Uri    string `json:"uri"`
	Status string `json:"status"`
}

// GlobalStat 全局下载统计
type GlobalStat struct {
	DownloadSpeed      string `json:"downloadSpeed"`
	UploadSpeed        string `json:"uploadSpeed"`
	NumActive          string `json:"numActive"`
	NumWaiting         string `json:"numWaiting"`
	NumStopped         string `json:"numStopped"`
	NumStoppedTotal    string `json:"numStoppedTotal"`
}

// VersionInfo aria2 版本信息
type VersionInfo struct {
	Version  string   `json:"version"`
	Features []string `json:"enabledFeatures"`
}

// AddOptions AddURI 时的可选参数
type AddOptions struct {
	Dir      string `json:"dir,omitempty"`
	Out      string `json:"out,omitempty"`
	Header   string `json:"header,omitempty"`
	Split    string `json:"split,omitempty"`
	Position string `json:"position,omitempty"`
}

// ProcessStatus aria2 进程运行状态
type ProcessStatus struct {
	Running  bool   // 进程是否运行中
	PID      int    // 进程 PID
	Uptime   string // 运行时长（人类可读）
	Tasks    int    // 当前活动任务数
	Version  string // aria2 版本号
}

// HealthInfo 综合健康检查信息
type HealthInfo struct {
	Aria2Running bool   // aria2 是否运行
	Aria2Version string // aria2 版本
	RPCOK        bool   // RPC 连接是否正常
	DiskFree     string // 下载目录剩余空间（人类可读）
	DiskPercent  float64 // 磁盘使用百分比
}
