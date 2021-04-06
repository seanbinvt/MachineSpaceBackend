package fileManagement

import (
	"fmt"
	"log"
	//"os"
	//"net/http"
	"os/exec"
	"strings"
	"strconv"

	//"bytes"
	//"io"
	//"strings"

	//"golang.org/x/crypto/bcrypt"

	//"go.mongodb.org/mongo-driver/mongo"
)

var prevPort = 5959

type CreateVMRequest struct {
	Username     string `json:"username" bson:"Username"`
}


type DeleteSnapshotRequest struct {
	Username     string `json:"username" bson:"Username"`
	SnapshotName string `json:"snapshotName" bson:"SnapshotName"`
}

type LoadSnapshotRequest struct {
	Username     string `json:"username" bson:"Username"`
	SnapshotName string `json:"snapshotName" bson:"SnapshotName"`
}



/*
Creates VM for user when they click to create their first snapshot.
*/
func CreateVM() {
	/*
		args, err := ioutil.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil {
			log.Fatal(err)
		}
	*/

	vmName := "Bob"

	//Possible command to make it run in background?
	status := runCommand("virt-install --connect=qemu:///system --name "+vmName+" --os-type=Linux --os-variant=ubuntu20.04 --memory=2048 --vcpus=1 --disk path=/var/lib/libvirt/images/"+vmName+".qcow2,bus=virtio,size=5 --graphics spice --cdrom /var/lib/kimchi/isos/ubuntu-20.04.1-desktop-amd64.iso --qemu-commandline=env=SPICE_DEBUG_ALLOW_MC=1")

	if status == true {
		fmt.Println("VM Created")
	} else {
		fmt.Println("VM name already in use (ERROR)")
	}
}

/*
Starts VM of given username (port number taken from user document in DB)
*/
func StartVM() {
	vmName := "vm1"

	status := runCommand("virsh -c qemu:///system start "+vmName)

	if status == true {
		fmt.Println("VM Started")

		// Number will be from DB as user is assigned here and stored in DB.
		portNumber := prevPort

		//If port number is already taken, check if user has it and increment by 1. FOR LOOP HERE

		out, _ := exec.Command("virsh", "-c", "qemu:///system", "domdisplay", "--type", "spice",vmName).Output()

		address := "localhost:"+string(out[len(out)-5:len(out)-1])

		fmt.Println("/websockify/websockify.py."+strconv.Itoa(portNumber)+"."+address+".")

		args := [3]string{"/websockify/websockify.py",strconv.Itoa(portNumber),address}
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Start()

		if portNumber > 6000 {
			prevPort = 5959
		} else {
			prevPort = portNumber+1
		}

		fmt.Println("VM connected to websockify")

	} else {
		fmt.Println("VM already started or doesn't exist (ERROR)")
	}

}

/*
Check status of VM given username (on/off)

(CAN FINISH LATER)
*/
func CheckVMStatus() {
	vmName := "vm1"

	s := "virsh list --all | grep "+vmName
	fmt.Println(s)
	args := strings.Split(s, " ")

	out, err := exec.Command(args[0], args[1:]...).Output()
	if err != nil {
		log.Fatal(err)
	} else {
		fmt.Printf("%s", out)
	}
}

/*
Starts shutdown process on VM of given username.
*/
func ShutdownVM() {
	vmName := "vm1"

	status := runCommand("virsh -c qemu:///system shutdown "+vmName)

	if status == true {
		fmt.Println("VM Shutdown")
	} else {
		fmt.Println("VM already shutdown (ERROR)")
	}
}

/*
Save snapshot for given username of given snapshot name at the VM's current state.
*/
func CreateSnapshot() {
	vmName := "vm1"
	snapshotName := "test"

	status := runCommand("virsh -c qemu:///system snapshot-create-as --domain "+vmName+" --name "+snapshotName)

	if status == true {
		fmt.Println("Snapshot Created")
	} else {
		fmt.Println("Snapshot name already in use (ERROR)")
	}
}

/*
Load snapshot for given username and given snapshot into the state of user's VM.
*/
func LoadSnapshot() {
	vmName := "vm1"
	snapshotName := "test"

	status := runCommand("virsh snapshot-revert --domain "+vmName+" --snapshotname "+snapshotName)

	if status == true {
		fmt.Println("Snapshot Loaded")
	} else {
		fmt.Println("Snapshot doesn't exist (ERROR)")
	}
}

/*
Deletes snapshot of given name from dashboard.
*/
//func DeleteSnapshot(w http.ResponseWriter, r *http.Request, db *mongo.Database) {
func DeleteSnapshot() {
	/*
		args, err := ioutil.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil {
			log.Fatal(err)
		}
	*/

	vmName := "Kevin"
	snapshotName := "deletemepls"

	status := runCommand("virsh -c qemu:///system snapshot-delete --domain "+vmName+" --snapshotname "+snapshotName)

	if status == true {
		fmt.Println("Snapshot deleted")
	} else {
		fmt.Println("Snapshot not found (ERROR)")
	}
}

// Error = true, Complete run = false
func runCommand(s string) bool {
	args := strings.Split(s, " ")

	fmt.Println(args)

	cmd := exec.Command(args[0], args[1:]...)

	err := cmd.Run()

	if err != nil {
		//log.Fatal(err)
		return false
	}
	return true

}
