package main

import (
	"stvz.io/coral/pkg/cmd"
)

// const (
// 	DefaultCertDir              string = "/etc/webhook/tls"
// 	DefaultEnableLeaderElection bool   = false
// 	DefaultSkipInsecureVerify   bool   = false
// )

// func init() {
// 	_ = v1.AddToScheme(scheme)
// 	_ = corev1.AddToScheme(scheme)
// 	_ = appsv1.AddToScheme(scheme)
// }

func main() {
	root := cmd.NewRoot()
	root.Execute()
}
