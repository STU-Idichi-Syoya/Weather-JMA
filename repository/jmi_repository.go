package repository

import (
	"encoding/json"
	"fmt"
	"image/color"
	"image/png"
	"math"
	"net/http"
	"strconv"
	"time"
)

type JMIRepository struct {
}

//参考：https://www.trail-note.net/tech/coordinate/
// 返却値はタイル座標，ピクセル座標x,y
/*

in python
def latlon2tile(lon, lat, z):
	x = int((lon / 180 + 1) * 2**z / 2) # x座標
	y = int(((-log(tan((45 + lat / 2) * pi / 180)) + pi) * 2**z / (2 * pi))) # y座標
	print [y,x]

*/
func Latlon2TileCoordinate(lat float64, lon float64,zoom int) (int, int,int,int) {
	
	zoomf:=float64(zoom)

	x := int((lon / 180 + 1) * math.Pow(2,zoomf) / 2) // x座標
	y := int(((-math.Log(math.Tan((45 + lat / 2) * math.Pi / 180)) + math.Pi) * math.Pow(2,zoomf) / (2 * math.Pi))) // y座標
	
	px:=x%256
	py:=y%256

	//show result
	return x,y,px,py
}

/*
	json response:
	[
  {"basetime": "20220817033500", "validtime": "20220817043500", "elements": ["hrpns", "hrpns_nd"]}
 ,{"basetime": "20220817033500", "validtime": "20220817043000", "elements": ["hrpns", "hrpns_nd"]}
 ,{"basetime": "20220817033500", "validtime": "20220817042500", "elements": ["hrpns", "hrpns_nd"]}
 ,{"basetime": "20220817033500", "validtime": "20220817042000", "elements": ["hrpns", "hrpns_nd"]}
 ,{"basetime": "20220817033500", "validtime": "20220817041500", "elements": ["hrpns", "hrpns_nd"]}
 ,{"basetime": "20220817033500", "validtime": "20220817041000", "elements": ["hrpns", "hrpns_nd"]}
 ,{"basetime": "20220817033500", "validtime": "20220817040500", "elements": ["hrpns", "hrpns_nd"]}
 ,{"basetime": "20220817033500", "validtime": "20220817040000", "elements": ["hrpns", "hrpns_nd"]}
 ,{"basetime": "20220817033500", "validtime": "20220817035500", "elements": ["hrpns", "hrpns_nd"]}
 ,{"basetime": "20220817033500", "validtime": "20220817035000", "elements": ["hrpns", "hrpns_nd"]}
 ,{"basetime": "20220817033500", "validtime": "20220817034500", "elements": ["hrpns", "hrpns_nd"]}
 ,{"basetime": "20220817033500", "validtime": "20220817034000", "elements": ["hrpns", "hrpns_nd"]}
]
*/
type JMIWeatherTime struct {
	Basetime string `json:"basetime"`
	Validtime string `json:"validtime"`
	Elements []string `json:"elements"`
}

func (jmiRepository *JMIRepository) GetAvailableTime() ([]JMIWeatherTime, error) {
	const url= "https://www.jma.go.jp/bosai/jmatile/data/nowc/targetTimes_N2.json"
	res,err:=http.Get(url)
	defer res.Body.Close()
	if err != nil {
		return nil, err
	}
	var targetTimes_N2 []JMIWeatherTime
	err = json.NewDecoder(res.Body).Decode(&targetTimes_N2)
	if err != nil {
		return nil, err
	}
	targetTimes_N2 = append(targetTimes_N2, JMIWeatherTime{
		Basetime: targetTimes_N2[0].Basetime,
		Validtime: targetTimes_N2[0].Basetime,
	} )
	return targetTimes_N2, nil
}

type ResultInfo struct {
	Count       int     `json:"Count"`
	Total       int     `json:"Total"`
	Start       int     `json:"Start"`
	Status      int     `json:"Status"`
	Latency     float64 `json:"Latency"`
	Description string  `json:"Description"`
	Copyright   string  `json:"Copyright"`
}
type Geometry struct {
	Type        string `json:"Type"`
	Coordinates string `json:"Coordinates"`
}
type Weather struct {
	Type     string  `json:"Type"`
	Date     int64   `json:"Date"`
	Rainfall float64 `json:"Rainfall"`
}
type WeatherList struct {
	Weather []Weather `json:"Weather"`
}
type Property struct {
	WeatherAreaCode int         `json:"WeatherAreaCode"`
	WeatherList     WeatherList `json:"WeatherList"`
}
type Feature struct {
	ID       string   `json:"Id"`
	Name     string   `json:"Name"`
	Geometry Geometry `json:"Geometry"`
	Property Property `json:"Property"`
}
type Ydf struct {
	ResultInfo ResultInfo `json:"ResultInfo"`
	Feature    Feature    `json:"Feature"`
}

func (jmiRepository *JMIRepository) GetWeather(lat float64, lon float64) (*Ydf, error) {
	const zoom=14
	startTime:=time.Now()
	x,y,px,py:=Latlon2TileCoordinate(lat,lon,zoom)
	atime,err:= jmiRepository.GetAvailableTime()
	if err!=nil{
		return &Ydf{},err
	}
	ydf:=&Ydf{
		ResultInfo:ResultInfo{
			Count:1,
			Total:10,
			Start:0,
			Status:200,
			Latency:0.0,
			Description:"",
			Copyright: "STU-Idichi-Syoya",
		},
		Feature:Feature{
			ID:fmt.Sprintf("%d_%d_%d_%d",x,y,px,py),
			Name:fmt.Sprintf("%d_%d_%d_%d",x,y,px,py),
			Geometry:Geometry{
				Type:"Point",
				Coordinates:fmt.Sprintf("%f,%f",lat,lon),
			},
			Property:Property{
				WeatherAreaCode:0,//固定
				WeatherList:WeatherList{
					Weather:nil,
				},
			},
		},
	}

	weatherList:=make([]Weather,10)

	for i,v:=range atime{
		// url:=fmt.Sprintf("https://www.jma.go.jp/bosai/jmatile/data/nowc/20220816111500/none/20220816111500/surf/hrpns/10/896/407.png")
		// todo cache
		url:=fmt.Sprintf("https://www.jma.go.jp/bosai/jmatile/data/nowc/%s/none/%s/surf/hrpns/%d/%d/%d.png",v.Basetime,v.Validtime,zoom,x,y)
		res,err:=http.Get(url)
		defer res.Body.Close()
		if err!=nil{
			return &Ydf{},err
		}
		img,err:=png.Decode(res.Body)
		if err!=nil{
			return &Ydf{},err
		}
		color:=color.RGBAModel.Convert(img.At(px,py)).(color.RGBA)
		value:=0
		if color.A == 255 {
			if color.R == 242 && color.G == 242 && color.B == 255 {
				//0~1mm
				value = 1
			} else if color.R == 160 && color.G == 210 && color.B == 255 {
				//1~5mm
				value = 5
			} else if color.R == 33 && color.G == 140 && color.B == 255 {
				//5~10mm
				value = 10
			} else if color.R == 0 && color.G == 65 && color.B == 255 {
				//10~20mm
				value = 20
			} else if color.R == 250 && color.G == 245 && color.B == 0 {
				//20~30mm
				value = 30
			} else if color.R == 255 && color.G == 153 && color.B == 0 {
				//30~50mm
				value = 50
			} else if color.R == 255 && color.G == 40 && color.B == 0 {
				//50~80mm
				value = 80
			} else if color.R == 180 && color.G == 0 && color.B == 104 {
				//80mm~
				value = 100
			}
		}

		

		date,err:=strconv.ParseInt(v.Basetime,10,64)
		if err!=nil{
			return &Ydf{},err
		}
		Weather:=Weather{
			Type:     "",
			Date:     date,
			Rainfall: float64(value),
		}
		if i==0{
			Weather.Type="observation"
		}else{
			Weather.Type="forcast"
		}
		weatherList=append(weatherList,Weather)
		
	}
	ydf.ResultInfo.Latency=float64(time.Since(startTime).Milliseconds())
	ydf.Feature.Property.WeatherList.Weather=weatherList

	return ydf,nil
}