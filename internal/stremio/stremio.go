package stremio

import "scrapper/internal/config"

type Manifest struct {
	ID          string
	Version     string
	Name        string
	Description string
	Resources   []string
	Types       []string
	Catalogs    []any
	Logo        string
}

func NewManifest(manifestConfig config.ManifestConfig) Manifest {
	return Manifest{
		ID:          manifestConfig.ID,
		Version:     manifestConfig.Version,
		Name:        manifestConfig.Name,
		Description: manifestConfig.Description,
		Resources:   manifestConfig.Resources,
		Types:       manifestConfig.Types,
		Catalogs:    manifestConfig.Catalogs,
		Logo:        manifestConfig.Logo,
	}
}
