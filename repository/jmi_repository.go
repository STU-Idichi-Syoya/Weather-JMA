package repository

import (
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"math"
	"net/http"
	"strconv"
)

type JMIRepository struct {
}

//参考：https://www.trail-note.net/tech/coordinate/
// 返却値はタイル座標，ピクセル座標x,y
/*

in js
var latLon2tile = function(lat, lon, zoom) {
    lat = parseFloat(lat); // 緯度
    lon = parseFloat(lon); // 経度
    zoom = parseInt(zoom); // 尺度

    var pixelX = parseInt(Math.pow(2, zoom + 7) * (lon / 180 + 1));
    var tileX = parseInt(pixelX / 256);

    var pixelY = parseInt((Math.pow(2, zoom + 7) / Math.PI) * ((-1 * Math.atanh(Math.sin((Math.PI / 180) * lat))) + Math.atanh(Math.sin((Math.PI / 180) * L))));
    var tileY = parseInt(pixelY / 256);

    $('#pixelX').val(pixelX); // ピクセル座標X
    $('#tileX').val(tileX); // タイル座標X
    $('#pixelY').val(pixelY); // ピクセル座標Y
    $('#tileY').val(tileY); // タイル座標Y
};

*/
func latlon2TileCoordinate(lat float64, lon float64,zoom int) (int, int,int,int) {

	pixelX := int(math.Pow(2, float64(zoom) + 7) * (lon / 180 + 1))
	tileX := int(pixelX % 256)
	pixelY := int((math.Pow(2, float64(zoom) / math.Pi) * ((-1 * math.Atanh(math.Sin((math.Pi / 180) * lat))) + math.Atanh(math.Sin((math.Pi / 180) * lon)))))
	tileY := int(pixelY % 256)
	return pixelX,pixelY,tileX,tileY
}

type TargetTime struct {
	BaseTime string `json:"baseTime"`
	ValidTime string `json:"validTime"`
	Elements []string `json:"elements"`
}

type targetTimes_N2 struct {
	TargetTime []TargetTime 
}


func (jmiRepository *JMIRepository) GetAvailableTime() (targetTimes_N2, error) {
	const url= "https://www.jma.go.jp/bosai/jmatile/data/nowc/targetTimes_N2.json"
	res,err:=http.Get(url)
	defer res.Body.Close()
	var data targetTimes_N2

	if err!=nil{
		return data,err
	}
	
	if err:=json.NewDecoder(res.Body).Decode(&data);err!=nil{
		return data,err
	}

	
	data.TargetTime = append(data.TargetTime, TargetTime{
		BaseTime: data.TargetTime[0].BaseTime,
		ValidTime: data.TargetTime[0].BaseTime,
	})
	
	return data,nil

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
	x,y,px,py:=latlon2TileCoordinate(lat,lon,zoom)
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

	for i,v:=range atime.TargetTime{
		// url:=fmt.Sprintf("https://www.jma.go.jp/bosai/jmatile/data/nowc/20220816111500/none/20220816111500/surf/hrpns/10/896/407.png")
		// todo cache
		url:=fmt.Sprintf("https://www.jma.go.jp/bosai/jmatile/data/nowc/%s/none/%s/surf/hrpns/%d/%d/%d.png",v.BaseTime,v.ValidTime,zoom,x,y)
		res,err:=http.Get(url)
		defer res.Body.Close()
		if err!=nil{
			return &Ydf{},err
		}
		img,_,err:=image.Decode(res.Body)
		if err!=nil{
			return &Ydf{},err
		}
		color:=img.At(px,py).(color.RGBA)
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

		

		date,err:=strconv.ParseInt(v.BaseTime,10,64)
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
	ydf.Feature.Property.WeatherList.Weather=weatherList

	return ydf,nil
}