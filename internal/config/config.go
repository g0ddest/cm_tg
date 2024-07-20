package config

import (
	"log"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	TelegramToken      string
	DynamoDBTableName  string
	SQSQueueURL        string
	AWSRegion          string
	AWSAccessKeyID     string
	AWSSecretAccessKey string
	AllowedUserIDs     []int64
}

func LoadConfig() Config {
	allowedUserIDsStr := os.Getenv("ALLOWED_USER_IDS")
	allowedUserIDs := []int64{}

	for _, idStr := range strings.Split(allowedUserIDsStr, ",") {
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err == nil {
			allowedUserIDs = append(allowedUserIDs, id)
		}
	}

	config := Config{
		TelegramToken:      os.Getenv("TELEGRAM_TOKEN"),
		DynamoDBTableName:  os.Getenv("DYNAMODB_TABLE_NAME"),
		SQSQueueURL:        os.Getenv("SQS_QUEUE_URL"),
		AWSRegion:          os.Getenv("AWS_REGION"),
		AWSAccessKeyID:     os.Getenv("AWS_ACCESS_KEY_ID"),
		AWSSecretAccessKey: os.Getenv("AWS_SECRET_ACCESS_KEY"),
		AllowedUserIDs:     allowedUserIDs,
	}

	if config.TelegramToken == "" ||
		config.DynamoDBTableName == "" ||
		config.SQSQueueURL == "" ||
		config.AWSRegion == "" ||
		config.AWSAccessKeyID == "" ||
		config.AWSSecretAccessKey == "" {
		log.Fatalf("One or more environment variables are missing")
	}

	return config
}
