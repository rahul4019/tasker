package cron

import (
	"fmt"

	"github.com/hibiken/asynq"
	"github.com/rahul4019/tasker/internal/config"
	"github.com/rahul4019/tasker/internal/database"
	"github.com/rahul4019/tasker/internal/logger"
	"github.com/rahul4019/tasker/internal/repository"
	"github.com/rahul4019/tasker/internal/server"
	"github.com/redis/go-redis/v9"
)

type JobContext struct {
	Config        *config.Config
	Server        *server.Server
	JobClient     *asynq.Client
	Repositories  *repository.Repositories
	LoggerService *logger.LoggerService
}

func NewJobContext() (*JobContext, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	loggerService := logger.NewLoggerService(cfg.Observability)
	loggerInstance := logger.NewLoggerWithService(cfg.Observability, loggerService)

	db, err := database.New(cfg, &loggerInstance, loggerService)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr: cfg.Redis.Address,
	})

	srv := &server.Server{
		Config:        cfg,
		Logger:        &loggerInstance,
		LoggerService: loggerService,
		DB:            db,
		Redis:         redisClient,
	}

	jobClient, err := initJobClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize job client: %w", err)
	}

	repositories := repository.NewRepositories(srv)

	return &JobContext{
		Config:        cfg,
		Server:        srv,
		JobClient:     jobClient,
		Repositories:  repositories,
		LoggerService: loggerService,
	}, nil

}

func (c *JobContext) Close() {
	if c.Server != nil && c.Server.DB != nil {
		c.Server.DB.Pool.Close()
	}
	if c.Server != nil && c.Server.Redis != nil {
		c.Server.Redis.Close()
	}
	if c.JobClient != nil {
		c.JobClient.Close()
	}
	if c.LoggerService != nil {
		c.LoggerService.Shutdown()
	}
}

func initJobClient(cfg *config.Config) (*asynq.Client, error) {
	redisOpt := asynq.RedisClientOpt{
		Addr: cfg.Redis.Address,
	}
	client := asynq.NewClient(redisOpt)
	return client, nil
}
