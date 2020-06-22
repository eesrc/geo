package geometry

import (
	"testing"
)

func TestPolygon_ShapeInside(t *testing.T) {
	type args struct {
		shape Shape
	}

	polygonSquare10by10 := NewPolygonFromPoints([]Point{
		Point{0, 0},
		Point{10, 0},
		Point{10, 10},
		Point{0, 10},
	})
	polygonSquare10by10.Triangles = [][]Point{
		[]Point{
			Point{0, 0},
			Point{10, 0},
			Point{0, 10},
		},
		[]Point{
			Point{10, 0},
			Point{0, 10},
			Point{10, 10},
		},
	}

	polygonTriangle := NewPolygonFromPoints([]Point{
		Point{0, 0},
		Point{10, 0},
		Point{0, 10},
	})
	polygonTriangle.Triangles = [][]Point{polygonTriangle.AlphaShape}

	circle1RadiusBottomLeft := Circle{
		Radius: 1,
		Origo:  Point{X: 1, Y: 1},
	}
	circle1RadiusInCenter := Circle{
		Radius: 1,
		Origo:  Point{X: 5, Y: 5},
	}
	circle1RadiusToCloseToEdge := Circle{
		Radius: 1,
		Origo:  Point{X: 5, Y: 9.1},
	}
	circle1RadiusOutsideSquare := Circle{
		Radius: 1,
		Origo:  Point{X: 5, Y: 15},
	}
	circle1RadiusTouchingOuterEdge := Circle{
		Radius: 1,
		Origo:  Point{X: 5, Y: 11},
	}
	circle100RadiusEncompassing := Circle{
		Radius: 100,
		Origo:  Point{X: 0, Y: 0},
	}

	tests := []struct {
		name   string
		fields Polygon
		args   args
		want   bool
	}{
		{
			name:   "should return true when circle is within triangle",
			fields: polygonTriangle,
			args: args{
				shape: &circle1RadiusBottomLeft,
			},
			want: true,
		},
		{
			name:   "should return false when circle is on triangle vertice",
			fields: polygonTriangle,
			args: args{
				shape: &circle1RadiusInCenter,
			},
			want: false,
		},
		{
			name:   "should return true when circle is within polygon",
			fields: polygonSquare10by10,
			args: args{
				shape: &circle1RadiusInCenter,
			},
			want: true,
		},
		{
			name:   "should return false when circle is too close to edge of square",
			fields: polygonSquare10by10,
			args: args{
				shape: &circle1RadiusToCloseToEdge,
			},
			want: false,
		},
		{
			name:   "should return false when circle is outside square",
			fields: polygonSquare10by10,
			args: args{
				shape: &circle1RadiusOutsideSquare,
			},
			want: false,
		},
		{
			name:   "should return false when circle is outside square, yet touching outer edge",
			fields: polygonSquare10by10,
			args: args{
				shape: &circle1RadiusTouchingOuterEdge,
			},
			want: false,
		},
		{
			name:   "should return false when circle is encompassing square",
			fields: polygonSquare10by10,
			args: args{
				shape: &circle100RadiusEncompassing,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			polygon := &Polygon{
				ID:          tt.fields.ID,
				Name:        tt.fields.Name,
				Description: tt.fields.Description,
				AlphaShape:  tt.fields.AlphaShape,
				Triangles:   tt.fields.Triangles,
				BoundingBox: tt.fields.BoundingBox,
			}
			if got := polygon.ShapeInside(tt.args.shape); got != tt.want {
				t.Errorf("Polygon.ShapeInside() = %v, want %v", got, tt.want)
			}
		})
	}
}
