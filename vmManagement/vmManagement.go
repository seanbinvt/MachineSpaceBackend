package fileManagement

import (
	"exec"
	"fmt"
	"log"
	"net/http"

	//"bytes"
	//"io"
	//"strings"

	//"golang.org/x/crypto/bcrypt"

	"go.mongodb.org/mongo-driver/mongo"
)

type FileInfo struct {
	Name        string `json:"name" bson:"Name"`
	Description string `json:"description" bson:"Description"`
}

func CreateSnapshot(w http.ResponseWriter, r *http.Request, db *mongo.Database) {

}

func DeleteSnapshot(w http.ResponseWriter, r *http.Request, db *mongo.Database) {
	out, err := exec.Command("ls").Output()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s", out)
}
