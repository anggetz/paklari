package models

type ExecEntry struct {
	Key    string   `json:"key"`
	Cmd    []string `json:"cmd"`
	Name   string   `json:"name"`
	Dir    string   `json:"dir"`
	Status string   `json:"status"`
}
