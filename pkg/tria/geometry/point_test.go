package geometry

import (
	"testing"
)

func TestPoint_DistanceTo(t *testing.T) {
	type fields struct {
		X float64
		Y float64
	}
	type args struct {
		p2 *Point
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   float64
	}{
		{
			name: "Should show correct distance to for a simple horizontal line",
			fields: fields{
				X: 0,
				Y: 0,
			},
			args: args{
				&Point{
					X: 4,
					Y: 0,
				},
			},
			want: 4.0,
		},
		{
			name: "Should show correct distance to for a simple vertical line",
			fields: fields{
				X: 0,
				Y: 0,
			},
			args: args{
				&Point{
					X: 0,
					Y: 4,
				},
			},
			want: 4.0,
		},
		{
			name: "Should show correct distance to for a simple (reverse) vertical line",
			fields: fields{
				X: 4,
				Y: 0,
			},
			args: args{
				&Point{
					X: 0,
					Y: 0,
				},
			},
			want: 4.0,
		},
		{
			name: "Should show correct distance to for a simple AB line",
			fields: fields{
				X: 0,
				Y: 0,
			},
			args: args{
				&Point{
					X: 4,
					Y: 3,
				},
			},
			want: 5.0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p1 := &Point{
				X: tt.fields.X,
				Y: tt.fields.Y,
			}
			if got := p1.DistanceTo(tt.args.p2); got != tt.want {
				t.Errorf("Point.DistanceTo() = %v, want %v", got, tt.want)
			}
		})
	}
}

func BenchmarkPoint_DistanceTo(b *testing.B) {
	point := Point{X: 2, Y: 2}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		point.DistanceTo(&Point{X: 0, Y: 0})
	}
}

func TestPoint_DistanceToLine(t *testing.T) {
	type fields struct {
		X float64
		Y float64
	}
	type args struct {
		a *Point
		b *Point
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   float64
	}{
		{
			name: "Should give correct distance on a simple horizontal line",
			fields: fields{
				X: 2,
				Y: 2,
			},
			args: args{
				&Point{X: 0, Y: 0},
				&Point{X: 4, Y: 0},
			},
			want: 2.0,
		},
		{
			name: "Should give correct distance on a simple vertical line",
			fields: fields{
				X: 2,
				Y: 2,
			},
			args: args{
				&Point{X: 0, Y: 0},
				&Point{X: 0, Y: 4},
			},
			want: 2.0,
		},
		{
			name: "Should give correct distance where point is out of bounds of horizontal line",
			fields: fields{
				X: -2,
				Y: 0,
			},
			args: args{
				&Point{X: 0, Y: 0},
				&Point{X: 4, Y: 0},
			},
			want: 2.0,
		},
		{
			name: "Should give correct distance where point is out of bounds of horizontal line",
			fields: fields{
				X: 0,
				Y: 6,
			},
			args: args{
				&Point{X: 0, Y: 4},
				&Point{X: 0, Y: 0},
			},
			want: 2.0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p1 := &Point{
				X: tt.fields.X,
				Y: tt.fields.Y,
			}
			if got := p1.DistanceToLine(tt.args.a, tt.args.b); got != tt.want {
				t.Errorf("Point.DistanceToLine() = %v, want %v", got, tt.want)
			}
		})
	}
}

func BenchmarkPoint_DistanceToLine(b *testing.B) {
	point := Point{X: 2, Y: 2}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Simple
		point.DistanceToLine(&Point{X: 0, Y: 0}, &Point{X: 4, Y: 0})
		// Out of bounds
		point.DistanceToLine(&Point{X: 4, Y: 2}, &Point{X: 6, Y: 2})
	}
}
