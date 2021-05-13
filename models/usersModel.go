package models

import (
	jwt "github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
	"golangblog/config"
	"golangblog/schemas"
	"log"
	"os"
	"strings"
	"time"
)

var (
	jwtKey = []byte(os.Getenv("JWT_SECRET_KEY"))
)

type User struct {
	ID        uint64    `json:"id"`
	Username  string    `json:"username" binding:"required,min=3,max=100"`
	Email     string    `json:"email" binding:"required,email"`
	Password  string    `json:"password" binding:"required,min=6,max=100"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type SendReset struct {
	Email string `json:"email" binding:"required,email"`
}

type Claims struct {
	UserId string `json:"email"`
	jwt.StandardClaims
}

type UserGoogle struct {
	ID            uint64 `json:"id"`
	Email         string `json:"email"`
	Username      string `json:"name"`
	VerifiedEmail bool   `json:"verified_email"`
	Avatar        string `json:"picture"`
}

type UserFacebook struct {
	ID       uint64 `json:"id"`
	Email    string `json:"email"`
	Username string `json:"name"`
	Avatar   string `json:"avatar"`
}

type ListUser struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Create token
func CreateAuthToken(user_id string) (string, error) {

	expiredTime := time.Now().Add(time.Minute * 15).Unix() // token expired after 15 minutes
	claims := &Claims{
		UserId: user_id,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expiredTime,
		},
	}
	// Generate token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// sign token with secret key encoding
	tokenString, err := token.SignedString(jwtKey)

	return tokenString, err
}

func DecodeAuthToken(tokenStr string) (string, error) {
	claims := &Claims{}

	tkn, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			return "", err
		}
		return "", err
	}

	if !tkn.Valid {
		return "", err
	}

	return claims.UserId, nil
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

	db := config.InitDB()
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
	u.Email = payload.Email
	u.ID = lastId

	return u, nil
}

func VerifyLogin(email, password string) (User, error) {
	var user User

	db := config.InitDB()
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

	db := config.InitDB()
	defer db.Close()

	query := `select count(id) from users where email = $1`

	db.QueryRow(query, email).Scan(&counter)

	return counter
}

func FindUserByEmail(email string) (uint64, error) {
	var u User

	db := config.InitDB()
	defer db.Close()

	query := `select id from users where email = $1`
	row := db.QueryRow(query, email)
	row.Scan(&u.ID)

	return u.ID, nil
}

func VerifyAccountModel(email string) error {
	db := config.InitDB()
	defer db.Close()

	query := `update confirmation_users set activated = $1 where user_id = $2`

	stmt, err := db.Prepare(query)
	if err != nil {
		panic(err)
	}

	user_id, _ := FindUserByEmail(email)
	stmt.Exec(true, user_id)

	return nil
}

func SaveGoogleUser(u UserGoogle) (UserGoogle, error) {
	db := config.InitDB()
	defer db.Close()

	query := `insert into users(username,email) values($1,$2) returning id`

	stmt, err := db.Prepare(query)
	if err != nil {
		panic(err)
	}

	var lastId uint64
	queryErr := stmt.QueryRow(u.Username, u.Email).Scan(&lastId)
	if queryErr != nil {
		log.Fatal(queryErr)
	}
	u.ID = lastId

	return u, nil
}

func SaveFacebookUser(username, email string) (UserFacebook, error) {
	var u UserFacebook

	db := config.InitDB()
	defer db.Close()

	query := `insert into users(username,email) values($1,$2) returning id`

	stmt, err := db.Prepare(query)
	if err != nil {
		panic(err)
	}

	var lastId uint64
	queryErr := stmt.QueryRow(username, email).Scan(&lastId)
	if queryErr != nil {
		log.Fatal(queryErr)
	}
	u.ID = lastId
	u.Username = username
	u.Email = email

	return u, nil
}

func GetUser() []ListUser {

	// create list of user
	users := []ListUser{}

	db := config.InitDB()
	defer db.Close()

	query := `select username,email,password from users`

	rows, _ := db.Query(query)
	for rows.Next() {
		var u ListUser

		rows.Scan(&u.Username, &u.Email, &u.Password)
		users = append(users, u)
	}

	return users
}

func CheckEmailPasswordReset(email string) int {
	var counter int

	db := config.InitDB()
	defer db.Close()

	query := `select count(id) from password_resets where email = $1`

	db.QueryRow(query, email).Scan(&counter)

	return counter
}
