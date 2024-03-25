package cmd

import "time"

const (
	DefaultCertDir              string        = "/etc/webhook/tls"
	DefaultEnableLeaderElection bool          = false
	DefaultSkipInsecureVerify   bool          = false
	DefaultLogLevel             int8          = 0
	DefaultPullInterval         time.Duration = 10 * time.Second
	DefaultCleanInterval        time.Duration = 5 * time.Second
	DefaultContainerdAddr       string        = "unix:///kubelet/containerd/containerd.sock"
	DefaultNamespace            string        = ""
)
