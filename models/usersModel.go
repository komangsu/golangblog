package models

import (
	jwt "github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
	"golangblog/database"
	"golangblog/schemas"
	"log"
	"os"
	"strings"
	"time"
)

type User struct {
	ID        uint64    `json:"id"`
	Username  string    `json:"username" binding:"required,min=3,max=100"`
	Email     string    `json:"email" binding:"required,email"`
	Password  string    `json:"password" binding:"required,min=6,max=100"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type LoginUser struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6,max=100"`
}

// Create token
func CreateToken(user_id uint64) (string, error) {

	claims := jwt.MapClaims{}

	claims["authorized"] = true
	claims["user_id"] = user_id
	claims["exp"] = time.Now().Add(time.Minute * 15).Unix() // token expired after 15 minutes

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET_KEY")))
	return tokenString, err
}

func Hash(password string) ([]byte, error) {
	return bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
}

func CheckPassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

// Prepare strips user input of any white spaces
func (u *User) Prepare() {
	u.Username = strings.TrimSpace(u.Username)
	u.Email = strings.TrimSpace(u.Email)
}

// Create user
func SaveUser(payload schemas.RegisterUser) (User, error) {
	var u User

	db := database.InitDB()
	defer db.Close()

	query := `insert into users(username,email,password) values($1,$2,$3) returning id`

	stmt, err := db.Prepare(query)
	if err != nil {
		panic(err)
	}

	// hash password
	hashedPassword, errHash := Hash(payload.Password)
	if errHash != nil {
		log.Fatal(errHash)
	}
	payload.Password = string(hashedPassword)

	var lastId uint64

	queryErr := stmt.QueryRow(payload.Username, payload.Email, payload.Password).Scan(&lastId)
	if queryErr != nil {
		log.Fatal(queryErr)
	}
	u.Username = payload.Username
	u.ID = lastId

	return u, nil
}

func VerifyLogin(email, password string) (User, error) {
	var user User

	db := database.InitDB()
	defer db.Close()

	query := `select id,email,password from users where email = $1`

	row := db.QueryRow(query, email)
	row.Scan(&user.ID, &user.Email, &user.Password)

	// Check password
	err := CheckPassword(user.Password, password)

	return user, err
}

func CheckEmailExists(email string) int {

	var counter int

	db := database.InitDB()
	defer db.Close()

	query := `select count(id) from users where email = $1`

	db.QueryRow(query, email).Scan(&counter)

	return counter
}

func FindUserByEmail(email string) (uint64, error) {
	var u User

	db := database.InitDB()
	defer db.Close()

	query := `select id from users where email = $1`
	db.QueryRow(query, email)

	id := u.ID

	return id, nil
}
