package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// SyncClient 同步客户端
type SyncClient struct {
	CloudAPIURL string       // 云端API地址，例如 "http://your-server.com:8080"
	HTTPClient  *http.Client // HTTP客户端
}

// NewSyncClient 创建同步客户端
func NewSyncClient(cloudAPIURL string) *SyncClient {
	if cloudAPIURL == "" {
		return nil
	}
	return &SyncClient{
		CloudAPIURL: cloudAPIURL,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// SyncToCloud 同步本地数据到云端
func (sc *SyncClient) SyncToCloud() error {
	if sc == nil {
		return fmt.Errorf("同步客户端未初始化")
	}

	// 1. 获取未同步的数据
	localData, err := GetUnsyncedData()
	if err != nil {
		return fmt.Errorf("获取未同步数据失败: %w", err)
	}

	// 如果没有数据需要同步，直接返回
	if len(localData.WorthRecords) == 0 && len(localData.FlowRecords) == 0 {
		return nil
	}

	// 2. 序列化数据
	jsonData, err := json.Marshal(localData)
	if err != nil {
		return fmt.Errorf("序列化数据失败: %w", err)
	}

	// 3. 发送POST请求到云端
	url := fmt.Sprintf("%s/api/v1/sync/upload", sc.CloudAPIURL)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := sc.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("请求云端失败: %w", err)
	}
	defer resp.Body.Close()

	// 4. 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取响应失败: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("云端返回错误: %s (状态码: %d)", string(body), resp.StatusCode)
	}

	// 5. 标记本地数据为已同步
	var worthIDs []uint
	var flowIDs []uint
	for _, worth := range localData.WorthRecords {
		worthIDs = append(worthIDs, worth.ID)
	}
	for _, flow := range localData.FlowRecords {
		flowIDs = append(flowIDs, flow.ID)
	}

	if err := MarkDataAsSynced(worthIDs, flowIDs); err != nil {
		// 即使标记失败，也不影响同步，只记录错误
		logger.Warnf("标记数据为已同步失败: %v", err)
	}

	return nil
}

// SyncFromCloud 从云端同步数据到本地
func (sc *SyncClient) SyncFromCloud(lastSyncTime time.Time) error {
	if sc == nil {
		return fmt.Errorf("同步客户端未初始化")
	}

	// 1. 构建请求URL
	url := fmt.Sprintf("%s/api/v1/sync/download", sc.CloudAPIURL)
	if !lastSyncTime.IsZero() {
		url += "?last_sync_time=" + lastSyncTime.Format(time.RFC3339)
	}

	// 2. 发送GET请求
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}

	resp, err := sc.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("请求云端失败: %w", err)
	}
	defer resp.Body.Close()

	// 3. 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取响应失败: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("云端返回错误: %s (状态码: %d)", string(body), resp.StatusCode)
	}

	// 4. 解析响应
	var response struct {
		Code int      `json:"code"`
		Msg  string   `json:"msg"`
		Data SyncData `json:"data"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return fmt.Errorf("解析响应失败: %w", err)
	}

	// 5. 合并数据到本地
	if len(response.Data.WorthRecords) > 0 || len(response.Data.FlowRecords) > 0 {
		if err := MergeDataToLocal(&response.Data); err != nil {
			return fmt.Errorf("合并数据到本地失败: %w", err)
		}
	}

	return nil
}

// FullSync 完整同步（先上传，再下载）
// 修复后的逻辑：
// 1. 在上传前获取最后同步时间（避免上传过程中的边界情况）
// 2. 上传本地未同步数据到云端
// 3. 从云端下载该时间点之后的所有更新
// 4. 更新最后同步时间为当前时间
func (sc *SyncClient) FullSync() error {
	// 0. 在上传前获取最后同步时间（关键修复：避免遗漏数据）
	lastSyncTime := GetLastSyncTime()

	// 1. 先上传本地数据到云端
	if err := sc.SyncToCloud(); err != nil {
		return fmt.Errorf("上传到云端失败: %w", err)
	}

	// 2. 从云端下载数据（使用上传前的时间，确保不遗漏）
	if err := sc.SyncFromCloud(lastSyncTime); err != nil {
		return fmt.Errorf("从云端下载失败: %w", err)
	}

	// 3. 更新最后同步时间为当前时间
	if err := UpdateLastSyncTime(time.Now()); err != nil {
		// 即使更新失败，也不影响同步，只记录警告
		logger.Warnf("更新最后同步时间失败: %v", err)
	}

	return nil
}
