package shell

// ShellType represents supported shell types
type ShellType string

const (
	Bash       ShellType = "bash"
	Zsh        ShellType = "zsh"
	Fish       ShellType = "fish"
	PowerShell ShellType = "powershell"
	Cmd        ShellType = "cmd"
)

// IsSupported checks if the shell type is supported
func IsSupported(shellType string) bool {
	switch ShellType(shellType) {
	case Bash, Zsh, Fish, PowerShell:
		return true
	default:
		return false
	}
}

// GetFileExtension returns the appropriate file extension for the shell
func GetFileExtension(shellType string) string {
	switch ShellType(shellType) {
	case PowerShell:
		return ".ps1"
	case Fish:
		return ".fish"
	default:
		return ".sh"
	}
}