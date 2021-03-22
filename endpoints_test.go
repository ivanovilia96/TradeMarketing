package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/mux"
)

// удаляет данные из таблицы по id
func deleteRowAfterTest(id int64) {
	sqlQuery := "DELETE FROM Statistics WHERE id=" + strconv.Itoa(int(id)) + ";"
	_, err := ConnectedDataBase.Exec(sqlQuery)
	if err != nil {
		panic(err.Error())
	}
}

// создает стандартные данные для теста и возвращает id
func createRowInTableForTest() int64 {
	sqlQuery := `insert into Statistics(date,views,clicks,cost,cpc,cpm) values ('2000-01-30',1000,100,"1.12","1.11","1.12")`
	result, err := ConnectedDataBase.Exec(sqlQuery)
	if err != nil {
		panic(err.Error())
	}
	lastInsertId, err := result.LastInsertId()
	if err != nil {
		panic(err.Error())
	}
	return lastInsertId
}

// проверяем то, что возвращается массив а не 1 запись, который состоит по меньшей мере из того кол-ва  row, которые мы добавляем в ручную в тесте
func TestGetStatisticsConutOFReturnedRows(t *testing.T) {
	createdRowId := createRowInTableForTest()
	createdRowId2 := createRowInTableForTest()
	defer deleteRowAfterTest(createdRowId)
	defer deleteRowAfterTest(createdRowId2)
	req, err := http.NewRequest("GET", "/statistics-from=1999-12-30,to=2001-12-32,sortField=date", nil)
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()
	router := mux.NewRouter()
	router.HandleFunc("/statistics-from={fromDate},to={toDate},sortField={sortField}", GetStatistics)
	router.ServeHTTP(recorder, req)

	requestBody, err := ioutil.ReadAll(recorder.Body)
	if err != nil {
		panic(err)
	}

	expectedResult := []Statistics{}
	err = json.Unmarshal(requestBody, &expectedResult)

	// Check the count which we except.
	if len(expectedResult) < 2 {
		t.Error("excepten at least 2 row returned from query")

	}
}

// проверяю работает ли валидация в урле по полям (fromDate , toDate)
func TestGetStatisticsDateValidation(t *testing.T) {
	createdRowId := createRowInTableForTest()
	defer deleteRowAfterTest(createdRowId)
	fromDate := "1999-12-30"
	toDate := "2001-12-30"
	req, err := http.NewRequest("GET", "/statistics-from="+fromDate+",to="+toDate+",sortField=date", nil)
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()
	router := mux.NewRouter()
	router.HandleFunc("/statistics-from={fromDate},to={toDate},sortField={sortField}", GetStatistics)
	router.ServeHTTP(recorder, req)

	requestBody, err := ioutil.ReadAll(recorder.Body)
	if err != nil {
		panic(err)
	}

	expectedResult := []Statistics{}
	err = json.Unmarshal(requestBody, &expectedResult)
	layout := "2006-01-02"
	parsedFromDate, err := time.Parse(layout, fromDate)
	if err != nil {
		t.Error("invalid data")
	}
	parsedToDate, err := time.Parse(layout, toDate)
	if err != nil {
		t.Error("invalid data")
	}
	for _, v := range expectedResult {
		parsedTime, err := time.Parse(layout, v.Date)
		if err != nil {
			t.Error("invalid data")
		}
		// check on valid date of returned rows
		if (!parsedTime.After(parsedFromDate) || !parsedTime.Before(parsedToDate)) && !(v.Date == fromDate || v.Date == toDate) {
			t.Error("data in not between theese datas : " + fromDate + " - " + toDate)
		}
	}

	// Check the count which we except.
	if len(expectedResult) < 1 {
		t.Error("except at least 1 row returned from query")
	}
}

//Проверяю что запись добавляется в Бд с валидными данными
func TestPutStatistics(t *testing.T) {
	dataForAdd := []interface{}{"2000-01-30", 1000, 100, "1.12", "", ""}
	statistics := Statistics{
		dataForAdd[0].(string), dataForAdd[1].(int), dataForAdd[2].(int), dataForAdd[3].(string), dataForAdd[4].(string), dataForAdd[5].(string),
	}

	data, err := json.Marshal(statistics)
	if err != nil {
		t.Fatal(err)
	}
	res, err := http.NewRequest("PUT", "/statistics", bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(PutStatistics)
	handler.ServeHTTP(rr, res)
	// т.к. по идее требовалось сделать mock бд с помощью docker, в связи с нехваткой времени пока так
	id := strings.Split(rr.Body.String(), `"id":`)
	id = strings.Split(id[1], "}")
	idInt, err := strconv.Atoi(id[0])
	if err != nil {
		t.Fatal(err)
	}
	defer deleteRowAfterTest(int64(idInt))
	if status := rr.Code; status != http.StatusAccepted {
		t.Errorf("handler PutStatistics returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
	//запрашиваем то, что мы только что добавили в бд
	addedRow, err := ConnectedDataBase.Query(fmt.Sprintf("select date, views, clicks, cost, cpc, Cpm from statistics where id = %v", id[0]))
	if err != nil {
		t.Fatal(err.Error())
	}

	// проверяем что то, что мы добавили присутствует и соответствует ожиданиям
	for addedRow.Next() {
		oneRowData := Statistics{}
		err := addedRow.Scan(&oneRowData.Date, &oneRowData.Views, &oneRowData.Clicks, &oneRowData.Cost, &oneRowData.Cpc, &oneRowData.Cpm)
		if err != nil {
			panic(err)
		}
		cost, err := strconv.ParseFloat(oneRowData.Cost, 64)
		if err != nil {
			panic(err)
		}
		floatCpc, err := strconv.ParseFloat(oneRowData.Cpc, 64)
		if err != nil {
			panic(err)
		}
		Cpc := fmt.Sprintf("%.2f", floatCpc)
		floatCpm, err := strconv.ParseFloat(oneRowData.Cpm, 64)
		if err != nil {
			panic(err)
		}
		Cpm := fmt.Sprintf("%.2f", floatCpm)
		// идет проверка значений которые мы получили с бд, тем значениям, которые должны были получиться в результате вычислений по нашим условиям (считаем тут фактически)
		if oneRowData.Date != dataForAdd[0].(string) || oneRowData.Views != dataForAdd[1].(int) || oneRowData.Clicks != dataForAdd[2].(int) || oneRowData.Cost != dataForAdd[3].(string) ||
			Cpc != fmt.Sprintf("%.2f", cost/float64(oneRowData.Clicks)) || Cpm != fmt.Sprintf("%.2f", cost/float64(oneRowData.Views)*1000) {
			t.Error("there is an error in math arifmetics in PutStatistics func ")

		}
	}
}

// проверяем то, что бд очищается
func TestDeleteStatistics(t *testing.T) {
	createdRowId := createRowInTableForTest()
	createdRowId2 := createRowInTableForTest()
	defer deleteRowAfterTest(createdRowId)
	defer deleteRowAfterTest(createdRowId2)
	req, err := http.NewRequest("DELETE", "/statistics", nil)
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()
	router := mux.NewRouter()
	router.HandleFunc("/statistics", DeleteStatistics)
	router.ServeHTTP(recorder, req)

	requestBody, err := ioutil.ReadAll(recorder.Body)
	if err != nil {
		panic(err)
	}

	expectedResult := []Statistics{}
	err = json.Unmarshal(requestBody, &expectedResult)

	// Check the count which we except.
	if len(expectedResult) > 0 {
		t.Error("excepted empty DB t")
	}
}
