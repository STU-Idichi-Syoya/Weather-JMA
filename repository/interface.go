package repository

type IJMIRepository interface{
	GetWeather(lat float64, lon float64) (string, error)
	GetAvailableTime() (string, error)
}

