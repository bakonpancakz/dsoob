package include

import "embed"

//go:embed templates/*.txt
var Templates embed.FS

//go:embed DatabaseSchema.sql
var DatabaseSchema string

//go:embed DatabaseGeolocate.kani.gz
var DatabaseGeolocate []byte
