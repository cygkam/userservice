package service

import (
	"context"
	"encoding/json"
	"fmt"

	"user-service/internal/cache"
	"user-service/internal/config"
	"user-service/internal/db"

	"github.com/sirupsen/logrus"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type UserService struct {
	DBConnection *pgx.Conn
	Cache        *cache.Cache
}

func NewUserService(db *pgx.Conn) *UserService {
	us := &UserService{DBConnection: db}

	cacheCfg := &cache.CacheConfig{
		DistributionEnabled: config.GetBool("cache.distribution.enabled"),
		Port:                config.GetString("cache.peer.port"),
		PodIP:               config.GetString("cache.pod.ip"),
		Namespace:           config.GetString("cache.pod.namespace"),
		Selector:            config.GetString("cache.pod.selector"),
	}

	c, err := cache.New(cacheCfg, us)
	if err != nil {
		logrus.Fatalf("Cache initialization failed: %v", err)
	}
	us.Cache = c

	return us
}

func (us *UserService) GetUser(c context.Context, uuid string) (db.User, error) {
	var user db.User

	if bytes, hit := us.Cache.CachePool.Get(c, uuid); hit {
		err := json.Unmarshal(bytes, &user)
		if err != nil {
			return user, err
		}
		return user, nil
	}

	return user, fmt.Errorf("unexpected error")
}

func (us *UserService) Fetch(c context.Context, id string) ([]byte, error) {

	uuid, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}

	q := db.New(us.DBConnection)
	user, err := q.GetUser(c, uuid)
	if err != nil {
		return nil, err
	}

	bytes, err := json.Marshal(user)
	if err != nil {
		return nil, err
	}

	return bytes, nil
}
