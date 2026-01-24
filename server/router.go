package server

import (
	"github.com/gin-gonic/gin"
)

func RegisterFinanceRouter(r *gin.RouterGroup) {
	financeGroup := r.Group("/finance")
	{
		// 利润相关
		financeGroup.GET("/profit/", QyeryProfitView)
		financeGroup.GET("/profit/history/", QueryHistoryProfitView)
		financeGroup.GET("/enum/", FEnumView)
		// 现值记录
		financeGroup.POST("/worth/", CreateWorthView)
		// 流水记录
		financeGroup.POST("/flow-record/", CreateFlowRecordView)
	}
}

// RegisterSyncRouter 注册同步路由（用于云端服务器）
func RegisterSyncRouter(r *gin.RouterGroup) {
	syncGroup := r.Group("/sync")
	{
		// 上传本地数据到云端
		syncGroup.POST("/upload", SyncUploadView)
		// 从云端下载数据
		syncGroup.GET("/download", SyncDownloadView)
	}
}
