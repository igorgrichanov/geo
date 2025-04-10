package dadata

import (
	"geo/internal/config"
	"geo/internal/service/geo"
	"reflect"
	"testing"
)

func TestGeoService_AddressSearch(t *testing.T) {
	conf := config.MustLoadConfig("../../../../config/local.yaml")
	geoService := NewGeoService(conf.Dadata.ApiKey, conf.Dadata.ApiSecret)

	type args struct {
		input string
	}
	tests := []struct {
		name    string
		args    args
		want    []*geo.Address
		wantErr bool
	}{
		{
			name: "success",
			args: args{
				input: "г Москва, ул Снежная",
			},
			want: []*geo.Address{
				{
					City:   "Москва",
					Street: "Снежная",
					House:  "",
					Lat:    "55.852405",
					Lon:    "37.646947",
				},
				{
					City:   "Москва",
					Street: "Снежная",
					House:  "",
					Lat:    "55.475475",
					Lon:    "36.902316",
				},
				{
					City:   "Москва",
					Street: "Снежная",
					House:  "",
					Lat:    "55.520941",
					Lon:    "37.307258",
				},
				{
					City:   "Москва",
					Street: "Снежная",
					House:  "1",
					Lat:    "55.849384",
					Lon:    "37.64015",
				},
				{
					City:   "Москва",
					Street: "Снежная",
					House:  "1А",
					Lat:    "55.846724",
					Lon:    "37.639545",
				},
				{
					City:   "Москва",
					Street: "Снежная",
					House:  "3А",
					Lat:    "55.8495825",
					Lon:    "37.6409167",
				},
				{
					City:   "Москва",
					Street: "Снежная",
					House:  "4",
					Lat:    "55.8481373",
					Lon:    "37.6414907",
				},
				{
					City:   "Москва",
					Street: "Снежная",
					House:  "5",
					Lat:    "55.849247",
					Lon:    "37.641514",
				},
				{
					City:   "Москва",
					Street: "Снежная",
					House:  "6",
					Lat:    "55.84864",
					Lon:    "37.642159",
				},
				{
					City:   "Москва",
					Street: "Снежная",
					House:  "7",
					Lat:    "55.84959",
					Lon:    "37.642051",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := geoService.AddressSearch(tt.args.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("AddressSearch() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AddressSearch() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGeoService_AddressGeoCode(t *testing.T) {
	conf := config.MustLoadConfig("../../../../config/local.yaml")
	geoService := NewGeoService(conf.Dadata.ApiKey, conf.Dadata.ApiSecret)

	type args struct {
		lat string
		lng string
	}
	tests := []struct {
		name    string
		args    args
		want    []*geo.Address
		wantErr bool
	}{
		{
			name: "success",
			args: args{
				lat: "55.8481373",
				lng: "37.6414907",
			},
			want: []*geo.Address{
				{
					City:   "Москва",
					Street: "Снежная",
					House:  "4",
					Lat:    "55.8481373",
					Lon:    "37.6414907",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := geoService.AddressGeoCode(tt.args.lat, tt.args.lng)
			if (err != nil) != tt.wantErr {
				t.Errorf("AddressGeoCode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got[0], tt.want[0]) {
				t.Errorf("AddressGeoCode() got = %v, want %v", got, tt.want)
			}
		})
	}
}
