package preloader

import (
	"github.com/gin-gonic/gin"
)

func CreateAddHandler[T any, U any, R any](
	cfg *Config[R],
	handlerFunc func(c *gin.Context, u *U, r *R) *T,
) gin.HandlerFunc {
	return Preload(
		cfg,
		func(c *gin.Context, u *U, r *R) {

			data := handlerFunc(c, u, r)
			if data == nil {
				return
			}

			if err := cfg.Base.DB.Create(data).Error; err != nil {
				c.JSON(500, cfg.Base.RespFunc("资源创建失败", err, nil))
				return
			}

			c.JSON(200, cfg.Base.RespFunc("数据创建成功", nil, data))

		},
	)
}
