package geometry

import (
	"math/rand"
	"testing"
)

func Test_lineIntersects(t *testing.T) {
	type args struct {
		line1 []Point
		line2 []Point
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "Intersecting lines",
			args: args{
				line1: []Point{Point{0, 0}, Point{1, 1}},
				line2: []Point{Point{1, 0}, Point{0, 1}},
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "Intersecting lines reversed",
			args: args{
				line1: []Point{Point{1, 0}, Point{0, 1}},
				line2: []Point{Point{0, 0}, Point{1, 1}},
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "Coincident/overlapping lines",
			args: args{
				line1: []Point{Point{-1, -1}, Point{2, 2}},
				line2: []Point{Point{0, 0}, Point{1, 1}},
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "Coincident/overlapping lines reversed",
			args: args{
				line1: []Point{Point{0, 0}, Point{1, 1}},
				line2: []Point{Point{-1, -1}, Point{2, 2}},
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "Parallell lines",
			args: args{
				line1: []Point{Point{1, 1}, Point{2, 1}},
				line2: []Point{Point{0, 0}, Point{2, 0}},
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "Parallell lines reversed",
			args: args{
				line1: []Point{Point{0, 0}, Point{2, 0}},
				line2: []Point{Point{1, 1}, Point{2, 1}},
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "Wrong parameters line 1",
			args: args{
				line1: []Point{},
				line2: []Point{Point{0, 0}, Point{1, 1}},
			},
			want:    false,
			wantErr: true,
		},
		{
			name: "Wrong parameters line 2",
			args: args{
				line1: []Point{Point{0, 0}, Point{1, 1}},
				line2: []Point{},
			},
			want:    false,
			wantErr: true,
		},
		{
			name: "Non-touching lines outside bounding box",
			args: args{
				line1: []Point{Point{10.418364, 63.42602}, Point{32.078053, 63.42602}},
				line2: []Point{Point{6.521944, 61.852493}, Point{6.530277, 61.862778}},
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "Non-touching lines reversed outside bounding box",
			args: args{
				line1: []Point{Point{6.521944, 61.852493}, Point{6.530277, 61.862778}},
				line2: []Point{Point{10.418364, 63.42602}, Point{32.078053, 63.42602}},
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "Non-touching within bounding box",
			args: args{
				line1: []Point{Point{0, 0}, Point{4, 6}},
				line2: []Point{Point{3, 2}, Point{7, 0}},
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "Non-touching within bounding box reversed",
			args: args{
				line1: []Point{Point{3, 2}, Point{7, 0}},
				line2: []Point{Point{0, 0}, Point{4, 6}},
			},
			want:    false,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := LineIntersects(tt.args.line1, tt.args.line2)
			if (err != nil) != tt.wantErr {
				t.Errorf("lineIntersects() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("lineIntersects() = %v, want %v", got, tt.want)
			}
		})
	}
}

func BenchmarkLineIntersects(b *testing.B) {
	randomPoints := make([][]Point, 1000)

	for i := range randomPoints {
		randomPoints[i] = []Point{
			Point{rand.Float64(), rand.Float64()},
			Point{rand.Float64(), rand.Float64()},
			Point{rand.Float64(), rand.Float64()},
			Point{rand.Float64(), rand.Float64()},
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, randomPointsXYXY := range randomPoints {
			_, _ = LineIntersects(
				[]Point{randomPointsXYXY[0], randomPointsXYXY[1]},
				[]Point{randomPointsXYXY[2], randomPointsXYXY[3]},
			)
		}
	}
}
