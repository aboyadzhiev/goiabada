package data

import (
	"fmt"
	"strings"

	"github.com/leodip/goiabada/internal/entities"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"golang.org/x/exp/slog"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Database struct {
	DB *gorm.DB
}

func NewDatabase() (*Database, error) {

	dsn := fmt.Sprintf("%v:%v@tcp(%v:%v)/%v?charset=utf8mb4&parseTime=True&loc=UTC",
		viper.GetString("DB.Username"),
		viper.GetString("DB.Password"),
		viper.GetString("DB.Host"),
		viper.GetInt("DB.Port"),
		viper.GetString("DB.DbName"))

	logMsg := strings.Replace(dsn, viper.GetString("DB.Password"), "******", -1)
	slog.Info(fmt.Sprintf("using database: %v", logMsg))

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, errors.Wrap(err, "unable to open database")
	}

	var database = &Database{
		DB: db,
	}

	err = database.migrate()
	if err != nil {
		return nil, err
	}

	err = database.seed()
	if err != nil {
		return nil, err
	}

	return database, nil
}

func (d *Database) migrate() error {
	err := d.DB.AutoMigrate(
		&entities.Client{},
		&entities.Permission{},
		&entities.User{},
		&entities.UserConsent{},
		&entities.UserSession{},
		&entities.UserSessionClient{},
		&entities.RedirectURI{},
		&entities.Code{},
		&entities.KeyPair{},
		&entities.Settings{},
		&entities.PreRegistration{},
		&entities.Resource{},
		&entities.Group{},
		&entities.GroupAttribute{},
		&entities.UserAttribute{},
	)
	if err != nil {
		return errors.Wrap(err, "unable to migrate entities")
	}
	return err
}

func (d *Database) isDbEmpty() bool {
	if err := d.DB.First(&entities.Settings{}).Error; errors.Is(err, gorm.ErrRecordNotFound) {
		return true
	}
	return false
}

func (d *Database) GetClientByClientIdentifier(clientIdentifier string) (*entities.Client, error) {
	var client entities.Client

	result := d.DB.
		Preload("RedirectURIs").
		Preload("Permissions").
		Preload("Permissions.Resource").
		Where("client_identifier = ?", clientIdentifier).First(&client)

	if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
		return nil, errors.Wrap(result.Error, "unable to fetch client from database")
	}

	if result.RowsAffected == 0 {
		return nil, nil
	}

	return &client, nil
}

func (d *Database) GetUserByUsername(username string) (*entities.User, error) {
	var user entities.User

	result := d.DB.
		Preload(clause.Associations).
		Where("username = ?", username).First(&user)

	if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
		return nil, errors.Wrap(result.Error, "unable to fetch user from database")
	}

	if result.RowsAffected == 0 {
		return nil, nil
	}

	return &user, nil
}

func (d *Database) GetUserById(id uint) (*entities.User, error) {
	var user entities.User

	result := d.DB.
		Preload(clause.Associations).
		Where("id = ?", id).First(&user)

	if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
		return nil, errors.Wrap(result.Error, "unable to fetch user from database")
	}

	if result.RowsAffected == 0 {
		return nil, nil
	}

	return &user, nil
}

func (d *Database) GetUserBySubject(subject string) (*entities.User, error) {
	var user entities.User

	result := d.DB.
		Preload(clause.Associations).
		Where("subject = ?", subject).First(&user)

	if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
		return nil, errors.Wrap(result.Error, "unable to fetch user from database")
	}

	if result.RowsAffected == 0 {
		return nil, nil
	}

	return &user, nil
}

func (d *Database) GetUserByEmail(email string) (*entities.User, error) {
	var user entities.User

	result := d.DB.
		Preload(clause.Associations).
		Where("email = ?", email).First(&user)

	if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
		return nil, errors.Wrap(result.Error, "unable to fetch user from database")
	}

	if result.RowsAffected == 0 {
		return nil, nil
	}

	return &user, nil
}

func (d *Database) CreateCode(code *entities.Code) (*entities.Code, error) {
	result := d.DB.Create(code)

	if result.Error != nil {
		return nil, errors.Wrap(result.Error, "unable to create code in database")
	}

	return code, nil
}

func (d *Database) GetCode(code string, used bool) (*entities.Code, error) {
	var c entities.Code

	result := d.DB.
		Preload("Client").
		Preload("User").
		Preload("User.Permissions").
		Preload("User.Groups").
		Where("code = ? and used = ?", code, used).First(&c)

	if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
		return nil, errors.Wrap(result.Error, "unable to fetch code from database")
	}

	if result.RowsAffected == 0 {
		return nil, nil
	}

	return &c, nil
}

func (d *Database) GetSigningKey() (*entities.KeyPair, error) {
	var c entities.KeyPair

	result := d.DB.Order("Id desc").First(&c) // most recent

	if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
		return nil, errors.Wrap(result.Error, "unable to fetch keypair from database")
	}

	if result.RowsAffected == 0 {
		return nil, nil
	}

	return &c, nil
}

func (d *Database) GetSettings() (*entities.Settings, error) {
	var settings entities.Settings

	var result = d.DB.First(&settings)
	if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
		return nil, errors.Wrap(result.Error, "unable to fetch settings from database")
	}

	if result.RowsAffected == 0 {
		return nil, nil
	}

	return &settings, nil
}

func (d *Database) GetAllResources() ([]entities.Resource, error) {
	var resources []entities.Resource

	result := d.DB.Find(&resources)

	if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
		return nil, errors.Wrap(result.Error, "unable to fetch resources from database")
	}

	if result.RowsAffected == 0 {
		return []entities.Resource{}, nil
	}

	return resources, nil
}

func (d *Database) GetResourceById(id uint) (*entities.Resource, error) {
	var res entities.Resource

	result := d.DB.
		Where("id = ?", id).First(&res)

	if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
		return nil, errors.Wrap(result.Error, "unable to fetch resource from database")
	}

	if result.RowsAffected == 0 {
		return nil, nil
	}

	return &res, nil
}

func (d *Database) GetResourceByResourceIdentifier(resourceIdentifier string) (*entities.Resource, error) {
	var res entities.Resource

	result := d.DB.
		Where("resource_identifier = ?", resourceIdentifier).First(&res)

	if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
		return nil, errors.Wrap(result.Error, "unable to fetch resource from database")
	}

	if result.RowsAffected == 0 {
		return nil, nil
	}

	return &res, nil
}

func (d *Database) GetResourcePermissions(resourceId uint) ([]entities.Permission, error) {
	var permissions []entities.Permission

	result := d.DB.Where("resource_id = ?", resourceId).Find(&permissions)

	if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
		return nil, errors.Wrap(result.Error, "unable to fetch resource permissions from database")
	}

	if result.RowsAffected == 0 {
		return []entities.Permission{}, nil
	}

	return permissions, nil
}

func (d *Database) GetUserSessionBySessionIdentifier(sessionIdentifier string) (*entities.UserSession, error) {
	var userSession entities.UserSession

	result := d.DB.
		Preload(clause.Associations).
		Preload("Clients.Client").
		Where("session_identifier = ?", sessionIdentifier).First(&userSession)

	if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
		return nil, errors.Wrap(result.Error, "unable to fetch user session from database")
	}

	if result.RowsAffected == 0 {
		return nil, nil
	}

	return &userSession, nil
}

func (d *Database) GetUserSessionsByUserId(userId uint) ([]entities.UserSession, error) {
	var userSessions []entities.UserSession

	result := d.DB.
		Preload(clause.Associations).
		Preload("Clients.Client").
		Where("user_id = ?", userId).Find(&userSessions)

	if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
		return nil, errors.Wrap(result.Error, "unable to fetch user sessions from database")
	}

	if result.RowsAffected == 0 {
		return []entities.UserSession{}, nil
	}
	return userSessions, nil
}

func (d *Database) UpdateUserSession(userSession *entities.UserSession) (*entities.UserSession, error) {

	result := d.DB.Session(&gorm.Session{FullSaveAssociations: true}).Updates(&userSession)

	if result.Error != nil {
		return nil, errors.Wrap(result.Error, "unable to update user session in database")
	}

	return userSession, nil
}

func (d *Database) CreateUserSession(userSession *entities.UserSession) (*entities.UserSession, error) {

	result := d.DB.Save(userSession)

	if result.Error != nil {
		return nil, errors.Wrap(result.Error, "unable to create user session in database")
	}

	return userSession, nil
}

func (d *Database) UpdateUser(user *entities.User) (*entities.User, error) {

	result := d.DB.Save(user)

	if result.Error != nil {
		return nil, errors.Wrap(result.Error, "unable to update user in database")
	}

	return user, nil
}

func (d *Database) GetUserConsent(userId uint, clientId uint) (*entities.UserConsent, error) {
	var consent *entities.UserConsent

	result := d.DB.
		Preload("Client").
		Preload("User").
		Where("user_id = ? and client_id = ?", userId, clientId).First(&consent)

	if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
		return nil, errors.Wrap(result.Error, "unable to fetch user consent from database")
	}

	if result.RowsAffected == 0 {
		return nil, nil
	}

	return consent, nil
}

func (d *Database) GetUserConsents(userId uint) ([]entities.UserConsent, error) {
	var consents []entities.UserConsent

	result := d.DB.
		Preload("Client").
		Where("user_id = ?", userId).Find(&consents)

	if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
		return nil, errors.Wrap(result.Error, "unable to fetch user consents from database")
	}

	if result.RowsAffected == 0 {
		return []entities.UserConsent{}, nil
	}
	return consents, nil
}

func (d *Database) DeleteUserConsent(consentId uint) error {
	result := d.DB.Unscoped().Delete(&entities.UserConsent{}, consentId)

	if result.Error != nil {
		return errors.Wrap(result.Error, "unable to delete user consent from database")
	}

	return nil
}

func (d *Database) DeleteUserSession(userSessionId uint) error {

	// user session clients
	result := d.DB.Exec("DELETE FROM user_session_clients WHERE user_session_id = ?", userSessionId)
	if result.Error != nil {
		return errors.Wrap(result.Error, "unable to delete user session clients from database")
	}

	result = d.DB.Unscoped().Delete(&entities.UserSession{}, userSessionId)

	if result.Error != nil {
		return errors.Wrap(result.Error, "unable to delete user session from database")
	}

	return nil
}

func (d *Database) SaveUserConsent(userConsent *entities.UserConsent) (*entities.UserConsent, error) {

	result := d.DB.Save(userConsent)

	if result.Error != nil {
		return nil, errors.Wrap(result.Error, "unable to update user consent in database")
	}

	return userConsent, nil
}

func (d *Database) UpdateCode(code *entities.Code) (*entities.Code, error) {

	result := d.DB.Save(code)

	if result.Error != nil {
		return nil, errors.Wrap(result.Error, "unable to update code in database")
	}

	return code, nil
}

func (d *Database) CreatePreRegistration(preRegistration *entities.PreRegistration) (*entities.PreRegistration, error) {
	result := d.DB.Create(preRegistration)

	if result.Error != nil {
		return nil, errors.Wrap(result.Error, "unable to create pre registration in database")
	}

	return preRegistration, nil
}

func (d *Database) GetPreRegistrationByEmail(email string) (*entities.PreRegistration, error) {
	var preRegistration entities.PreRegistration

	result := d.DB.
		Preload(clause.Associations).
		Where("email = ?", email).First(&preRegistration)

	if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
		return nil, errors.Wrap(result.Error, "unable to fetch pre registration from database")
	}

	if result.RowsAffected == 0 {
		return nil, nil
	}

	return &preRegistration, nil
}

func (d *Database) CreateUser(user *entities.User) (*entities.User, error) {

	result := d.DB.Save(user)

	if result.Error != nil {
		return nil, errors.Wrap(result.Error, "unable to create user in database")
	}

	return user, nil
}

func (d *Database) DeletePreRegistration(preRegistrationId uint) error {
	result := d.DB.Unscoped().Delete(&entities.PreRegistration{}, preRegistrationId)

	if result.Error != nil {
		return errors.Wrap(result.Error, "unable to delete pre registration from database")
	}

	return nil
}

func (d *Database) GetClients() ([]entities.Client, error) {
	var clients []entities.Client

	result := d.DB.Find(&clients)

	if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
		return nil, errors.Wrap(result.Error, "unable to fetch clients from database")
	}

	if result.RowsAffected == 0 {
		return []entities.Client{}, nil
	}

	return clients, nil
}

func (d *Database) GetClientById(id uint) (*entities.Client, error) {
	var client entities.Client

	result := d.DB.
		Preload(clause.Associations).
		Preload("Permissions.Resource").
		Where("id = ?", id).First(&client)

	if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
		return nil, errors.Wrap(result.Error, "unable to fetch client from database")
	}

	if result.RowsAffected == 0 {
		return nil, nil
	}

	return &client, nil
}

func (d *Database) UpdateClient(client *entities.Client) (*entities.Client, error) {

	result := d.DB.Save(client)

	if result.Error != nil {
		return nil, errors.Wrap(result.Error, "unable to update client in database")
	}

	return client, nil
}

func (d *Database) CreateRedirectURI(redirectURI *entities.RedirectURI) (*entities.RedirectURI, error) {
	result := d.DB.Create(redirectURI)

	if result.Error != nil {
		return nil, errors.Wrap(result.Error, "unable to create redirect uri in database")
	}

	return redirectURI, nil
}

func (d *Database) DeleteRedirectURI(redirectURIId uint) error {
	result := d.DB.Unscoped().Delete(&entities.RedirectURI{}, redirectURIId)

	if result.Error != nil {
		return errors.Wrap(result.Error, "unable to delete redirect uri from database")
	}

	return nil
}

func (d *Database) GetPermissionById(id uint) (*entities.Permission, error) {
	var permission entities.Permission

	result := d.DB.
		Preload(clause.Associations).
		Where("id = ?", id).First(&permission)

	if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
		return nil, errors.Wrap(result.Error, "unable to fetch permission from database")
	}

	if result.RowsAffected == 0 {
		return nil, nil
	}

	return &permission, nil
}

func (d *Database) DeleteClientPermission(clientId uint, permissionId uint) error {

	client, err := d.GetClientById(clientId)
	if err != nil {
		return err
	}

	permission, err := d.GetPermissionById(permissionId)
	if err != nil {
		return err
	}

	err = d.DB.Model(&client).Association("Permissions").Delete(permission)

	if err != nil {
		return errors.Wrap(err, "unable to delete client permission from database")
	}

	return nil
}

func (d *Database) AddClientPermission(clientId uint, permissionId uint) error {

	client, err := d.GetClientById(clientId)
	if err != nil {
		return err
	}

	permission, err := d.GetPermissionById(permissionId)
	if err != nil {
		return err
	}

	err = d.DB.Model(&client).Association("Permissions").Append(permission)

	if err != nil {
		return errors.Wrap(err, "unable to append client permission in database")
	}

	return nil
}

func (d *Database) DeleteClient(clientId uint) error {

	// delete user consents
	result := d.DB.Unscoped().Where("client_id = ?", clientId).Delete(&entities.UserConsent{})
	if result.Error != nil {
		return errors.Wrap(result.Error, "unable to delete user consents from database (to delete a client)")
	}

	// delete redirect uris
	result = d.DB.Unscoped().Where("client_id = ?", clientId).Delete(&entities.RedirectURI{})
	if result.Error != nil {
		return errors.Wrap(result.Error, "unable to delete redirect uris from database (to delete a client)")
	}

	// delete codes
	result = d.DB.Unscoped().Where("client_id = ?", clientId).Delete(&entities.Code{})
	if result.Error != nil {
		return errors.Wrap(result.Error, "unable to delete codes from database (to delete a client)")
	}

	// delete permissions assigned to client
	client, err := d.GetClientById(clientId)
	if err != nil {
		return err
	}

	for _, permission := range client.Permissions {
		err = d.DeleteClientPermission(clientId, permission.Id)
		if err != nil {
			return err
		}
	}

	// delete client
	result = d.DB.Unscoped().Delete(&entities.Client{}, clientId)

	if result.Error != nil {
		return errors.Wrap(result.Error, "unable to delete client from database")
	}

	return nil
}

func (d *Database) CreateClient(client *entities.Client) (*entities.Client, error) {
	result := d.DB.Create(client)

	if result.Error != nil {
		return nil, errors.Wrap(result.Error, "unable to create client in database")
	}

	return client, nil
}

func (d *Database) UpdateResource(resource *entities.Resource) (*entities.Resource, error) {

	result := d.DB.Save(resource)

	if result.Error != nil {
		return nil, errors.Wrap(result.Error, "unable to update resource in database")
	}

	return resource, nil
}

func (d *Database) GetPermissionByPermissionIdentifier(permissionIdentifier string) (*entities.Permission, error) {
	var permission entities.Permission

	result := d.DB.
		Preload(clause.Associations).
		Where("permission_identifier = ?", permissionIdentifier).First(&permission)

	if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
		return nil, errors.Wrap(result.Error, "unable to fetch permission from database")
	}

	if result.RowsAffected == 0 {
		return nil, nil
	}

	return &permission, nil
}

func (d *Database) CreatePermission(permission *entities.Permission) (*entities.Permission, error) {
	result := d.DB.Create(permission)

	if result.Error != nil {
		return nil, errors.Wrap(result.Error, "unable to create permission in database")
	}

	return permission, nil
}

func (d *Database) UpdatePermission(permission *entities.Permission) (*entities.Permission, error) {

	result := d.DB.Save(permission)

	if result.Error != nil {
		return nil, errors.Wrap(result.Error, "unable to update permission in database")
	}

	return permission, nil
}

func (d *Database) DeletePermission(permissionId uint) error {

	result := d.DB.Exec("DELETE FROM users_permissions WHERE permission_id = ?", permissionId)
	if result.Error != nil {
		return errors.Wrap(result.Error, "unable to delete user permissions from database")
	}

	result = d.DB.Exec("DELETE FROM groups_permissions WHERE permission_id = ?", permissionId)
	if result.Error != nil {
		return errors.Wrap(result.Error, "unable to delete group permissions from database")
	}

	result = d.DB.Exec("DELETE FROM clients_permissions WHERE permission_id = ?", permissionId)
	if result.Error != nil {
		return errors.Wrap(result.Error, "unable to delete client permissions from database")
	}

	result = d.DB.Unscoped().Delete(&entities.Permission{}, permissionId)
	if result.Error != nil {
		return errors.Wrap(result.Error, "unable to delete permission from database")
	}

	return nil
}

func (d *Database) DeleteResource(resourceId uint) error {

	permissions, err := d.GetResourcePermissions(resourceId)
	if err != nil {
		return err
	}

	for _, permission := range permissions {
		err = d.DeletePermission(permission.Id)
		if err != nil {
			return err
		}
	}

	result := d.DB.Unscoped().Delete(&entities.Resource{}, resourceId)

	if result.Error != nil {
		return errors.Wrap(result.Error, "unable to delete resource from database")
	}

	return nil
}

func (d *Database) GetAllGroups() ([]entities.Group, error) {
	var groups []entities.Group

	result := d.DB.Find(&groups)

	if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
		return nil, errors.Wrap(result.Error, "unable to fetch groups from database")
	}

	if result.RowsAffected == 0 {
		return []entities.Group{}, nil
	}

	return groups, nil
}

func (d *Database) GetGroupById(id uint) (*entities.Group, error) {
	var group entities.Group

	result := d.DB.
		Preload("Permissions").
		Where("id = ?", id).First(&group)

	if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
		return nil, errors.Wrap(result.Error, "unable to fetch group from database")
	}

	if result.RowsAffected == 0 {
		return nil, nil
	}

	return &group, nil
}

func (d *Database) GetGroupByGroupIdentifier(groupIdentifier string) (*entities.Group, error) {
	var group entities.Group

	result := d.DB.
		Where("group_identifier = ?", groupIdentifier).First(&group)

	if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
		return nil, errors.Wrap(result.Error, "unable to fetch group from database")
	}

	if result.RowsAffected == 0 {
		return nil, nil
	}

	return &group, nil
}

func (d *Database) UpdateGroup(group *entities.Group) (*entities.Group, error) {

	result := d.DB.Save(group)

	if result.Error != nil {
		return nil, errors.Wrap(result.Error, "unable to update group in database")
	}

	return group, nil
}

func (d *Database) GetGroupMembers(groupId uint, page int, pageSize int) ([]entities.User, int, error) {
	var users []entities.User

	result := d.DB.Raw("SELECT users.* FROM users_groups "+
		"INNER JOIN users ON users_groups.user_id = users.id "+
		"WHERE users_groups.group_id = ? "+
		"ORDER BY users.given_name ASC "+
		"LIMIT ?, ?", groupId, (page-1)*pageSize, pageSize).Scan(&users)

	if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
		return nil, 0, errors.Wrap(result.Error, "unable to fetch users from database")
	}

	if result.RowsAffected == 0 {
		return []entities.User{}, 0, nil
	}

	var total int64
	d.DB.Raw("SELECT COUNT(*) FROM users_groups WHERE users_groups.group_id = ?", groupId).Count(&total)

	return users, int(total), nil
}

func (d *Database) GetUserSessionsByClientId(clientId uint, page int, pageSize int) ([]entities.UserSession, int, error) {
	var userSessions []entities.UserSession

	result := d.DB.
		Preload(clause.Associations).
		Preload("Clients.Client").
		Joins("JOIN user_session_clients ON user_session_clients.user_session_id = user_sessions.id").
		Where("user_session_clients.client_id = ?", clientId).
		Order("user_sessions.last_accessed DESC").
		Limit(pageSize).
		Offset((page - 1) * pageSize).
		Find(&userSessions)

	if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
		return nil, 0, errors.Wrap(result.Error, "unable to fetch user sessions from database")
	}

	if result.RowsAffected == 0 {
		return []entities.UserSession{}, 0, nil
	}

	var total int64
	d.DB.Raw("SELECT COUNT(*) FROM user_sessions "+
		"INNER JOIN user_session_clients ON user_session_clients.user_session_id = user_sessions.id "+
		"WHERE user_session_clients.client_id = ? ", clientId).Count(&total)

	return userSessions, int(total), nil
}

func (d *Database) CreateResource(resource *entities.Resource) (*entities.Resource, error) {
	result := d.DB.Create(resource)

	if result.Error != nil {
		return nil, errors.Wrap(result.Error, "unable to create resource in database")
	}

	return resource, nil
}

func (d *Database) AddUserToGroup(user *entities.User, group *entities.Group) error {

	err := d.DB.Model(&user).Association("Groups").Append(group)

	if err != nil {
		return errors.Wrap(err, "unable to append user to group in database")
	}

	return nil
}

func (d *Database) RemoveUserFromGroup(user *entities.User, group *entities.Group) error {

	err := d.DB.Model(&user).Association("Groups").Delete(group)

	if err != nil {
		return errors.Wrap(err, "unable to remove user from group in database")
	}

	return nil
}

func (d *Database) CreateGroup(group *entities.Group) (*entities.Group, error) {
	result := d.DB.Create(group)

	if result.Error != nil {
		return nil, errors.Wrap(result.Error, "unable to create group in database")
	}

	return group, nil
}

func (d *Database) CountMembers(groupId uint) (int, error) {
	var total int64
	d.DB.Raw("SELECT COUNT(*) FROM users_groups WHERE users_groups.group_id = ?", groupId).Count(&total)

	return int(total), nil
}

func (d *Database) DeleteGroup(groupId uint) error {

	result := d.DB.Exec("DELETE FROM users_groups WHERE group_id = ?", groupId)
	if result.Error != nil {
		return errors.Wrap(result.Error, "unable to delete user groups from database")
	}

	result = d.DB.Exec("DELETE FROM group_attributes WHERE group_id = ?", groupId)
	if result.Error != nil {
		return errors.Wrap(result.Error, "unable to delete group attributes from database")
	}

	result = d.DB.Exec("DELETE FROM groups_permissions WHERE group_id = ?", groupId)
	if result.Error != nil {
		return errors.Wrap(result.Error, "unable to delete group permissions from database")
	}

	result = d.DB.Unscoped().Delete(&entities.Group{}, groupId)
	if result.Error != nil {
		return errors.Wrap(result.Error, "unable to delete group from database")
	}

	return nil
}

func (d *Database) GetGroupAttributes(groupId uint) ([]entities.GroupAttribute, error) {
	var attributes []entities.GroupAttribute

	result := d.DB.
		Preload(clause.Associations).
		Where("group_id = ?", groupId).Order("`key` ASC").Find(&attributes)

	if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
		return nil, errors.Wrap(result.Error, "unable to fetch attributes from database")
	}

	if result.RowsAffected == 0 {
		return []entities.GroupAttribute{}, nil
	}

	return attributes, nil
}

func (d *Database) DeleteGroupAttributeById(groupAttributeId uint) error {

	result := d.DB.Unscoped().Delete(&entities.GroupAttribute{}, groupAttributeId)

	if result.Error != nil {
		return errors.Wrap(result.Error, "unable to delete group attribute from database")
	}

	return nil
}

func (d *Database) CreateGroupAttribute(groupAttribute *entities.GroupAttribute) (*entities.GroupAttribute, error) {
	result := d.DB.Create(groupAttribute)

	if result.Error != nil {
		return nil, errors.Wrap(result.Error, "unable to create group attribute in database")
	}

	return groupAttribute, nil
}

func (d *Database) GetGroupAttributeById(attributeId uint) (*entities.GroupAttribute, error) {
	var attr entities.GroupAttribute

	result := d.DB.
		Preload(clause.Associations).
		Where("id = ?", attributeId).First(&attr)

	if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
		return nil, errors.Wrap(result.Error, "unable to fetch group attribute from database")
	}

	if result.RowsAffected == 0 {
		return nil, nil
	}

	return &attr, nil
}

func (d *Database) UpdateGroupAttribute(groupAttribute *entities.GroupAttribute) (*entities.GroupAttribute, error) {

	result := d.DB.Save(groupAttribute)

	if result.Error != nil {
		return nil, errors.Wrap(result.Error, "unable to update group attribute in database")
	}

	return groupAttribute, nil
}

func (d *Database) AddGroupPermission(groupId uint, permissionId uint) error {

	group, err := d.GetGroupById(groupId)
	if err != nil {
		return err
	}

	permission, err := d.GetPermissionById(permissionId)
	if err != nil {
		return err
	}

	err = d.DB.Model(&group).Association("Permissions").Append(permission)

	if err != nil {
		return errors.Wrap(err, "unable to append group permission in database")
	}

	return nil
}

func (d *Database) DeleteGroupPermission(groupId uint, permissionId uint) error {

	group, err := d.GetGroupById(groupId)
	if err != nil {
		return err
	}

	permission, err := d.GetPermissionById(permissionId)
	if err != nil {
		return err
	}

	err = d.DB.Model(&group).Association("Permissions").Delete(permission)

	if err != nil {
		return errors.Wrap(err, "unable to delete group permission from database")
	}

	return nil
}

func (d *Database) GetUsers(query string, page int, pageSize int) ([]entities.User, int, error) {
	var users []entities.User

	var result *gorm.DB
	var where string

	query = strings.TrimSpace(query)
	if query == "" {
		// no search filter
		result = d.DB.
			Preload("Groups").
			Limit(pageSize).
			Offset((page - 1) * pageSize).
			Find(&users)

	} else {
		// with search filter
		where = "subject LIKE ? OR " +
			"username LIKE ? OR " +
			"given_name LIKE ? OR " +
			"middle_name LIKE ? OR " +
			"family_name LIKE ? OR " +
			"email LIKE ? "
		query = "%" + query + "%"

		result = d.DB.
			Preload("Groups").
			Limit(pageSize).
			Offset((page-1)*pageSize).
			Where(where, query, query, query, query, query, query).
			Find(&users)
	}

	if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
		return nil, 0, errors.Wrap(result.Error, "unable to fetch users from database")
	}

	if result.RowsAffected == 0 {
		return []entities.User{}, 0, nil
	}

	var total int64
	if query == "" {
		d.DB.Raw("SELECT COUNT(*) FROM users").Count(&total)
	} else {
		d.DB.Raw("SELECT COUNT(*) FROM users WHERE "+where,
			query, query, query, query, query, query).Count(&total)
	}

	return users, int(total), nil
}

func (d *Database) AddUserPermission(userId uint, permissionId uint) error {

	user, err := d.GetUserById(userId)
	if err != nil {
		return err
	}

	permission, err := d.GetPermissionById(permissionId)
	if err != nil {
		return err
	}

	err = d.DB.Model(&user).Association("Permissions").Append(permission)

	if err != nil {
		return errors.Wrap(err, "unable to append user permission in database")
	}

	return nil
}

func (d *Database) DeleteUserPermission(userId uint, permissionId uint) error {

	user, err := d.GetUserById(userId)
	if err != nil {
		return err
	}

	permission, err := d.GetPermissionById(permissionId)
	if err != nil {
		return err
	}

	err = d.DB.Model(&user).Association("Permissions").Delete(permission)

	if err != nil {
		return errors.Wrap(err, "unable to delete user permission from database")
	}

	return nil
}

func (d *Database) GetUserAttributes(userId uint) ([]entities.UserAttribute, error) {
	var attributes []entities.UserAttribute

	result := d.DB.
		Preload(clause.Associations).
		Where("user_id = ?", userId).Order("`key` ASC").Find(&attributes)

	if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
		return nil, errors.Wrap(result.Error, "unable to fetch attributes from database")
	}

	if result.RowsAffected == 0 {
		return []entities.UserAttribute{}, nil
	}

	return attributes, nil
}

func (d *Database) DeleteUserAttributeById(userAttributeId uint) error {

	result := d.DB.Unscoped().Delete(&entities.UserAttribute{}, userAttributeId)

	if result.Error != nil {
		return errors.Wrap(result.Error, "unable to delete user attribute from database")
	}

	return nil
}

func (d *Database) CreateUserAttribute(userAttribute *entities.UserAttribute) (*entities.UserAttribute, error) {
	result := d.DB.Create(userAttribute)

	if result.Error != nil {
		return nil, errors.Wrap(result.Error, "unable to create user attribute in database")
	}

	return userAttribute, nil
}

func (d *Database) GetUserAttributeById(attributeId uint) (*entities.UserAttribute, error) {
	var attr entities.UserAttribute

	result := d.DB.
		Preload(clause.Associations).
		Where("id = ?", attributeId).First(&attr)

	if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
		return nil, errors.Wrap(result.Error, "unable to fetch user attribute from database")
	}

	if result.RowsAffected == 0 {
		return nil, nil
	}

	return &attr, nil
}

func (d *Database) UpdateUserAttribute(userAttribute *entities.UserAttribute) (*entities.UserAttribute, error) {

	result := d.DB.Save(userAttribute)

	if result.Error != nil {
		return nil, errors.Wrap(result.Error, "unable to update user attribute in database")
	}

	return userAttribute, nil
}

func (d *Database) DeleteUser(user *entities.User) error {

	// codes
	result := d.DB.Exec("DELETE FROM codes WHERE user_id = ?", user.Id)
	if result.Error != nil {
		return errors.Wrap(result.Error, "unable to delete user codes from database")
	}

	// user attributes
	result = d.DB.Exec("DELETE FROM user_attributes WHERE user_id = ?", user.Id)
	if result.Error != nil {
		return errors.Wrap(result.Error, "unable to delete user attributes from database")
	}

	// user consents
	result = d.DB.Exec("DELETE FROM user_consents WHERE user_id = ?", user.Id)
	if result.Error != nil {
		return errors.Wrap(result.Error, "unable to delete user consents from database")
	}

	// user sessions
	result = d.DB.Exec("DELETE FROM user_sessions WHERE user_id = ?", user.Id)
	if result.Error != nil {
		return errors.Wrap(result.Error, "unable to delete user sessions from database")
	}

	// user groups
	result = d.DB.Exec("DELETE FROM users_groups WHERE user_id = ?", user.Id)
	if result.Error != nil {
		return errors.Wrap(result.Error, "unable to delete user groups from database")
	}

	// user permissions
	result = d.DB.Exec("DELETE FROM users_permissions WHERE user_id = ?", user.Id)
	if result.Error != nil {
		return errors.Wrap(result.Error, "unable to delete user permissions from database")
	}

	// delete user
	result = d.DB.Unscoped().Delete(user)

	if result.Error != nil {
		return errors.Wrap(result.Error, "unable to delete user from database")
	}

	return nil

}

func (d *Database) SaveSettings(settings *entities.Settings) (*entities.Settings, error) {

	result := d.DB.Save(settings)

	if result.Error != nil {
		return nil, errors.Wrap(result.Error, "unable to update settings in database")
	}

	return settings, nil
}
