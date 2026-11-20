// Copyright The Notary Project Authors.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package plugin

import (
	"fmt"

	. "github.com/notaryproject/notation/test/e2e/internal/notation"
	"github.com/notaryproject/notation/test/e2e/internal/utils"
	. "github.com/notaryproject/notation/test/e2e/suite/common"
	. "github.com/onsi/ginkgo/v2"
)

var _ = Describe("notation plugin sign and verify image types", func() {
	It("with OCI image manifest", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
			// the default artifact is an OCI image manifest
			addPluginKey(notation, vhost)

			signAndVerifyWithPlugin(notation, artifact, "jws")
			signAndVerifyWithPlugin(notation, artifact, "cose")
		})
	})

	It("with Docker manifest v2", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, _ *Artifact, vhost *utils.VirtualHost) {
			artifact := GenerateArtifact("e2e-docker-manifest", "")
			addPluginKey(notation, vhost)

			signAndVerifyWithPlugin(notation, artifact, "jws")
			signAndVerifyWithPlugin(notation, artifact, "cose")
		})
	})

	It("with OCI image index", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, _ *Artifact, vhost *utils.VirtualHost) {
			artifact := GenerateArtifact("e2e-image-index", "")
			addPluginKey(notation, vhost)

			signAndVerifyWithPlugin(notation, artifact, "jws")
			signAndVerifyWithPlugin(notation, artifact, "cose")
		})
	})
})

// addPluginKey installs the e2e-plugin and adds a plugin key with the
// SIGNATURE_GENERATOR.ENVELOPE capability. The generated signature embeds the
// verification plugin header so that verification is also done by the plugin.
func addPluginKey(notation *utils.ExecOpts, vhost *utils.VirtualHost) {
	vhost.SetOption(AddPlugin(NotationE2EPluginPath))
	notation.Exec("key", "add", "plugin-key", "--id", "key1", "--plugin", PluginName,
		"--plugin-config", fmt.Sprintf("%s=true", CapabilityEnvelopeGenerator),
		"--plugin-config", fmt.Sprintf("%s=%s", HeaderVerificationPlugin, PluginName)).
		MatchKeyWords("plugin-key")
}

// signAndVerifyWithPlugin signs the artifact with the plugin in the given
// signature envelope format and verifies it with plugin verification.
func signAndVerifyWithPlugin(notation *utils.ExecOpts, artifact *Artifact, sigFormat string) {
	notation.WithDescription(fmt.Sprintf("sign with plugin in %s format", sigFormat)).
		Exec("sign", artifact.ReferenceWithDigest(), "--key", "plugin-key", "--signature-format", sigFormat, "-d").
		MatchErrKeyWords(
			"Plugin get-plugin-metadata request",
			"Plugin generate-envelope request",
		).
		MatchKeyWords(SignSuccessfully)

	notation.WithDescription(fmt.Sprintf("verify %s signature with plugin", sigFormat)).
		Exec("verify", artifact.ReferenceWithDigest(), "-d").
		MatchErrKeyWords(
			"Plugin verify-signature request",
			"Plugin verify-signature response",
		).
		MatchKeyWords(VerifySuccessfully)
}
