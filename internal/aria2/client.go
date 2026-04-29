package aria2

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	// 单个 RPC 调用超时
	rpcTimeout = 10 * time.Second
	// 请求 ID 随机字节数
	idBytes = 8
)

// Client aria2 JSON-RPC HTTP 客户端
type Client struct {
	url    string
	secret string
	http   *http.Client
}

// NewClient 创建 aria2 RPC 客户端
func NewClient(host string, port int, secret string) *Client {
	return &Client{
		url:    fmt.Sprintf("http://%s:%d/jsonrpc", host, port),
		secret: secret,
		http:   &http.Client{Timeout: rpcTimeout},
	}
}

// generateID 生成随机请求 ID
func generateID() (string, error) {
	b := make([]byte, idBytes)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("生成请求ID失败: %w", err)
	}
	return hex.EncodeToString(b), nil
}

// call 核心 RPC 调用方法，构建 JSON-RPC 请求、发送 HTTP POST、解析响应
func (c *Client) call(method string, params []any, result any) error {
	id, err := generateID()
	if err != nil {
		return err
	}

	// 构建请求，首位参数为 token:secret 认证
	fullParams := make([]any, 0, len(params)+1)
	fullParams = append(fullParams, "token:"+c.secret)
	fullParams = append(fullParams, params...)

	req := RpcRequest{
		Jsonrpc: "2.0",
		Method:  "aria2." + method,
		Id:      id,
		Params:  fullParams,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("序列化 RPC 请求失败: %w", err)
	}

	httpReq, err := http.NewRequest("POST", c.url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("创建 HTTP 请求失败: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(httpReq)
	if err != nil {
		return fmt.Errorf("aria2 RPC 请求失败: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取 RPC 响应失败: %w", err)
	}

	var rpcResp RpcResponse
	if err := json.Unmarshal(respBody, &rpcResp); err != nil {
		return fmt.Errorf("解析 RPC 响应失败: %w", err)
	}

	if rpcResp.Error != nil {
		return fmt.Errorf("aria2 错误 [%d]: %s", rpcResp.Error.Code, rpcResp.Error.Message)
	}

	// 如果调用者不需要结果，直接返回
	if result == nil {
		return nil
	}

	// 将 Result 重新序列化再反序列化到目标结构体
	resultJSON, err := json.Marshal(rpcResp.Result)
	if err != nil {
		return fmt.Errorf("序列化 RPC 结果失败: %w", err)
	}
	if err := json.Unmarshal(resultJSON, result); err != nil {
		return fmt.Errorf("解析 RPC 结果失败: %w", err)
	}
	return nil
}

// ===== 下载操作 =====

// AddURI 添加 HTTP/FTP 下载任务，返回 GID
func (c *Client) AddURI(uris []string, opts *AddOptions) (string, error) {
	params := []any{uris}
	if opts != nil {
		params = append(params, opts)
	}
	var gid string
	if err := c.call("addUri", params, &gid); err != nil {
		return "", err
	}
	return gid, nil
}

// AddTorrent 添加种子文件下载，filepath 为服务器本地路径
func (c *Client) AddTorrent(filepath string) (string, error) {
	var gid string
	// aria2.addTorrent 接受 base64 编码的种子内容
	if err := c.call("addTorrent", []any{filepath}, &gid); err != nil {
		return "", err
	}
	return gid, nil
}

// AddMetalink 添加 Metalink 下载
func (c *Client) AddMetalink(uri string) (string, error) {
	var gid string
	if err := c.call("addMetalink", []any{uri}, &gid); err != nil {
		return "", err
	}
	return gid, nil
}

// Pause 暂停指定任务
func (c *Client) Pause(gid string) error {
	return c.call("pause", []any{gid}, nil)
}

// PauseAll 暂停全部活动任务
func (c *Client) PauseAll() error {
	return c.call("pauseAll", nil, nil)
}

// Resume 恢复指定任务
func (c *Client) Resume(gid string) error {
	return c.call("resume", []any{gid}, nil)
}

// ResumeAll 恢复全部暂停任务
func (c *Client) ResumeAll() error {
	return c.call("resumeAll", nil, nil)
}

// Remove 删除指定任务（包含文件）
func (c *Client) Remove(gid string) error {
	return c.call("remove", []any{gid}, nil)
}

// RemoveResult 删除已完成任务记录
func (c *Client) RemoveResult(gid string) error {
	return c.call("removeDownloadResult", []any{gid}, nil)
}

// PurgeResult 清理所有已完成/失败/已删除的任务记录
func (c *Client) PurgeResult() error {
	return c.call("purgeDownloadResult", nil, nil)
}

// ===== 查询操作 =====

// TellStatus 获取指定任务状态
func (c *Client) TellStatus(gid string) (*StatusInfo, error) {
	var info StatusInfo
	if err := c.call("tellStatus", []any{gid}, &info); err != nil {
		return nil, err
	}
	return &info, nil
}

// TellActive 获取所有活动任务列表
func (c *Client) TellActive() ([]*StatusInfo, error) {
	var list []*StatusInfo
	if err := c.call("tellActive", nil, &list); err != nil {
		return nil, err
	}
	return list, nil
}

// TellWaiting 获取等待中任务列表（offset 起始位置，num 数量）
func (c *Client) TellWaiting(offset, num int) ([]*StatusInfo, error) {
	var list []*StatusInfo
	if err := c.call("tellWaiting", []any{offset, num}, &list); err != nil {
		return nil, err
	}
	return list, nil
}

// TellStopped 获取已停止任务列表（offset 起始位置，num 数量）
func (c *Client) TellStopped(offset, num int) ([]*StatusInfo, error) {
	var list []*StatusInfo
	if err := c.call("tellStopped", []any{offset, num}, &list); err != nil {
		return nil, err
	}
	return list, nil
}

// ===== 配置操作 =====

// ChangeGlobalOption 修改全局配置项
func (c *Client) ChangeGlobalOption(opts map[string]string) error {
	return c.call("changeGlobalOption", []any{opts}, nil)
}

// ChangeOption 修改指定任务配置项
func (c *Client) ChangeOption(gid string, opts map[string]string) error {
	return c.call("changeOption", []any{gid, opts}, nil)
}

// GetOption 获取指定任务配置项
func (c *Client) GetOption(gid string) (map[string]string, error) {
	var opts map[string]string
	if err := c.call("getOption", []any{gid}, &opts); err != nil {
		return nil, err
	}
	return opts, nil
}

// GetGlobalStat 获取全局下载统计
func (c *Client) GetGlobalStat() (*GlobalStat, error) {
	var stat GlobalStat
	if err := c.call("getGlobalStat", nil, &stat); err != nil {
		return nil, err
	}
	return &stat, nil
}

// GetVersion 获取 aria2 版本信息
func (c *Client) GetVersion() (*VersionInfo, error) {
	var info VersionInfo
	if err := c.call("getVersion", nil, &info); err != nil {
		return nil, err
	}
	return &info, nil
}
