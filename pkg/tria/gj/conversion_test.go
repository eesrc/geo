package gj

import (
	"reflect"
	"testing"
)

func Test_convertUTMtoLatLong(t *testing.T) {
	type args struct {
		coordinates []float64
		zone        int
		latZone     string
	}
	tests := []struct {
		name string
		args args
		want []float64
	}{
		{
			name: "Simple convertion of UTM to LatLong",
			args: args{
				coordinates: []float64{510000, 7042000},
				zone:        33,
				latZone:     "Z",
			},
			want: []float64{15.20090999682305, 63.506393858088},
		},
		{
			name: "Simple convertion of UTM to LatLong with decimals",
			args: args{
				coordinates: []float64{435664.6399999987, 7294165.270000309},
				zone:        33,
				latZone:     "Z",
			},
			want: []float64{13.595453790407474, 65.76282088756047},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ConvertUTMtoLatLong(tt.args.coordinates, tt.args.zone, tt.args.latZone); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("convertUTMtoLatLong() = %v, want %v", got, tt.want)
			}
		})
	}
}
