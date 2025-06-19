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

type Logger interface {
	Resp(c *gin.Context, code int, msg string, err error)
}

type Config struct {
	DB            *gorm.DB
	JWTKey        string
	JWTExpiry     time.Duration
	UserTableName string
	AdminColName  string
	Logger
}

type Option struct {
	Bind       BindOption
	Permission Permission
	Tables     []string
}

type BindOption uint32

const (
	Uri BindOption = 1 << iota
	Query
	JSON
	Multipart
)

func Preload[R any, U any](
	cfg *Config,
	opt *Option,
	defReq *R,
	handlerFunc func(c *gin.Context, u *U, r *R),
) gin.HandlerFunc {
	return func(c *gin.Context) {

		var u *U
		var req *R

		if opt == nil {
			opt = new(Option)
		}

		if opt.Permission >= Login {
			rawToken := c.GetHeader("Authorization")
			if len(rawToken) < 8 {
				cfg.Logger.Resp(c, 400, "token无效", nil)
				return
			}

			token := rawToken[7:]

			jwtToken, err := jwt.Parse(
				token,
				func(t *jwt.Token) (any, error) {
					return []byte(cfg.JWTKey), nil
				},
			)
			if err != nil {
				cfg.Logger.Resp(c, 400, "token秘钥不正确", err)
				return
			}

			claims, ok := jwtToken.Claims.(jwt.MapClaims)
			if !ok || !jwtToken.Valid {
				cfg.Logger.Resp(c, 400, "token格式不正确", nil)
				return
			}

			userId := uint(claims["userId"].(float64))

			c.Set("userId", userId)

			newToken, err := GetJwt(userId, cfg.JWTKey, cfg.JWTExpiry)

			if err != nil {
				cfg.Logger.Resp(c, 400, "生成新token失败", err)
				return
			}
			c.Header("Authorization", newToken)

			var user U
			query := cfg.DB.Table(cfg.UserTableName).Where("id = ?", userId).Session(new(gorm.Session))

			var admin bool
			if err := query.Select(cfg.AdminColName).Scan(&admin).Error; err != nil {
				cfg.Logger.Resp(c, 500, "没有admin字段", err)
				return
			}

			if opt.Permission >= Admin && !admin {
				cfg.Logger.Resp(c, 403, "你不是管理员", nil)
				return
			}

			if opt.Tables != nil {
				for _, value := range opt.Tables {
					query = query.Preload(value)
				}
			}

			if err := query.First(&user).Error; errors.Is(
				err, gorm.ErrRecordNotFound,
			) {
				cfg.Logger.Resp(c, 400, "用户不存在", err)
				return
			} else if err != nil {
				cfg.Logger.Resp(c, 500, "查询用户失败", err)
				return
			}

			u = &user
		}

		if opt.Bind != 0 {

			if defReq != nil {
				temp := *defReq
				req = &temp
			} else {
				req = new(R)
			}

			if opt.Bind&Uri != 0 {
				if err := c.ShouldBindUri(req); err != nil {
					cfg.Logger.Resp(c, 400, "路径参数有误", err)
					return
				}
			}

			if opt.Bind&Query != 0 {
				if err := c.ShouldBindQuery(req); err != nil {
					cfg.Logger.Resp(c, 400, "查询字符串参数有误", err)
					return
				}
			}

			if opt.Bind&JSON != 0 {
				if err := c.ShouldBindJSON(req); err != nil {
					cfg.Logger.Resp(c, 400, "请求体格式有误", err)
					return
				}
			}

			if opt.Bind&Multipart != 0 {
				if err := c.ShouldBind(req); err != nil {
					cfg.Logger.Resp(c, 400, "请求表单格式有误", err)
					return
				}
			}

		}

		handlerFunc(c, u, req)
	}

}
