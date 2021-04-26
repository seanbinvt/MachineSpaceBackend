package userAuth

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	//"encoding/base64"
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"
	"strings"
	"os"

	//"time"
	//"io"

	"golang.org/x/crypto/bcrypt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type LoginCred struct {
	Username string `json:"username", db:"username"`
	Password string `json:"password", db:"password"`
}

type LoginCredInfo struct {
	Username  string   `json:"username", db:"username"`
	Password  string   `json:"password", db:"password"`
}

/*
0 - No error (Success)
1 - Username already in used (Fail)
2 - Other error (Fail)
*/
type RegisterResponse struct {
	ErrorCode int `json:"errorCode"`
}

/*
func main() {
	for {
		// Enter a password and generate a salted hash
		pwd := getPwd()
		hash := hashAndSalt(pwd)

		// Enter the same password again and compare it with the
		// first password entered
		pwd2 := getPwd()
		pwdMatch := comparePasswords(hash, pwd2)
		fmt.Println("Passwords Match?", pwdMatch)
	}
}

func getPwd(string pwd) []byte { // Prompt the user to enter a password
	fmt.Println("Enter a password") // We will use this to store the users input
	//var pwd string                  // Read the users input
	//_, err := fmt.Scan(&pwd)
	//if err != nil {
		//log.Println(err)
	//} // Return the users input as a byte slice which will save us
	// from having to do this conversion later on
	return []byte(pwd)
}
*/
func CompareToken(token string, username string, collection *mongo.Collection) bool {
	err := collection.FindOne(context.TODO(), bson.M{"username": username, "authToken": token})
	if err != nil {
		return false
	}
	return true
}

func Signup(w http.ResponseWriter, r *http.Request, db *mongo.Database) {
	args, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		log.Fatal(err)
	}

	var logs LoginCred
	err = json.Unmarshal(args, &logs)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	logs.Password = HashAndSalt([]byte(logs.Password))

	var result map[string]interface{}
	jsonRes := ``

	if logs.Username != "" && logs.Password != "" {

		collection := db.Collection("users")
		log.Println(logs)

		//dbInfo, _ := json.Marshal(logs)
		//var emptyArray []string
		id, err := collection.InsertOne(context.TODO(), bson.D{{"Username", logs.Username}, {"Password", logs.Password}})
		if err != nil {
			// If error then the username is already taken
			jsonRes = `{"ErrorCode": 1}`
		} else {
			// Success
			log.Println("New user created at ID ", id)
			jsonRes = `{"ErrorCode": 0, "Username": "` + logs.Username + `"}`
		}
	} else {
		jsonRes = `{"ErrorCode": 1}`
	}
	json.Unmarshal([]byte(jsonRes), &result)
	marRes, _ := json.Marshal(result)
	w.Write(marRes)
}

func Login(w http.ResponseWriter, r *http.Request, db *mongo.Database) {
	for _, cookie := range r.Cookies() {
		log.Println("Found a cookie named:", cookie.Name)
	}
	args, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		log.Fatal(err)
	}

	var logs LoginCred        // Logs from the user
	var logsOut LoginCredInfo // Logs from the DB for given Username

	err = json.Unmarshal(args, &logs)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	var result map[string]interface{}
	jsonRes := ``

	if logs.Username != "" && logs.Password != "" {

		collection := db.Collection("users")
		log.Println(logs)

		//dbInfo, _ := json.Marshal(logs)
		errD := collection.FindOne(context.TODO(), bson.D{{"Username", logs.Username}}).Decode(&logsOut)

		log.Println(logsOut)
		if errD != nil {
			// If error then the username doesn't exist
			jsonRes = `{"errorCode": 1}`
		} else {
			// Success
			compare := ComparePasswords(logsOut.Password, []byte(logs.Password))

			if compare {
				// Login success
				jsonRes = `{"errorCode": 0, "username": "` + logsOut.Username + `"}`

				sessionTokenRaw, err := exec.Command("uuidgen").Output()

				log.Println("Session:", string(sessionTokenRaw))

				sessionToken := strings.TrimSuffix(string(sessionTokenRaw), "\n")

				tokenAge := 86400 //One day
				domain := "." + os.Getenv("FRONTEND")
				if err == nil {
					sessionCookie := &http.Cookie{
						Name:     "session",
						Value:    sessionToken,
						Domain:   domain, // "/" for frontend on localhost
						Path:     "/",
						HttpOnly: false,
						MaxAge:   tokenAge, //Expires after 1 day
						SameSite: http.SameSiteLaxMode,
					}

					usernameCookie := &http.Cookie{
						Name:     "username",
						Value:    logsOut.Username,
						Domain:   domain,
						Path:     "/", // "/" for frontend on localhost
						HttpOnly: false,
						MaxAge:   tokenAge, //Expires after 1 day
						SameSite: http.SameSiteLaxMode,
					}

					fmt.Println(time.Now().Add(time.Second * time.Duration(tokenAge)))
					update := bson.M{"$set": bson.M{"AuthToken": sessionToken, "Port": 0, "Expiration": time.Now().Add(time.Second * time.Duration(tokenAge))}}
					filter := bson.M{"Username": bson.M{"$eq": logsOut.Username}}

					_, err := collection.UpdateOne(
						context.TODO(),
						filter,
						update,
					)

					if err != nil {
						log.Fatal(err)
					} else {
						log.Printf("Updated user %s to authToken %s!\n", logsOut.Username, string(sessionToken))
					}

					http.SetCookie(w, sessionCookie)
					http.SetCookie(w, usernameCookie)
				}
			} else {
				// User exists, but password is wrong
				//log.Println("User exists but password is wrong")
				jsonRes = `{"errorCode": 1}`
			}
		}
	} else {
		jsonRes = `{"errorCode": 1}`
	}
	json.Unmarshal([]byte(jsonRes), &result)
	marRes, err := json.Marshal(result)
	//log.Println(jsonRes)

	w.Write(marRes)
}

func HashAndSalt(pwd []byte) string {
	// Use GenerateFromPassword to hash & salt pwd
	// MinCost is just an integer constant provided by the bcrypt
	// package along with DefaultCost & MaxCost.
	// The cost can be any value you want provided it isn't lower
	// than the MinCost (4)
	hash, err := bcrypt.GenerateFromPassword(pwd, bcrypt.MinCost)
	if err != nil {
		log.Println(err)
	} // GenerateFromPassword returns a byte slice so we need to
	// convert the bytes to a string and return it
	log.Println("Hash ", string(hash))
	return string(hash)
}
func ComparePasswords(hashedPwd string, plainPwd []byte) bool { // Since we'll be getting the hashed password from the DB it
	// will be a string so we'll need to convert it to a byte slice
	byteHash := []byte(hashedPwd)
	//log.Println("Comparing ", hashedPwd, " to ", string(plainPwd))
	err := bcrypt.CompareHashAndPassword(byteHash, plainPwd)
	if err != nil {
		log.Println(err)
		return false
	}
	return true
}
