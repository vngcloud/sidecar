package models

import "k8s.io/apimachinery/pkg/types"

type FileInfo struct {
	ResourceName    string
	ResourceUID     types.UID
	ResourceVersion string
	Content         []byte
	Index           int
}
