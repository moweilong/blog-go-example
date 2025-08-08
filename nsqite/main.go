package main

import (
	"encoding/json"
	"log"
	"log/slog"
	"time"

	"github.com/ixugo/nsqite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// 定义消息处理器
type UserActionHandler struct{}

func (h *UserActionHandler) HandleMessage(message *nsqite.Message) error {
	var action struct {
		UserID    string `json:"user_id"`
		Action    string `json:"action"`
		ContentID string `json:"content_id"`
		Timestamp string `json:"timestamp"`
	}
	if err := json.Unmarshal(message.Body, &action); err != nil {
		return err
	}
	// 数据清洗和统计分析
	return h.analyzeUserAction(action)
}

func (h *UserActionHandler) analyzeUserAction(action struct {
	UserID    string `json:"user_id"`
	Action    string `json:"action"`
	ContentID string `json:"content_id"`
	Timestamp string `json:"timestamp"`
}) error {
	// 数据清洗
	// 统计分析
	slog.Info("analyzeUserAction", "action", action)
	// 存储分析结果
	return nil
}

func main() {
	db, err := gorm.Open(sqlite.Open("user_actions.db"), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}
	// 设置 GORM 数据库连接
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal(err)
	}
	nsqite.SetDB(nsqite.DriverNameSQLite, sqlDB)

	const topic = "user_actions"
	// 创建生产者
	p := nsqite.NewProducer()
	// 创建消费者，设置最大重试次数为 5
	c := nsqite.NewConsumer(topic, "consumer1", nsqite.WithMaxAttempts(5))
	// 添加 5 个并发处理器
	c.AddConcurrentHandlers(&UserActionHandler{}, 5)

	// 在事务中发布消息
	db.Transaction(func(tx *gorm.DB) error {
		// 业务操作
		if err := doSomeBusiness(tx); err != nil {
			return err
		}
		// 发布消息
		action := map[string]interface{}{
			"user_id":    "123",
			"action":     "view",
			"content_id": "456",
			"timestamp":  time.Now().Format(time.RFC3339),
		}
		body, _ := json.Marshal(action)
		return p.PublishTx(func(v *nsqite.Message) error {
			return db.Create(v).Error
		}, topic, body)
	},
	)

}

func doSomeBusiness(tx *gorm.DB) error {
	// 业务操作
	slog.Info("doSomeBusiness")
	return nil
}
