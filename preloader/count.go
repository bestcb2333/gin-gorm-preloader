package preloader

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func CreateCountHandler[T any, U any, R any](
	cfg *Config[R],
	queryFunc func(q *gorm.DB, c *gin.Context, u *U, r *R) *gorm.DB,
) gin.HandlerFunc {
	return Preload(
		cfg,
		func(c *gin.Context, u *U, r *R) {
			query := cfg.Base.DB.Model(new(T))
			query = queryFunc(query, c, u, r)

			var total int64
			if err := query.Count(&total).Error; err != nil {
				c.JSON(500, cfg.Base.RespFunc("获取总数失败", err, nil))
				return
			}

			c.JSON(200, cfg.Base.RespFunc("数据查询成功", nil, total))
		},
	)
}
