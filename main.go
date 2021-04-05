package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	userAuth "machineSpaceAPI/userAuth"
	fileManagement "machineSpaceAPI/fileManagement"

	// encryption/decryption
	"github.com/gorilla/mux" // http router used

	// for .env variables compatability
	"github.com/joho/godotenv"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func signup(w http.ResponseWriter, r *http.Request) {
	scheme := ""
	if (*r).Header["Referer"] != nil {
		scheme = (*r).Header["Referer"][0][0:5]
	}
	allowOpts(&w, scheme)
	if (*r).Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	} else {
		userAuth.Signup(w, r, db)
	}
}

func login(w http.ResponseWriter, r *http.Request) {
	scheme := ""
	if (*r).Header["Referer"] != nil {
		scheme = (*r).Header["Referer"][0][0:5]
	}
	allowOpts(&w, scheme)
	if (*r).Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	} else {
		userAuth.Login(w, r, db)
	}
}

func upload(w http.ResponseWriter, r *http.Request) {
	scheme := ""
	if (*r).Header["Referer"] != nil {
		scheme = (*r).Header["Referer"][0][0:5]
	}
	allowOpts(&w, scheme)
	if (*r).Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	} else {
		fileManagement.FileUpload(w, r, db)
	}
}

func testAuth(w http.ResponseWriter, r *http.Request) {
	scheme := ""
	fmt.Println("here")
	if (*r).Header["Referer"] != nil {
		scheme = (*r).Header["Referer"][0][0:5]
	}
	allowOpts(&w, scheme)
	if (*r).Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	} else {
		//userAuth.CompareToken(w, r, db)
	}
}

func main() {
	environment := "dev"

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if environment == "dev" {
		godotenv.Load(".env")
	}

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(os.Getenv("ATLAS_URI")))

	if err != nil {
		log.Fatal(err)
	}
	db = client.Database("machine_space")

	handler()
}

var db *mongo.Database

var environment string = "dev"

func handler() {
	port := os.Getenv("PORT")
	log.Println("Server running on port: ", port)

	r := mux.NewRouter()
	//r.HandleFunc("/battlereport/{server}/{reportID}", viewBattleReport).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/login", login).Methods("POST", "OPTIONS")
	r.HandleFunc("/api/register", signup).Methods("POST", "OPTIONS")
	
	r.HandleFunc("/api/upload", upload).Methods("POST", "OPTIONS")
	log.Fatal(http.ListenAndServe(":"+port, r)) // If error then log to console
	fmt.Println("Running on port", port)
}

func allowOpts(w *http.ResponseWriter, ref string) {
	if ref == "" {
		return
	} else if ref[4] != 's' {
		//log.Println("http://" + os.Getenv("FRONTEND"))
		(*w).Header().Set("Access-Control-Allow-Origin", "http://"+os.Getenv("FRONTEND"))
	} else {
		(*w).Header().Set("Access-Control-Allow-Origin", "https://"+os.Getenv("FRONTEND"))
	}

	(*w).Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
    (*w).Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
    (*w).Header().Set("Access-Control-Allow-Credentials", "true")
}
