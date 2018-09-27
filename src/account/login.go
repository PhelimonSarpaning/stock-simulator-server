package account

import (
	"errors"
	"github.com/stock-simulator-server/src/portfolio"
	"github.com/stock-simulator-server/src/session"
	"github.com/stock-simulator-server/src/utils"
)

/**
Return a user provided the username and Password
If the Password is correct return user, else return err
*/
func ValidateUser(username, password string) (string, error){
	userListLock.Acquire("get-user")
	defer userListLock.Release()
	userUuid, exists := uuidList[username]
	if !exists {
		return "", errors.New("user does not exist")
	}
	user := userList[userUuid]

	if !comparePasswords(user.Password, password) {
		return "", errors.New("password is incorrect")
	}
	sessionToken := session.NewSessionToken(user.Uuid)
	return sessionToken, nil
}

/**
Renew a user user a session token
 */
func ConnectUser(sessionToken string)(*User, error) {
	userId, err := session.GetUserId(sessionToken)
	if err != nil{
		return nil, err
	}
	userListLock.Acquire("renew-user")
	defer userListLock.Release()
	user, exists := userList[userId]
	if !exists{
		return nil, errors.New("user found in session list but not in current users")
	}
	user.Active = true
	UpdateChannel.Offer(user)
	return user, nil
}

/**
Build a new user
set their Password to that provided
*/
func NewUser(username, password string) (string, error) {
	uuid := utils.PseudoUuid()
	if len(username) > 20{
		return "", errors.New("username too long")
	}
	if len(username) < 4{
		return "", errors.New("username too short")
	}

	if len(password) < minPasswordLength{
		return "", errors.New("password too short")
	}
	hashedPassword := hashAndSalt(password)
	user, err := MakeUser(uuid, username, username, hashedPassword, "", `{"swag":"420"}`)
	if err != nil {
		utils.RemoveUuid(uuid)
		return "", err
	}
	port, _ := portfolio.NewPortfolio(uuid)
	user.PortfolioId = port.UUID
	sessionToken := session.NewSessionToken(user.Uuid)
	return sessionToken, nil
}