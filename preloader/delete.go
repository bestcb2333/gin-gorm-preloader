package preloader

import (
	"github.com/gin-gonic/gin"
)

func CreateDeleteHandler[T any, U any](
	cfg *Config,
	opt *Option,
) gin.HandlerFunc {
	return Preload(
		cfg,
		opt,
		&struct{}{},
		func(c *gin.Context, u *U, r *struct{}) {

			var req struct {
				IDs []uint `form:"id"`
			}

			if err := cfg.DB.Where(
				"id IN ?", req.IDs,
			).Delete(new(T)).Error; err != nil {
				cfg.Logger.Resp(c, 500, "删除失败", err, nil)
				return
			}

			cfg.Logger.Resp(c, 200, "删除成功", nil, nil)
		},
	)
}
