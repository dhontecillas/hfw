/**

Example of a basic web application using the **HFW** framework

*/

package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"

	"github.com/dhontecillas/hfw/pkg/bundler"
	"github.com/dhontecillas/hfw/pkg/config"
	"github.com/dhontecillas/hfw/pkg/ginfw"
	ginfwconfig "github.com/dhontecillas/hfw/pkg/ginfw/config"
	"github.com/dhontecillas/hfw/pkg/ginfw/web"
	"github.com/dhontecillas/hfw/pkg/ginfw/web/session"
	"github.com/dhontecillas/hfw/pkg/ginfw/web/wusers"
	"github.com/dhontecillas/hfw/pkg/obs/metrics"
)

const (
	// ConfAppPrefix has the value of the prefix to be prepended
	// to all HFW config vars, so a project can have its own prefix
	// for everything and avoid clashing names with other apps running
	// in the same evironment.
	ConfAppPrefix string = "WEBEXAMPLE_"

	// ConfStaticPath holds the directory where we store static assets.
	// This is not HFW specific, but some config var that we want for
	// our app.
	ConfStaticPath string = "WEB_STATICPATH"
)

// WebConfig contains the configuration to find
// the file assets for the application: here we
// put the app specific configuration.
type WebConfig struct {
	staticPath string
}

func readWebConfig() *WebConfig {
	var wc WebConfig
	wc.staticPath = os.Getenv(ConfStaticPath)
	return &wc
}

func main() {
	if err := config.InitConfig(ConfAppPrefix); err != nil {
		panic(err.Error())
	}

	router := gin.Default()

	// ReadInsightsConfig reads the configuration about where to
	// send logs and metrics.
	insConfig := config.ReadInsightsConfig(ConfAppPrefix)

	// From the config, we can create a function to instatiate the
	// insights object to send metrics and logs, and also a function
	// to have a clean shutdown of the reporting (that is sending
	// pending metrics and logs before closing the app)
	insBuilder, insFlush := config.CreateInsightsBuilder(insConfig,
		metrics.Defs{})
	es := config.BuildExternalServices(ConfAppPrefix, insBuilder, insFlush)
	defer es.Shutdown()

	// Apply the db migrations that will create the required tables
	// to register users.
	ins := es.ExtServices().Ins
	if err := bundler.ApplyMigrationsFromConfig(
		"up", viper.GetViper(), ins.L, ConfAppPrefix); err != nil {
		panic(err)
	}

	router.Use(ginfw.ExtServicesMiddleware(es))

	// set the web dependecies:
	redisConf := config.ReadRedisConfig(ConfAppPrefix)
	sessionConf, err := ginfwconfig.ReadSessionConf(ins, ConfAppPrefix, &redisConf)
	if err != nil {
		panic(err)
	}
	session.Use(router, sessionConf)

	// read the specific configuration for aur app, in this case
	// just the static assets folder.
	wbcfg := readWebConfig()
	router.Static("/static", wbcfg.staticPath)

	router.GET("/", home)

	// configure the routes for user registration
	actionPaths := wusers.ActionPaths{
		BasePath:          "/users/",
		ActivationPath:    "/users/activate/",
		ResetPasswordPath: "/users/resetpassword",
	}
	wusers.Routes(router.Group("/users"), actionPaths)

	// to use it in production:
	// router.HTMLRender = web.NewMultiRenderEngineFromDirs(
	//	"../../pkg/ginfw/web/wusers/", "./")
	router.HTMLRender = web.NewHTMLRender(
		"../../pkg/ginfw/web/wusers/", "./")
	err = router.Run()
	if err != nil {
		fmt.Printf("error running router: %s\n", err.Error())
	}
}

func home(c *gin.Context) {
	userID := session.GetUserID(c)
	c.HTML(http.StatusOK, "landing.html",
		gin.H{
			"wuser_registration_url": "/users/register",
			"foo":                    "bar",
			"userID":                 userID,
		})
}
