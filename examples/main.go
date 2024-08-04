package main

import (
	"fmt"
	"strings"

	"github.com/ic-it/csvadapter"
)

type Email string

func (e *Email) UnmarshalText(text []byte) error {
	*e = Email(text)
	return nil
}

type User struct {
	Password  string `csvadapter:"password"`
	ID        int    `csvadapter:"id"`
	Username  string `csvadapter:"user"`
	Email     Email  `csvadapter:"email"`
	SomeOther bool   `csvadapter:"someother,omitempty"`
}

func (u User) String() string {
	return fmt.Sprintf("User(ID: %d, Username: %s, Email: %s, Password: %s, SomeOther: %t)", u.ID, u.Username, u.Email, u.Password, u.SomeOther)
}

func main() {
	adapter, err := csvadapter.NewCSVAdapter[User]()
	if err != nil {
		panic(err)
	}
	fmt.Println(adapter)

	reader := strings.NewReader(`id,user,password,email,someother,
1,admin,123456,test@mail.cc,true,
2,asdasd,asdasdad,test@mail.cc,,
`)
	var users []User
	err = adapter.EatCSV(reader, &users)
	if err != nil {
		panic(err)
	}

	fmt.Println(users)
}
