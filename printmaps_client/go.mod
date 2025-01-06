module github.com/printmaps/printmaps/printmaps_client

go 1.23

require (
	github.com/StefanSchroeder/Golang-Ellipsoid v0.0.0-20241222205244-01047a497824
	github.com/davecgh/go-spew v1.1.1
	github.com/im7mortal/UTM v1.4.0
	github.com/paulmach/orb v0.11.1
	github.com/printmaps/printmaps/pd v1.0.0
	github.com/yuin/gopher-lua v1.1.1
	gopkg.in/yaml.v2 v2.4.0
)

require go.mongodb.org/mongo-driver v1.17.1 // indirect

replace github.com/printmaps/printmaps/pd => ../pd
