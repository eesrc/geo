package triangulation

import (
	"testing"

	"github.com/eesrc/geo/pkg/tria/geometry"
)

func TestGeoPolygon_PointInside(t *testing.T) {
	type fields struct {
		Name        string
		Description string
		AlphaShape  []geometry.Point
		Triangles   [][]geometry.Point
		BoundingBox geometry.BoundingBox
	}
	type args struct {
		p geometry.Point
	}

	geoPolygon, _ := NewTriangulatedPolygonFromPoints([]geometry.Point{
		geometry.Point{X: 0, Y: 0},
		geometry.Point{X: 10, Y: 0},
		geometry.Point{X: 10, Y: 10},
		geometry.Point{X: 5, Y: 10},
		geometry.Point{X: 5, Y: 5},
		geometry.Point{X: 0, Y: 5},
	})

	testFields := fields{
		Name:        geoPolygon.Name,
		Description: geoPolygon.Description,
		AlphaShape:  geoPolygon.AlphaShape,
		Triangles:   geoPolygon.Triangles,
		BoundingBox: geoPolygon.BoundingBox,
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name:   "returns true if point is inside polygon",
			fields: testFields,
			args: args{
				p: geometry.Point{X: 0, Y: 1},
			},
			want: true,
		},
		{
			name:   "returns false if point is inside polygon",
			fields: testFields,
			args: args{
				p: geometry.Point{X: 4, Y: 7},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			polygon := geometry.Polygon{
				Name:        tt.fields.Name,
				Description: tt.fields.Description,
				AlphaShape:  tt.fields.AlphaShape,
				Triangles:   tt.fields.Triangles,
				BoundingBox: tt.fields.BoundingBox,
			}
			if got := geometry.PointInsidePolygon(&tt.args.p, &polygon); got != tt.want {
				t.Errorf("geometry.Polygon.PointInside() = %v, want %v", got, tt.want)
			}
		})
	}
}

func BenchmarkPointInside(b *testing.B) {
	geoPolygon, _ := NewTriangulatedPolygonFromPoints([]geometry.Point{
		geometry.Point{X: 0, Y: 0},
		geometry.Point{X: 10, Y: 0},
		geometry.Point{X: 10, Y: 10},
		geometry.Point{X: 5, Y: 10},
		geometry.Point{X: 5, Y: 5},
		geometry.Point{X: 0, Y: 5},
	})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		geoPolygon.PointInside(&geometry.Point{X: 4, Y: 7})
	}
}
