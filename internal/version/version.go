package version

var (
	buildTime  string
	commitHash string
	entryPoint string
)

func BuildTime() string {
	if buildTime == "" {
		return "unknown"
	}
	return buildTime
}

func CommitHash() string {
	if commitHash == "" {
		return "unknown"
	}
	return commitHash
}

func EntryPoint() string {
	if entryPoint == "" {
		return "bird-lg-go"
	}
	return entryPoint
}
