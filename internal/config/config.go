package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"fmt"
)

const configFileName = "gatorconfig.json"

func getConfigFilePath() (string,error) {	
	currDir,err := os.Getwd()
	if err != nil {
		fmt.Println("Error getting home dir:", err)
        return "",err
	}
	
	filePath := filepath.Join(currDir,"gatorconfig.json")
	return filePath,nil
}

type Config struct {
	DbUrl				string 		`json:"db_url"`
	CurrentUserName		string		`json:"current_user_name"`
}


func Read() Config {

	filePath,err := getConfigFilePath()
	if err != nil {
		return Config{}
	}

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

func write(cfg Config) error {
	filePath,err := getConfigFilePath()
	if err != nil {
		return err
	}
	data,err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		fmt.Println("error marshaling the config",err)
	}
	err = os.WriteFile(filePath, data, 0644)
    if err != nil {
        fmt.Println("Error writing file:", err) 
		return err
    }
	return nil
}

func (cfg Config) SetUser(username string) {
	cfg.CurrentUserName = username
	write(cfg)
}













