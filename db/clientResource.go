package db

import (
	"fmt"
	"strings"

	"github.com/parkhomchik/oauth2/model"
	"github.com/parkhomchik/oauth2/utils"
	uuid "github.com/satori/go.uuid"
)

//client
func (dbm *DBManager) GetClient() []model.Client {
	var clients []model.Client
	dbm.DB.Find(&clients)
	return clients
}

func (dbm *DBManager) GetClientByID(id, userID uuid.UUID) (model.Client, error) {
	var client model.Client
	err := dbm.DB.Where("id = ? and user_id = ?", id, userID).Find(&client).Error
	return client, err
}

func (dbm *DBManager) GetScopesClient(id uuid.UUID) []model.Scope {
	var clientScopeID []model.ClientScopes
	var scopes []model.Scope
	dbm.DB.Where("client_id = ?", id).Find(&clientScopeID)
	var scopeIDs []uuid.UUID

	for _, uid := range clientScopeID {
		scopeIDs = append(scopeIDs, uid.ScopeID)
	}

	dbm.DB.Where("id in (?)", scopeIDs).Find(&scopes)

	return scopes
}

func (dbm *DBManager) PostScopesClient(clientScope model.ClientScopes) error {
	return dbm.DB.Save(&clientScope).Error
}

func (dbm *DBManager) ClientScope(clientID, role string) error {
	var client model.Client
	dbm.DB.Where("ID = ?", clientID).First(&client)
	var scope model.Scope
	dbm.DB.Where("name = ?", role).First(&scope)
	var clientScope model.ClientScopes
	return dbm.DB.Where("client_id = ? AND scope_id = ?", client.ID, scope.ID).First(&clientScope).Error
}

func (dbm *DBManager) RegistrationClient(client model.Client, roles string) (model.Client, error) {
	var clientScope model.ClientScopes
	err := dbm.DB.Create(&client).Error
	if err == nil {
		client.Secret = utils.GenerateSecret()
		err = dbm.DB.Save(&client).Error
		if roles != "" {
			scopes := strings.Split(roles, ",")
			for _, s := range scopes {
				scope, err := dbm.GetScopeByName(s)
				if err == nil {
					clientScope.ClientID = client.ID
					clientScope.ScopeID = scope.ID
					err = dbm.PostScopesClient(clientScope)
					if err != nil {
						fmt.Println("Error client post scope: ", err)
					}
				}
			}
		}
	}
	return client, err
}

func (dbm *DBManager) DeleteClient(client model.Client) error {
	return dbm.DB.Delete(&client).Error
}

func (dbm *DBManager) DeleteClientScope(client model.Client) error {
	var clientScopes model.ClientScopes
	if err := dbm.DB.Where("client_id = ?", client.ID).First(&clientScopes).Error; err != nil {
		return err
	}
	return dbm.DB.Where("user_id = ?", clientScopes.ClientID).Delete(&clientScopes).Error
}
