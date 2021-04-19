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

var jwtKey = []byte(os.Getenv("JWT_SECRET_KEY"))

type User struct {
	ID        uint64    `json:"id"`
	Username  string    `json:"username" binding:"required,min=3,max=100"`
	Email     string    `json:"email" binding:"required,email"`
	Password  string    `json:"password" binding:"required,min=6,max=100"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Claims struct {
	UserId string `json:"email"`
	jwt.StandardClaims
}

type TokenDetails struct {
	AccessToken  string
	RefreshToken string
}

// Create token
func CreateToken(user_id string) (string, error) {

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
	u.Email = payload.Email
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
	row := db.QueryRow(query, email)
	row.Scan(&u.ID)

	return u.ID, nil
}

func VerifyAccountModel(email string) error {
	db := database.InitDB()
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
