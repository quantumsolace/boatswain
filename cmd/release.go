// Copyright © 2017 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var colorNone = "\033[00m"
var colorYellow = "\033[01;33m"
var colorGreen = "\033[01;32m"

var env string
var dryrun bool
var ns string
var packfile string
var xdebug bool
var noExecute bool

type Config struct {
	ReleasePath string
}

// releaseCmd represents the release command
var releaseCmd = &cobra.Command{
	Use:   "release <appName>",
	Short: "Deploys application to minikube, staging, or production cluster",
	Long: `Wraps around a helm install command to automate common helm configuration
	options. Sets packageId, environment, and other important values.Execute
	
	Examples:
	boatswain release medbridge -x
	Release medbridge in the minikube cluster with XDebug enabled

	release medbridge -e staging
	Release medbridge in the staging cluster

	release medbridge -e staging -n test
	Release medbridge in the staging cluster, test namespace
	`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Println("Required argument: releaseName")
			return
		}

		releaseName := args[0]

		var xdebugHost string
		var packageId string
		var fullReleaseName string

		environments := []string{"development", "dev", "staging", "stage", "production", "prod"}

		// based on environment, set default packageId, packfile, and context
		switch env {
		case environments[0], environments[1]:
			if len(packageId) == 0 {
				packageId = "dev"
			}
			if len(packfile) == 0 {
				packfile = "values.env.yaml"
			}
			useK8sCurrContext("minikube")

		case environments[2], environments[3]:
			if len(packageId) == 0 {
				packageId = "staging"
			}
			if len(packfile) == 0 {
				packfile = "values.staging.yaml"
			}
			useK8sCurrContext("staging")
		case environments[4], environments[5]:
			if len(packageId) == 0 {
				packageId = "prod"
			}
			if len(packfile) == 0 {
				packfile = "values.prod.yaml"
			}
			useK8sCurrContext("production")
		default:
			fmt.Println("Invalid environment: " + env)
			os.Exit(1)
		}

		fullReleaseName = packageId + "-" + releaseName

		//xdebug option turned on, so get the xdebug host ip address
		if xdebug {

			var (
				cmdOut []byte
				err    error
			)
			if cmdOut, err = getXDebugHost(); err != nil {
				fmt.Fprintln(os.Stderr, "There was an error running ipconfig command: ", err)
				os.Exit(1)
			}
			xdebugHost = string(cmdOut)
			fmt.Println("Output: ", xdebugHost)
		}

		releasePath := viper.GetString("ReleasePath")
		appPath := releasePath + "/" + releaseName

		fmt.Printf("Deploying: %s\n", appPath)

		//build helm cmd
		setValues := "environment=" + env + ",packageId=" + packageId
		if xdebug {
			setValues += ",xdebugHost=" + xdebugHost
		}

		//fully qualified path
		packfile = appPath + "/" + packfile

		execHelmUpgradeCmd(fullReleaseName, appPath, setValues, packfile, ns)
	},
}

func init() {
	RootCmd.AddCommand(releaseCmd)

	//set option flags
	releaseCmd.Flags().StringVarP(&env, "environment", "e", "development", "Target environment for the release. 'production', 'staging', and 'development' are valid options")
	releaseCmd.Flags().BoolVarP(&dryrun, "dry-run", "d", false, "Dry run. Outputs the generated yaml files without deploying")
	releaseCmd.Flags().StringVarP(&ns, "namespace", "n", "default", "Namespace to deploy to")
	releaseCmd.Flags().StringVarP(&packfile, "packfile", "p", "", "The values yaml file to use")
	releaseCmd.Flags().BoolVarP(&xdebug, "xdebug", "x", false, "Enables xdebug (for dev environments only)")
	releaseCmd.Flags().BoolVar(&noExecute, "no-execute", false, "Echoes helm upgrade command, but does not execute")
}

func getK8sCurrContext() ([]byte, error) {
	cmdName := "kubectl"
	cmdArgs := []string{"config", "current-context"}
	cmdOut, err := exec.Command(cmdName, cmdArgs...).Output()
	check(err)
	return cmdOut, err
}

func getXDebugHost() ([]byte, error) {
	cmdName := "ipconfig"
	cmdArgs := []string{"getifaddr", "en0"}
	cmdOut, err := exec.Command(cmdName, cmdArgs...).Output()
	check(err)
	return cmdOut, err
}

func useK8sCurrContext(context string) ([]byte, error) {
	cmdName := "kubectl"
	cmdArgs := []string{"config", "use-context", context}
	cmdOut, err := exec.Command(cmdName, cmdArgs...).Output()
	check(err)
	return cmdOut, err
}

func execHelmUpgradeCmd(fullReleaseName string, appPath string, setValues string, packfile string, ns string) {
	dryRunOpt := ""
	debugOpt := ""
	if dryrun {
		dryRunOpt = "--dry-run"
		debugOpt = "--debug"
	}

	cmdName := "helm"
	cmdArgs := []string{
		"upgrade", fullReleaseName,
		"--install", appPath,
		"--set", setValues,
		"--values", packfile,
		"--namespace", ns,
		dryRunOpt, debugOpt}

	cmd := exec.Command(cmdName, cmdArgs...)
	cmdString := strings.Join(cmd.Args, " ")
	echoGoodMessage(cmdString)

	if !noExecute {
		fmt.Println("\n\nRunning helm upgrade...\n\n")
		out, _ := cmd.CombinedOutput()
		fmt.Printf("%s", out)
	}
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func pathExists(dirPath string) bool {
	_, err := os.Stat(dirPath)
	return !os.IsNotExist(err)

}

func echoWarningMessage(msg string) {
	fmt.Printf("%s%s%s", colorYellow, msg, colorNone)
}

func echoGoodMessage(msg string) {
	fmt.Printf("%s%s%s", colorGreen, msg, colorNone)
}