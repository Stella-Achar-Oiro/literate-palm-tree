// internal/service/config.go
package service

const (
    MapboxAccessToken  = "pk.eyJ1Ijoic3RlbGxhYWNoYXJvaXJvIiwiYSI6ImNtMWhmZHNlODBlc3cybHF5OWh1MDI2dzMifQ.wk3v-v7IuiSiPwyq13qdHw"
    MapboxGeocodingAPI = "https://api.mapbox.com/geocoding/v5/mapbox.places"
)

func GetMapboxAccessToken() string {
    return MapboxAccessToken
}

func GetMapboxGeocodingAPI() string {
    return MapboxGeocodingAPI
}