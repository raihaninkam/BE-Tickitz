package utils

import "github.com/gin-gonic/gin"

func HandleResponse(ctx *gin.Context, status int, response any) {
	ctx.JSON(status, response)
}
