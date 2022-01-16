package core

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/logrusorgru/aurora"
)

type ExecEntry struct {
	Key     string        `json:"key"`
	Cmd     []string      `json:"cmd"`
	Name    string        `json:"name"`
	Dir     string        `json:"dir"`
	Status  ProcessStatus `json:"status"`
	CmdExec *exec.Cmd
}

type ProcessStatus int

const (
	// NotStarted 0
	NotStarted ProcessStatus = iota
	// Running 1
	Running
	// Error 2
	Error
	// Done 3
	Done
)

func (s ProcessStatus) String() string {
	return [...]string{"Not Started", "Running", "Error", "Done"}[s]
}

type Exec interface {
	ReadEntries(string) Exec
	Run(string)
	Start() error
}

const (
	// CommandRun     = "run"
	CommandRun = "run"
	// CommandRestart = "restart"
	CommandRestart = "restart"
	// CommandStop    = "stop"
	CommandStop = "stop"
	// CommandStatus  = "status"
	CommandStatus = "status"
)

type execImpl struct {
	EntriesValue []ExecEntry
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

// Status of the entries
func (u *execImpl) Status() {
	for _, v := range u.EntriesValue {
		var status aurora.Value

		switch v.Status {
		case NotStarted:
			status = aurora.White("[ " + v.Status.String() + " ]")
		case Running:
			status = aurora.Green("[ " + v.Status.String() + " ]")
		case Error:
			status = aurora.Red("[ " + v.Status.String() + " ]")
		case Done:
			status = aurora.Cyan("[ " + v.Status.String() + " ]")
		}

		fmt.Println(status, v.Name)
	}
}

//run command based on entries value
func (ei *execImpl) Run(nameEntry string) {
	for i, ex := range ei.EntriesValue {

		if nameEntry == "" || (nameEntry != "" && nameEntry == ex.Name) {
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

			ei.EntriesValue[i].Status = Running

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

}

// StopEntry by given name
func (u *execImpl) StopEntry(nameEntry string) error {
	for _, ex := range u.EntriesValue {

		if nameEntry == "" || (nameEntry != "" && nameEntry == ex.Name) {

			fmt.Println("Stopping:", ex.Name)

			err := ex.CmdExec.Process.Kill()
			if err != nil {
				return err
			}

			ex.CmdExec = nil
		}
	}

	return nil

}

// Start the runner
func (u *execImpl) Start() error {
	fmt.Println(aurora.Green("Pak Lariiii!"))
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("> ")
		text, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}

		text = strings.TrimSpace(text)
		comps := strings.Split(text, " ")

		switch comps[0] {
		case CommandRun:
			if len(comps) >= 2 {
				if comps[1] != "all" {
					go u.Run(comps[1])
				} else {
					go u.Run("")
				}
			}
		case CommandRestart:
		case CommandStop:
			if len(comps) >= 2 {
				if comps[1] != "all" {
					u.StopEntry(comps[1])
				} else {
					u.StopEntry("")
				}
			}
		case CommandStatus:
			u.Status()
		default:
		}
	}
}
