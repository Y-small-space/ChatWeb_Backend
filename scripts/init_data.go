package main

import (
	"chatweb/config"
	"chatweb/internal/model"
	"chatweb/internal/repository/mongodb"
	"context"
	"fmt"
	"log"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	// 加载配置
	cfg := config.LoadConfig()

	// 连接数据库
	mongodb.InitMongoDB(cfg.MongoDB.URI, cfg.MongoDB.Database)

	// 创建用户数据
	users := []model.User{
		{
			Username:  "user002",
			Email:     "user002@example.com",
			Phone:     "18700000002",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			Username:  "user003",
			Email:     "user003@example.com",
			Phone:     "18700000003",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			Username:  "user004",
			Email:     "user004@example.com",
			Phone:     "18700000004",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			Username:  "user005",
			Email:     "user005@example.com",
			Phone:     "18700000005",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			Username:  "user006",
			Email:     "user006@example.com",
			Phone:     "18700000006",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			Username:  "user007",
			Email:     "user007@example.com",
			Phone:     "18700000007",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			Username:  "user008",
			Email:     "user008@example.com",
			Phone:     "18700000008",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			Username:  "user009",
			Email:     "user009@example.com",
			Phone:     "18700000009",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			Username:  "user010",
			Email:     "user010@example.com",
			Phone:     "18700000010",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			Username:  "user011",
			Email:     "user011@example.com",
			Phone:     "18700000011",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	// 为所有用户设置相同的密码
	password := "password123"
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal(err)
	}

	// 插入用户数据
	collection := mongodb.GetUserCollection()
	for _, user := range users {
		user.Password = string(hashedPassword)
		result, err := collection.InsertOne(context.Background(), user)
		if err != nil {
			log.Printf("Failed to insert user %s: %v", user.Username, err)
			continue
		}
		fmt.Printf("Inserted user %s with ID: %v\n", user.Username, result.InsertedID)
	}

	fmt.Println("Data initialization completed!")
}
