package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"crontab/master/common"
	"crontab/master/model"
)

// AuthMiddleware JWT 认证
func AuthMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		var (
			redisToken string
		)

		// 从 header 中获取 token
		//tokenString := ctx.GetHeader("Authorization")

		// 从 cookie 中获取token
		tokenString, err := ctx.Cookie("jwt_token")

		// validate token formate
		if tokenString == "" {
			ctx.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "您没有权限"})
			ctx.Abort()
			return
		}

		//tokenString = tokenString[7:]
		token, claims, err := common.ParseToken(tokenString)
		if err != nil || !token.Valid {
			ctx.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "您没有权限"})
			ctx.Abort()
			return
		}

		// 验证通过后获取claim 中的userId
		userId := claims.UserId
		var user model.User
		common.GMsql.DB.First(&user, userId)

		// 用户
		if user.ID == 0 {
			ctx.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "您没有权限"})
			ctx.Abort()
			return
		}

		if redisToken, err = common.GRdb.RDB.Get(ctx, string(user.ID)).Result(); err != nil || redisToken != token.Raw {
			ctx.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "您没有权限"})
			ctx.Abort()
			return
		}

		// 用户存在 将user 的信息写入上下文
		ctx.Set("user", user)

		ctx.Next()
	}
}
