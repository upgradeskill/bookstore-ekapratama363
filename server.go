package main

import (
	"fmt"
	"myapp/config"
	"myapp/model"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type Response struct {
	ErrorCode int         `json:"error_code" form:"error_code"`
	Message   string      `json:"message" form:"message"`
	Data      interface{} `json:"data"`
}
type jwtCustomClaims struct {
	Name  string `json:"name"`
	Admin bool   `json:"admin"`
	jwt.StandardClaims
}

func login(c echo.Context) error {
	username := c.FormValue("username")
	password := c.FormValue("password")

	// Throws unauthorized error
	if username != "jon" || password != "shhh!" {
		return echo.ErrUnauthorized
	}

	// Set custom claims
	claims := &jwtCustomClaims{
		"Jon Snow",
		true,
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 72).Unix(),
		},
	}

	// Create token with claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Generate encoded token and send it as response.
	t, err := token.SignedString([]byte("secret"))
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, echo.Map{
		"token": t,
	})
}

func main() {
	config.ConnectDB()
	route := echo.New()

	// Middleware
	route.Use(middleware.Logger())
	route.Use(middleware.Recover())

	// Login route
	route.POST("/login", login)

	// Restricted group
	r := route.Group("/user")
	// Configure middleware with the custom claims type
	config := middleware.JWTConfig{
		Claims:     &jwtCustomClaims{},
		SigningKey: []byte("secret"),
	}
	r.Use(middleware.JWTWithConfig(config))

	r.GET("", func(c echo.Context) error {
		response := new(Response)
		users, err := model.GetAll(c.QueryParam("keywords")) // method get all
		if err != nil {
			response.ErrorCode = 10
			response.Message = "Gagal melihat data user"
			return c.JSON(http.StatusBadRequest, response)
		} else {
			response.ErrorCode = 0
			response.Message = "Sukses melihat data user"
			response.Data = users
			return c.JSON(http.StatusOK, response)
		}
	})

	r.POST("/store", func(c echo.Context) error {
		user := new(model.Users)
		c.Bind(user)
		contentType := c.Request().Header.Get("Content-type")
		if contentType == "application/json" {
			fmt.Println("Request dari json")
		}
		response := new(Response)
		if user.CreateUser() != nil { // method create user
			response.ErrorCode = 10
			response.Message = "Gagal create data user"
			return c.JSON(http.StatusBadRequest, response)
		} else {
			response.ErrorCode = 0
			response.Message = "Sukses create data user"
			response.Data = *user
			return c.JSON(http.StatusOK, response)
		}
	})

	r.GET("/show/:email", func(c echo.Context) error {
		user, err := model.GetOneByEmail(c.Param("email")) // method get by email
		response := new(Response)

		if err != nil {
			response.ErrorCode = 10
			response.Message = "Gagal melihat data user"
			return c.JSON(http.StatusBadRequest, response)
		} else {
			response.ErrorCode = 0
			response.Message = "Sukses melihat data user"
			response.Data = user
			return c.JSON(http.StatusOK, response)
		}
	})

	r.PUT("/update/:email", func(c echo.Context) error {
		user := new(model.Users)
		c.Bind(user)
		response := new(Response)
		if user.UpdateUser(c.Param("email")) != nil { // method update user
			response.ErrorCode = 10
			response.Message = "Gagal update data user"
			return c.JSON(http.StatusBadRequest, response)
		} else {
			response.ErrorCode = 0
			response.Message = "Sukses update data user"
			response.Data = *user
			return c.JSON(http.StatusOK, response)
		}
	})

	r.DELETE("/delete/:email", func(c echo.Context) error {
		user, _ := model.GetOneByEmail(c.Param("email")) // method get by email
		response := new(Response)

		if user.DeleteUser() != nil {
			response.ErrorCode = 10
			response.Message = "User tidak ditemukan"
			return c.JSON(http.StatusNotFound, response)
		} else {
			response.ErrorCode = 0
			response.Message = "Sukses menghapus data user"
			return c.JSON(http.StatusOK, response)
		}
	})

	route.Start(":9000")
}
