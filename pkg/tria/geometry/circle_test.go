package geometry

import (
	"testing"
)

func TestCircle_PolygonInside(t *testing.T) {
	simpleCircle := Circle{
		Radius: 10,
		Origo:  Point{X: 0, Y: 0},
	}

	polygonInsideCircle := NewPolygonFromPoints([]Point{
		Point{X: 0, Y: 0},
		Point{X: 1, Y: 0},
		Point{X: 1, Y: 1},
	})
	polygonOverlappingCircle := NewPolygonFromPoints([]Point{
		Point{X: 0, Y: 0},
		Point{X: 100, Y: 0},
		Point{X: 100, Y: 1},
	})
	polygonOutsideCircle := NewPolygonFromPoints([]Point{
		Point{X: 100, Y: 0},
		Point{X: 101, Y: 0},
		Point{X: 101, Y: 1},
	})

	type args struct {
		polygon *Polygon
	}
	tests := []struct {
		name   string
		fields Circle
		args   args
		want   bool
	}{
		{
			name:   "polygon inside of circle",
			fields: simpleCircle,
			args: args{
				polygon: &polygonInsideCircle,
			},
			want: true,
		},
		{
			name:   "polygon overlapping circle",
			fields: simpleCircle,
			args: args{
				polygon: &polygonOverlappingCircle,
			},
			want: false,
		},
		{
			name:   "polygon outside circle",
			fields: simpleCircle,
			args: args{
				polygon: &polygonOutsideCircle,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			circle := tt.fields
			if got := circle.PolygonInside(tt.args.polygon); got != tt.want {
				t.Errorf("Circle.PolygonInside() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCircle_CircleInside(t *testing.T) {
	tinyCircle := Circle{
		Radius: 1,
		Origo:  Point{X: 0, Y: 0},
	}
	simpleCircle := Circle{
		Radius: 10,
		Origo:  Point{X: 0, Y: 0},
	}
	overlappingCircle := Circle{
		Radius: 2,
		Origo:  Point{X: 10, Y: 0},
	}

	type args struct {
		circleCandidate *Circle
	}
	tests := []struct {
		name   string
		fields Circle
		args   args
		want   bool
	}{
		{
			name:   "circle inside circle",
			fields: simpleCircle,
			args: args{
				circleCandidate: &tinyCircle,
			},
			want: true,
		},
		{
			name:   "circle inside itself",
			fields: simpleCircle,
			args: args{
				circleCandidate: &simpleCircle,
			},
			want: true,
		},
		{
			name:   "circle overlapping circle completely",
			fields: tinyCircle,
			args: args{
				circleCandidate: &simpleCircle,
			},
			want: false,
		},
		{
			name:   "circle overlapping circle partially",
			fields: simpleCircle,
			args: args{
				circleCandidate: &overlappingCircle,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			circle := &Circle{
				Name:   tt.fields.Name,
				Origo:  tt.fields.Origo,
				Radius: tt.fields.Radius,
			}
			if got := circle.CircleInside(tt.args.circleCandidate); got != tt.want {
				t.Errorf("Circle.CircleInside() = %v, want %v", got, tt.want)
			}
		})
	}
}
