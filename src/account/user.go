package account

import (
	"errors"
	"github.com/stock-simulator-server/src/change"
	"github.com/stock-simulator-server/src/lock"
	"github.com/stock-simulator-server/src/portfolio"
	"github.com/stock-simulator-server/src/utils"
)

// keep the uuid to user
var userList = make(map[string]*User)

// keep the username to uuid list
var uuidList = make(map[string]string)
var userListLock = lock.NewLock("user-list")

/*
User Object
Represents a unique individual of the system
*/
type User struct {
	UserName      string     `json:"-"`
	Password      string     `json:"-"`
	DisplayName   string     `json:"display_name" change:"-"`
	Uuid          string     `json:"-"`
	Active        bool       `json:"active" change:"-"`
	ActiveClients int64      `json:"-"`
	Lock          *lock.Lock `json:"-"`
}

/**
Return a user provided the username and Password
If the Password is correct return user, else return err
*/
func GetUser(username, password string) (*User, error) {
	userListLock.Acquire("get-user")
	defer userListLock.Release()
	userUuid, exists := uuidList[username]
	if !exists {
		return nil, errors.New("user does not exist")
	}
	user := userList[userUuid]
	if user.Password != password {
		return nil, errors.New("Password is incorrect")
	}
	user.Active = true
	change.SubscribeUpdateInputs.Offer(user)
	return user, nil
}

/**
Build a new user
set their Password to that provided
*/
func NewUser(username, password string) (*User, error) {
	uuid := utils.PseudoUuid()
	return MakeUser(uuid, username, username, password)
}

func MakeUser(uuid, username, displayName, password string) (*User, error) {
	userListLock.Acquire("new-user")
	defer userListLock.Release()
	_, userNameExists := uuidList[username]
	if userNameExists {
		return nil, errors.New("username already exists")
	}
	uuidList[username] = uuid
	portfolio.NewPortfolio(uuid, username)
	userList[uuid] = &User{
		UserName:    username,
		DisplayName: displayName,
		Password:    password,
		Uuid:        uuid,
		Lock:        lock.NewLock("user"),
		Active:      true,
	}
	change.NewSubscribeCreated.Offer(userList[uuid])
	utils.RegisterUuid(uuid, userList[uuid])
	return userList[uuid], nil

}

/**
Logout the user and decrement the active client count
*/
func (user *User) LogoutUser() {
	user.Lock.Acquire("logout")
	defer user.Lock.Release()
	user.ActiveClients -= 1
	if user.ActiveClients < 0 {
		user.ActiveClients = 0
	}
	if user.ActiveClients == 0 {
		user.Active = false
	}
	change.SubscribeUpdateInputs.Offer(user)
}

func (user *User) GetId() string {
	return user.Uuid
}
func (user *User) GetType() string {
	return "user"
}

/**
Turn the user map into a list so they can be sent to a rx client
*/
func GetAllUsers() []*User {
	userListLock.Acquire("get all users")
	defer userListLock.Release()
	lst := make([]*User, len(userList))
	i := 0
	for _, val := range userList {
		lst[i] = val
		i += 1
	}
	return lst
}
