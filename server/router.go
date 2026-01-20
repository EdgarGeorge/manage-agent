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
