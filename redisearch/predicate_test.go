package redisearch

import (
	"reflect"
	"testing"
)

func TestInRange(t *testing.T) {
	type args struct {
		property     string
		min          interface{}
		max          interface{}
		minInclusive bool
		maxInclusive bool
	}
	tests := []struct {
		name string
		args args
		want Predicate
	}{
		{
			name: "",
			args: args{
				property:     "start",
				min:          0,
				max:          1,
				minInclusive: false,
				maxInclusive: true,
			},
			want: Predicate{
				Property:     "start",
				min:          0,
				max:          1,
				minInclusive: false,
				maxInclusive: true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := InRange(tt.args.property, tt.args.min, tt.args.max, tt.args.minInclusive, tt.args.maxInclusive); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("InRange() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEquals(t *testing.T) {
	type args struct {
		property string
		value    interface{}
	}
	tests := []struct {
		name string
		args args
		want Predicate
	}{
		{
			name: "equal",
			args: args{
				property: "start",
				value:    1,
			},
			want: Predicate{
				Property:     "start",
				min:          1,
				max:          1,
				minInclusive: true,
				maxInclusive: true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Equals(tt.args.property, tt.args.value); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Equals() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLessThan(t *testing.T) {
	type args struct {
		property string
		value    interface{}
	}
	tests := []struct {
		name string
		args args
		want Predicate
	}{
		{
			name: "<=",
			args: args{
				property: "start",
				value:    1,
			},
			want: Predicate{
				Property:     "start",
				min:          "-inf",
				max:          1,
				minInclusive: true,
				maxInclusive: false,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := LessThan(tt.args.property, tt.args.value); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LessThan() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLessThanEquals(t *testing.T) {
	type args struct {
		property string
		value    interface{}
	}
	tests := []struct {
		name string
		args args
		want Predicate
	}{
		{
			name: "n",
			args: args{
				property: "n",
				value:    1,
			},
			want: Predicate{
				Property:     "n",
				min:          "-inf",
				max:          1,
				minInclusive: true,
				maxInclusive: true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := LessThanEquals(tt.args.property, tt.args.value); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LessThanEquals() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGreaterThan(t *testing.T) {
	type args struct {
		property string
		value    interface{}
	}
	tests := []struct {
		name string
		args args
		want Predicate
	}{
		{
			name: "n",
			args: args{
				property: "n",
				value:    1,
			},
			want: Predicate{
				Property:     "n",
				min:          1,
				max:          "+inf",
				minInclusive: false,
				maxInclusive: true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GreaterThan(tt.args.property, tt.args.value); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GreaterThan() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGreaterThanEquals(t *testing.T) {
	type args struct {
		property string
		value    interface{}
	}
	tests := []struct {
		name string
		args args
		want Predicate
	}{
		{
			name: "n",
			args: args{
				property: "n",
				value:    1,
			},
			want: Predicate{
				Property:     "n",
				min:          1,
				max:          "+inf",
				minInclusive: true,
				maxInclusive: true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GreaterThanEquals(tt.args.property, tt.args.value); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GreaterThanEquals() = %v, want %v", got, tt.want)
			}
		})
	}
}
