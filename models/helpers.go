package models

import "github.com/jinzhu/gorm"

// first will query using the provided gorm.DB and
// will get the first item returned and place it into
// dst.  It will return error not found if nothing is found
func first(db *gorm.DB, dst interface{}) error {
	err := db.First(dst).Error
	if err == gorm.ErrRecordNotFound {
		return ErrNotFound
	}
	return err
}
