package preloader

import (
	"github.com/gin-gonic/gin"
)

type PreloadField struct {
	Name   string
	Fields []string
}

func CreateGetHandler[T any, U any](
	cfg *Config,
	opt *Option,
	preloads []PreloadField,
	selects []string,
) gin.HandlerFunc {
	opt.Bind = &BindOption{Query: true}
	return Preload(
		cfg,
		opt,
		&struct{}{},
		func(c *gin.Context, u *U, r *struct{}) {

			q := cfg.DB.Model(new(T))

			for _, preload := range preloads {
				q = q.Preload(preload.Name, Select(preload.Fields...))
			}

			id := c.Param("id")

			if selects != nil {
				for _, sel := range selects {
					q = q.Select(sel)
				}
			}

			var data T
			if err := q.First(&data, "id = ?", id).Error; err != nil {
				cfg.Logger.Resp(c, 500, "查询失败", err, nil)
				return
			}

			cfg.Logger.Resp(c, 200, "数据查询成功", nil, &data)
		},
	)
}
