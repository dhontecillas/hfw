package wtokenapi

import (
	"github.com/gin-gonic/gin"

	"github.com/dhontecillas/hfw/pkg/ginfw"
	"github.com/dhontecillas/hfw/pkg/tokenapi"
)

func buildController(c *gin.Context) tokenapi.TokenAPI {
	ed := ginfw.ExtServices(c)
	repo := tokenapi.NewRepoSQLX(ed.Ins, ed.SQL)
	return tokenapi.NewTokenAPI(ed.Ins, repo)
}
