package server

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SyncData 同步数据结构
type SyncData struct {
	WorthRecords []WorthModel      `json:"worth_records"`
	FlowRecords  []FlowRecordModel `json:"flow_records"`
	LastSyncTime time.Time         `json:"last_sync_time"`
}

// SyncUploadView 上传本地数据到云端
func SyncUploadView(c *gin.Context) {
	var syncData SyncData
	if err := c.ShouldBindJSON(&syncData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"msg":  "请求参数错误: " + err.Error(),
			"code": http.StatusBadRequest,
		})
		return
	}

	// 合并数据到云端数据库
	err := MergeDataToCloud(syncData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"msg":  "同步失败: " + err.Error(),
			"code": http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"msg":  "同步成功",
		"code": http.StatusOK,
	})
}

// SyncDownloadView 从云端下载数据
func SyncDownloadView(c *gin.Context) {
	lastSyncTimeStr := c.DefaultQuery("last_sync_time", "")

	var lastSyncTime time.Time
	var err error

	if lastSyncTimeStr != "" {
		lastSyncTime, err = time.Parse(time.RFC3339, lastSyncTimeStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"msg":  "时间格式错误，请使用 RFC3339 格式",
				"code": http.StatusBadRequest,
			})
			return
		}
	}

	data, err := GetCloudData(lastSyncTime)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"msg":  "获取数据失败: " + err.Error(),
			"code": http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"msg":  "success",
		"code": http.StatusOK,
		"data": data,
	})
}

// MergeDataToCloud 合并数据到云端（冲突解决策略：时间戳优先）
func MergeDataToCloud(syncData SyncData) error {
	return Mysql.Transaction(func(tx *gorm.DB) error {
		// 1. 处理现值记录
		for _, worth := range syncData.WorthRecords {
			var existing WorthModel
			// 根据时间查找是否已存在
			err := tx.Where("time = ?", worth.Time).First(&existing).Error

			if err == gorm.ErrRecordNotFound {
				// 新记录，直接插入
				worth.SyncStatus = 1
				worth.ID = 0 // 重置ID，让数据库自动分配
				if err := tx.Create(&worth).Error; err != nil {
					return fmt.Errorf("插入现值记录失败: %w", err)
				}
			} else if err != nil {
				return fmt.Errorf("查询现值记录失败: %w", err)
			} else {
				// 记录已存在，检查时间戳决定是否更新
				if existing.UpdatedAt.Before(worth.UpdatedAt) {
					// 本地版本更新，使用本地数据
					worth.ID = existing.ID
					worth.SyncStatus = 1
					if err := tx.Save(&worth).Error; err != nil {
						return fmt.Errorf("更新现值记录失败: %w", err)
					}
				}
				// 否则保留云端版本（不更新）
			}
		}

		// 2. 处理流水记录
		for _, flow := range syncData.FlowRecords {
			var existing FlowRecordModel
			// 根据时间和类型查找是否已存在（同一时间同一类型的记录视为同一笔）
			err := tx.Where("time = ? AND type = ? AND value = ?",
				flow.Time, flow.Type, flow.Value).First(&existing).Error

			if err == gorm.ErrRecordNotFound {
				// 新记录，直接插入
				flow.SyncStatus = 1
				flow.ID = 0 // 重置ID
				if err := tx.Create(&flow).Error; err != nil {
					return fmt.Errorf("插入流水记录失败: %w", err)
				}
			} else if err != nil {
				return fmt.Errorf("查询流水记录失败: %w", err)
			} else {
				// 记录已存在，检查时间戳
				if existing.UpdatedAt.Before(flow.UpdatedAt) {
					// 本地版本更新
					flow.ID = existing.ID
					flow.SyncStatus = 1
					if err := tx.Save(&flow).Error; err != nil {
						return fmt.Errorf("更新流水记录失败: %w", err)
					}
				}
			}
		}

		return nil
	})
}

// GetCloudData 从云端获取数据（获取指定时间之后的数据）
func GetCloudData(lastSyncTime time.Time) (*SyncData, error) {
	syncData := &SyncData{
		WorthRecords: []WorthModel{},
		FlowRecords:  []FlowRecordModel{},
		LastSyncTime: time.Now(),
	}

	// 获取现值记录（更新时间晚于 lastSyncTime）
	var worths []WorthModel
	query := Mysql.Where("updated_at > ?", lastSyncTime)
	if err := query.Find(&worths).Error; err != nil {
		return nil, fmt.Errorf("查询现值记录失败: %w", err)
	}
	syncData.WorthRecords = worths

	// 获取流水记录（更新时间晚于 lastSyncTime）
	var flows []FlowRecordModel
	query = Mysql.Where("updated_at > ?", lastSyncTime)
	if err := query.Find(&flows).Error; err != nil {
		return nil, fmt.Errorf("查询流水记录失败: %w", err)
	}
	syncData.FlowRecords = flows

	return syncData, nil
}

// GetUnsyncedData 获取本地未同步的数据
func GetUnsyncedData() (*SyncData, error) {
	syncData := &SyncData{
		WorthRecords: []WorthModel{},
		FlowRecords:  []FlowRecordModel{},
		LastSyncTime: time.Now(),
	}

	// 获取未同步的现值记录
	var worths []WorthModel
	if err := Mysql.Where("sync_status = 0 OR sync_status IS NULL").Find(&worths).Error; err != nil {
		return nil, fmt.Errorf("查询未同步现值记录失败: %w", err)
	}
	syncData.WorthRecords = worths

	// 获取未同步的流水记录
	var flows []FlowRecordModel
	if err := Mysql.Where("sync_status = 0 OR sync_status IS NULL").Find(&flows).Error; err != nil {
		return nil, fmt.Errorf("查询未同步流水记录失败: %w", err)
	}
	syncData.FlowRecords = flows

	return syncData, nil
}

// MarkDataAsSynced 标记数据为已同步
func MarkDataAsSynced(worthIDs []uint, flowIDs []uint) error {
	return Mysql.Transaction(func(tx *gorm.DB) error {
		if len(worthIDs) > 0 {
			if err := tx.Model(&WorthModel{}).
				Where("id IN ?", worthIDs).
				Update("sync_status", 1).Error; err != nil {
				return fmt.Errorf("标记现值记录失败: %w", err)
			}
		}

		if len(flowIDs) > 0 {
			if err := tx.Model(&FlowRecordModel{}).
				Where("id IN ?", flowIDs).
				Update("sync_status", 1).Error; err != nil {
				return fmt.Errorf("标记流水记录失败: %w", err)
			}
		}

		return nil
	})
}

// MergeDataToLocal 合并云端数据到本地（冲突解决策略：时间戳优先）
func MergeDataToLocal(cloudData *SyncData) error {
	return Mysql.Transaction(func(tx *gorm.DB) error {
		// 1. 处理现值记录
		for _, worth := range cloudData.WorthRecords {
			var existing WorthModel
			err := tx.Where("time = ?", worth.Time).First(&existing).Error

			if err == gorm.ErrRecordNotFound {
				// 新记录，直接插入
				worth.SyncStatus = 1
				worth.ID = 0
				if err := tx.Create(&worth).Error; err != nil {
					return fmt.Errorf("插入现值记录失败: %w", err)
				}
			} else if err != nil {
				return fmt.Errorf("查询现值记录失败: %w", err)
			} else {
				// 记录已存在，检查时间戳
				if existing.UpdatedAt.Before(worth.UpdatedAt) {
					// 云端版本更新，使用云端数据
					worth.ID = existing.ID
					worth.SyncStatus = 1
					if err := tx.Save(&worth).Error; err != nil {
						return fmt.Errorf("更新现值记录失败: %w", err)
					}
				}
			}
		}

		// 2. 处理流水记录
		for _, flow := range cloudData.FlowRecords {
			var existing FlowRecordModel
			err := tx.Where("time = ? AND type = ? AND value = ?",
				flow.Time, flow.Type, flow.Value).First(&existing).Error

			if err == gorm.ErrRecordNotFound {
				// 新记录，直接插入
				flow.SyncStatus = 1
				flow.ID = 0
				if err := tx.Create(&flow).Error; err != nil {
					return fmt.Errorf("插入流水记录失败: %w", err)
				}
			} else if err != nil {
				return fmt.Errorf("查询流水记录失败: %w", err)
			} else {
				// 记录已存在，检查时间戳
				if existing.UpdatedAt.Before(flow.UpdatedAt) {
					// 云端版本更新
					flow.ID = existing.ID
					flow.SyncStatus = 1
					if err := tx.Save(&flow).Error; err != nil {
						return fmt.Errorf("更新流水记录失败: %w", err)
					}
				}
			}
		}

		return nil
	})
}
