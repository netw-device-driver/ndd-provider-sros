module github.com/netw-device-driver/ndd-provider-sros

go 1.16

require (
	github.com/netw-device-driver/ndd-core v0.1.8
	github.com/netw-device-driver/ndd-grpc v0.1.6
	github.com/netw-device-driver/ndd-runtime v0.3.78
	github.com/netw-device-driver/netwdevpb v0.1.27
	github.com/pkg/errors v0.9.1
	github.com/spf13/cobra v1.2.1
	golang.org/x/sys v0.0.0-20210630005230-0f9fa26af87c // indirect
	golang.org/x/tools v0.1.5 // indirect
	google.golang.org/grpc v1.39.0
	k8s.io/api v0.21.3
	k8s.io/apiextensions-apiserver v0.21.3
	k8s.io/apimachinery v0.21.3
	k8s.io/client-go v0.21.3
	sigs.k8s.io/controller-runtime v0.9.3
)
