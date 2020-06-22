package geometry

import (
	"reflect"
	"testing"
)

func Test_normalizePoints(t *testing.T) {
	type args struct {
		points []Point
	}
	tests := []struct {
		name string
		args args
		want []Point
	}{
		{
			name: "Should return equals if points are less than 3",
			args: args{
				points: []Point{
					Point{X: 0, Y: 0},
					Point{X: 1, Y: 0},
				},
			},
			want: []Point{
				Point{X: 0, Y: 0},
				Point{X: 1, Y: 0},
			},
		},
		{
			name: "Should remove last element if equals to first",
			args: args{
				points: []Point{
					Point{X: 0, Y: 0},
					Point{X: 1, Y: 0},
					Point{X: 1, Y: 1},
					Point{X: 0, Y: 0},
				},
			},
			want: []Point{
				Point{X: 0, Y: 0},
				Point{X: 1, Y: 0},
				Point{X: 1, Y: 1},
			},
		},
		{
			name: "Should remove uneccesary points on a line of polygon",
			args: args{
				points: []Point{
					Point{X: 0, Y: 0},
					Point{X: 1, Y: 0},
					Point{X: 2, Y: 0},
					Point{X: 2, Y: 1},
					Point{X: 2, Y: 2},
					Point{X: 2, Y: 3},
				},
			},
			want: []Point{
				Point{X: 0, Y: 0},
				Point{X: 2, Y: 0},
				Point{X: 2, Y: 3},
			},
		},
		{
			name: "Should reverse polygon list if it's clockwise",
			args: args{
				points: []Point{
					Point{X: 0, Y: 0},
					Point{X: 0, Y: 1},
					Point{X: 1, Y: 1},
				},
			},
			want: []Point{
				Point{X: 1, Y: 1},
				Point{X: 0, Y: 1},
				Point{X: 0, Y: 0},
			},
		},
		{
			name: "Should correctly remove following duplicates in polygon",
			args: args{
				points: []Point{
					Point{X: 0, Y: 0},
					Point{X: 0, Y: 1},
					Point{X: 0, Y: 1},
					Point{X: 0, Y: 1},
					Point{X: 0, Y: 1},
					Point{X: 0, Y: 1},
					Point{X: 0, Y: 1},
					Point{X: 1, Y: 1},
				},
			},
			want: []Point{
				Point{X: 1, Y: 1},
				Point{X: 0, Y: 1},
				Point{X: 0, Y: 0},
			},
		},
		{
			name: "Should correctly remove following duplicates in polygon",
			args: args{
				points: []Point{
					Point{X: 0, Y: 0},
					Point{X: 0, Y: 1},
					Point{X: 0, Y: 1},
					Point{X: 1, Y: 1},
				},
			},
			want: []Point{
				Point{X: 1, Y: 1},
				Point{X: 0, Y: 1},
				Point{X: 0, Y: 0},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NormalizePoints(tt.args.points); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("normalizePoints() = %v, want %v", got, tt.want)
			}
		})
	}
}
