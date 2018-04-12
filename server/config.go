package server

type Config struct {
	// Server configuration
	Host            string
	Port            uint
	ExecConfigRoots []string

	// File system configurations
	AccountName             string
	AccountKey              string
	Container               string
	SourceWorktreeId        string
	WorktreeId              string
	MountPoint              string
	AutoSaveIntervalSeconds uint
}
