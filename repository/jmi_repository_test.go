package repository

import "testing"

func Test_Coordinate(t *testing.T){
		
	zoom:=14
	lat:=36.104665
	lon:=140.087099
	ansTileX:=14567
	ansTileY:=6427
	x,y,_,_:=Latlon2TileCoordinate(lat,lon,zoom)

	if ansTileX==x && ansTileY==y{
	}else{
		t.Errorf("ansTileX:%d,x:%d,ansTileY:%d,y:%d",ansTileX,x,ansTileY,y)
		t.Errorf("latitude should be between -90 and 90")
	}
}
