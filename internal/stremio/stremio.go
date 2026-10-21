package stremio

type ManifestOptions struct {
	ID          string
	Version     string
	Name        string
	Description string
	Resources   []string
	Types       []string
	Catalogs    []any
	Logo        string
}

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

func NewManifest(opt ManifestOptions) Manifest {
	return Manifest{
		ID:          opt.ID,
		Version:     opt.Version,
		Name:        opt.Name,
		Description: opt.Description,
		Resources:   opt.Resources,
		Types:       opt.Types,
		Catalogs:    opt.Catalogs,
		Logo:        opt.Logo,
	}
}
