package fileManagement

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	//"bytes"
	//"io"
	//"strings"
	//"golang.org/x/crypto/bcrypt"
	//"go.mongodb.org/mongo-driver/mongo"
)

const maxPort = 6000
const minPort = 5959

var prevPort = minPort

type Username struct {
	Username string `json:"username" bson:"Username"`
}

type UserAndSnapshot struct {
	Username     string `json:"username" bson:"Username"`
	SnapshotName string `json:"snapshotName" bson:"SnapshotName"`
}

type Document struct {
	Username   string    `json:"username" bson:"Username"`
	Password   string    `json:"password" bson:"Password"`
	AuthToken  string    `json:"authToken" bson:"AuthToken"`
	Snapshots  []string  `json:"snapshots" bson:"snapshots"`
	Port       int       `json:"port" bson:"Port"`
	Expiration time.Time `json:"expiration" bson:"Expiration"` // Date that the port and authToken expires
}

/*
Incrementally go through possible ports 5959-6000 and assign an open one to user.
*/
func getPort(username string, collection *mongo.Collection) int {
	var out Document

	// Check if user is already assigned a port
	if err := collection.FindOne(context.TODO(), bson.M{"username": username}).Decode(&out); err != nil {
		log.Fatal(err)
	}

	if out.Port > 0 {
		// Username already assigned port
		return out.Port
	}

	// All user documents that have a currently assigned port and haven't already expired
	var docs []Document
	var cursor *mongo.Cursor
	cursor, _ = collection.Find(context.TODO(), bson.M{"Port": bson.M{"$ne": 0}, "Expiration": bson.M{"$lte": time.Now()}})
	cursor.All(context.TODO(), &docs)

	// Iterate though possible ports and assign
	portNumber := prevPort
	for n := 0; n < maxPort-minPort; n++ {
		found := false
		// Iterate through assigned user ports.
		for i := 0; i < len(docs); i++ {
			if docs[i].Port == portNumber {
				found = true
				break
			}
		}

		if portNumber == maxPort {
			portNumber = minPort
		} else if !found {
			// An open port is found, return and update user port in DB
			if _, err := collection.UpdateOne(context.TODO(), bson.M{"Username": username}, bson.D{{"$set": bson.D{{"Port": portNumber}}}}); err != nil {
				fmt.Println(err)
			}
			return portNumber
		} else {
			portNumber++
		}
	}
	// All ports taken.
	return 0
}

/*
Creates VM for user when they click to create their first snapshot.

error 0: Creation successful.
error 1: VM name already created.
*/
func CreateVM(w http.ResponseWriter, r *http.Request, db *mongo.Database) {
	args, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		log.Fatal(err)
	}

	var usernameStruct Username
	err = json.Unmarshal(args, &usernameStruct)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	//Possible command to make it run in background?
	status := runCommand("virt-install --connect=qemu:///system --name " + usernameStruct.Username + " --os-type=Linux --os-variant=ubuntu20.04 --memory=2048 --vcpus=1 --disk path=/var/lib/libvirt/images/" + usernameStruct.Username + ".qcow2,bus=virtio,size=5 --graphics spice --cdrom /var/lib/kimchi/isos/ubuntu-20.04.1-desktop-amd64.iso --qemu-commandline=env=SPICE_DEBUG_ALLOW_MC=1")

	if status {
		fmt.Println("VM Created")
		sendReturn(`{"error": 0}`, w)
	} else {
		fmt.Println("VM name already in use (ERROR)")
		sendReturn(`{"error": 1}`, w)
	}
}

/*
Starts VM of given username (port number taken from user document in DB)

error 0: VM started and connected to websockify.
error 1: VM started or doesn't exist.
*/
func StartVM(w http.ResponseWriter, r *http.Request, db *mongo.Database) {
	args, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		log.Fatal(err)
	}

	var usernameStruct Username
	err = json.Unmarshal(args, &usernameStruct)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	//vmName := "vm1"

	status := runCommand("virsh -c qemu:///system start " + usernameStruct.Username)

	portNumber := getPort(usernameStruct.Username, db.Collection("users"))

	if portNumber == 0 {
		// Unable to assign port to user to start VM
		sendReturn(`{"error": 2, "port":`+strconv.Itoa(portNumber)+`}`, w)
	}

	if status {
		fmt.Println("VM Started")

		// Number will be from DB as user is assigned here and stored in DB.

		//If port number is already taken, check if user has it and increment by 1. FOR LOOP HERE

		out, _ := exec.Command("virsh", "-c", "qemu:///system", "domdisplay", "--type", "spice", usernameStruct.Username).Output()

		address := "localhost:" + string(out[len(out)-5:len(out)-1])

		fmt.Println("/websockify/websockify.py." + strconv.Itoa(portNumber) + "." + address + ".")

		args := [3]string{"/websockify/websockify.py", strconv.Itoa(portNumber), address}
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Start()

		fmt.Println("VM connected to websockify")
		sendReturn(`{"error": 0}`, w)

	} else {
		fmt.Println("VM already started or doesn't exist (ERROR)")
		sendReturn(`{"error": 1, "port":`+strconv.Itoa(portNumber)+`}`, w)
	}

}

/*
Check status of VM given username (on/off)

(CAN FINISH LATER)

func CheckVMStatus(w http.ResponseWriter, r *http.Request, db *mongo.Database) {
	vmName := "vm1"

	s := "virsh list --all | grep " + vmName
	fmt.Println(s)
	args := strings.Split(s, " ")

	out, err := exec.Command(args[0], args[1:]...).Output()
	if err != nil {
		log.Fatal(err)
	} else {
		fmt.Printf("%s", out)
	}
}
*/

/*
Starts shutdown process on VM of given username.

error 0: VM shutdown successful.
error 1: VM already shutdown.
*/
func ShutdownVM(w http.ResponseWriter, r *http.Request, db *mongo.Database) {
	args, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		log.Fatal(err)
	}

	var usernameStruct Username
	err = json.Unmarshal(args, &usernameStruct)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	status := runCommand("virsh -c qemu:///system shutdown " + usernameStruct.Username)

	if status {
		fmt.Println("VM Shutdown")
		sendReturn(`{"error": 0}`, w)
	} else {
		fmt.Println("VM already shutdown (ERROR)")
		sendReturn(`{"error": 1}`, w)
	}
}

/*
Save snapshot for given username of given snapshot name at the VM's current state.

error 0: Snapshot created.
error 1: Snapshot name already in use.
*/
func CreateSnapshot(w http.ResponseWriter, r *http.Request, db *mongo.Database) {
	args, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		log.Fatal(err)
	}

	var userSnapStruct UserAndSnapshot
	err = json.Unmarshal(args, &userSnapStruct)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	status := runCommand("virsh -c qemu:///system snapshot-create-as --domain " + userSnapStruct.Username + " --name " + userSnapStruct.SnapshotName)

	if status {
		fmt.Println("Snapshot Created")
		sendReturn(`{"error": 0}`, w)
	} else {
		fmt.Println("Snapshot name already in use (ERROR)")
		sendReturn(`{"error": 1}`, w)
	}
}

/*
Load snapshot for given username and given snapshot into the state of user's VM.

error 0: Snapshot loaded.
error 1: Snapshot name doesn't exist.
*/
func LoadSnapshot(w http.ResponseWriter, r *http.Request, db *mongo.Database) {
	args, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		log.Fatal(err)
	}

	var userSnapStruct UserAndSnapshot
	err = json.Unmarshal(args, &userSnapStruct)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	status := runCommand("virsh snapshot-revert --domain " + userSnapStruct.Username + " --snapshotname " + userSnapStruct.SnapshotName)

	if status {
		fmt.Println("Snapshot Loaded")
		sendReturn(`{"error": 0}`, w)
	} else {
		fmt.Println("Snapshot doesn't exist (ERROR)")
		sendReturn(`{"error": 1}`, w)
	}
}

/*
Deletes snapshot of given name from dashboard.

error 0: Snapshot deleted.
error 1: Snapshot doesn't exist.
*/
func DeleteSnapshot(w http.ResponseWriter, r *http.Request, db *mongo.Database) {
	args, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		log.Fatal(err)
	}

	var userSnapStruct UserAndSnapshot
	err = json.Unmarshal(args, &userSnapStruct)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	status := runCommand("virsh -c qemu:///system snapshot-delete --domain " + userSnapStruct.Username + " --snapshotname " + userSnapStruct.SnapshotName)

	if status {
		fmt.Println("Snapshot deleted")
		sendReturn(`{"error": 0}`, w)
	} else {
		fmt.Println("Snapshot not found (ERROR)")
		sendReturn(`{"error": 1}`, w)
	}
}

/*
Sends the given JSON back as a respose to request
*/
func sendReturn(jsonRes string, w http.ResponseWriter) {
	var result map[string]interface{}

	json.Unmarshal([]byte(jsonRes), &result)
	marRes, _ := json.Marshal(result)

	w.Write(marRes)
}

// Error = true, Complete run = false
func runCommand(s string) bool {
	args := strings.Split(s, " ")

	fmt.Println(args)

	cmd := exec.Command(args[0], args[1:]...)

	err := cmd.Run()

	return err == nil

}
