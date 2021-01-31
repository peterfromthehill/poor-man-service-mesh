package main

type MeshConfig struct {
	IncomePort   int
	OutgoingPort int
	SslKeyPath   string
	SslCertPath  string
	ServiceCidr  string
	PodCidr      string
	ExitProxy    string
	Secure       bool
	CA           string
	DirectoryURL string
	ExitNode     bool
}
