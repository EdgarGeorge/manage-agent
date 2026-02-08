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
		// 现值记录
		financeGroup.POST("/worth/", CreateWorthView)
		financeGroup.GET("/worth/", QueryWorthView)
		financeGroup.DELETE("/worth/:id", DeleteWorthView)
		// 流水记录
		financeGroup.POST("/flow-record/", CreateFlowRecordView)
		financeGroup.GET("/flow-record/", QueryFlowRecordView)
		financeGroup.DELETE("/flow-record/:id", DeleteFlowRecordView)
		// 资金类型管理（统一使用 enum 接口获取列表，ftype 接口用于增删改）
		financeGroup.GET("/enum/", FEnumView)
		financeGroup.POST("/ftype/", CreateFTypeView)
		financeGroup.PUT("/ftype/:id", UpdateFTypeView)
		financeGroup.DELETE("/ftype/:id", DeleteFTypeView)
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
