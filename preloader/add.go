package preloader

import (
	"github.com/gin-gonic/gin"
)

func CreateAddHandler[T any, U any, R any](
	cfg *Config,
	opt *Option,
	req *R,
	handlerFunc func(c *gin.Context, u *U, r *R) *T,
) gin.HandlerFunc {
	opt.Bind = &BindOption{JSON: true}
	return Preload(
		cfg,
		opt,
		req,
		func(c *gin.Context, u *U, r *R) {

			data := handlerFunc(c, u, r)
			if data == nil {
				return
			}

			if err := cfg.DB.Create(data).Error; err != nil {
				cfg.Logger.Resp(c, 500, "资源创建失败", err, nil)
				return
			}

			cfg.Logger.Resp(c, 200, "数据创建成功", nil, data)

		},
	)
}
