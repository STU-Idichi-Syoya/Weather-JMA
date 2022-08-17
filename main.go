package main

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/STU-Idichi-Syoya/Weather-JMA/repository"
)

type API_Params struct {
	Lat float64 `json:"lat" validate:"required"`
	Lon float64 `json:"lon" validate:"required"`
}

func getParams(r *http.Request, params *API_Params) error {
	// get lat
	lat, err := strconv.ParseFloat(r.URL.Query().Get("lat"), 64)
	if err != nil {
		return err
	}
	params.Lat = lat
	// get lon
	lon, err := strconv.ParseFloat(r.URL.Query().Get("lon"), 64)
	if err != nil {
		return err
	}
	params.Lon = lon
	return nil
}

func main(){
	jmiRepository := &repository.JMIRepository{}

	// http server
	http.HandleFunc("/api/weather/V1/place", func(w http.ResponseWriter, r *http.Request) {
		// get params
		params := API_Params{}
		if err := getParams(r, &params); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		// get weather
		weather, err := jmiRepository.GetWeather(params.Lat, params.Lon)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		res,err:=json.Marshal(weather)
		if err!=nil{
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}


		// return weather json
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(res)
	})
						
	http.ListenAndServe(":8080", nil)
}