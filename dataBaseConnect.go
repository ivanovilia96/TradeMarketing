package main

import (
	"bufio"
	"database/sql"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

func dataBaseConnect() *sql.DB {
	var (
		connectionMethod                                      = "@tcp"
		hostname                                              = "127.0.0.1"
		port                                                  = "3306"
		DBName                                                = "firstDB"
		login, password, getDataFromConsole                   string
		errOnReadingFromCMDLogin, errOnReadingFromCMDPassword error
	)
	println("Do you want to enter your username and password or download from a file? response options [file or cmd ] default is file :")
	getDataFromConsole, _ = bufio.NewReader(os.Stdin).ReadString('\n')
	if len(getDataFromConsole) > 2 && getDataFromConsole[:len(getDataFromConsole)-2] == "cmd" {
		println("Enter login for mySql connection:")
		login, errOnReadingFromCMDLogin = bufio.NewReader(os.Stdin).ReadString('\n')
		println("Enter password for mySql connection:")
		password, errOnReadingFromCMDPassword = bufio.NewReader(os.Stdin).ReadString('\n')
	} else {
		data, err := ioutil.ReadFile("informationForConnection.txt")
		if err != nil {
			log.Fatal(err)
		}
		stringData := string(data)
		splittedData := strings.Split(stringData, ",password:")
		login = strings.Split(splittedData[0], "login:")[1] + "  "
		password = splittedData[1] + "  "
		println(login, "login")
		println(password, "password")
	}

	if errOnReadingFromCMDLogin != nil {
		panic("Error when  you try pass your login" + errOnReadingFromCMDLogin.Error())
	}

	if errOnReadingFromCMDPassword != nil {
		panic("Error when  you try pass your password" + errOnReadingFromCMDPassword.Error())
	}

	db, err := sql.Open("mysql", login[:len(login)-2]+":"+password[:len(password)-2]+connectionMethod+"("+hostname+":"+port+")/")
	if err != nil {
		panic(err)
	}
	_, err = db.Exec("CREATE DATABASE IF NOT EXISTS " + DBName)
	if err != nil {
		panic("Error on creating dataBase:" + err.Error())
	}

	_, err = db.Exec("USE " + DBName)
	if err != nil {
		panic("Error when we call USE dataBase : " + err.Error())
	}

	sqlQueryCreateNotes := `CREATE TABLE if not exists Statistics(
		date DATE NOT NULL,
		views int ,
		clicks int ,
		cost DECIMAL(19 , 2 ),
		cpc DECIMAL(19 , 2 ),
		cpm DECIMAL(19 , 2 ) 
		);`
	_, err = db.Exec(sqlQueryCreateNotes)
	if err != nil {
		panic("Error on creating Statistics table: " + err.Error())
	}

	return db
}

var ConnectedDataBase = dataBaseConnect()
