package app

import (
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"cosy-console/config"
	"cosy-console/internal/app/adapters"
	itemrepo "cosy-console/internal/domain/catalog/repositories/item"
	itemusecase "cosy-console/internal/domain/catalog/usecases/item"
	orderrepo "cosy-console/internal/domain/orders/repositories/order"
	orderusecase "cosy-console/internal/domain/orders/usecases/order"
	"cosy-console/internal/infrastructure/cache"
	dbinfra "cosy-console/internal/infrastructure/db"
	"cosy-console/pkg/transaction"
)

type Container struct {
	Shared  Shared
	Orders  OrdersDomain
	Catalog CatalogDomain
}

type Shared struct {
	DB    *gorm.DB
	Tx    transaction.Manager
	Redis *redis.Client
}

type OrdersDomain struct {
	Repos    OrdersRepos
	Usecases OrdersUsecases
}

type OrdersRepos struct {
	OrderRepo *orderrepo.OrderRepo
}

type OrdersUsecases struct {
	Order *orderusecase.OrderUseCase
}

type CatalogDomain struct {
	Repos    CatalogRepos
	Usecases CatalogUsecases
}

type CatalogRepos struct {
	ItemRepo *itemrepo.ItemRepo
}

type CatalogUsecases struct {
	Item *itemusecase.ItemUseCase
}

func NewContainer(cfg *config.Config, dbOpts ...dbinfra.Option) *Container {
	db := dbinfra.Connect(cfg.DatabaseURL, dbOpts...)
	tx := dbinfra.NewTxManager(db)
	redisClient := cache.NewRedisClient(cfg.RedisURL)

	orderRepo := orderrepo.New(db)
	itemRepo := itemrepo.New(db)

	pricer := adapters.NewCatalogPricer(itemRepo)
	itemsCache := adapters.NewRedisItemsCache(redisClient)

	return &Container{
		Shared: Shared{
			DB:    db,
			Tx:    tx,
			Redis: redisClient,
		},
		Orders: OrdersDomain{
			Repos: OrdersRepos{
				OrderRepo: orderRepo,
			},
			Usecases: OrdersUsecases{
				Order: orderusecase.New(orderRepo, pricer, tx),
			},
		},
		Catalog: CatalogDomain{
			Repos: CatalogRepos{
				ItemRepo: itemRepo,
			},
			Usecases: CatalogUsecases{
				Item: itemusecase.New(itemRepo, itemsCache),
			},
		},
	}
}
