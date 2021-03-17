package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
)

func DeleteStatistics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	sqlStatement := `DELETE FROM statistics`

	results, err := ConnectedDataBase.Query(sqlStatement)
	defer results.Close()
	fmt.Printf("%v - results", results)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message": "невозможно удалить"}`))

	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "статистика успешно очищена"}`))
}

func GetStatistics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	var (
		vars      = mux.Vars(r)
		sortField = vars["sortField"]
		fromDate  = vars["fromDate"]
		toDate    = vars["toDate"]
	)

	sqlStatement := fmt.Sprintf(
		`select date, views , clicks , cost , cpc , cpm  from statistics where date >='%v' and date <= '%v'  order by %v`, fromDate, toDate, sortField,
	)

	results, err := ConnectedDataBase.Query(sqlStatement)
	defer results.Close()
	fmt.Printf("%v - results", results)
	resultData := []Statistics{}
	for results.Next() {
		oneRowData := Statistics{}
		err := results.Scan(&oneRowData.Date, &oneRowData.Views, &oneRowData.Clicks, &oneRowData.Cost, &oneRowData.Cpc, &oneRowData.Cpm)
		if err != nil {
			panic(err)
		}
		resultData = append(resultData, oneRowData)
	}
	if err != nil {
		panic(err.Error())
	}

	resultDataJson, err := json.Marshal(resultData)
	if err != nil {
		panic(err.Error())
	}
	w.WriteHeader(http.StatusOK)
	w.Write(resultDataJson)
}

func PutStatistics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	requestBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}

	// устанавливаю default value
	bodyInStructure := Statistics{"", 0, 0, "0.00", "", ""}
	err = json.Unmarshal(requestBody, &bodyInStructure)
	if err != nil {
		panic(err)
	}

	if bodyInStructure.Date == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message": "Дата не является опциональным полем"}`))
	}

	splittedMoney := strings.Split(bodyInStructure.Cost, ".")

	// валидация копеек
	if len(splittedMoney) == 2 && len(splittedMoney[1]) > 2 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message": "Неверный формат копеек, их не может быть больше 100 в 1 рубле"}`))
	}

	// для вычислений преобразуем в float64, изначально в нем не принимаю потому что будет куча нулей после числа и валидация на копейки затруднится
	// + если введем не число оно тоже выбросится
	cost, err := strconv.ParseFloat(bodyInStructure.Cost, 64)
	if err != nil {
		panic("Вы передали не правильную стоимость")
	}

	cpcField := 0.0
	if !math.IsNaN(cost / float64(bodyInStructure.Clicks)) {
		cpcField = cost / float64(bodyInStructure.Clicks)
	}

	cpmField := 0.0
	if !math.IsNaN(cost / float64(bodyInStructure.Views)) {
		cpmField = cost / float64(bodyInStructure.Views) * 1000
	}
	sqlQuery := fmt.Sprintf(`insert into Statistics(date, views, clicks, cost, cpc, cpm) values ('%v', %v, %v, "%v", "%v", "%v")`,
		bodyInStructure.Date, bodyInStructure.Views, bodyInStructure.Clicks,
		bodyInStructure.Cost, cpcField, cpmField)

	_, err = ConnectedDataBase.Exec(sqlQuery)
	if err != nil {
		panic(err.Error())
	}

	dataJson, err := json.Marshal(struct {
		Text   string
		Status int
	}{"Запись добавлена успешно", http.StatusAccepted})
	if err != nil {
		panic(err.Error())
	}
	w.WriteHeader(http.StatusAccepted)
	w.Write(dataJson)
}
