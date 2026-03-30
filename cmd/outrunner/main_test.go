package main

import (
	"testing"

	"github.com/actions/scaleset"
)

func TestLabelsMatch(t *testing.T) {
	label := func(names ...string) []scaleset.Label {
		var labels []scaleset.Label
		for _, n := range names {
			labels = append(labels, scaleset.Label{Name: n, Type: "User"})
		}
		return labels
	}

	tests := []struct {
		name     string
		existing []scaleset.Label
		desired  []scaleset.Label
		want     bool
	}{
		{"identical", label("linux", "x64"), label("linux", "x64"), true},
		{"same set different order", label("x64", "linux"), label("linux", "x64"), true},
		{"different labels", label("linux"), label("windows"), false},
		{"extra label", label("linux", "x64"), label("linux"), false},
		{"missing label", label("linux"), label("linux", "x64"), false},
		{"both empty", nil, nil, true},
		{"existing empty", nil, label("linux"), false},
		{"desired empty", label("linux"), nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := labelsMatch(tt.existing, tt.desired)
			if got != tt.want {
				t.Errorf("labelsMatch() = %v, want %v", got, tt.want)
			}
		})
	}
}
