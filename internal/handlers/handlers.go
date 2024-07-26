package handlers

import (
	"cm_tg/internal/config"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/google/uuid"
)

type Message struct {
	ID         string `json:"id"`
	Service    string `json:"service"`
	CreatedAt  string `json:"created_at"`
	RawMessage string `json:"raw_message"`
	Source     Source `json:"source"`
}

type Source struct {
	Channel    string `json:"channel"`
	SourceURI  string `json:"source_uri"`
	SenderName string `json:"sender_name"`
	SenderURI  string `json:"sender_uri"`
}

func HandleMessage(cfg config.Config, message *tgbotapi.Message) {
	// Проверка ID отправителя
	isAllowed := false
	if cfg.AllowedUserIDs != nil && len(cfg.AllowedUserIDs) > 0 {
		for _, allowedID := range cfg.AllowedUserIDs {
			if message.From.ID == int(allowedID) {
				isAllowed = true
				break
			}
		}
	} else {
		isAllowed = true
	}

	if !isAllowed {
		return
	}

	// Генерация уникального идентификатора
	var id = uuid.New().String()

	// Форматирование даты
	createdAt := time.Unix(int64(message.Date), 0).Format(time.RFC3339)

	// Формирование ссылки на сообщение в чате
	var messageLink string
	if message.Chat.IsPrivate() {
		messageLink = fmt.Sprintf("https://t.me/c/%d/%d", message.Chat.ID, message.MessageID)
	} else {
		messageLink = fmt.Sprintf("https://t.me/%s/%d", message.Chat.UserName, message.MessageID)
	}

	// Формирование ссылки на пользователя
	var userLink string
	if message.From.UserName != "" {
		userLink = fmt.Sprintf("https://t.me/%s", message.From.UserName)
	} else {
		userLink = fmt.Sprintf("tg://user?id=%d", message.From.ID)
	}

	// Создание сообщения для SQS
	sqsMsg := Message{
		ID:         id,
		Service:    cfg.ServiceName,
		CreatedAt:  createdAt,
		RawMessage: message.Text,
		Source: Source{
			Channel:    "telegram",
			SourceURI:  messageLink,
			SenderName: fmt.Sprintf("%s %s", message.From.FirstName, message.From.LastName),
			SenderURI:  userLink,
		},
	}

	// Подготовка элемента для DynamoDB
	dynamoMsg := map[string]*dynamodb.AttributeValue{
		"id":         {S: aws.String(id)},
		"mp":         {S: aws.String(strings.ToLower(cfg.ServiceName) + "_ms:" + id)},
		"service":    {S: aws.String(sqsMsg.Service)},
		"created_at": {S: aws.String(sqsMsg.CreatedAt)},
		"source": {
			M: map[string]*dynamodb.AttributeValue{
				"channel":     {S: aws.String(sqsMsg.Source.Channel)},
				"source_uri":  {S: aws.String(sqsMsg.Source.SourceURI)},
				"sender_name": {S: aws.String(sqsMsg.Source.SenderName)},
				"sender_uri":  {S: aws.String(sqsMsg.Source.SenderURI)},
			},
		},
		"raw_message": {S: aws.String(sqsMsg.RawMessage)},
	}

	// Создание новой сессии AWS
	sess := session.Must(session.NewSession(&aws.Config{
		Region:      aws.String(cfg.AWSRegion),
		Credentials: credentials.NewStaticCredentials(cfg.AWSAccessKeyID, cfg.AWSSecretAccessKey, ""),
	}))
	svc := dynamodb.New(sess)

	// Вставка элемента в таблицу DynamoDB
	_, err := svc.PutItem(&dynamodb.PutItemInput{
		TableName: aws.String(cfg.DynamoDBTableName),
		Item:      dynamoMsg,
	})

	if err != nil {
		log.Fatalf("Ошибка при вставке элемента в DynamoDB: %s", err)
	}

	// Сериализация сообщения в JSON для отправки в SQS
	jsonMsg, err := json.Marshal(sqsMsg)
	if err != nil {
		log.Fatalf("Ошибка при сериализации сообщения: %s", err)
	}

	sqsSvc := sqs.New(sess)

	// Отправка сообщения в очередь SQS
	_, err = sqsSvc.SendMessage(&sqs.SendMessageInput{
		QueueUrl:    aws.String(cfg.SQSQueueURL),
		MessageBody: aws.String(string(jsonMsg)),
	})

	if err != nil {
		log.Fatalf("Ошибка при отправке сообщения в SQS: %s", err)
	}
}
