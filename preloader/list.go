package preloader

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type PageConfiger interface {
	GetPageConfig() PageConfig
}

type PageConfig struct {
	Page     int `form:"page"`
	PageSize int `form:"page_size"`
}

func (pc PageConfig) GetPageConfig() PageConfig {
	return pc
}

func CreateListHandler[T any, U any, R PageConfiger](
	cfg *Config,
	opt *Option,
	req *R,
	queryFunc func(query *gorm.DB, c *gin.Context, u *U, r *R) *gorm.DB,
) gin.HandlerFunc {
	opt.Bind = &BindOption{Query: true}
	return Preload(
		cfg,
		opt,
		req,
		func(c *gin.Context, u *U, r *R) {
			query := cfg.DB.Model(new(T))
			query = queryFunc(query, c, u, r)

			var total int64
			if err := query.Count(&total).Error; err != nil {
				cfg.Logger.Resp(c, 500, "获取总数失败", err, nil)
				return
			}

			pc := (*r).GetPageConfig()

			query = query.Limit(pc.PageSize).Offset((pc.Page - 1) * pc.PageSize)

			var data []T
			if err := query.Find(&data).Error; err != nil {
				cfg.Logger.Resp(c, 500, "数据查询失败", err, nil)
				return
			}

			cfg.Logger.Resp(c, 200, "数据查询成功", nil, gin.H{
				"data":  data,
				"total": total,
			})
		},
	)
}
