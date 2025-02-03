package app

import (
	"context"
	"log"
	"messanger/config"
	"messanger/controller/http"
	"messanger/data/cache/redis"
	"messanger/data/repository/mysql"
	sms "messanger/data/sms/cmd_sms"
	"messanger/domain/service/auth"
	"messanger/domain/service/chats"
	"messanger/domain/service/groups"
	"messanger/domain/service/messages"
	"messanger/domain/service/phone"
	"messanger/domain/service/users"
	"messanger/pkg/db"
	"messanger/pkg/http_server"
	"messanger/pkg/redis"
	"os"
)

func Run(cfgPath string) {
	log.SetFlags(log.Lshortfile | log.Ldate | log.Ltime)
	errorsLogger := log.New(os.Stderr, "", log.Ldate|log.Ltime)

	cfg, err := config.GetConfig(cfgPath)
	if err != nil {
		log.Fatal("get config error: ", err)
	}

	DB, err := db.Connect(cfg.MySQL)
	if err != nil {
		log.Fatal("database connect error: ", err)
	}
	defer DB.Close()
	TxDB := db.NewDBWithTx(DB)

	r, err := redis.Connect(cfg.Redis)
	if err != nil {
		log.Fatal("r connect error: ", err)
	}
	defer r.Close()

	smsSender := sms.NewCmdSmsAdapter()

	chatsRepo := mysql.NewChats(TxDB)
	groupsRepo := mysql.NewGroups(TxDB)
	contactsRepo := mysql.NewContacts(TxDB)
	userRepo := mysql.NewUsers(TxDB)
	messagesRepo := mysql.NewMessages(TxDB)
	c := cache.NewCache(r)

	phoneConf := phone.NewPhoneService(smsSender, c)

	authService := auth.NewAuthService(c, userRepo, phoneConf, cfg.AuthService)
	userService := users.NewUsersService(userRepo, contactsRepo, chatsRepo, phoneConf)
	chatService := chats.NewChatService(chatsRepo, groupsRepo)
	groupService := groups.NewGroupService(chatsRepo, groupsRepo)
	messagesService := messages.NewMessagesService(messagesRepo, chatsRepo)

	h := http.NewHandler(
		authService,
		userService,
		messagesService,
		chatService,
		groupService,
		errorsLogger,
	)

	h.InitRouter()

	server := http_server.NewHttpServer(h, cfg.HttpServer)
	defer func() {
		if err := server.Shutdown(context.Background()); err != nil {
			log.Println("shutdown server error: ", err)
		}
	}()

	log.Printf("http://%s/", server.Addr)
	if err := server.ListenAndServe(); err != nil {
		log.Fatal("http server error: ", err)
	}
}
