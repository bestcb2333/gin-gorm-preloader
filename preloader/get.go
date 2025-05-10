package preloader

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func CreateGetHandler[T any, U any](
	cfg *Config[struct{}],
	queryFunc func(q *gorm.DB, c *gin.Context, u *U) *gorm.DB,
) gin.HandlerFunc {
	cfg.Bind = &BindConfig{Query: true}
	return Preload(
		cfg,
		func(c *gin.Context, u *U, r *struct{}) {

			q := cfg.Base.DB.Model(new(T))

			q = queryFunc(q, c, u)
			if q == nil {
				return
			}

			id := c.Param("id")

			var data T
			if err := q.First(&data, "id = ?", id).Error; err != nil {
				c.JSON(500, cfg.Base.RespFunc("查询失败", err, nil))
				return
			}

			c.JSON(200, cfg.Base.RespFunc("数据查询成功", nil, &data))
		},
	)
}
