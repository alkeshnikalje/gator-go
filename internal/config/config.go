package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"fmt"
)

type Config struct {
	DbUrl				string 		`json:"db_url"`
	CurrentUserName		string		`json:"current_user_name"`
}


func Read() Config {
	currDir,err := os.Getwd()
	if err != nil {
		fmt.Println("Error getting home dir:", err)
        return Config{}
	}
	
	filePath := filepath.Join(currDir,"gatorconfig.json")

	file,err := os.Open(filePath)
	if err != nil {
		fmt.Println("Error opening a file", err)
		return Config{}
	}
	defer file.Close()
	
	var config Config
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
	    fmt.Println("Error decoding JSON:", err)
        return Config{}
	}
	
	return config 
}


func (cfg Config) SetUser(username string) {
	cfg.CurrentUserName = username
	curDir, _ := os.Getwd()
	filePath := filepath.Join(curDir, "gatorconfig.json")
	data, _ := json.MarshalIndent(cfg, "", "  ")
	err := os.WriteFile(filePath, data, 0644)
    if err != nil {
        fmt.Println("Error writing file:", err) 
    }
}













