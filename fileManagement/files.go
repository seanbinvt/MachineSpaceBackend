package fileManagement

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"fmt"
	"log"
	"path/filepath"
	//"bytes"
	//"io"
	//"strings"

	//"golang.org/x/crypto/bcrypt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type FileInfo struct {
	Name   string      `json:"name" bson:"Name"`
    Description    string      `json:"description" bson:"Description"`
}

func FileUpload(w http.ResponseWriter, r *http.Request, db *mongo.Database) {
	fmt.Println("FILE UPLOAD")

	r.ParseMultipartForm(32 << 20) // limit your max input length!
    file, handler, err := r.FormFile("file") // Retrieve the file from form data
	name := r.FormValue("fileName")
	description := r.FormValue("fileDesc")
	user := r.FormValue("user")
	ext := filepath.Ext(handler.Filename)

	fmt.Println("Name:", name, "Desc:", description, "User:", user)


    if err != nil {
        fmt.Println(err)
    }
	defer file.Close() // Close the file when we finish
	
	fmt.Printf("Uploaded File: %+v\n", handler.Filename)
    fmt.Printf("File Size: %+v\n", handler.Size)
	fmt.Printf("MIME Header: %+v\n", handler.Header)
	
	collection := db.Collection("users")

	opts := options.Update().SetUpsert(true)
	match := bson.M{"Username": user}
	change := bson.M{"$push": bson.M{"Files": bson.M{"Name": name+ext, "Description": description}}}


	result, err := collection.UpdateOne(context.TODO(), match, change, opts)
	
	if err != nil {
		log.Fatal(err)
	}

	var res map[string]interface{}
	jsonRes := ``

	// Do file updating on success
	if (result.UpsertedCount != 0 || result.MatchedCount != 0) {
		if _, err := os.Stat("files/"+user); os.IsNotExist(err) {
			os.Mkdir("files/"+user, 0777)
		}
		_ = os.Mkdir("files/"+user, 0777)

		//tempFile, err := ioutil.TempFile("files/"+user, name)
		pwd, _ := os.Getwd()
		tempFile, err := os.Create(pwd+"/files/"+user+"/"+name+ext)
		if err != nil {
			fmt.Println(err)
			jsonRes = `{"errorCode": 1}`
			return
		}
		defer tempFile.Close()

		fileBytes, err := ioutil.ReadAll(file)
		if err != nil {
			fmt.Println(err)
			jsonRes = `{"errorCode": 1}`
		}
		tempFile.Write(fileBytes)
		jsonRes = `{"errorCode": 0, "name": "`+name+ext+`", "description": "`+description+`"}`
	}
	json.Unmarshal([]byte(jsonRes), &res)
	marRes, _ := json.Marshal(res)
	w.Write(marRes)
}