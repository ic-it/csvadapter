package main

import (
	"fmt"
	"slices"
	"strings"

	"github.com/ic-it/csvadapter"
)

type Email string

func (e *Email) UnmarshalText(text []byte) error {
	*e = Email(text)
	return nil
}

type User struct {
	Password  string `csva:"password"`
	ID        int    `csva:"id"`
	Username  string `csva:"user"`
	Email     Email  `csva:"email"`
	SomeOther bool   `csva:"someother,omitempty"`
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
	users, err := adapter.FromCSV(reader)
	if err != nil {
		panic(err)
	}

	fmt.Println("Users:")
	for user, err := range users {
		if err != nil {
			panic(err)
		}
		fmt.Println(user)
	}

	users2 := []User{
		{ID: 1, Username: "admin", Password: "123456", Email: "asd@asd.cc", SomeOther: true},
		{ID: 2, Username: "asdasd", Password: "asdasdad", Email: "dsa@dsa.cc", SomeOther: false},
	}

	writer := strings.Builder{}
	err = adapter.ToCSV(&writer, slices.Values(users2))
	if err != nil {
		panic(err)
	}

	fmt.Println("CSV:")
	fmt.Println(writer.String())
}
