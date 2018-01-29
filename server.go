package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/satori/go.uuid"

	"github.com/parkhomchik/oauth2/model"
	"github.com/parkhomchik/oauth2/oauth"

	"github.com/gin-gonic/gin"
	"github.com/go-oauth2/gin-server"
	"github.com/jinzhu/gorm"
	"gopkg.in/oauth2.v3/manage"
	aserver "gopkg.in/oauth2.v3/server"
	"gopkg.in/oauth2.v3/store"

	"golang.org/x/crypto/sha3"

	_ "github.com/jinzhu/gorm/dialects/postgres"
)

//Db переменная для конекции
var Db *gorm.DB

//PortalDb переменная для конекции
var PortalDb *gorm.DB

func main() {
	//fmt.Println("password: ", encryptPassword("test"))
	initDB()

	manager := manage.NewDefaultManager()

	manager.MustTokenStorage(store.NewMemoryTokenStore())
	clientStore := store.NewClientStore()

	clients := getClient()
	for _, c := range clients {
		scopes := getScopesClient(c.ID)
		var scopesString []string
		for _, s := range scopes {
			scopesString = append(scopesString, s.Name)
		}
		client := &oauth.Client{
			ID:     fmt.Sprint(c.ID),
			Secret: c.Secret,
			Domain: c.Domain,
			Scope:  scopesString,
			UserID: c.UserID,
		}
		fmt.Println("Scope: ", client.Scope)
		clientStore.Set(client.ID, client)
	}
	users := getUsers()
	for _, u := range users {
		scopes := getScopesUser(u.ID)

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
	g.OPTIONS("/registrationuser", func(c *gin.Context) { c.Next() })
	g.POST("/registrationuser", func(c *gin.Context) {
		var userInf model.User
		if err := c.Bind(&userInf); err != nil {
			c.JSON(http.StatusBadRequest, err)
			return
		}
		if userInf.Password != "" {
			userInf.Password = encryptPassword(userInf.Password)
		}
		userInf, err := registrationUser(userInf)
		if err != nil {
			c.String(http.StatusBadRequest, err.Error())
			return
		}
		c.JSON(http.StatusOK, userInf)
	})

	g.POST("/registrationclient", func(c *gin.Context) {
		var clientInf model.Client
		if err := c.Bind(&clientInf); err != nil {
			c.JSON(http.StatusBadRequest, err)
			return
		}
		if clientInf.Secret != "" {
			clientInf.Secret = encryptPassword(clientInf.Secret)
		}
		clientInf, err := registrationClient(clientInf)
		if err != nil {
			c.String(http.StatusBadRequest, err.Error())
			return
		}
		c.JSON(http.StatusOK, clientInf)
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

	api := g.Group("api")
	{
		api.Use(server.HandleTokenVerify())
		api.GET("/test", func(c *gin.Context) {
			ti, exists := c.Get("AccessToken")
			if exists {
				c.JSON(http.StatusOK, ti)
				return
			}
			c.String(http.StatusOK, "not found")
		})
	}

	connect := g.Group("connect")
	{
		connect.OPTIONS("/userinfo", func(c *gin.Context) {
			c.Next()
		})

		connect.GET("/userinfo", server.HandleTokenVerify(), func(c *gin.Context) {
			ti, exists := c.Get("AccessToken")
			if exists {
				var tokenInfo model.TokenInfo
				var user model.User

				bodyBytes, err := json.Marshal(ti)
				if err != nil {
					fmt.Println(err)
				}
				json.Unmarshal(bodyBytes, &tokenInfo)

				if err := Db.Where("id = ?", tokenInfo.UserID).Find(&user).Error; err != nil {
					c.JSON(404, err)
					return
				}
				c.JSON(200, &user)
				return
			}
			c.String(http.StatusOK, "not found")
		})

		connect.PUT("/setuserinfo", server.HandleTokenVerify(), func(c *gin.Context) {
			var userInf model.User
			if err := c.Bind(&userInf); err != nil {
				c.JSON(http.StatusBadRequest, err)
				return
			}
			if userInf.Password != "" {
				userInf.Password = encryptPassword(userInf.Password)
			}
			userInf, err := setUserInfo(userInf)
			if err != nil {
				c.String(http.StatusBadRequest, err.Error())
				return
			}
			c.JSON(http.StatusOK, userInf)
		})
	}
	g.Run(":9096")
}

func clientScopeHandler(clientID, scope string) (allowed bool, err error) {
	scopes := strings.Split(scope, " ")
	for _, s := range scopes {
		if err := clientScope(clientID, s); err != nil {
			return false, err
		}
	}
	return true, nil
}

func passwordAuthorizationHandler(username, password string) (userID string, err error) {
	user, err := login(username, encryptPassword(password))
	return fmt.Sprint(user.ID), err
}

func encryptPassword(password string) string {
	h := sha3.New512()
	h.Write([]byte(password))
	b := h.Sum(nil)
	return base64.StdEncoding.EncodeToString(b)
}

func initDB() {
	var configuration model.Configuration
	configuration.Load()
	dbinfo := fmt.Sprintf("host=%s user=%s dbname=%s password=%s sslmode=disable", configuration.DbHost, configuration.DbUser, configuration.DbName, configuration.DbPass)
	//Db, _ = gorm.Open("postgres", dbinfo)
	var err error
	Db, err = gorm.Open("postgres", dbinfo)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	//defer Db.Close()

	Db.LogMode(true)
	Db.AutoMigrate(&model.User{}, &model.Client{}, &model.Scope{})
}

func initPortalDB() {
	var configuration model.Configuration
	configuration.Load()
	dbinfo := fmt.Sprintf("host=%s user=%s dbname=%s password=%s sslmode=disable", configuration.DbPortalHost, configuration.DbPortalUser, configuration.DbPortalName, configuration.DbPortalPass)
	var err error
	PortalDb, err = gorm.Open("postgres", dbinfo)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	PortalDb.LogMode(true)
}

func getClient() []model.Client {
	var clients []model.Client
	Db.Find(&clients)
	return clients
}

func getScopesClient(id uuid.UUID) []model.Scope {
	var clientScopeID []model.ClientScopes
	var scopes []model.Scope
	Db.Where("client_id = ?", id).Find(&clientScopeID)
	var scopeIDs []uuid.UUID

	for _, uid := range clientScopeID {
		scopeIDs = append(scopeIDs, uid.ScopeID)
	}

	Db.Where("id in (?)", scopeIDs).Find(&scopes)

	return scopes
}

func getUsers() (users []model.User) {
	Db.Find(&users)
	return
}

func getScopesUser(id uuid.UUID) []model.Scope {
	var userScopeID []model.UserScopes
	var scopes []model.Scope
	Db.Where("user_id = ?", id).Find(&userScopeID)
	var scopeIDs []uuid.UUID
	for _, uid := range userScopeID {
		scopeIDs = append(scopeIDs, uid.ScopeID)
	}
	Db.Where("id in (?)", scopeIDs).Find(&scopes)
	return scopes
}

func login(username, password string) (model.User, error) {
	var user model.User
	err := Db.Where("login = ? AND password = ?", username, password).First(&user).Error
	return user, err
}

func clientScope(clientID, role string) error {
	var client model.Client
	Db.Where("ID = ?", clientID).First(&client)
	var scope model.Scope
	Db.Where("name = ?", role).First(&scope)
	var clientScope model.ClientScopes
	return Db.Where("client_id = ? AND scope_id = ?", client.ID, scope.ID).First(&clientScope).Error
}

func setCORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "authorization, accesstoken, content-type")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, PUT")
		c.Next()
	}
}

func registrationUser(user model.User) (model.User, error) {
	if err := Db.Create(&user).Error; err != nil {
		return user, err
	}
	/*
		initPortalDB()
		var staff model.Staff
		staff.Name = user.Name

		err := PortalDb.Create(&staff).Error
		PortalDb.Close()
	*/
	return user, nil
}

func registrationClient(client model.Client) (model.Client, error) {
	err := Db.Create(&client).Error
	return client, err
}

func setUserInfo(userInf model.User) (model.User, error) {
	var user model.User
	if err := Db.Where("id = ?", userInf.ID).First(&user).Error; err != nil {
		return user, err
	}
	user = userInf
	if err := Db.Save(&userInf).Error; err != nil {
		return user, err
	}
	return user, nil
}
