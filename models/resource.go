package models

type Label struct {
	Name  string `yaml:"name" validate:"required"`
	Value string `yaml:"value" validate:"required"`
}
type Resource struct {
	Type   string  `yaml:"type" validate:"required,configmap|secret|both"`
	Path   string  `yaml:"path" validate:"required,dir"`
	Labels []Label `yaml:"labels" validate:"required,min=1"`
}
type Resources struct {
	Resources []Resource `yaml:"resources" validate:"required,min=1"`
}
