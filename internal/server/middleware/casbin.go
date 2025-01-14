package middleware

import (
	"net/http"

	"github.com/casbin/casbin/v2"
	"github.com/gin-gonic/gin"
	"github.com/nilorg/naas/internal/model"
	naasCasbin "github.com/nilorg/naas/internal/module/casbin"
	"github.com/nilorg/naas/internal/pkg/contexts"
	"github.com/nilorg/naas/internal/service"
	"github.com/nilorg/oauth2"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// CasbinAuthRequired 身份验证
func CasbinAuthRequired(enforcer casbin.IEnforcer) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		tokenClaims := ctx.MustGet("token").(*oauth2.JwtClaims)
		openID := model.ConvertStringToID(tokenClaims.Subject)

		roles, _ := service.Role.GetAllRoleByUserID(contexts.WithGinContext(ctx), openID)
		if len(roles) > 0 {
			ctx.Set("current_role", roles)
		} else {
			ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "permission denied",
			})
			return
		}
		allow := false
		for _, role := range roles {
			logrus.Debugf("openid: %d, role code: %s", openID, role.RoleCode)
			check, checkErr := naasCasbin.EnforceRoute(role, viper.GetString("naas.resource.id"), ctx.Request, enforcer)
			if checkErr != nil {
				logrus.Errorf("casbin enforce web route:", checkErr)
				ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{
					"error": checkErr.Error(),
				})
				return
			}
			if check {
				allow = true
				break
			}
		}
		if !allow {
			ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "permission denied",
			})
			return
		}

		usr, userInfo, err := service.User.GetInfoOneByCache(contexts.WithGinContext(ctx), openID)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": err.Error(),
			})
			return
		}

		ctx.Set("current_user", &model.SessionAccount{
			UserID:   usr.ID,
			UserName: usr.Username,
			Nickname: userInfo.Nickname,
			Picture:  userInfo.Picture,
		})
		ctx.Next()
	}
}
