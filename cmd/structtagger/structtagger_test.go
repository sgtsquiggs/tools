package main

import (
	"reflect"
	"testing"
)

func Test_tagValueGetter(t *testing.T) {
	type args struct {
		tag  reflect.StructTag
		name string
	}
	tests := []struct {
		name   string
		args   args
		want   string
		wantOk bool
	}{
		{
			name: "json",
			args: args{
				tag:  `protobuf:"bytes,1,opt,name=contentLeft,proto3" json:"content_left,omitempty"`,
				name: "json",
			},
			want:   "content_left",
			wantOk: true,
		},
		{
			name: "json",
			args: args{
				tag:  `json:"content_left"`,
				name: "json",
			},
			want:   "content_left",
			wantOk: true,
		},
		{
			name: "protobuf",
			args: args{
				tag:  `protobuf:"bytes,1,opt,name=contentLeft,proto3" json:"content_left,omitempty"`,
				name: "protobuf",
			},
			want:   "contentLeft",
			wantOk: true,
		},
		{
			name: "bson",
			args: args{
				tag:  `json:"id,omitempty" bson:"_id"`,
				name: "bson",
			},
			want:   "_id",
			wantOk: true,
		},
		{
			name: "undefined tag",
			args: args{
				tag:  `json:"id,omitempty" bson:"_id"`,
				name: "gson",
			},
			want:   "",
			wantOk: false,
		},
		{
			name: "defined yet unknown tag",
			args: args{
				tag:  `cooltag:"the coolness"`,
				name: "cooltag",
			},
			want:   "",
			wantOk: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := tagValueGetter(tt.args.tag, tt.args.name)
			if got != tt.want {
				t.Errorf("tagValueGetter() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.wantOk {
				t.Errorf("tagValueGetter() got1 = %v, want %v", got1, tt.wantOk)
			}
		})
	}
}
