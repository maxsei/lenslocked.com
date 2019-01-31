package main

import (
	"fmt"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"lenslocked.com/models"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "$m0kycat"
	dbname   = "lenslocked_dev"
)

// User is an example
type User struct {
	gorm.Model
	Name  string
	Email string `gorm:"not null;unique_index"`
}

func main() {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname,
	)
	us, err := models.NewUserService(psqlInfo)
	if err != nil {
		panic(err)
	}
	defer us.Close()
	// newUser := models.User{
	// 	Name:  "Michael Scott",
	// 	Email: "michael@papermifflin.com",
	// }
	// if err = us.Create(&newUser); err != nil {
	// 	panic(err)
	// }
	// user, err := us.ByEmail("michael@papermifflin.com")
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println(user)
	// if err := us.Delete(1); err != nil {
	// 	panic(err)
	// }
	// user.Email = "michael@anotherpaperco.com"
	// if err := us.Update(user); err != nil {
	// 	panic(err)
	// }

	// us.DestructiveReset()
}
