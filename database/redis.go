package database

import (
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"os"
	"time"
)

var (
	access_secret  = []byte(os.Getenv("ACCESS_SECRET"))
	refresh_secret = []byte(os.Getenv("REFRESH_SECRET"))
)

var (
	client = &redisClient{}
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

func InitRedis() *redisClient {
	// initializing redis
	c := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	_, err := c.Ping(c.Context()).Result()
	if err != nil {
		panic(err)
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
	atClaims["access_uuid"] = td.AccessUuid
	atClaims["user_id"] = user_id
	atClaims["exp"] = td.AtExpires
	at := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)
	td.AccessToken, err = at.SignedString(access_secret)

	if err != nil {
		return nil, err
	}

	// refresh token
	rtClaims := jwt.MapClaims{}
	rtClaims["refresh_uuid"] = td.RefreshUuid
	rtClaims["user_id"] = user_id
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

	_, _, _ = now, at, rt

	return nil
}
