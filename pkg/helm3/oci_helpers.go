package helm3

import (
	"fmt"
	"regexp"
	"strings"
)

var ociChartRegex *regexp.Regexp = regexp.MustCompile(`oci:\/\/([^:/]+)\/([^:]+)(:([^:]+))?`)

func IsOciInstall(step InstallStep) bool {
	return ociChartRegex.MatchString(step.Chart)
}

// Only call this function if IsOciInstall == true
func LoginToOciRegistryIfNecessary(step InstallStep, m *Mixin) error {
	if (step.RegistryAuth.Username == "" && step.RegistryAuth.Password != "") ||
		(step.RegistryAuth.Username != "" && step.RegistryAuth.Password == "") {
		return fmt.Errorf("either password or username is empty but not both")
	}

	subGroups := ociChartRegex.FindStringSubmatch(step.Chart)

	registryUrl := subGroups[1]

	cmd := m.NewCommand(
		"helm3",
		"registry",
		"login",
		registryUrl)

	if step.RegistryAuth.Username != "" && step.RegistryAuth.Password != "" {
		cmd.Args = append(cmd.Args,
			"-u",
			step.RegistryAuth.Username,
			"-p",
			step.RegistryAuth.Password)
	}

	return m.RunCmd(cmd, false)
}

// Only call this function if IsOciInstall == true
func GetFullOciResourceName(step InstallStep) string {
	subGroups := ociChartRegex.FindStringSubmatch(step.Chart)

	fullOciResourceName := subGroups[1] + "/" + subGroups[2]
	tag := ":" + subGroups[4]
	if subGroups[4] == "" {
		tag = ":" + step.Version
	}
	fullOciResourceName = fullOciResourceName + tag
	return fullOciResourceName
}

// Only call this function if IsOciInstall == true
func PullChartFromOciRegistry(step InstallStep, m *Mixin) error {
	cmd := m.NewCommand(
		"helm3",
		"chart",
		"pull",
		GetFullOciResourceName(step))

	return m.RunCmd(cmd, true)
}

// Only call this function if IsOciInstall == true
func ExportOciChartToTempPath(step InstallStep, m *Mixin) (string, error) {
	tempChartPath := "/tmp/"
	cmd := m.NewCommand(
		"helm3",
		"chart",
		"export",
		GetFullOciResourceName(step),
		"--destination",
		tempChartPath)

	cmd.Stderr = m.Err

	// format the command with all arguments
	prettyCmd := fmt.Sprintf("%s %s", cmd.Path, strings.Join(cmd.Args, " "))
	fmt.Fprintln(m.Out, prettyCmd)

	// Here where really the command get executed
	output, err := cmd.Output()
	m.Out.Write(output)
	// Exit on error
	if err != nil {
		return "", fmt.Errorf("could not execute command, %s: %s", prettyCmd, err)
	}

	storagePathRegex := regexp.MustCompile(`Exported chart to (.*)`)
	subGroups := storagePathRegex.FindStringSubmatch(string(output))

	if len(subGroups) == 0 || subGroups[1] == "" {
		return "", fmt.Errorf("could not extract export path using regex %s", storagePathRegex.String())
	}

	return subGroups[1], nil
}

// Only call this function if IsOciInstall == true
func RemoveLocalOciExport(step InstallStep, m *Mixin) error {
	cmd := m.NewCommand(
		"rm",
		"-rf",
		step.Chart)
	return m.RunCmd(cmd, true)
}
