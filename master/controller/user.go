package controller

import (
	"golang.org/x/crypto/bcrypt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"crontab/master/common"
	"crontab/master/model"
	"crontab/master/response"
)

func Register(ctx *gin.Context) {

	var (
		user model.User
	)

	// 获取参数
	username := ctx.PostForm("username")
	password := ctx.PostForm("password")
	telephone := ctx.PostForm("telephone")
	email := ctx.PostForm("email")

	common.GMsql.DB.Where("username=?", username).First(&user)
	if user.ID != 0 {
		response.Response(ctx, http.StatusInternalServerError, 400, nil, "用户名已存在")
		return
	}

	// 创建用户
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		response.Response(ctx, http.StatusInternalServerError, 500, nil, "密码加密错误，注册失败")
		return
	}

	newUser := model.User{
		Username:  username,
		Password:  string(hashedPassword),
		Telephone: telephone,
		Email:     email,
	}

	common.GMsql.DB.Create(&newUser)

	// 发放token
	token, err := common.ReleaseToken(newUser)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "系统异常"})
		log.Printf("token generate error : %v", err)
		return
	}

	if err = common.GRdb.RDB.Set(ctx, string(user.ID), token, 0).Err(); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "系统异常"})
		return
	}

	response.Success(ctx, gin.H{"token": token}, "注册成功")
	return

}

func Login(ctx *gin.Context) {

	var (
		user model.User
	)
	username := ctx.PostForm("username")
	password := ctx.PostForm("password")

	common.GMsql.DB.Where("username=?", username).First(&user)
	if user.ID == 0 {
		ctx.JSON(http.StatusUnprocessableEntity, gin.H{"code": 200, "msg": "用户不存在"})
		return
	}

	// 判断密码收否正确
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "密码错误"})
		return
	}

	// 发放token
	token, err := common.ReleaseToken(user)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "系统异常"})
		log.Printf("token generate error : %v", err)
		return
	}

	if err := common.GRdb.RDB.Set(ctx, string(user.ID), token, 0).Err(); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "系统异常"})
		return
	}

	/*
	   name		string	cookie名字
	   value	string	cookie值
	   maxAge	int		有效时间，单位是秒，MaxAge=0 忽略MaxAge属性，MaxAge<0 相当于删除cookie, 通常可以设置-1代表删除，MaxAge>0 多少秒后cookie失效
	   path		string	cookie路径
	   domain	string	cookie作用域
	   secure	bool	Secure=true，那么这个cookie只能用https协议发送给服务器
	   httpOnly	bool	设置HttpOnly=true的cookie不能被js获取到
	*/
	ctx.SetCookie("jwt_token", token, common.GConfig.Http.LoginExpire, "/", "127.0.0.1", false, true)
	// 返回结果
	response.Success(ctx, gin.H{"userInfo": common.ToUserDto(user)}, "登录成功")
	return
}

func Logout(ctx *gin.Context) {

	user, _ := ctx.Get("user")

	if err := common.GRdb.RDB.Del(ctx, string(user.(model.User).ID)).Err(); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "系统异常"})
		return
	}

	response.Success(ctx, nil, "退出登录")
	return
}

func UserInfo(ctx *gin.Context) {
	user, _ := ctx.Get("user")

	ctx.JSON(http.StatusOK, gin.H{"code": 200, "data": gin.H{"user": common.ToUserDto(user.(model.User))}})
}

func UserList(ctx *gin.Context) {

	var (
		userArr []model.User
	)
	common.GMsql.DB.Find(&userArr)
	// 返回结果
	response.Success(ctx, gin.H{"users": userArr}, "查询成功")
	return
}
