package triangulation

import (
	"reflect"
	"testing"

	"github.com/eesrc/geo/pkg/tria/geometry"
)

func TestTriangulateByEarCut(t *testing.T) {
	type args struct {
		polygon []geometry.Point
	}
	tests := []struct {
		name    string
		args    args
		want    [][]geometry.Point
		wantErr bool
	}{
		{
			name: "Should triangulate a simple square", args: args{[]geometry.Point{
				geometry.Point{X: 0, Y: 0},
				geometry.Point{X: 1, Y: 0},
				geometry.Point{X: 1, Y: 1},
				geometry.Point{X: 0, Y: 1},
			}},
			want: [][]geometry.Point{
				[]geometry.Point{geometry.Point{X: 0, Y: 1}, geometry.Point{X: 0, Y: 0}, geometry.Point{X: 1, Y: 0}},
				[]geometry.Point{geometry.Point{X: 1, Y: 0}, geometry.Point{X: 1, Y: 1}, geometry.Point{X: 0, Y: 1}},
			},
			wantErr: false,
		},
		{
			name: "Should triangulate a simple polygon", args: args{[]geometry.Point{
				geometry.Point{X: 0, Y: 0},
				geometry.Point{X: 1, Y: 0},
				geometry.Point{X: 1, Y: 1},
				geometry.Point{X: 2, Y: 2},
				geometry.Point{X: 0, Y: 1},
			}},
			want: [][]geometry.Point{
				[]geometry.Point{geometry.Point{X: 0, Y: 1}, geometry.Point{X: 0, Y: 0}, geometry.Point{X: 1, Y: 0}},
				[]geometry.Point{geometry.Point{X: 1, Y: 1}, geometry.Point{X: 2, Y: 2}, geometry.Point{X: 0, Y: 1}},
				[]geometry.Point{geometry.Point{X: 0, Y: 1}, geometry.Point{X: 1, Y: 0}, geometry.Point{X: 1, Y: 1}},
			},
			wantErr: false,
		},
		{
			name: "Should triangulate a concave polygon", args: args{[]geometry.Point{
				geometry.Point{X: 0, Y: 0},
				geometry.Point{X: 1, Y: 0},
				geometry.Point{X: .25, Y: 1},
				geometry.Point{X: 3, Y: 2},
				geometry.Point{X: 4, Y: 1},
				geometry.Point{X: 3, Y: 3},
				geometry.Point{X: 0, Y: 2},
			}}, want: [][]geometry.Point{
				[]geometry.Point{geometry.Point{X: 0, Y: 0}, geometry.Point{X: 1, Y: 0}, geometry.Point{X: .25, Y: 1}},
				[]geometry.Point{geometry.Point{X: 3, Y: 2}, geometry.Point{X: 4, Y: 1}, geometry.Point{X: 3, Y: 3}},
				[]geometry.Point{geometry.Point{X: 0, Y: 2}, geometry.Point{X: 0, Y: 0}, geometry.Point{X: .25, Y: 1}},
				[]geometry.Point{geometry.Point{X: .25, Y: 1}, geometry.Point{X: 3, Y: 2}, geometry.Point{X: 3, Y: 3}},
				[]geometry.Point{geometry.Point{X: 3, Y: 3}, geometry.Point{X: 0, Y: 2}, geometry.Point{X: .25, Y: 1}},
			}, wantErr: false,
		},
		{
			name: "Should triangulate a polygon", args: args{[]geometry.Point{
				geometry.Point{X: 0, Y: 0},
				geometry.Point{X: 1, Y: 0},
				geometry.Point{X: 1, Y: 1},
				geometry.Point{X: 0, Y: 1},
				geometry.Point{X: -1, Y: 1},
				geometry.Point{X: -14, Y: 2},
				geometry.Point{X: -16, Y: -2},
			}}, want: [][]geometry.Point{
				[]geometry.Point{geometry.Point{X: 0, Y: 0}, geometry.Point{X: 1, Y: 0}, geometry.Point{X: 1, Y: 1}},
				[]geometry.Point{geometry.Point{X: -1, Y: 1}, geometry.Point{X: -14, Y: 2}, geometry.Point{X: -16, Y: -2}},
				[]geometry.Point{geometry.Point{X: -16, Y: -2}, geometry.Point{X: 0, Y: 0}, geometry.Point{X: 1, Y: 1}},
				[]geometry.Point{geometry.Point{X: 1, Y: 1}, geometry.Point{X: -1, Y: 1}, geometry.Point{X: -16, Y: -2}},
			}, wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := TriangulateByEarCut(tt.args.polygon)
			if (err != nil) != tt.wantErr {
				t.Errorf("TriangulateByEarCut() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TriangulateByEarCut() = %v, want %v", got, tt.want)
			}
		})
	}
}

func BenchmarkTriangulateByEarCut(b *testing.B) {
	poly := []geometry.Point{
		geometry.Point{X: 0, Y: 0},
		geometry.Point{X: 1, Y: 0},
		geometry.Point{X: 1, Y: 1},
		geometry.Point{X: 0, Y: 1},
		geometry.Point{X: -1, Y: 1},
		geometry.Point{X: -14, Y: 2},
		geometry.Point{X: -16, Y: -2},
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = TriangulateByEarCut(poly)
	}
}
