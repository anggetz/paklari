package core

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"

	"paklari/internal/models"
)

type Exec interface {
	ReadEntries(string) Exec
	Run()
}

var (
	Error   = "Error"
	Running = "Running"
	Done    = "Done"
)

type execImpl struct {
	EntriesValue []models.ExecEntry
}

//intiate object Execution
func NewExec() Exec {
	return &execImpl{}
}

//read entries file based on entrieslocation path
func (ei *execImpl) ReadEntries(entriesLocation string) Exec {
	// Open our jsonFile
	jsonFile, err := os.Open(entriesLocation)
	// if we os.Open returns an error then handle it
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("Successfully Opened %s \n", entriesLocation)
	// defer the closing of our jsonFile so that we can parse it later on
	defer jsonFile.Close()

	// read our opened jsonFile as a byte array.
	byteValue, _ := ioutil.ReadAll(jsonFile)

	// we initialize our Users array

	// we unmarshal our byteArray which contains our
	// jsonFile's content into 'users' which we defined above
	json.Unmarshal(byteValue, &ei.EntriesValue)

	// return the object itself.
	return ei
}

//run command based on entries value
func (ei *execImpl) Run() {
	for _, ex := range ei.EntriesValue {

		done := make(chan error)

		args := ex.Cmd
		args[0], _ = exec.LookPath(args[0])

		comm := exec.Command(args[0], args[1:]...)
		if ex.Dir != "" {
			comm.Dir = ex.Dir
		}
		comm.Stderr = os.Stderr
		outPipe, err := comm.StdoutPipe()
		if err != nil {
			log.Fatal(err.Error())
		}

		go func() {
			done <- comm.Run()
		}()

		go func() {
			s := bufio.NewScanner(outPipe)
			for s.Scan() {
				fmt.Println("["+ex.Name+"\t]", string(s.Bytes()))
			}
			if s.Err() != nil {
				log.Println("scan:", s.Err())
			}
		}()

		select {
		case err := <-done:
			if err != nil {
				ex.Status = Error
				log.Fatal(err.Error())
			}

			ex.Status = Done
		}

	}

}
