package app

import (
	"context"
	"log"
	cache "messanger/adapters/cache/redis"
	"messanger/adapters/repository/mysql"
	"messanger/app/db_conn"
	"messanger/app/http_server"
	"messanger/app/redis_conn"
	"messanger/config"
	"messanger/controller/httpAPI"
	"messanger/core/service"
	"os"
)

func Run(cfgPath string) {
	log.SetFlags(log.Lshortfile | log.Ldate | log.Ltime)
	errorsLogger := log.New(os.Stderr, "", log.Ldate|log.Ltime)

	cfg, err := config.GetConfig(cfgPath)
	if err != nil {
		log.Fatal("get config error: ", err)
	}

	db, err := db_conn.Connect(cfg.MySQL)
	if err != nil {
		log.Fatal("database connect error: ", err)
	}
	defer db.Close()

	redis, err := redis_conn.Connect(cfg.Redis)
	if err != nil {
		log.Fatal("redis connect error: ", err)
	}

	userRepo := mysql.NewUsers(db)
	chatsRepo := mysql.NewChats(db)
	messagesRepo := mysql.NewMessages(db)
	c := cache.NewCache(redis)

	emailService := service.NewEmailService(cfg.Email)
	authService := service.NewAuthService(c, userRepo, cfg.AuthService)
	userService := service.NewUsersService(userRepo, emailService)
	chatService := service.NewChatService(chatsRepo)
	messagesService := service.NewMessagesService(messagesRepo, userRepo)

	go func() {
		for {
			select {
			case chatId := <-chatService.OnDeleteChat:
				if err := messagesService.OnDeleteChat(chatId); err != nil {
					errorsLogger.Println(err)
				}

			case chatId := <-userService.DeleteChat:
				ctx := context.WithValue(context.Background(), "IsSystemCall", struct{}{})
				if err := chatService.Delete(ctx, chatId); err != nil {
					errorsLogger.Println(err)
				}
			}
		}
	}()

	h := httpAPI.NewHandler(
		authService,
		userService,
		messagesService,
		chatService,
		emailService,
		errorsLogger,
	)

	h.InitRouter()

	server := http_server.NewHttpServer(h, cfg.HttpServer)
	log.Printf("http://%s/", server.Addr)
	if err := server.ListenAndServe(); err != nil {
		log.Fatal("http server error: ", err)
	}
}
