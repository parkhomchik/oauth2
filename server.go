package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/satori/go.uuid"

	"github.com/parkhomchik/oauth2/db"
	"github.com/parkhomchik/oauth2/model"
	"github.com/parkhomchik/oauth2/oauth"
	"github.com/parkhomchik/oauth2/utils"

	"github.com/gin-gonic/gin"
	"github.com/go-oauth2/gin-server"
	"gopkg.in/oauth2.v3/manage"
	aserver "gopkg.in/oauth2.v3/server"
	"gopkg.in/oauth2.v3/store"

	_ "github.com/jinzhu/gorm/dialects/postgres"
)

//DBManager Для работы с базой
var DBManager db.DBManager

func main() {
	DBManager.InitDB()

	manager := manage.NewDefaultManager()

	manager.MustTokenStorage(store.NewMemoryTokenStore())
	clientStore := store.NewClientStore()

	clients := DBManager.GetClient()
	for _, c := range clients {
		scopes := DBManager.GetScopesClient(c.ID)
		var scopesString []string
		for _, s := range scopes {
			scopesString = append(scopesString, s.Name)
		}
		client := &oauth.Client{
			ID:     fmt.Sprint(c.ID),
			Secret: c.Secret,
			Domain: c.Domain,
			Scope:  scopesString,
			UserID: fmt.Sprint(c.UserID),
		}
		fmt.Println("Scope: ", client.Scope)
		clientStore.Set(client.ID, client)
	}
	users := DBManager.GetUsers()
	for _, u := range users {
		scopes := DBManager.GetScopesUser(u.ID)

		var scopesString []string
		for _, s := range scopes {
			scopesString = append(scopesString, s.Name)
		}

		user := &oauth.Client{
			ID:       u.Login,
			Password: u.Password,
			Secret:   u.Password,
			Domain:   "http://localhost",
			Scope:    scopesString,
			UserID:   u.Login,
		}
		clientStore.Set(user.ID, user)
	}

	manager.MapClientStorage(clientStore)

	server.InitServer(manager)
	server.SetAllowGetAccessRequest(false)
	server.SetClientInfoHandler(aserver.ClientFormHandler)

	//Доступные типы авторизации
	server.SetAllowedGrantType("client_credentials", "password")
	server.SetAllowedResponseType("code", "token")

	//Обработка ролей
	server.SetClientScopeHandler(clientScopeHandler)

	server.SetPasswordAuthorizationHandler(passwordAuthorizationHandler)

	g := gin.Default()
	g.Use(setCORSMiddleware())

	g.GET("/.well-known/openid-configuration", func(c *gin.Context) {
		data, err := ioutil.ReadFile("config/openid-configuration")
		if err != nil {
			fmt.Println(err)
		}
		c.String(http.StatusOK, string(data))
	})
	g.POST("/registrationuser", func(c *gin.Context) {
		roles := c.Query("roles")
		var userInf model.User
		if err := c.Bind(&userInf); err != nil {
			c.JSON(http.StatusBadRequest, err)
			return
		}
		if userInf.Password != "" {
			userInf.Password = utils.EncryptPassword(userInf.Password)
		}

		userInf, err := DBManager.RegistrationUser(userInf, roles)
		if err != nil {
			c.String(http.StatusBadRequest, err.Error())
			return
		}
		c.JSON(http.StatusOK, userInf)
	})

	auth := g.Group("/oauth2")
	{
		auth.GET("/token", server.HandleTokenRequest)  //Получение токена client_credentials, password
		auth.POST("/token", server.HandleTokenRequest) //Получение токена client_credentials, password

		auth.GET("")
		checkToken := auth.Group("/check") //Проверка токена
		checkToken.Use(server.HandleTokenVerify())
		checkToken.GET("", func(c *gin.Context) {
			ti, exists := c.Get("AccessToken")
			if exists {
				c.JSON(http.StatusOK, ti)
				return
			}
		})
	}

	connect := g.Group("connect")
	{
		connect.GET("/userinfo", server.HandleTokenVerify(), func(c *gin.Context) {
			ti, _ := c.Get("AccessToken")

			var tokenInfo model.TokenInfo
			var user model.User

			bodyBytes, _ := json.Marshal(ti)
			json.Unmarshal(bodyBytes, &tokenInfo)
			user, err := DBManager.GetUserByID(tokenInfo.UserID)
			if err != nil {
				c.JSON(404, err)
				return
			}
			c.JSON(200, &user)
		})

		connect.GET("/clientinfo/:clientid", server.HandleTokenVerify(), func(c *gin.Context) {
			id, err := uuid.FromString(c.Param("clientid"))
			if err != nil {
				c.JSON(400, err)
				return
			}
			ti, _ := c.Get("AccessToken")
			var tokenInfo model.TokenInfo
			var client model.Client

			bodyBytes, _ := json.Marshal(ti)
			json.Unmarshal(bodyBytes, &tokenInfo)
			client, err = DBManager.GetClientByID(id, tokenInfo.UserID)
			if err != nil {
				c.JSON(404, err)
				return
			}
			c.JSON(200, &client)
		})

		connect.PUT("/setuserinfo", server.HandleTokenVerify(), func(c *gin.Context) {
			var userInf model.User
			if err := c.Bind(&userInf); err != nil {
				c.JSON(http.StatusBadRequest, err)
				return
			}
			if userInf.Password != "" {
				userInf.Password = utils.EncryptPassword(userInf.Password)
			}
			userInf, err := DBManager.SetUserInfo(userInf)
			if err != nil {
				c.String(http.StatusBadRequest, err.Error())
				return
			}
			c.JSON(http.StatusOK, userInf)
		})

		connect.POST("/registrationclient", server.HandleTokenVerify(), func(c *gin.Context) {
			roles := c.Query("roles")
			var tokenInfo model.TokenInfo
			ti := c.MustGet("AccessToken")
			bodyBytes, _ := json.Marshal(ti)
			json.Unmarshal(bodyBytes, &tokenInfo)

			userID := tokenInfo.UserID
			permissionCheck, err := userScopeHandler(userID, "write")
			if permissionCheck {
				var clientInf model.Client
				clientInf.UserID = userID
				clientInf, err = DBManager.RegistrationClient(clientInf, roles)
				if err != nil {
					c.String(http.StatusBadRequest, err.Error())
					return
				}
				c.JSON(http.StatusOK, clientInf)
			} else {
				c.JSON(550, err)
			}
		})

		connect.DELETE("/client/:id", server.HandleTokenVerify(), func(c *gin.Context) {
			var tokenInfo model.TokenInfo
			ti := c.MustGet("AccessToken")
			bodyBytes, _ := json.Marshal(ti)
			json.Unmarshal(bodyBytes, &tokenInfo)
			userID := tokenInfo.UserID
			permissionCheck, err := userScopeHandler(userID, "write")
			if permissionCheck {
				id, err := uuid.FromString(c.Param("id"))
				if err != nil {
					c.JSON(400, err)
					return
				}
				var client model.Client
				client, err = DBManager.GetClientByID(id, userID)
				if err != nil {
					c.JSON(404, err)
					return
				}
				scopeErr := DBManager.DeleteClientScope(client)
				if scopeErr != nil {
					fmt.Println("Delete client scope error: ", scopeErr)
				}
				if err := DBManager.DeleteClient(client); err != nil {
					c.JSON(500, err)
					return
				}
				c.JSON(http.StatusOK, scopeErr)
			} else {
				c.JSON(550, err)
			}
		})

		connect.DELETE("/user/:id", server.HandleTokenVerify(), func(c *gin.Context) {
			id, err := uuid.FromString(c.Param("id"))
			if err != nil {
				c.JSON(400, err)
				return
			}
			var user model.User
			user, err = DBManager.GetUserByID(id)
			if err != nil {
				c.JSON(404, err)
				return
			}
			scopeErr := DBManager.DeleteUserScope(user) //аналогично клиенту
			if scopeErr != nil {
				fmt.Println("Delete user scope error: ", scopeErr)
			}
			if err := DBManager.DeleteUser(user); err != nil {
				c.JSON(500, err)
				return
			}
			c.JSON(http.StatusOK, scopeErr)
		})
	}
	g.Run(":9096")
}

func setCORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "authorization, accesstoken, content-type")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, PUT")
		if c.Request.Method != "OPTIONS" {
			c.Next()
		} else {
			c.AbortWithStatus(http.StatusOK)
		}
	}
}

func clientScopeHandler(clientID, scope string) (allowed bool, err error) {
	scopes := strings.Split(scope, " ")
	for _, s := range scopes {
		if err := DBManager.ClientScope(clientID, s); err != nil {
			return false, err
		}
	}
	return true, nil
}

func userScopeHandler(userID uuid.UUID, scope string) (allowed bool, err error) {
	scopes := strings.Split(scope, " ")
	for _, s := range scopes {
		if err := DBManager.UserScope(userID, s); err != nil {
			return false, err
		}
	}
	return true, nil
}

func passwordAuthorizationHandler(username, password string) (userID string, err error) {
	user, err := DBManager.Login(username, utils.EncryptPassword(password))
	return fmt.Sprint(user.ID), err
}
