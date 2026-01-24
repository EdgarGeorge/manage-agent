package server

import (
	"fmt"

	"gorm.io/gorm"
)

// 记录现值
type WorthModel struct {
	gorm.Model         // 内置模型结构体，包含 ID、CreatedAt、UpdatedAt、DeletedAt 字段
	Time       string  `gorm:"column:time; type:varchar(100); comment:现值时刻"`
	Cash       float64 `gorm:"column:cash; type:decimal(12,2); comment:现金"`
	StockA     float64 `gorm:"column:stock_a; type:decimal(12,2); comment:A股"`
	StockM     float64 `gorm:"column:stock_m; type:decimal(12,2); comment:M股"`
	Hongli     float64 `gorm:"column:hongli; type:decimal(12,2); comment:红利"`
	Bond       float64 `gorm:"column:bond; type:decimal(12,2); comment:债券"`
	Debt       float64 `gorm:"column:debt; type:decimal(12,2); comment:债权"`
	// 同步相关字段（用于数据同步）
	SyncStatus int `gorm:"column:sync_status; type:int; default:0; comment:同步状态 0-未同步 1-已同步" json:"sync_status,omitempty"`
}

// FlowRecordModel 资金流动记录表
type FlowRecordModel struct {
	gorm.Model         // 内置模型结构体，包含 ID、CreatedAt、UpdatedAt、DeletedAt 字段
	Time       string  `gorm:"column:time;type:varchar(100);comment:时间" json:"time"`
	Type       string  `gorm:"column:type;type:varchar(100);comment:类型" json:"type"`
	Value      float64 `gorm:"column:value;type:decimal(12,2);comment:金额" json:"value"`
	// 同步相关字段（用于数据同步）
	SyncStatus int `gorm:"column:sync_status; type:int; default:0; comment:同步状态 0-未同步 1-已同步" json:"sync_status,omitempty"`
}

func (w WorthModel) Create() error {
	result := Mysql.Create(&w)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (f FlowRecordModel) Create() error {
	result := Mysql.Create(&f)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (f FlowRecordModel) CreateInBatch(records []FlowRecordModel) error {
	// 确保所有记录的同步状态为未同步
	for i := range records {
		records[i].SyncStatus = 0
	}
	return Mysql.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&records).Error; err != nil {
			return err
		}
		return nil
	})
}

func (w WorthModel) GetLatestWorth() (*WorthModel, error) {
	var latestWorth WorthModel
	// Order("time DESC") 按自定义时间字段降序
	result := Mysql.Order("time DESC").Take(&latestWorth)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("未找到任何现值记录")
		}
		return nil, fmt.Errorf("查询最新现值记录失败：%v", result.Error)
	}
	return &latestWorth, nil
}

// 查询指定时间之间的所有记录（包含时间边界）
func (w WorthModel) QueryByTime(startDate, endDate string) ([]WorthModel, error) {
	var worths []WorthModel
	result := Mysql.Where("time >= ? AND time <= ?", startDate, endDate).Find(&worths)
	if result.Error != nil {
		return nil, result.Error
	}
	return worths, nil
}

// 查询指定时间之间的所有记录（包含时间边界）
func (f FlowRecordModel) QueryByTime(startDate, endDate string) ([]FlowRecordModel, error) {
	var flows []FlowRecordModel
	result := Mysql.Where("time >= ? AND time <= ?", startDate, endDate).Find(&flows)
	if result.Error != nil {
		return nil, result.Error
	}
	return flows, nil
}

type RequestRecordModel struct {
	gorm.Model

	RequestID          string `gorm:"column:request_id; type:string; size:100; index; not null; comment:请求ID;"`
	UserName           string `gorm:"column:username; type:string; size:100; index; not null; comment:用户名;"`
	RequestURL         string `gorm:"column:request_url; type:string; size:100; index; comment:请求URL;"`
	ClientIP           string `gorm:"column:client_ip; type:string; size:100; comment:请求IP;"`
	RequestMethod      string `gorm:"column:request_method; type:string; size:100; comment:请求方法;"`
	RequestBody        string `gorm:"column:request_body; type:string; size:1000; comment:请求Body;"`
	RequestQueryString string `gorm:"column:request_query_string; type:string; size:1000; comment:请求查询参数;"`
	ResponseStatus     string `gorm:"column:response_status; type:string; size:100; comment:响应状态;"`
	ResponseBody       string `gorm:"column:response_body; type:string; size:1000; comment:响应体;"`
	Latency            string `gorm:"column:latency; type:string; size:100; comment:耗时;"`
}

func (f *RequestRecordModel) Create() {
	Mysql.Create(&f)
}
