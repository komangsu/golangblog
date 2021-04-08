package models

import (
	jwt "github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
	"golangblog/database"
	"log"
	"os"
	"strings"
	"time"
)

type User struct {
	ID              uint64    `json:"id"`
	Username        string    `json:"username" binding:"required,min=3,max=100"`
	Email           string    `json:"email" binding:"required,email"`
	Password        string    `json:"password" binding:"required,min=6,max=100"`
	ConfirmPassword string    `json:"confirm_password" binding:"required,eqfield=Password"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
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
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func Hash(password string) ([]byte, error) {
	return bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
}

func CheckPassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

// hash user password
func (u *User) BeforeSave() error {
	hashedPassword, err := Hash(u.Password)
	if err != nil {
		return err
	}
	u.Password = string(hashedPassword)
	return nil

}

// Prepare strips user input of any white spaces
func (u *User) Prepare() {
	u.Username = strings.TrimSpace(u.Username)
	u.Email = strings.TrimSpace(u.Email)
}

// Create user
func (u *User) SaveUser() (*User, error) {
	db := database.InitDB()
	defer db.Close()

	query := `INSERT INTO users(username,email,password) VALUES($1,$2,$3) RETURNING id`

	stmt, err := db.Prepare(query)
	if err != nil {
		panic(err)
	}

	// hash password
	hashErr := u.BeforeSave()
	if hashErr != nil {
		log.Fatal(hashErr)
	}

	queryErr := stmt.QueryRow(&u.Username, &u.Email, &u.Password).Scan(&u.ID)
	if queryErr != nil {
		panic(queryErr)
	}

	return u, nil
}

func GetUserByUsername(email string) (User, error) {
	var u User

	db := database.InitDB()
	defer db.Close()

	query := `SELECT id,username,email,password FROM users WHERE email=$1`

	row := db.QueryRow(query, email)
	errRow := row.Scan(&u.ID, &u.Username, &u.Email, &u.Password)
	if errRow != nil {
		log.Fatal("email not found")
	}

	return u, nil
}

// Login User
func SignIn(email, password string) (User, error) {
	var user User

	db := database.InitDB()
	defer db.Close()

	query := `SELECT id,username,email,password FROM users WHERE email = $1`

	row := db.QueryRow(query, email)
	errRow := row.Scan(&user.ID, &user.Username, &user.Email, &user.Password)
	if errRow != nil {
		log.Fatal("user not found.")
	}

	// Check password
	err := CheckPassword(user.Password, password)
	if err != nil && err == bcrypt.ErrMismatchedHashAndPassword { // mismatch return when a password and hash do not match.
		log.Fatal("Invalid login credentials")
	}

	return user, nil
}
