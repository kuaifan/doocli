package db

import (
	"database/sql"
	"doocli/utils"
	_ "modernc.org/sqlite"
	"os"
	"strconv"
)

func Init() {
	db, err := Instance()
	if err != nil {
		utils.PrintError(err.Error())
		os.Exit(1)
	}
	defer db.Close()
	//
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS config(name VARCHAR(512) PRIMARY KEY NOT NULL, value TEXT NOT NULL)")
	if err != nil {
		utils.PrintError(err.Error())
		os.Exit(1)
	}
}

func Instance() (*sql.DB, error) {
	_ = utils.Mkdir(utils.RunDir("/.cache"), os.ModePerm)
	return sql.Open("sqlite", utils.RunDir("/.cache/cli.db"))
}

func GetConfigString(name string) string {
	value, err := GetConfig(name)
	if err != nil {
		return ""
	}
	return value
}

func GetConfigInt(name string) int {
	value, err := GetConfig(name)
	if err != nil {
		return 0
	}
	i, err := strconv.Atoi(value)
	if err != nil {
		return 0
	}
	return i
}

func GetConfig(name string) (string, error) {
	db, err := Instance()
	if err != nil {
		return "", err
	}
	defer db.Close()
	//
	var value string
	err = db.QueryRow("SELECT value FROM config WHERE name=?", name).Scan(&value)
	if err != nil {
		return "", err
	}
	return value, nil
}

func SetConfig(name, value string) error {
	db, err := Instance()
	if err != nil {
		return err
	}
	defer db.Close()
	//
	_, err = db.Exec("INSERT OR REPLACE INTO config(name, value) VALUES(?, ?)", name, value)
	if err != nil {
		return err
	}
	return nil
}

func DelConfig(name string) error {
	db, err := Instance()
	if err != nil {
		return err
	}
	defer db.Close()
	//
	_, err = db.Exec("DELETE FROM config WHERE name=?", name)
	if err != nil {
		return err
	}
	return nil
}
