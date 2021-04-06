package fileManagement

import (
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"strconv"
	"strings"

	//"bytes"
	//"io"
	//"strings"

	//"golang.org/x/crypto/bcrypt"

	"go.mongodb.org/mongo-driver/mongo"
)

type CreateSnapshotRequest struct {
	Username     string `json:"username" bson:"Username"`
	SnapshotName string `json:"snapshotName" bson:"SnapshotName"`
	Description  string `json:"description" bson:"Description"`
}

type DeleteSnapshotRequest struct {
	Username     string `json:"username" bson:"Username"`
	SnapshotName string `json:"snapshotName" bson:"SnapshotName"`
}

type LoadSnapshotRequest struct {
	Username     string `json:"username" bson:"Username"`
	SnapshotName string `json:"snapshotName" bson:"SnapshotName"`
}

func CreateSnapshot(w http.ResponseWriter, r *http.Request, db *mongo.Database) {
	/*
		args, err := ioutil.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil {
			log.Fatal(err)
		}
	*/

	status := runCommand("virt-install --connect=qemu:///system --name (vm) --os-type=Linux --os-variant=ubuntu20.04 --memory=2048 --vcpus=1 --disk path=/var/lib/libvirt/images/(vm).qcow2,bus=virtio,size=5 --graphics spice --cdrom /var/lib/kimchi/isos/ubuntu-20.04.1-desktop-amd64.iso --qemu-commandline=env=SPICE_DEBUG_ALLOW_MC=1 > /dev/null 2>&1 &")

	fmt.Println(strconv.FormatBool(status))
}

func DeleteSnapshot(w http.ResponseWriter, r *http.Request, db *mongo.Database) {
	/*
		args, err := ioutil.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil {
			log.Fatal(err)
		}
	*/

	status := runCommand("")

	fmt.Println(strconv.FormatBool(status))
}

func runCommand(s string) bool {
	args := strings.Split(s, " ")

	cmd := exec.Command(args[0], args[1:]...)

	err := cmd.Run()

	if err != nil {
		log.Fatal(err)
		return false
	}
	return true

}
