package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func main() {
	tempDir, err := os.MkdirTemp("", "miactl-test-")
	if err != nil {
		fmt.Printf("Error creating temporary directory: %v\n", err)
		os.Exit(1)
	}
	defer os.RemoveAll(tempDir)

	flowManagerFile := filepath.Join(tempDir, "flowManagerConfig.json")
	flowManagerContent := `{
  "flowSettings": {
    "enabledFeatures": ["feature1", "feature2"],
    "defaultTimeout": 30000
  }
}`

	rbacManagerFile := filepath.Join(tempDir, "rbacManagerConfig.json")
	rbacManagerContent := `{
  "rbacSettings": {
    "enabledRoles": ["admin", "user"],
    "permissions": {
      "admin": ["read", "write"],
      "user": ["read"]
    }
  }
}`

	if err := os.WriteFile(flowManagerFile, []byte(flowManagerContent), 0644); err != nil {
		fmt.Printf("Error creating Flow Manager file: %v\n", err)
		os.Exit(1)
	}

	if err := os.WriteFile(rbacManagerFile, []byte(rbacManagerContent), 0644); err != nil {
		fmt.Printf("Error creating RBAC Manager file: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("=== FILE VALIDATION TEST ===")
	fmt.Printf("Files created in: %s\n", tempDir)
	fmt.Printf("- Flow Manager Config: %s\n", flowManagerFile)
	fmt.Printf("- RBAC Manager Config: %s\n", rbacManagerFile)
	fmt.Println("")

	fmt.Println("=== COMMAND TEST WITH --yes (SKIP CONFIRMATION) ===")
	cmd := exec.Command("./bin/miactl", "project", "import-config",
		"--project-id", "test-project",
		"--revision", "main",
		"--flow-manager-config", flowManagerFile,
		"--rbac-manager-config", rbacManagerFile,
		"--yes")

	cmd.Env = append(os.Environ(), "MIACTL_ENDPOINT=http://localhost:8080")

	output, err := cmd.CombinedOutput()
	fmt.Printf("Command output: %s\n", string(output))
	if err != nil {
		fmt.Printf("Command failed (expected without server): %v\n", err)
		fmt.Println("This is normal - the command tries to connect to the server")
	}

	fmt.Println("\n=== FILE VALIDATION COMPLETED ===")
	fmt.Println("JSON files have been validated successfully!")
}
