package service

import (
	"fmt"
	"reflect"
	"socialai/backend"
	"socialai/constants"
	"socialai/model"

	"github.com/olivere/elastic/v7"
)

// 1. username + password => Good
// 2. username existed, pwd wrong
// 3. username not existed
// 4. connect to ES: error
func CheckUser(username, password string) (bool, error) {
	query := elastic.NewBoolQuery()
	query.Must(elastic.NewTermQuery("username", username))
	query.Must(elastic.NewTermQuery("password", password))
	searchResult, err := backend.ESBackend.ReadFromES(query, constants.USER_INDEX)
	if err != nil {
		return false, err
	}

	var utype model.User
	for _, item := range searchResult.Each(reflect.TypeOf(utype)) {
		u := item.(model.User)
		if u.Password == password {
			fmt.Printf("Login as %s\n", username)
			return true, nil
		}
	}
	return false, nil
}

func AddUser(user *model.User) (bool, error) {
	// 1. verify username existed
	query := elastic.NewTermQuery("username", user.Username)
	searchResult, err := backend.ESBackend.ReadFromES(query, constants.USER_INDEX)
	if err != nil {
		return false, err
	}
	if searchResult.TotalHits() > 0 {
		return false, nil
	}

	// if not existe, add to ES
	err = backend.ESBackend.SaveToES(user, constants.USER_INDEX, user.Username)
	if err != nil {
		return false, err
	}
	fmt.Printf("User is added: %s\n", user.Username)
	return true, nil
}
