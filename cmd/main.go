package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/BurntSushi/toml"
	"github.com/Vladroon22/Test-Task-BackDev/config"
	"github.com/Vladroon22/Test-Task-BackDev/internal/database"
	"github.com/Vladroon22/Test-Task-BackDev/internal/handlers"
	"github.com/Vladroon22/Test-Task-BackDev/internal/service"

	"github.com/gorilla/mux"
)

var (
	Toml string
)

func main() {
	flag.Parse()

	flag.StringVar(&Toml, "path-to-toml", "./config/conf.toml", "path-to-toml")

	cnf := config.CreateConfig()
	_, err := toml.DecodeFile(Toml, cnf)
	if err != nil {
		log.Panicln(err)
		return
	}

	db := database.NewDB(cnf)
	if err := db.Connect(); err != nil {
		log.Fatalln(err)
		return
	}
	log.Println("Database connected!")

	repo := database.NewRepo(db)        // sql
	srv := service.NewService(repo)     // sql - interface
	h := handlers.NewHandler(repo, srv) // ручки

	router := mux.NewRouter()
	router.HandleFunc("/getTokenPair/{id:[0-9]+}", h.GetPair).Methods("GET")
	router.HandleFunc("/makeRefresh/{id:[0-9]+}", h.MakeRefresh).Methods("GET")
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
