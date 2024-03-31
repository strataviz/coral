package cmd

import "time"

const (
	DefaultCertDir              string        = "/etc/webhook/tls"
	DefaultEnableLeaderElection bool          = false
	DefaultSkipInsecureVerify   bool          = false
	DefaultLogLevel             int8          = 0
	DefaultPollInterval         time.Duration = 10 * time.Second
	DefaultContainerdAddr       string        = "unix:///kubelet/containerd/containerd.sock"
	DefaultNamespace            string        = ""
)
