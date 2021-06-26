module github.com/printmaps/printmaps/printmaps_webservice

go 1.16

require (
	github.com/JamesMilnerUK/pip-go v0.0.0-20180711171552-99c4cbbc7deb // indirect
	github.com/gofrs/uuid v4.0.0+incompatible // indirect
	github.com/julienschmidt/httprouter v1.3.0 // indirect
	github.com/printmaps/printmaps/pd v1.0.0
	github.com/rs/cors v1.7.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)

replace github.com/printmaps/printmaps/pd => ../pd
