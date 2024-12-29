package utils

import (
	cfg "application/config/gen"
	"application/pkg/utils/log"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"math/rand"
	"os"
	"time"
)

var (
	LubanTables   *cfg.Tables
	TimeZone      *time.Location
	TelegramToken string
	BotUsername   string
	Develop       = "develop"
	Test          = "test"
	Produce       = "produce"
	NetworkTRC20  = "TRC20"
	NetworkTON    = "TON"
	InitBalance   = int64(500000) //临时
	InitUSDT      = int64(0)      //临时 (美分为单位)
	InitVoucher   = int64(0)      //临时
)

func init() {
	InitConfig("./config/", "yaml", []string{"common"})
	//InitConfig("./config/", "yaml", []string{"common_prod"})
	TelegramToken = "7369353290:AAEgnAXlCIXdHMOhy2AZoqaD5pOB_fj5Hi4"
	BotUsername = "XbinSky_bot"
	rand.Seed(time.Now().UnixNano())
}

func InitTimeZone() {
	var err error
	TimeZone, err = time.LoadLocation("Asia/Shanghai")
	if err != nil {
		log.Fatalf("Failed to load America/Mazatlan time zone: %v", err)
	}
}

func InitConfig(path string, fileType string, fileNames []string) {
	viper.SetConfigType(fileType)
	viper.AddConfigPath(path)
	for index, fileName := range fileNames {
		if index == 0 {
			viper.SetConfigName(fileName)
			err := viper.ReadInConfig() // Find and read the config file
			if err != nil {             // Handle errors rezading the config file
				panic(fmt.Errorf("fatal ecode config file: %w", err))
			}
		} else {
			viper.SetConfigName(fileName)
			err := viper.MergeInConfig()
			if err != nil { // Handle errors rezading the config file
				panic(fmt.Errorf("fatal ecode config file: %w", err))
			}
		}
	}
	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		fmt.Println("Config file changed:", e.Name)
	})
}

func loader(file string) ([]map[string]interface{}, error) {
	if bytes, err := os.ReadFile("./config/output_json/" + file + ".json"); err != nil {
		return nil, errors.New(fmt.Sprintf("%s: %s", file, err.Error()))
	} else {
		jsonData := make([]map[string]interface{}, 0)
		if err = json.Unmarshal(bytes, &jsonData); err != nil {
			return nil, err
		}
		return jsonData, nil
	}
}

func InitLubanTables() {
	var err error
	LubanTables, err = cfg.NewTables(loader)
	if err != nil {
		panic(fmt.Errorf("fatal on load config tables: %w", err))
	}
}
