package database

import (
	"context"
	"fmt"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	ctx    = context.Background()
	client = &redisClient{}

	access_secret  = []byte(os.Getenv("ACCESS_SECRET"))
	refresh_secret = []byte(os.Getenv("REFRESH_SECRET"))

	redisAddr = os.Getenv("REDIS_DB_HOST") + ":6379"
)

type redisClient struct {
	c *redis.Client
}

type TokenDetails struct {
	AccessToken  string
	RefreshToken string
	AccessUuid   string
	RefreshUuid  string
	AtExpires    int64
	RtExpires    int64
}

type AccessDetails struct {
	AccessUuid string
	UserId     uint64
}

func InitRedis() *redisClient {
	// initializing redis
	c := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	_, err := c.Ping(ctx).Result()
	if err != nil {
		log.Fatal("Failed to connect to redis")
	}

	client.c = c
	return client
}

func CreateToken(user_id uint64) (*TokenDetails, error) {
	td := &TokenDetails{}

	td.AtExpires = time.Now().Add(time.Minute * 15).Unix() // set token expires after 15 minutes
	td.AccessUuid = uuid.New().String()

	td.RtExpires = time.Now().Add(time.Hour * 24 * 7).Unix() // set token expires after 7 days
	td.RefreshUuid = uuid.New().String()

	var err error
	// access token
	atClaims := jwt.MapClaims{}
	atClaims["authorized"] = true
	atClaims["jti"] = td.AccessUuid
	atClaims["identity"] = user_id
	atClaims["exp"] = td.AtExpires
	at := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)
	td.AccessToken, err = at.SignedString(access_secret)

	if err != nil {
		return nil, err
	}

	// refresh token
	rtClaims := jwt.MapClaims{}
	rtClaims["jti"] = td.RefreshUuid
	rtClaims["identity"] = user_id
	rtClaims["exp"] = td.RtExpires
	rt := jwt.NewWithClaims(jwt.SigningMethodHS256, rtClaims)
	td.RefreshToken, err = rt.SignedString(refresh_secret)
	if err != nil {
		return nil, err
	}

	return td, nil
}

func CreateAuth(user_id uint64, td *TokenDetails) error {
	at := time.Unix(td.AtExpires, 0)
	rt := time.Unix(td.RtExpires, 0)
	now := time.Now()

	errAccess := client.c.Set(ctx, td.AccessUuid, strconv.Itoa(int(user_id)), at.Sub(now)).Err()
	if errAccess != nil {
		return errAccess
	}

	errRefresh := client.c.Set(ctx, td.RefreshUuid, strconv.Itoa(int(user_id)), rt.Sub(now)).Err()
	if errRefresh != nil {
		return errRefresh
	}

	return nil
}

func ExtractToken(r *http.Request) string {
	bearToken := r.Header.Get("Authorization")
	strArr := strings.Split(bearToken, " ")
	if len(strArr) == 2 {
		return strArr[1]
	}
	return ""
}

func VerifyToken(r *http.Request) (*jwt.Token, error) {
	tokenString := ExtractToken(r)
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return access_secret, nil
	})

	if err != nil {
		return nil, err
	}
	return token, nil
}

func TokenValid(r *http.Request) error {
	token, err := VerifyToken(r)
	if err != nil {
		return err
	}

	if _, ok := token.Claims.(jwt.Claims); !ok && !token.Valid {
		return err
	}
	return nil
}

func ExtractTokenMetadata(r *http.Request) (*AccessDetails, error) {
	token, err := VerifyToken(r)
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if ok && token.Valid {
		accessUuid, ok := claims["jti"].(string)
		if !ok {
			return nil, err
		}

		userId, err := strconv.ParseUint(fmt.Sprintf("%.f", claims["identity"]), 10, 64)
		if err != nil {
			return nil, err
		}
		return &AccessDetails{
			AccessUuid: accessUuid,
			UserId:     userId,
		}, nil
	}
	return nil, err
}

func FetchAuth(authD *AccessDetails) (uint64, error) {
	userid, err := client.c.Get(ctx, authD.AccessUuid).Result()
	if err != nil {
		return 0, err
	}

	userID, _ := strconv.ParseUint(userid, 10, 64)
	return userID, nil
}

func DeleteAuth(jti string) (int64, error) {
	deleted, err := client.c.Del(ctx, jti).Result()
	if err != nil {
		return 0, err
	}
	return deleted, nil
}
