package main

import (
	"dripapp/configs"
	"dripapp/internal/dripapp/middleware"
	"dripapp/internal/pkg/session"
	_sessionUcase "dripapp/internal/pkg/session/usecase"
	_userDelivery "dripapp/internal/pkg/user/delivery"
	_userRepo "dripapp/internal/pkg/user/repository"
	_userUsecase "dripapp/internal/pkg/user/usecase"
	"log"
	"net/http"
	"os"

	_ "dripapp/docs"

	"github.com/gorilla/mux"
)

const StatusEmailAlreadyExists = 1001

// @title Drip API
// @version 1.0
// @description API for Drip.
// @termsOfService http://swagger.io/terms/

// @host api.ijia.me
// @BasePath /api/v1

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Set-Cookie
func main() {
	// logfile
	logFile, err := os.OpenFile("logs.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer func(logFile *os.File) {
		err := logFile.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(logFile)

	configs.SetConfig()

	// router
	router := mux.NewRouter()

	// repository
	userRepo, err := _userRepo.NewPostgresUserRepository(configs.Postgres)
	if err != nil {
		log.Fatal(err)
	}

	// userRepo := _userRepo.NewMockDB()
	// userRepo.Init()
	// sm := session.NewSessionDB()

	sm, err := session.NewTarantoolConnection(configs.Tarantool)
	if err != nil {
		log.Fatal(err)
	}

	timeoutContext := configs.Timeouts.ContextTimeout

	// usecase
	userUCase := _userUsecase.NewUserUsecase(userRepo, sm, timeoutContext)
	sessionUcase := _sessionUcase.NewSessionUsecase(sm, timeoutContext)

	// delivery
	_userDelivery.SetRouting(router, userUCase, sessionUcase)

	// middleware
	middleware.NewMiddleware(router, sm, logFile)

	srv := &http.Server{
		Handler:      router,
		Addr:         configs.Server.Port,
		WriteTimeout: http.DefaultClient.Timeout,
		ReadTimeout:  http.DefaultClient.Timeout,
	}

	staticHandler := http.StripPrefix(
		"/media/",
		http.FileServer(http.Dir("./media")),
	)
	http.Handle("/media/", staticHandler)

	log.Println("starting server at :9999")

	go func() {
		err := http.ListenAndServe(":9999", nil)
		if err != nil {
			log.Println("media server died:\n", err)
		}
	}()

	log.Printf("STD starting server at %s\n", srv.Addr)

	log.Fatal(srv.ListenAndServe())
	// log.Fatal(srv.ListenAndServeTLS(certFile, keyFile))
}
