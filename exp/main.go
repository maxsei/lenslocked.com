package main

import (
	"fmt"

	"lenslocked.com/models"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "$m0kycat"
	dbname   = "lenslocked_dev"
)

func main() {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname,
	)
	us, err := models.NewUserService(psqlInfo)
	us.DestructiveReset()
	if err != nil {
		panic(err)
	}
	defer us.Close()
	us.AutoMigrate()
	user1 := models.User{
		Name:     "max schulte",
		Email:    "max@max.com",
		Password: "max",
		Remember: "1235",
	}
	err = us.Create(&user1)
	if err != nil {
		panic(err)
	}
	user2, err := us.ByRemember("1235")
	if err != nil {
		panic(err)
	}
	fmt.Printf("User 1: %#v\n", user1)
	fmt.Printf("User 2: %#v\n", user2)
	user2.Password = "$m0kycat"
	user2.Remember = "something else"
	if err := us.Update(user2); err != nil {
		panic(err)
	}
	fmt.Printf("User 2 updated: %#v\n", user2)
}
