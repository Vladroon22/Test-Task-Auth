package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/Vladroon22/Test-Task-BackDev/config"
	"github.com/Vladroon22/Test-Task-BackDev/internal/database"
	"github.com/Vladroon22/Test-Task-BackDev/internal/handlers"
	"github.com/Vladroon22/Test-Task-BackDev/internal/service"
	"github.com/Vladroon22/Test-Task-BackDev/internal/sessions"

	"github.com/gorilla/mux"
)

func main() {

	cnf := config.CreateConfig()

	db := database.NewDB(cnf)
	if err := db.Connect(); err != nil {
		log.Fatalln(err)
		return
	}
	log.Println("Database connected!")

	repo := database.NewRepo(db)
	sess := sessions.NewSessions(repo)
	srv := service.NewService(repo)
	h := handlers.NewHandler(repo, srv, sess, cnf)

	router := mux.NewRouter()
	router.HandleFunc("/getTokenPair/{id:[0-9]+}", h.GetPair).Methods("GET")
	router.HandleFunc("/makeRefresh/{id:[0-9]+}", h.MakeRefresh).Methods("POST")
	log.Println("Router established")

	log.Println("Server is listening --> localhost" + cnf.Addr_PORT)
	go http.ListenAndServe(cnf.Addr_PORT, router)

	exitSig := make(chan os.Signal, 1)
	signal.Notify(exitSig, syscall.SIGINT, syscall.SIGTERM)

	<-exitSig

	go func() {
		if err := db.CloseDB(); err != nil {
			log.Panicln(err)
			return
		}
	}()

	log.Println("Gracefull shutdown")
}
