package gj

import (
	"reflect"
	"testing"

	"github.com/eesrc/geo/pkg/tria/geometry"
	geojson "github.com/paulmach/go.geojson"
)

func TestGetFeatureFromPoints(t *testing.T) {
	simpleFeature := geojson.NewPolygonFeature([][][]float64{
		[][]float64{
			[]float64{0, 0},
			[]float64{1, 0},
			[]float64{1, 1},
			[]float64{0, 1},
			[]float64{0, 0},
		},
	})

	type args struct {
		points     []geometry.Point
		properties map[string]interface{}
	}
	tests := []struct {
		name string
		args args
		want *geojson.Feature
	}{
		{
			name: "Should correctly return a GeoJSON Feature with duplicated start and end from simple list of points",
			args: args{
				points: []geometry.Point{
					geometry.Point{X: 0, Y: 0},
					geometry.Point{X: 1, Y: 0},
					geometry.Point{X: 1, Y: 1},
					geometry.Point{X: 0, Y: 1},
				},
				properties: map[string]interface{}{},
			},
			want: simpleFeature,
		},
		{
			name: "Should correctly return a GeoJSON Feature from simple list with uplicated start and end",
			args: args{
				points: []geometry.Point{
					geometry.Point{X: 0, Y: 0},
					geometry.Point{X: 1, Y: 0},
					geometry.Point{X: 1, Y: 1},
					geometry.Point{X: 0, Y: 1},
					geometry.Point{X: 0, Y: 0},
				},
				properties: map[string]interface{}{},
			},
			want: simpleFeature,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetFeatureFromPoints(tt.args.points, tt.args.properties); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetFeatureFromPoints() = %v, want %v", got, tt.want)
			}
		})
	}
}
