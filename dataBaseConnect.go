package main

import (
	"database/sql"
)

func dataBaseConnect() *sql.DB {
	var (
		connectionMethod = "@tcp"
		hostname         = "host.docker.internal"
		port             = "3306"
		DBName           = "firstDB"
		login, password  = "root", "root"
	)
	db, err := sql.Open("mysql", login+":"+password+connectionMethod+"("+hostname+":"+port+")/")
	if err != nil {
		panic(err.Error())
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
		ID INT UNSIGNED NOT NULL AUTO_INCREMENT UNIQUE,
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
