package preloader

import (
	"errors"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

type Permission int

const (
	Public Permission = iota
	Login
	Admin
)

type BaseConfig struct {
	DB            *gorm.DB
	JWTKey        string
	JWTExpHours   time.Duration
	UserTableName string
	AdminColName  string
	RespFunc      func(message string, err error, data any) gin.H
}

type Config[R any] struct {
	Base       *BaseConfig
	Bind       *BindConfig
	Permission Permission
	Tables     []string
	DefReq     R
}

type BindConfig struct {
	Param     bool
	Query     bool
	JSON      bool
	Multipart bool
}

func Preload[R any, U any](
	cfg *Config[R],
	handlerFunc func(c *gin.Context, u *U, r *R),
) gin.HandlerFunc {
	return func(c *gin.Context) {

		var u *U

		if cfg.Permission >= Login {
			rawToken := c.GetHeader("Authorization")
			if len(rawToken) < 8 {
				c.AbortWithStatusJSON(400, cfg.Base.RespFunc("token无效", nil, nil))
				return
			}

			token := rawToken[7:]

			jwtToken, err := jwt.Parse(
				token,
				func(t *jwt.Token) (any, error) {
					return []byte(cfg.Base.JWTKey), nil
				},
			)
			if err != nil {
				c.AbortWithStatusJSON(400, cfg.Base.RespFunc("token秘钥不正确", err, nil))
				return
			}

			claims, ok := jwtToken.Claims.(jwt.MapClaims)
			if !ok || !jwtToken.Valid {
				c.AbortWithStatusJSON(400, cfg.Base.RespFunc("token格式不正确", nil, nil))
				return
			}

			userId := uint(claims["userId"].(float64))

			newToken, err := GetJwt(userId, cfg.Base.JWTKey, cfg.Base.JWTExpHours)

			if err != nil {
				c.AbortWithStatusJSON(400, cfg.Base.RespFunc("生成新token失败", err, nil))
				return
			}
			c.Header("Authorization", newToken)

			var user U
			query := cfg.Base.DB.Table(cfg.Base.UserTableName).Where("id = ?", userId)

			var admin bool
			if err := query.Pluck(cfg.Base.AdminColName, &admin).Error; err != nil {
				c.AbortWithStatusJSON(500, cfg.Base.RespFunc("没有admin字段", err, nil))
				return
			}

			if cfg.Permission >= Admin && !admin {
				c.AbortWithStatusJSON(403, cfg.Base.RespFunc("你不是管理员", nil, nil))
				return
			}

			if cfg.Tables != nil {
				for _, value := range cfg.Tables {
					query = query.Preload(value)
				}
			}

			if err := query.First(&user).Error; errors.Is(
				err, gorm.ErrRecordNotFound,
			) {
				c.AbortWithStatusJSON(400, cfg.Base.RespFunc("用户不存在", err, nil))
				return
			} else if err != nil {
				c.AbortWithStatusJSON(500, cfg.Base.RespFunc("查询用户失败", err, nil))
				return
			}

			u = &user
		}

		req := cfg.DefReq

		if cfg.Bind != nil {

			if cfg.Bind.Param {
				if err := c.ShouldBindUri(&req); err != nil {
					c.JSON(400, cfg.Base.RespFunc("路径参数有误", err, nil))
					return
				}
			}

			if cfg.Bind.Query {
				if err := c.ShouldBindQuery(&req); err != nil {
					c.JSON(400, cfg.Base.RespFunc("查询字符串参数有误", err, nil))
					return
				}
			}

			if cfg.Bind.JSON {
				if err := c.ShouldBindJSON(&req); err != nil {
					c.JSON(400, cfg.Base.RespFunc("请求体有误", err, nil))
					return
				}
			}

			if cfg.Bind.Multipart {
				if err := c.ShouldBind(&req); err != nil {
					c.JSON(400, cfg.Base.RespFunc("请求表单错误", err, nil))
					return
				}
			}

		}

		handlerFunc(c, u, &req)
	}

}
