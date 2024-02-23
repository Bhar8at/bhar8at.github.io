package main

import (
	"html/template"
	"net/http"
	"os"

	"github.com/Bhar8at/bhar8at.github.io/internal"
	socials "github.com/Bhar8at/bhar8at.github.io/internal/auth"
	"github.com/Bhar8at/bhar8at.github.io/middleware"
	"github.com/Bhar8at/bhar8at.github.io/routes"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

func index(c *gin.Context) {
	c.HTML(http.StatusOK, "indexT.html", nil)
}

func notFound(c *gin.Context) {

	c.HTML(http.StatusNotFound, "errorT.html", gin.H{
		"error":   "404 Not Found",
		"message": "The requested page was not found.",
	})

}

func main() {

	gin.SetMode(gin.ReleaseMode)

	app := gin.Default()

	// Redirects all URL's with trailing slash to the same URL without the trailing slash
	app.RedirectTrailingSlash = true

	// Makes it so that unavailable methods do not go ignored (405 status code returned)
	app.HandleMethodNotAllowed = true

	app.NoRoute(notFound)

	app.Static("/static", "./static")

	app.SetFuncMap(template.FuncMap{
		"formatAsTitle": internal.FormatAsTitle,
		"formatAsDate":  internal.FormatAsDate,
	})

	app.LoadHTMLGlob("templates/*")
	store := cookie.NewStore([]byte(os.Getenv("SECRET_KEY")))

	app.Use(sessions.Sessions("tsuki", store))
	// Storing the current session in a cookie
	app.Use(middleware.RecoveryMiddleware())
	// used to Recover from unexpected errors during handling of HTTP requests

	app.GET("/", index)
	app.GET("/signup", routes.SignUp)
	app.GET("/login", routes.Login)
	app.GET("/logout", routes.Logout)
	// there are more functions here :

	auth := app.Group("/auth")
	{
		auth.GET("/signup/google", socials.GoogleSignUp)
		auth.GET("/login/google", socials.GoogleLogin)
		auth.GET("/google", socials.GoogleAuth)

		auth.POST("/signup", routes.SignUp)
		auth.POST("/login", routes.Login)

	}

	if err := app.Run(); err != nil {
		panic(err)
	}

}
