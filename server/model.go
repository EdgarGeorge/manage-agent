package server

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

// FTypeModel 资金类型表（支持动态添加/删除类型）
type FTypeModel struct {
	gorm.Model
	Name  string `gorm:"column:name; type:varchar(50); not null; unique; comment:类型名称（英文，如 cash、stock_a）" json:"name"`
	Cname string `gorm:"column:cname; type:varchar(50); not null; comment:类型中文名称（如 现金、A股）" json:"cname"`
	Order int    `gorm:"column:order; type:int; default:0; comment:排序顺序" json:"order"`
}

// WorthModel 现值主表
type WorthModel struct {
	gorm.Model                  // 内置字段：ID, CreatedAt, UpdatedAt, DeletedAt
	Time       string           `gorm:"column:time; type:varchar(100); comment:现值时刻"`
	TypeWorths []TypeWorthModel `gorm:"foreignKey:WorthID; comment:该现值下的各类型价值列表"`
	SyncStatus int              `gorm:"column:sync_status; type:int; default:0; comment:同步状态 0-未同步 1-已同步" json:"sync_status,omitempty"`
}

// TypeWorthModel 各类型价值表（关联到现值主表，使用类型名而非外键）
type TypeWorthModel struct {
	gorm.Model
	WorthID  uint    `gorm:"column:worth_id; not null; index; comment:关联的现值主表ID"`
	TypeName string  `gorm:"column:type_name; type:varchar(50); not null; index; comment:资金类型名称（如 cash、stock_a）"`
	Value    float64 `gorm:"column:value; type:decimal(16,2); not null; default:0; comment:价值数值"`
}

// FlowRecordModel 资金流动记录表
type FlowRecordModel struct {
	gorm.Model         // 内置模型结构体，包含 ID、CreatedAt、UpdatedAt、DeletedAt 字段
	Time       string  `gorm:"column:time;type:varchar(100);comment:时间" json:"time"`
	Type       string  `gorm:"column:type;type:varchar(100);comment:类型" json:"type"`
	Value      float64 `gorm:"column:value;type:decimal(12,2);comment:金额" json:"value"`
	SyncStatus int     `gorm:"column:sync_status; type:int; default:0; comment:同步状态 0-未同步 1-已同步" json:"sync_status,omitempty"`
}

// Create 创建现值记录（包含关联的类型价值记录）
func (w WorthModel) Create() error {
	return Mysql.Transaction(func(tx *gorm.DB) error {
		// 先保存关联记录
		typeWorths := w.TypeWorths

		// 创建主记录时，使用 Select 只创建主记录字段，避免自动创建关联
		if err := tx.Omit("TypeWorths").Create(&w).Error; err != nil {
			return err
		}

		// 手动创建关联的类型价值记录
		if len(typeWorths) > 0 {
			for i := range typeWorths {
				typeWorths[i].WorthID = w.ID
				// 确保 ID 为 0，让 GORM 自动生成
				typeWorths[i].ID = 0
			}
			if err := tx.Create(&typeWorths).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// GetWithTypes 获取现值记录并加载关联的类型价值
func (w WorthModel) GetWithTypes() (*WorthModel, error) {
	var worth WorthModel
	if err := Mysql.Preload("TypeWorths").Where("id = ?", w.ID).First(&worth).Error; err != nil {
		return nil, err
	}
	return &worth, nil
}

// Query 分页查询现值记录（包含关联数据）
func (w WorthModel) Query(pageIndex, pageLimit int) ([]WorthModel, error) {
	var worthList []WorthModel
	offset := (pageIndex - 1) * pageLimit
	err := Mysql.Preload("TypeWorths").Order("time DESC").Offset(offset).Limit(pageLimit).Find(&worthList).Error
	return worthList, err
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

// Query 分页查询流水记录
func (f FlowRecordModel) Query(pageIndex, pageLimit int) ([]FlowRecordModel, error) {
	var recordList []FlowRecordModel
	offset := (pageIndex - 1) * pageLimit
	err := Mysql.Order("time DESC").Offset(offset).Limit(pageLimit).Find(&recordList).Error
	return recordList, err
}

// GetLatestWorth 获取最新的现值记录（包含关联的类型价值）
func (w WorthModel) GetLatestWorth() (*WorthModel, error) {
	var latestWorth WorthModel
	// Order("time DESC") 按自定义时间字段降序，并预加载关联的类型价值
	result := Mysql.Preload("TypeWorths").Order("time DESC").Take(&latestWorth)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("未找到任何现值记录")
		}
		return nil, fmt.Errorf("查询最新现值记录失败：%v", result.Error)
	}
	return &latestWorth, nil
}

// QueryByTime 查询指定时间之间的所有记录（包含时间边界，并加载关联的类型价值）
func (w WorthModel) QueryByTime(startDate, endDate string) ([]WorthModel, error) {
	var worths []WorthModel
	result := Mysql.Preload("TypeWorths").Where("time >= ? AND time <= ?", startDate, endDate).
		Order("time ASC").Find(&worths)
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

// SyncMetadataModel 同步元数据表（记录最后同步时间）
type SyncMetadataModel struct {
	gorm.Model
	LastSyncTime time.Time `gorm:"column:last_sync_time; type:datetime; comment:最后同步时间"`
}

func (f *RequestRecordModel) Create() {
	Mysql.Create(&f)
}

// FTypeModel 相关方法

// GetAllTypes 获取所有未删除的资金类型（按排序顺序）
func (f FTypeModel) GetAllTypes() ([]FTypeModel, error) {
	var types []FTypeModel
	if err := Mysql.Order("`order` ASC, id ASC").Find(&types).Error; err != nil {
		return nil, err
	}
	return types, nil
}

// InitDefaultTypes 初始化默认资金类型（确保 cash 类型存在）
func InitDefaultTypes() error {
	var existing FTypeModel
	err := Mysql.Where("name = ?", "cash").First(&existing).Error

	if err == gorm.ErrRecordNotFound {
		// cash 类型不存在，创建它
		cashType := FTypeModel{
			Name:  "cash",
			Cname: "现金",
			Order: 0, // 现金排在第一位
		}
		if err := cashType.Create(); err != nil {
			return fmt.Errorf("创建默认 cash 类型失败: %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("查询 cash 类型失败: %w", err)
	}
	// 如果已存在，不做任何操作

	return nil
}

// GetTypeByName 根据名称获取资金类型
func (f FTypeModel) GetTypeByName(name string) (*FTypeModel, error) {
	var ftype FTypeModel
	if err := Mysql.Where("name = ?", name).First(&ftype).Error; err != nil {
		return nil, err
	}
	return &ftype, nil
}

// Create 创建资金类型
func (f FTypeModel) Create() error {
	return Mysql.Create(&f).Error
}

// Update 更新资金类型
func (f FTypeModel) Update() error {
	return Mysql.Save(&f).Error
}

// Delete 删除资金类型（软删除）
func (f FTypeModel) Delete() error {
	return Mysql.Delete(&f).Error
}
