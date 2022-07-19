package wtokenapi

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/dhontecillas/hfw/pkg/ginfw/auth"
	"github.com/dhontecillas/hfw/pkg/ginfw/web/session"
	"github.com/dhontecillas/hfw/pkg/ids"
)

// WAPIRoutes sets up the routes to handle token api keys.
func WAPIRoutes(r gin.IRouter) {
	r.POST(PathAPIKeys,
		session.AuthRequired(),
		WAPICreate)
	r.GET(PathAPIKeys,
		session.AuthRequired(),
		WAPIList)
	r.DELETE(PathAPIKeys,
		session.AuthRequired(),
		WAPIDelete)
}

// OKRes is the response for a successful operation.
type OKRes struct {
	Success bool `json:"success"`
}

// FailRes is the response for a failed operation.
type FailRes struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}

// WAPICreate is the handler for a create api token endpoint.
func WAPICreate(c *gin.Context) {
	userID := auth.GetUserID(c)

	p := CreatePayload{}
	err := c.ShouldBindJSON(&p)
	if err != nil {
		c.JSON(http.StatusBadRequest, FailRes{Error: err.Error()})
		return
	}
	ctrl := buildController(c)
	res, err := ctrl.CreateKey(*userID, p.Description)
	if err != nil {
		// we do not leak the reason why the registration failed
		c.JSON(http.StatusOK, FailRes{Success: false, Error: err.Error()})
		return
	}

	jres := fromTokenAPI(res)
	c.JSON(http.StatusOK, *jres)
}

// WAPIList is the handler for the list api tokens endpoint.
func WAPIList(c *gin.Context) {
	userID := auth.GetUserID(c)
	if userID == nil {
		c.JSON(http.StatusBadRequest, nil)
		return
	}

	ctrl := buildController(c)
	res, err := ctrl.ListKeys(*userID, false)
	if err != nil {
		c.JSON(http.StatusBadRequest, FailRes{Error: err.Error()})
		return
	}
	jres := fromTokenAPISlice(res)
	c.JSON(http.StatusOK, jres)
}

// WAPIDelete is the handler for the delete api token endpoint.
func WAPIDelete(c *gin.Context) {
	userID := auth.GetUserID(c)
	// userID should never be nil, because the middleware takes care of that

	p := DeletePayload{}
	err := c.ShouldBindJSON(&p)
	if err != nil {
		c.JSON(http.StatusBadRequest, FailRes{
			Error: fmt.Sprintf("bad payload %s", err.Error())})
		return
	}

	var keyID ids.ID
	if err := keyID.FromShuffled(p.Key); err != nil {
		c.JSON(http.StatusBadRequest, FailRes{Error: fmt.Sprintf("key %s", err.Error())})
		return
	}

	ctrl := buildController(c)
	err = ctrl.DeleteKey(*userID, keyID)
	if err != nil {
		// we do not leak info about if the user exists or not
		c.JSON(http.StatusOK, OKRes{Success: true})
	}
	c.JSON(http.StatusOK, OKRes{Success: true})
}
