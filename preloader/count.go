package preloader

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func CreateCountHandler[T any, U any, R any](
	cfg *Config,
	opt *Option,
	req *R,
	queryFunc func(q *gorm.DB, c *gin.Context, u *U, r *R) *gorm.DB,
) gin.HandlerFunc {
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

			cfg.Logger.Resp(c, 200, "数据查询成功", nil, total)
		},
	)
}
