package repository

import "testing"

func TestGetWeather(t *testing.T) {
	repo:=JMIRepository{}
	weather,err:=repo.GetWeather(1.0,1.0)
	if err!=nil{
		t.Error(err)
	}
}

func TestGetAvailableTime(t *testing.T) {
	repo=JMIRepository{}
	_,err:=repo.GetAvailableTime()
	if err!=nil{
		t.Error(err)
	}
}

func TestGetWeather_error(t *testing.T) {
	repo=JMIRepository{}
	weather,err:=repo.GetWeather(1.0,1.0)
	if err==nil{
		t.Error("error expected")
	}
}
