package geometry

import (
	"reflect"
	"testing"
)

func TestBoundingBox_ContainsPoint(t *testing.T) {
	type fields struct {
		MaxX float64
		MaxY float64
		MinX float64
		MinY float64
	}
	type args struct {
		p *Point
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "point inside bounding box",
			fields: fields{
				MaxX: 10,
				MaxY: 10,
				MinX: 0,
				MinY: 0,
			},
			args: args{
				p: &Point{X: 5, Y: 5},
			},
			want: true,
		},
		{
			name: "point outside bounding box",
			fields: fields{
				MaxX: 10,
				MaxY: 10,
				MinX: 0,
				MinY: 0,
			},
			args: args{
				p: &Point{X: 11, Y: 11},
			},
			want: false,
		},
		{
			name: "point on side of bounding box",
			fields: fields{
				MaxX: 10,
				MaxY: 10,
				MinX: 0,
				MinY: 0,
			},
			args: args{
				p: &Point{X: 10, Y: 10},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			boundingBox := &BoundingBox{
				MaxX: tt.fields.MaxX,
				MaxY: tt.fields.MaxY,
				MinX: tt.fields.MinX,
				MinY: tt.fields.MinY,
			}
			if got := boundingBox.ContainsPoint(tt.args.p); got != tt.want {
				t.Errorf("BoundingBox.ContainsPoint() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_calculateBoundingBox(t *testing.T) {
	type args struct {
		points []Point
	}
	tests := []struct {
		name string
		args args
		want BoundingBox
	}{
		{
			name: "Should set correct BoundingBox based on simple polygon",
			args: args{
				[]Point{
					Point{X: 0, Y: 0},
					Point{X: 10, Y: 0},
					Point{X: 10, Y: 10},
					Point{X: 0, Y: 10},
				},
			},
			want: BoundingBox{
				MinX: 0,
				MaxX: 10,
				MinY: 0,
				MaxY: 10,
			},
		},
		{
			name: "Should set correct BoundingBox based on polygon",
			args: args{
				[]Point{
					Point{X: 12.89794921875, Y: 65.74416997811738},
					Point{X: 11.612548828125, Y: 65.36683689226321},
					Point{X: 12.7001953125, Y: 64.64270382119375},
					Point{X: 15.00732421875, Y: 64.84426759958093},
					Point{X: 14.809570312499998, Y: 65.58117863257927},
					Point{X: 13.55712890625, Y: 65.22910188319217},
					Point{X: 12.89794921875, Y: 65.74416997811738},
				},
			},
			want: BoundingBox{
				MinX: 11.612548828125,
				MaxX: 15.00732421875,
				MinY: 64.64270382119375,
				MaxY: 65.74416997811738,
			},
		},
		{
			name: "Should set correct BoundingBox based on line",
			args: args{
				[]Point{
					Point{X: 12.89794921875, Y: 65.74416997811738},
					Point{X: 11.612548828125, Y: 65.36683689226321},
				},
			},
			want: BoundingBox{
				MinX: 11.612548828125,
				MaxX: 12.89794921875,
				MinY: 65.36683689226321,
				MaxY: 65.74416997811738,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CalculateBoundingBox(tt.args.points); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("calculateBoundingBox() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBoundingBox_BoundingBoxIntersects(t *testing.T) {
	type fields struct {
		MaxX float64
		MaxY float64
		MinX float64
		MinY float64
	}
	type args struct {
		intersectingBoundingBox *BoundingBox
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "return true if bounding boxes intersect",
			fields: fields{
				MaxX: 10,
				MaxY: 10,
				MinX: 0,
				MinY: 0,
			},
			args: args{
				intersectingBoundingBox: &BoundingBox{
					MaxX: 15,
					MaxY: 15,
					MinX: 5,
					MinY: 5,
				},
			},
			want: true,
		},
		{
			name: "return true if bounding boxes intersect",
			fields: fields{
				MaxX: 10,
				MaxY: 10,
				MinX: 0,
				MinY: 0,
			},
			args: args{
				intersectingBoundingBox: &BoundingBox{
					MaxX: 15,
					MaxY: 5,
					MinX: 5,
					MinY: -5,
				},
			},
			want: true,
		},
		{
			name: "return true if bounding boxes intersect",
			fields: fields{
				MaxX: 10,
				MaxY: 10,
				MinX: 0,
				MinY: 0,
			},
			args: args{
				intersectingBoundingBox: &BoundingBox{
					MaxX: 5,
					MaxY: 5,
					MinX: -5,
					MinY: -5,
				},
			},
			want: true,
		},
		{
			name: "return true if bounding boxes intersect",
			fields: fields{
				MaxX: 10,
				MaxY: 10,
				MinX: 0,
				MinY: 0,
			},
			args: args{
				intersectingBoundingBox: &BoundingBox{
					MaxX: 5,
					MaxY: 15,
					MinX: -5,
					MinY: 5,
				},
			},
			want: true,
		},
		{
			name: "return false if bounding boxes doesn't intersect",
			fields: fields{
				MaxX: 10,
				MaxY: 10,
				MinX: 0,
				MinY: 0,
			},
			args: args{
				intersectingBoundingBox: &BoundingBox{
					MaxX: 15,
					MaxY: 15,
					MinX: 11,
					MinY: 11,
				},
			},
			want: false,
		},
		{
			name: "return true if bounding boxes share boundary",
			fields: fields{
				MaxX: 10,
				MaxY: 10,
				MinX: 0,
				MinY: 0,
			},
			args: args{
				intersectingBoundingBox: &BoundingBox{
					MaxX: 15,
					MaxY: 15,
					MinX: 10,
					MinY: 10,
				},
			},
			want: true,
		},
		{
			name: "return true if bounding boxes overlap",
			fields: fields{
				MaxX: 10,
				MaxY: 0,
				MinX: 0,
				MinY: 0,
			},
			args: args{
				intersectingBoundingBox: &BoundingBox{
					MaxX: 5,
					MaxY: 5,
					MinX: 5,
					MinY: -5,
				},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			boundingBox := &BoundingBox{
				MaxX: tt.fields.MaxX,
				MaxY: tt.fields.MaxY,
				MinX: tt.fields.MinX,
				MinY: tt.fields.MinY,
			}
			if got := boundingBox.BoundingBoxIntersects(tt.args.intersectingBoundingBox); got != tt.want {
				t.Errorf("BoundingBox.BoundingBoxIntersects() = %v, want %v", got, tt.want)
			}
		})
	}
}
