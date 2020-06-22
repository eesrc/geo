package geometry

import (
	"math/rand"
	"testing"
)

func TestIsClockwise(t *testing.T) {
	type args struct {
		poly []Point
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{name: "Should return true upon clockwise polygon", args: args{poly: []Point{
			Point{X: 0, Y: 0},
			Point{X: 0, Y: 1},
			Point{X: 1, Y: 1},
			Point{X: 1, Y: 0},
		}}, want: true,
		},
		{name: "Should return true upon clockwise triangle", args: args{poly: []Point{
			Point{X: 0, Y: 0},
			Point{X: 0, Y: 1},
			Point{X: 1, Y: 1},
		}}, want: true,
		},
		{name: "Should return false upon non-clockwise polygon", args: args{poly: []Point{
			Point{X: 0, Y: 0},
			Point{X: 1, Y: 0},
			Point{X: 1, Y: 1},
			Point{X: 0, Y: 1},
		}}, want: false,
		},
		{name: "Should return false upon non-clockwise triangle", args: args{poly: []Point{
			Point{X: 0, Y: 0},
			Point{X: 1, Y: 0},
			Point{X: 1, Y: 1},
		}}, want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsClockwise(tt.args.poly); got != tt.want {
				t.Errorf("IsClockwise() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_polygonArea(t *testing.T) {
	type args struct {
		data []Point
	}
	tests := []struct {
		name string
		args args
		want float64
	}{
		{
			name: "Simple triangle",
			args: args{
				[]Point{
					Point{0, 0},
					Point{0, 1},
					Point{1, 1},
				},
			}, want: 0.5,
		},
		{
			name: "Simple square",
			args: args{
				[]Point{
					Point{0, 0},
					Point{1, 0},
					Point{1, 1},
					Point{0, 1},
				},
			}, want: 1,
		},
		{
			name: "Random rectangle",
			args: args{
				[]Point{
					Point{2, 2},
					Point{11, 2},
					Point{9, 7},
					Point{4, 10},
				},
			}, want: 45.5,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := PolygonArea(tt.args.data); got != tt.want {
				t.Errorf("polygonArea() = %v, want %v", got, tt.want)
			}
		})
	}
}

func BenchmarkPointInsideTriangle(b *testing.B) {
	randomPoints := make([][]float64, 1000)

	for i := range randomPoints {
		randomPoints[i] = []float64{rand.Float64(), rand.Float64()}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, randomXY := range randomPoints {
			PointInsideTriangle(
				&Point{X: randomXY[0], Y: randomXY[1]},
				&Point{X: 0, Y: 0},
				&Point{X: 1, Y: 0},
				&Point{X: 1, Y: 1},
			)
		}
	}
}
