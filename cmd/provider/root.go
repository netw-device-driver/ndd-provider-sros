/*
Copyright 2021 Wim Henderickx.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package provider

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"

	ndrv1 "github.com/netw-device-driver/ndd-core/apis/dvr/v1"
	srosv1 "github.com/netw-device-driver/ndd-provider-sros/apis/sros/v1"
	"github.com/netw-device-driver/ndd-provider-sros/internal/initializer"
	"github.com/netw-device-driver/ndd-runtime/pkg/logging"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	//+kubebuilder:scaffold:imports
)

var (
	scheme = runtime.NewScheme()
	debug  bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "manager",
	Short: "provider ndd sros",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	zlog := zap.New(zap.JSONEncoder())
	rootCmd.SilenceUsage = true
	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "enable debug mode")

	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(srosv1.AddToScheme(scheme))
	utilruntime.Must(ndrv1.AddToScheme(scheme))
	//utilruntime.Must(v1.AddToScheme(scheme))
	utilruntime.Must(apiextensionsv1.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme

	cfg, err := ctrl.GetConfig()
	if err != nil {
		fmt.Printf("cannot get config %s\n", err)
	}

	cl, err := client.New(cfg, client.Options{Scheme: scheme})
	if err != nil {
		fmt.Printf("cannot create new kubernetes client %s\n", err)
	}
	/*
		ns := os.Getenv("POD_NAMESPACE")
		fmt.Printf("namespace %s\n", ns)

		podName := os.Getenv("POD_NAME")
		fmt.Printf("podName %s\n", podName)
		split := strings.Split(podName, "-")
		var depName string
		if len(split) > 2 {
			split = split[:len(split)-2]
			for i := 0; i < len(split); i++ {
				if i == len(split)-1 {
					depName += split[i]
				} else {
					depName += split[i] + "-"
				}
			}
		}
		fmt.Printf("depName %s\n", depName)

		dKey := types.NamespacedName{
			Namespace: ns,
			Name:      depName,
		}
		d := &appsv1.Deployment{}
		if err := cl.Get(context.TODO(), dKey, d); err != nil {
			fmt.Printf("cannot get deployment %s\n", err)
		}
		fmt.Printf("deployment: %v\n", *d)
	*/

	i := initializer.New(cl,
		initializer.NewCRDWaiter([]string{
			fmt.Sprintf("%s.%s", "registrations", srosv1.Group),
		}, time.Minute, time.Second, logging.NewLogrLogger(zlog.WithName("nddrbacinit"))),
		initializer.NewRegistrationObject(),
	)
	if err := i.Init(context.TODO()); err != nil {
		fmt.Printf("cannot initialize provider %s\n", err)
	}
	fmt.Printf("Initialization has been completed\n")
}
