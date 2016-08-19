/*
* Copyright 2016 Comcast Cable Communications Management, LLC
*
* Licensed under the Apache License, Version 2.0 (the "License");
* you may not use this file except in compliance with the License.
* You may obtain a copy of the License at
*
* http://www.apache.org/licenses/LICENSE-2.0
*
* Unless required by applicable law or agreed to in writing, software
* distributed under the License is distributed on an "AS IS" BASIS,
* WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
* See the License for the specific language governing permissions and
* limitations under the License.
 */

package zdddeploy_test

import (
	"github.com/cloudfoundry/cli/plugin/models"
	"github.com/cloudfoundry/cli/plugin/pluginfakes"
	"github.com/comcast/cf-zdd-plugin/canaryrepo"
	. "github.com/comcast/cf-zdd-plugin/zdddeploy"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("zddDeploy", func() {

	Describe(".init", func() {
		Context("when the package is imported", func() {
			It("should then be registered with the canary repo", func() {
				_, ok := canaryrepo.GetRegistry()[ZddDeployCmdName]
				Ω(ok).Should(BeTrue())
			})
		})
	})

	Describe("given: a valid run() method on a zdddeploy object which has been initialized with valid args", func() {
		var zddDeploy *ZddDeploy
		var ctrlAppName = "myTestApp#1.2.3-abcde"
		var ctrlManifestPath = "../fixtures/manifest.yml"
		var ctrlPathToArtifact = "path/to/artifact"
		var ctrlArgs = []string{ZddDeployCmdName, ctrlAppName, "-f", ctrlManifestPath, "-p", ctrlPathToArtifact}
		BeforeEach(func() {
			zddDeploy = new(ZddDeploy)
			zddDeploy.SetArgs(ctrlArgs)
		})

		Context("when called with a valid connection object for a new application deploy", func() {
			var fakeConnection *pluginfakes.FakeCliConnection
			var err error

			BeforeEach(func() {
				fakeConnection = new(pluginfakes.FakeCliConnection)
				returnModels := append([]plugin_models.GetAppsModel{}, plugin_models.GetAppsModel{
					Name: "1234",
				})
				fakeConnection.GetAppsReturns(returnModels, nil)
				err = zddDeploy.Run(fakeConnection)
			})
			It("should issue a single push of the application", func() {
				count := fakeConnection.CliCommandCallCount()
				Ω(err).ShouldNot(HaveOccurred())
				args := append([]string{"push"}, ctrlArgs[1:]...)
				Ω(fakeConnection.CliCommandArgsForCall(count - 1)).Should(Equal(args))
			})
		})
		Context("when called with a valid connection object for a new deploy", func() {
			var err error
			var fakeConnection *pluginfakes.FakeCliConnection
			BeforeEach(func() {
				zddDeploy.SetArgs([]string{ZddDeployCmdName, ctrlAppName, "-f", ctrlManifestPath, "-p", ctrlPathToArtifact, "480s"})
				fakeConnection = new(pluginfakes.FakeCliConnection)
				err = zddDeploy.Run(fakeConnection)
			})

			It("should not return an error", func() {
				Ω(err).ShouldNot(HaveOccurred())
			})
			It("should deploy an application in a scaleover way", func() {
				Ω(fakeConnection.CliCommandCallCount()).Should(Equal(1))
			})
		})
		Context("when called with a valid connection object for a new deploy and with additional scaleover arg", func() {
			var err error
			var fakeConnection *pluginfakes.FakeCliConnection
			BeforeEach(func() {
				zddDeploy.SetArgs([]string{ZddDeployCmdName, ctrlAppName, "-f", ctrlManifestPath, "-p", ctrlPathToArtifact, "60s"})
				fakeConnection = new(pluginfakes.FakeCliConnection)
				err = zddDeploy.Run(fakeConnection)
			})

			It("should not return an error", func() {
				Ω(err).ShouldNot(HaveOccurred())
			})
			It("should deploy an application in a scaleover way", func() {
				Ω(fakeConnection.CliCommandCallCount()).Should(Equal(1))
			})
		})

		Context("when called as a redeploy of the current app version", func() {
			var (
				fakeConnection *pluginfakes.FakeCliConnection
				returnModels   []plugin_models.GetAppsModel
				err            error
				callOrder      = map[string]int{
					"rename": 0,
					"push":   1,
					"delete": 2,
				}
				evaluateCallChainForAction = func(action string) {
					command := fakeConnection.CliCommandArgsForCall(callOrder[action])[0]
					Ω(command).Should(Equal(action))
				}
			)

			BeforeEach(func() {
				returnModels = make([]plugin_models.GetAppsModel, 0)
				fakeConnection = &pluginfakes.FakeCliConnection{}
				returnModels = append(returnModels,
					plugin_models.GetAppsModel{ctrlAppName, "", "", 0, 0, 0, 0, nil, nil},
				)
				zddDeploy = new(ZddDeploy)
				zddDeploy.SetArgs([]string{ZddDeployCmdName, ctrlAppName, "-f", ctrlManifestPath, "-p", ctrlPathToArtifact, "60s"})
				fakeConnection.GetAppsReturns(returnModels, nil)
				fakeConnection.CliCommandWithoutTerminalOutputReturns([]string{"requested state: stopped\ninstances: 0/10\nurls: app.cfapps.io"}, nil)
				err = zddDeploy.Run(fakeConnection)
			})

			It("should not return an error", func() {
				Ω(err).ShouldNot(HaveOccurred())
			})

			It("should have renamed the old version", func() {
				action := "rename"
				evaluateCallChainForAction(action)
			})

			It("should have pushed the new version", func() {
				action := "push"
				evaluateCallChainForAction(action)
			})

			It("should have deleted the old version", func() {
				action := "delete"
				evaluateCallChainForAction(action)
			})
		})
		Context("when called as a newly deployed app version", func() {
			var (
				fakeConnection *pluginfakes.FakeCliConnection
				returnModels   []plugin_models.GetAppsModel
				ctrlAppNameV2  string
				err            error
				callOrder      = map[string]int{
					"push":   0,
					"delete": 1,
				}
				evaluateCallChainForAction = func(action string) {
					command := fakeConnection.CliCommandArgsForCall(callOrder[action])[0]
					Ω(command).Should(Equal(action))
				}
			)

			BeforeEach(func() {
				returnModels = make([]plugin_models.GetAppsModel, 0)
				ctrlAppNameV2 = "myTestApp#1.2.2-abcde"
				returnModels = append(returnModels, plugin_models.GetAppsModel{ctrlAppNameV2, "", "", 0, 0, 0, 0, nil, nil})
				zddDeploy = new(ZddDeploy)
				fakeConnection = &pluginfakes.FakeCliConnection{}
				fakeConnection.GetAppsReturns(returnModels, nil)
				zddDeploy.SetArgs([]string{ZddDeployCmdName, ctrlAppName, "-f", ctrlManifestPath, "-p", ctrlPathToArtifact, "60s"})
				err = zddDeploy.Run(fakeConnection)
			})

			It("should not return an error", func() {
				Ω(err).ShouldNot(HaveOccurred())
			})

			It("should have pushed the new version", func() {
				action := "push"
				evaluateCallChainForAction(action)
			})

			It("should have deleted the old version", func() {
				action := "delete"
				evaluateCallChainForAction(action)
			})
		})
	})
})
