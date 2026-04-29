package aria2

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

// setupMockServer 创建模拟 aria2 JSON-RPC 服务器的 httptest.Server
func setupMockServer(t *testing.T, handler http.HandlerFunc) (*httptest.Server, *Client) {
	t.Helper()
	ts := httptest.NewServer(handler)

	client := &Client{
		url:    ts.URL + "/jsonrpc",
		secret: "testsecret",
		http:   ts.Client(),
	}
	return ts, client
}

// 测试：RPC 请求格式正确（jsonrpc 2.0、method 前缀、token:secret 认证）
func TestClient_CallRequestFormat(t *testing.T) {
	ts, client := setupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		var req RpcRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("解析请求失败: %v", err)
			return
		}

		// 验证 JSON-RPC 版本
		if req.Jsonrpc != "2.0" {
			t.Errorf("jsonrpc 应为 2.0, got=%s", req.Jsonrpc)
		}

		// 验证 method 前缀
		if !strings.HasPrefix(req.Method, "aria2.") {
			t.Errorf("method 应以 aria2. 开头, got=%s", req.Method)
		}

		// 验证 token:secret 认证
		if len(req.Params) == 0 || req.Params[0] != "token:testsecret" {
			t.Errorf("首个参数应为 token:testsecret, got=%v", req.Params)
		}

		// 验证 ID 不为空
		if req.Id == "" {
			t.Error("请求 ID 不应为空")
		}

		// 返回成功响应
		json.NewEncoder(w).Encode(RpcResponse{
			Jsonrpc: "2.0",
			Result:  "test-result",
			Id:      req.Id,
		})
	})
	defer ts.Close()

	// 触发一次 RPC 调用
	var result string
	err := client.call("testMethod", nil, &result)
	if err != nil {
		t.Fatalf("RPC 调用失败: %v", err)
	}
	if result != "test-result" {
		t.Errorf("结果应为 test-result, got=%s", result)
	}
}

// 测试：RPC 正常响应解析
func TestClient_CallSuccessResponse(t *testing.T) {
	ts, client := setupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		var req RpcRequest
		json.NewDecoder(r.Body).Decode(&req)

		json.NewEncoder(w).Encode(RpcResponse{
			Jsonrpc: "2.0",
			Result:  "b469d2d63f9a8b3e",
			Id:      req.Id,
		})
	})
	defer ts.Close()

	var gid string
	err := client.call("addUri", nil, &gid)
	if err != nil {
		t.Fatalf("RPC 调用失败: %v", err)
	}
	if gid != "b469d2d63f9a8b3e" {
		t.Errorf("GID 应为 b469d2d63f9a8b3e, got=%s", gid)
	}
}

// 测试：RPC 错误响应处理
func TestClient_CallErrorResponse(t *testing.T) {
	ts, client := setupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		var req RpcRequest
		json.NewDecoder(r.Body).Decode(&req)

		json.NewEncoder(w).Encode(RpcResponse{
			Jsonrpc: "2.0",
			Error: &RpcError{
				Code:    1,
				Message: "GID not found",
			},
			Id: req.Id,
		})
	})
	defer ts.Close()

	err := client.call("tellStatus", nil, nil)
	if err == nil {
		t.Fatal("RPC 错误时应返回 error")
	}
	if !strings.Contains(err.Error(), "aria2 错误") {
		t.Errorf("错误信息应包含 aria2 错误, got=%v", err)
	}
	if !strings.Contains(err.Error(), "GID not found") {
		t.Errorf("错误信息应包含原始错误, got=%v", err)
	}
}

// 测试：RPC 连接超时
func TestClient_CallTimeout(t *testing.T) {
	// 使用 TEST-NET 不可路由地址触发连接超时
	client := &Client{
		url:    "http://192.0.2.1:12345/jsonrpc",
		secret: "testsecret",
		http:   &http.Client{Timeout: 1 * time.Millisecond},
	}

	err := client.call("test", nil, nil)
	if err == nil {
		t.Fatal("超时应返回错误")
	}
}

// 测试：AddURI 方法正确调用 aria2.addUri
func TestClient_AddURI(t *testing.T) {
	ts, client := setupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		var req RpcRequest
		json.NewDecoder(r.Body).Decode(&req)

		if req.Method != "aria2.addUri" {
			t.Errorf("method 应为 aria2.addUri, got=%s", req.Method)
		}

		json.NewEncoder(w).Encode(RpcResponse{
			Jsonrpc: "2.0",
			Result:  "abc123",
			Id:      req.Id,
		})
	})
	defer ts.Close()

	gid, err := client.AddURI([]string{"http://example.com/file.iso"}, nil)
	if err != nil {
		t.Fatalf("AddURI 失败: %v", err)
	}
	if gid != "abc123" {
		t.Errorf("GID 应为 abc123, got=%s", gid)
	}
}

// 测试：所有 RPC 方法映射正确
func TestClient_AllMethodNames(t *testing.T) {
	// 捕获所有调用的 method 名称
	var methods []string

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req RpcRequest
		json.NewDecoder(r.Body).Decode(&req)
		methods = append(methods, req.Method)

		// 根据方法返回合适的响应类型
		result := any("ok")
		if strings.Contains(req.Method, "tellActive") || strings.Contains(req.Method, "tellWaiting") || strings.Contains(req.Method, "tellStopped") {
			result = []StatusInfo{}
		} else if strings.Contains(req.Method, "getVersion") {
			result = VersionInfo{Version: "1.36.0"}
		} else if strings.Contains(req.Method, "getGlobalStat") {
			result = GlobalStat{}
		} else if strings.Contains(req.Method, "getOption") {
			result = map[string]string{}
		} else if strings.Contains(req.Method, "tellStatus") {
			result = StatusInfo{Gid: "test"}
		}

		json.NewEncoder(w).Encode(RpcResponse{
			Jsonrpc: "2.0",
			Result:  result,
			Id:      req.Id,
		})
	}))
	defer ts.Close()

	client := &Client{
		url:    ts.URL + "/jsonrpc",
		secret: "secret",
		http:   ts.Client(),
	}

	// 调用所有公开方法
	client.AddURI([]string{"http://x.com"}, nil)
	tmpFile, _ := os.CreateTemp("", "test-torrent-*.torrent")
	tmpFile.Write([]byte("dummy-data"))
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())
	client.AddTorrent(tmpFile.Name())
	client.AddMetalink("http://x.com/metalink")
	client.Pause("gid1")
	client.PauseAll()
	client.Resume("gid1")
	client.ResumeAll()
	client.Remove("gid1")
	client.RemoveResult("gid1")
	client.PurgeResult()
	client.TellStatus("gid1")
	client.TellActive()
	client.TellWaiting(0, 10)
	client.TellStopped(0, 10)
	client.ChangeGlobalOption(map[string]string{"max-overall-download-limit": "1M"})
	client.ChangeOption("gid1", map[string]string{"max-download-limit": "1M"})
	client.GetOption("gid1")
	client.GetGlobalStat()
	client.GetVersion()

	expected := []string{
		"aria2.addUri",
		"aria2.addTorrent",
		"aria2.addMetalink",
		"aria2.pause",
		"aria2.pauseAll",
		"aria2.resume",
		"aria2.resumeAll",
		"aria2.remove",
		"aria2.removeDownloadResult",
		"aria2.purgeDownloadResult",
		"aria2.tellStatus",
		"aria2.tellActive",
		"aria2.tellWaiting",
		"aria2.tellStopped",
		"aria2.changeGlobalOption",
		"aria2.changeOption",
		"aria2.getOption",
		"aria2.getGlobalStat",
		"aria2.getVersion",
	}

	if len(methods) != len(expected) {
		t.Errorf("方法数量不匹配: got=%d, want=%d", len(methods), len(expected))
	}

	for i, exp := range expected {
		if i >= len(methods) {
			t.Errorf("缺少方法: %s", exp)
			continue
		}
		if methods[i] != exp {
			t.Errorf("方法[%d] 应为 %s, got=%s", i, exp, methods[i])
		}
	}
}

// 测试：响应中 result 为 null 时不报错
func TestClient_CallNullResult(t *testing.T) {
	ts, client := setupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		var req RpcRequest
		json.NewDecoder(r.Body).Decode(&req)

		// result 为 null
		resp := map[string]any{
			"jsonrpc": "2.0",
			"result":  nil,
			"id":      req.Id,
		}
		json.NewEncoder(w).Encode(resp)
	})
	defer ts.Close()

	err := client.Pause("test")
	if err != nil {
		t.Fatalf("result 为 null 时应成功: %v", err)
	}
}
