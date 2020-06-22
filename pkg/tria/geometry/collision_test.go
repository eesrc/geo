package geometry

import "testing"

func TestPointInsideTriangle(t *testing.T) {
	type args struct {
		p  Point
		ta Point
		tb Point
		tc Point
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{name: "should return true when same point as one of triangle edges", args: args{
			p:  Point{X: 0, Y: 0},
			ta: Point{X: 0, Y: 0},
			tb: Point{X: 1, Y: 0},
			tc: Point{X: 1, Y: 1},
		}, want: true},
		{name: "should return true on edge of vertex", args: args{
			p:  Point{X: .5, Y: .5},
			ta: Point{X: 0, Y: 0},
			tb: Point{X: 1, Y: 0},
			tc: Point{X: 1, Y: 1},
		}, want: true},
		{name: "should return true when inside triangle", args: args{
			p:  Point{X: .75, Y: .25},
			ta: Point{X: 0, Y: 0},
			tb: Point{X: 1, Y: 0},
			tc: Point{X: 1, Y: 1},
		}, want: true},
		{name: "should return false when point is outside of x", args: args{
			p:  Point{X: -1, Y: 0},
			ta: Point{X: 0, Y: 0},
			tb: Point{X: 1, Y: 0},
			tc: Point{X: 1, Y: 1},
		}, want: false},
		{name: "should return false when point is outside of y", args: args{
			p:  Point{X: 0, Y: 1},
			ta: Point{X: 0, Y: 0},
			tb: Point{X: 1, Y: 0},
			tc: Point{X: 1, Y: 1},
		}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := PointInsideTriangle(&tt.args.p, &tt.args.ta, &tt.args.tb, &tt.args.tc); got != tt.want {
				t.Errorf("PointInsideTriangle() = %v, want %v", got, tt.want)
			}
		})
	}
}
