package dm

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	fgorm "github.com/lzw5399/go-common-public/library/database/gorm"
	dbmodel "github.com/lzw5399/go-common-public/library/database/model"
	"github.com/lzw5399/go-common-public/library/database/repo"
)

func init() {
	Start(
		dbmodel.WithTenantTables(&HostApp{}),
	)
}

type HostAppStatus int32
type Platform int32
type BooleanValue int32

type HostApp struct {
	Id            uint64        `json:"id" gorm:"primary_key;column:id;type:BIGINT AUTO_INCREMENT;comment:'自增id'" sql:"auto_increment;primary_key"`
	HostAppId     string        `json:"hostAppId" gorm:"column:host_app_id;unique;type:varchar(64);comment:'宿主应用的id'"`                                                                // 宿主应用的id
	Name          string        `json:"name" gorm:"column:name;type:varchar(64);NOT NULL;default:'';comment:'应用名称'"`                                                                  // 应用名称
	DevOrganId    string        `json:"devOrganId" gorm:"column:dev_organ_id;type:varchar(64);index:idx_host_app_dev_organ_id;NOT NULL;default:'';comment:'开发端企业ID'"`                 // 开发端企业ID
	Desc          string        `json:"desc" gorm:"column:desc;type:varchar(1024);NOT NULL;default:'';comment:'应用描述'"`                                                                // 应用描述
	Owner         string        `json:"owner" gorm:"column:owner;type:varchar(64);NOT NULL;default:'';comment:'所属企业'"`                                                                // 所属企业
	Logo          string        `json:"logo" gorm:"column:logo;type:varchar(256);NOT NULL;default:'';comment:'应用图标网盘id'"`                                                             // 应用图标网盘id
	Expire        int64         `json:"expire" gorm:"column:expire;type:BIGINT;NOT NULL;default:0;comment:''"`                                                                        // 过期时间
	ApiServer     string        `json:"apiServer" gorm:"column:api_server;type:varchar(256);NOT NULL;default:'';comment:'子域名'"`                                                       // 子域名
	PlatForm      Platform      `json:"platform" gorm:"column:platform;type:smallint;NOT NULL;default:0;comment:'来源平台'"`                                                              // 来源平台
	StatusValue   HostAppStatus `json:"statusValue" gorm:"column:status_value;type:smallint;NOT NULL;default:0;comment:'应用状态'"`                                                       // 应用状态
	AutoBind      BooleanValue  `json:"autoBind" gorm:"column:auto_bind;type:smallint;NOT NULL;default:0;comment:'运营端创建专用.设置自动关联后，开发端所有的小程序都会默认关联该应用'"`                               // 设置自动关联后，企业端所有的小程序都会默认关联该应用
	HostAppPublic BooleanValue  `json:"hostAppPublic" gorm:"column:host_app_public;type:smallint;NOT NULL;default:0;comment:'是否设置为公开'"`                                               // 设置为公开后
	BundlePublic  BooleanValue  `json:"bundlePublic" gorm:"column:bundle_public;type:smallint;NOT NULL;default:0;comment:'共享给其他企业的场景下，开发端展示该应用 BundlelD、SDK Key、Sercret 等 SDK 集成信息'"` // 共享给其他企业的场景下，开发端展示该应用 BundlelD、SDK Key、Sercret 等 SDK 集成信息
	AuditableBase
	TenantBase
}

func (HostApp) TableName() string {
	return "host_app"
}

type TenantBase struct {
	OrganId string `json:"organId" bson:"organId" gorm:"column:organ_id;type:varchar(48);NOT NULL;default:'';comment:'企业id'"` // 租户id。 默认会创建索引。见shared/dbutil/migration.go
}

type AuditableBase struct {
	CreateBy   string `json:"createBy" bson:"create_by" gorm:"column:create_by;type:varchar(48);NOT NULL;default:'';comment:创建人"`
	UpdateBy   string `json:"updateBy" bson:"update_by" gorm:"column:update_by;type:varchar(48);NOT NULL;default:'';comment:更新人"`
	CreateTime int64  `json:"createTime" bson:"create_time" gorm:"column:create_time;type:BIGINT;NOT NULL;default:0;comment:创建时间"`
	UpdateTime int64  `json:"updateTime" bson:"update_time" gorm:"column:update_time;type:BIGINT;NOT NULL;default:0;comment:更新时间"`
}

func TestAutoMigrate(t *testing.T) {
	var err error
	err = fgorm.DB.Debug().AutoMigrate(&HostApp{})
	if err != nil {
		fmt.Printf("Error: failed to AutoMigrate: %v\n", err)
		return
	}
}

func TestCreate(t *testing.T) {
	err := fgorm.DB.Model(&HostApp{}).Debug().Create(&HostApp{
		HostAppId:     "2",
		Name:          "1",
		DevOrganId:    "1",
		Desc:          "1",
		Owner:         "1",
		Logo:          "1",
		Expire:        1,
		ApiServer:     "1",
		PlatForm:      1,
		StatusValue:   1,
		AutoBind:      1,
		HostAppPublic: 1,
		BundlePublic:  1,
		AuditableBase: AuditableBase{
			CreateBy:   "1",
			UpdateBy:   "1",
			CreateTime: 1,
			UpdateTime: 1,
		},
	}).Error

	if err != nil {
		fmt.Printf("Error: failed to Create: %v\n", err)
		return
	}
}

func TestSelect(t *testing.T) {

	ctx, option := repo.MergeRepoOption(context.Background(), repo.IgnoreTenant(false))
	tx := repo.GetGormTx(ctx, option)

	var countTest int64
	err := tx.Table("host_app").Where("status_value in ?", []HostAppStatus{4}).Count(&countTest).Error

	print(countTest)
	if err != nil {
		fmt.Printf("Error: failed to Where: %v\n", err)
		return
	}
}

func TestSelectAll(t *testing.T) {
	var data []HostApp
	err := fgorm.DB.Model(&HostApp{}).Debug().Find(&data).Error

	print(data)
	if err != nil {
		fmt.Printf("Error: failed to Where: %v\n", err)
		return
	}
}

func print(obj interface{}) {
	b, _ := json.Marshal(obj)
	fmt.Println(string(b))
}

func MergeRepoOption(ctx context.Context, opts ...RepoOptionFunc) (context.Context, *RepoOption) {
	option := &RepoOption{}
	for _, opt := range opts {
		opt(option)
	}

	ctx = context.WithValue(ctx, ignoreTenantKey{}, true)

	return ctx, option
}

type RepoOptionFunc func(*RepoOption)

type RepoOption struct {
	ignoreTenant bool // 是否忽略租户限制
	ignoreI18n   bool // 是否在gorm查询的时候，跳过i18n针对某些字段的自动翻译
}

type ignoreTenantKey struct{}
