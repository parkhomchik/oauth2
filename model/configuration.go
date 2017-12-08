package model

import (
	"encoding/json"
	"os"
)

//Configuration Настройки приложения
type Configuration struct {
	DbUser string
	DbPass string
	DbName string
	DbHost string
}

//Load Загрузка настроек из файла конфигураций
func (c *Configuration) Load() error {
	file, _ := os.Open("config.json")
	decoder := json.NewDecoder(file)
	err := decoder.Decode(&c)
	return err
}
