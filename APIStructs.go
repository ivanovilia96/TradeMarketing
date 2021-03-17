package main

type Statistics struct {
	Date   string `json:"date"`
	Views  int    `json:"views,omitempty"`
	Clicks int    `json:"clicks,omitempty"`
	Cost   string `json:"cost,omitempty"`
	Cpc    string `json:"cpc"`
	Cpm    string `json:"cpm"`
}
