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
				// 新记录，直接插入（包括关联的类型价值记录）
				worth.SyncStatus = 1
				worth.ID = 0 // 重置ID，让数据库自动分配
				// 重置关联记录的ID
				for i := range worth.TypeWorths {
					worth.TypeWorths[i].ID = 0
				}
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
					// 先删除旧的关联记录
					if err := tx.Where("worth_id = ?", existing.ID).Delete(&TypeWorthModel{}).Error; err != nil {
						return fmt.Errorf("删除旧关联记录失败: %w", err)
					}
					// 重置关联记录的ID并设置外键
					for i := range worth.TypeWorths {
						worth.TypeWorths[i].ID = 0
						worth.TypeWorths[i].WorthID = existing.ID
					}
					// 更新主记录和关联记录
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

// GetLastSyncTime 获取最后同步时间（从元数据表或本地最新记录）
func GetLastSyncTime() time.Time {
	// 1. 优先从元数据表获取
	var metadata SyncMetadataModel
	if err := Mysql.First(&metadata).Error; err == nil && !metadata.LastSyncTime.IsZero() {
		return metadata.LastSyncTime
	}

	// 2. 如果元数据表没有记录，从本地最新记录获取（兼容旧逻辑）
	var lastSyncTime time.Time

	// 查询 WorthModel 的最新更新时间
	var latestWorth WorthModel
	if err := Mysql.Model(&WorthModel{}).Order("updated_at DESC").Take(&latestWorth).Error; err == nil {
		lastSyncTime = latestWorth.UpdatedAt
	}

	// 查询 FlowRecordModel 的最新更新时间
	var latestFlow FlowRecordModel
	if err := Mysql.Model(&FlowRecordModel{}).Order("updated_at DESC").Take(&latestFlow).Error; err == nil {
		if latestFlow.UpdatedAt.After(lastSyncTime) {
			lastSyncTime = latestFlow.UpdatedAt
		}
	}

	return lastSyncTime
}

// UpdateLastSyncTime 更新最后同步时间
func UpdateLastSyncTime(syncTime time.Time) error {
	var metadata SyncMetadataModel
	err := Mysql.First(&metadata).Error

	if err == gorm.ErrRecordNotFound {
		// 创建新记录
		metadata.LastSyncTime = syncTime
		return Mysql.Create(&metadata).Error
	} else if err != nil {
		return fmt.Errorf("查询同步元数据失败: %w", err)
	}

	// 更新现有记录
	metadata.LastSyncTime = syncTime
	return Mysql.Save(&metadata).Error
}

// GetCloudData 从云端获取数据（获取指定时间之后的数据，使用 >= 避免遗漏）
func GetCloudData(lastSyncTime time.Time) (*SyncData, error) {
	syncData := &SyncData{
		WorthRecords: []WorthModel{},
		FlowRecords:  []FlowRecordModel{},
		LastSyncTime: time.Now(),
	}

	// 获取现值记录（更新时间 >= lastSyncTime，使用 >= 避免遗漏时间戳相同的记录）
	// 需要预加载关联的类型价值数据（使用类型名，不关联 FType）
	var worths []WorthModel
	query := Mysql.Preload("TypeWorths").Where("updated_at >= ?", lastSyncTime)
	if err := query.Find(&worths).Error; err != nil {
		return nil, fmt.Errorf("查询现值记录失败: %w", err)
	}
	syncData.WorthRecords = worths

	// 获取流水记录（更新时间 >= lastSyncTime）
	var flows []FlowRecordModel
	query = Mysql.Where("updated_at >= ?", lastSyncTime)
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
// 修复：确保所有合并的记录都标记为已同步
func MergeDataToLocal(cloudData *SyncData) error {
	return Mysql.Transaction(func(tx *gorm.DB) error {
		// 1. 处理现值记录
		for _, worth := range cloudData.WorthRecords {
			var existing WorthModel
			err := tx.Where("time = ?", worth.Time).First(&existing).Error

			if err == gorm.ErrRecordNotFound {
				// 新记录，直接插入并标记为已同步（包括关联的类型价值记录）
				worth.SyncStatus = 1
				worth.ID = 0
				// 重置关联记录的ID
				for i := range worth.TypeWorths {
					worth.TypeWorths[i].ID = 0
				}
				if err := tx.Create(&worth).Error; err != nil {
					return fmt.Errorf("插入现值记录失败: %w", err)
				}
			} else if err != nil {
				return fmt.Errorf("查询现值记录失败: %w", err)
			} else {
				// 记录已存在，检查时间戳
				if existing.UpdatedAt.Before(worth.UpdatedAt) {
					// 云端版本更新，使用云端数据并标记为已同步
					worth.ID = existing.ID
					worth.SyncStatus = 1
					// 先删除旧的关联记录
					if err := tx.Where("worth_id = ?", existing.ID).Delete(&TypeWorthModel{}).Error; err != nil {
						return fmt.Errorf("删除旧关联记录失败: %w", err)
					}
					// 重置关联记录的ID并设置外键
					for i := range worth.TypeWorths {
						worth.TypeWorths[i].ID = 0
						worth.TypeWorths[i].WorthID = existing.ID
					}
					// 更新主记录和关联记录
					if err := tx.Save(&worth).Error; err != nil {
						return fmt.Errorf("更新现值记录失败: %w", err)
					}
				} else if existing.UpdatedAt.Equal(worth.UpdatedAt) {
					// 时间戳相同，确保标记为已同步（避免重复同步）
					if existing.SyncStatus != 1 {
						if err := tx.Model(&existing).Update("sync_status", 1).Error; err != nil {
							return fmt.Errorf("标记现值记录为已同步失败: %w", err)
						}
					}
				}
				// 如果本地版本更新，保留本地数据（不更新）
			}
		}

		// 2. 处理流水记录
		for _, flow := range cloudData.FlowRecords {
			var existing FlowRecordModel
			err := tx.Where("time = ? AND type = ? AND value = ?",
				flow.Time, flow.Type, flow.Value).First(&existing).Error

			if err == gorm.ErrRecordNotFound {
				// 新记录，直接插入并标记为已同步
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
					// 云端版本更新，使用云端数据并标记为已同步
					flow.ID = existing.ID
					flow.SyncStatus = 1
					if err := tx.Save(&flow).Error; err != nil {
						return fmt.Errorf("更新流水记录失败: %w", err)
					}
				} else if existing.UpdatedAt.Equal(flow.UpdatedAt) {
					// 时间戳相同，确保标记为已同步（避免重复同步）
					if existing.SyncStatus != 1 {
						if err := tx.Model(&existing).Update("sync_status", 1).Error; err != nil {
							return fmt.Errorf("标记流水记录为已同步失败: %w", err)
						}
					}
				}
				// 如果本地版本更新，保留本地数据（不更新）
			}
		}

		return nil
	})
}
