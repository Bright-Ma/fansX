package database

type User struct {
	Id       int64  `gorm:"primary_key"`
	Phone    string `gorm:"size:11;unique;not null;index:phone"`
	Name     string `gorm:"unique;not null;"`
	Password string `gorm:"not null;"`
}
