package models

import (
	"github.com/creasty/defaults"
)

type Label struct {
	Name  string `yaml:"name"`
	Value string `yaml:"value"`
}
type Resource struct {
	Type      string  `yaml:"type" default:"configmap" validate:"oneof=configmap secret both"`
	Path      string  `yaml:"path" default:"\tmp" validate:"dir"`
	Labels    []Label `yaml:"labels" validate:"min=1"`
	Method    string  `yaml:"method" default:"watch" validate:"oneof=watch get"`
	Namespace string  `yaml:"namespace"`
}
type Resources struct {
	Resources []Resource `yaml:"resources" validate:"min=1"`
}

func (s *Resource) UnmarshalYAML(unmarshal func(interface{}) error) error {
	defaults.Set(s)
	type plain Resource
	if err := unmarshal((*plain)(s)); err != nil {
		return err
	}
	return nil
}
