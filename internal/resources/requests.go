// Copyright Mia srl
// SPDX-License-Identifier: Apache-2.0
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

package resources

type JWTTokenRequest struct {
	Code  string `json:"code"`
	State string `json:"state"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refreshToken"`
}

type ServiceAccountRequest struct {
	Name      string             `json:"name"`
	Type      string             `json:"tokenEndpointAuthMethod"` //nolint: tagliatelle
	Role      ServiceAccountRole `json:"role"`
	PublicKey PublicKey          `json:"publicKey,omitempty"`
}

type PublicKey struct {
	Type      string `json:"kty"` //nolint: tagliatelle
	Use       string `json:"use"` //nolint: tagliatelle
	Algorithm string `json:"alg"` //nolint: tagliatelle
	KeyID     string `json:"kid"` //nolint: tagliatelle
	Modulus   string `json:"n"`   //nolint: tagliatelle
	Exponent  string `json:"e"`   //nolint: tagliatelle
}

type ServiceAccountRole string

const (
	ServiceAccountBasic = "client_secret_basic"
	ServiceAccountJWT   = "private_key_jwt"

	ServiceAccountRoleGuest        = ServiceAccountRole("guest")
	ServiceAccountRoleReporter     = ServiceAccountRole("reporter")
	ServiceAccountRoleDeveloper    = ServiceAccountRole("developer")
	ServiceAccountRoleMaintainer   = ServiceAccountRole("maintainer")
	ServiceAccountRoleProjectAdmin = ServiceAccountRole("project-admin")
	ServiceAccountRoleCompanyOwner = ServiceAccountRole("company-owner")
)

type DeployProjectRequest struct {
	Environment string `json:"environment"`
	Revision    string `json:"revision"`
	Type        string `json:"deployType"`              //nolint: tagliatelle
	ForceDeploy bool   `json:"forceDeplpuWhenNoSemver"` //nolint: tagliatelle
}
