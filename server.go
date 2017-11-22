package main

import (
	"encoding/base64"
	"fmt"
	"net/http"

	"oauth2/model"
	"oauth2/oauth"

	"github.com/gin-gonic/gin"
	"github.com/go-oauth2/gin-server"
	"github.com/jinzhu/gorm"
	"gopkg.in/oauth2.v3/manage"
	aserver "gopkg.in/oauth2.v3/server"
	"gopkg.in/oauth2.v3/store"

	"golang.org/x/crypto/sha3"

	_ "github.com/jinzhu/gorm/dialects/postgres"
)

var Db *gorm.DB

func main() {
	fmt.Println("password: ", encryptPassword("test"))
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

	//Обработка ролей
	server.SetClientScopeHandler(clientScopeHandler)

	server.SetPasswordAuthorizationHandler(passwordAuthorizationHandler)

	g := gin.Default()

	auth := g.Group("/oauth2")
	{
		auth.GET("/token", server.HandleTokenRequest)  //Получение токена client_credentials, password
		auth.POST("/token", server.HandleTokenRequest) //Получение токена client_credentials, password
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
	g.Run(":9096")
}

func clientScopeHandler(clientID, scope string) (allowed bool, err error) {
	fmt.Println("Scope handler", clientID, scope)
	if scope == "read" {
		return true, nil
	}
	return false, nil
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
	Db, _ = gorm.Open("postgres", "host=localhost user=postgres dbname=oauth2 sslmode=disable password=parkhom4ik")
	Db.LogMode(true)
	Db.AutoMigrate(&model.User{}, &model.Client{}, &model.Scope{})
}

func getClient() []model.Client {
	var clients []model.Client
	Db.Find(&clients)
	return clients
}

func getScopesClient(id uint) []model.Scope {
	var clientScopeID []model.ClientScopes
	var scopes []model.Scope
	Db.Where("client_id = ?", id).Find(&clientScopeID)
	var clientIDs []uint

	for _, uid := range clientScopeID {
		clientIDs = append(clientIDs, uid.ClientID)
	}

	Db.Where("id in (?)", clientIDs).Find(&scopes)

	return scopes
}

func getUsers() (users []model.User) {
	Db.Find(&users)
	return
}

func getScopesUser(id uint) []model.Scope {
	var userScopeID []model.UserScopes
	var scopes []model.Scope
	Db.Where("user_id = ?", id).Find(&userScopeID)
	var userIDs []uint
	for _, uid := range userScopeID {
		userIDs = append(userIDs, uid.UserID)
	}
	Db.Where("id in (?)", userIDs).Find(&scopes)
	return scopes
}

func login(username, password string) (model.User, error) {
	var user model.User
	err := Db.Where("login = ? AND password = ?", username, password).First(&user).Error
	return user, err
}
