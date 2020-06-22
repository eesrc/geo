package event

import (
	"reflect"
	"testing"

	"github.com/eesrc/geo/pkg/model"
)

func TestDecodeEvent(t *testing.T) {
	parsedByte := []byte("aHR0cDovL2xvcmVtcGl4ZWwuY29tLzY0MC80ODAvY2F0cw==")

	type args struct {
		data interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    interface{}
		wantErr bool
	}{
		{
			name: "subscription message",
			args: args{
				data: []byte(`{"type":"subscription","data":{"subscriptionId":10,"position":{"ID":160787,"TrackerID":1,"Timestamp":1578681199053,"Lat":63.37183226679281,"Lon":10.3985595703125,"Alt":0,"Heading":0,"Speed":0,"Payload":"aHR0cDovL2xvcmVtcGl4ZWwuY29tLzY0MC80ODAvY2F0cw==","Precision":1},"details":{"movements":["inside"],"shapecollectionId":1,"shapeId":105}}}`),
			},
			want: SubscriptionEvent{
				Type: Subscription,
				Data: SubscriptionEventDetails{
					SubscriptionID: 10,
					Position: model.Position{
						ID:        160787,
						TrackerID: 1,
						Timestamp: 1578681199053,
						Lat:       63.37183226679281,
						Lon:       10.3985595703125,
						Alt:       0,
						Heading:   0,
						Speed:     0,
						Payload:   parsedByte,
						Precision: 1,
					},
					Details: TriggerDetails{
						Movements:         []string{"inside"},
						ShapecollectionID: 1,
						ShapeID:           105,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "position message",
			args: args{
				data: []byte(`{"type":"position","data":{"collectionId":1,"trackerId":1,"position":{"ID":160788,"TrackerID":1,"Timestamp":1578681199053,"Lat":63.37183226679281,"Lon":10.3985595703125,"Alt":0,"Heading":0,"Speed":0,"Payload":"aHR0cDovL2xvcmVtcGl4ZWwuY29tLzY0MC80ODAvY2F0cw==","Precision":1}}}`),
			},
			want: PositionEvent{
				Type: Position,
				Data: PositionEventDetails{
					CollectionID: 1,
					TrackerID:    1,
					Position: model.Position{
						ID:        160788,
						TrackerID: 1,
						Timestamp: 1578681199053,
						Lat:       63.37183226679281,
						Lon:       10.3985595703125,
						Alt:       0,
						Heading:   0,
						Speed:     0,
						Payload:   parsedByte,
						Precision: 1,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "lifecycle message",
			args: args{
				data: []byte(`{"type":"lifecycle","data":{"type":"updated","entityType":"collection","entityId":2}}`),
			},
			want: LifeCycleEvent{
				Type: LifeCycle,
				Data: LifecycleEventDetails{
					Type:       UpdatedEvent,
					EntityType: CollectionEntity,
					EntityID:   2,
				},
			},
			wantErr: false,
		},
		{
			name: "unknown type message",
			args: args{
				data: []byte(`{"type":"unknown","data":{"foo":"bar"}}`),
			},
			want: map[string]interface{}{
				"type": "unknown",
				"data": map[string]interface{}{
					"foo": "bar",
				},
			},
			wantErr: true,
		},
		{
			name: "no type in struct message",
			args: args{
				data: []byte(`{"noType":"missing"}`),
			},
			want: map[string]interface{}{
				"noType": "missing",
			},
			wantErr: true,
		},
		{
			name: "unmarshal error",
			args: args{
				data: []byte(`{noType":"missing"`),
			},
			want:    map[string]interface{}{},
			wantErr: true,
		},
		{
			name: "unmarshal error when type doesn't match",
			args: args{
				data: []byte(`{"type":{"foo":"bar"}}`),
			},
			want: map[string]interface{}{
				"type": map[string]interface{}{
					"foo": "bar",
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DecodeEvent(tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("DecodeEvent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DecodeEvent() = \n%#v\nwant\n%#v", got, tt.want)
			}
		})
	}
}
