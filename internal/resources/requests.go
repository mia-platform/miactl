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
	Name      string     `json:"name"`
	Type      string     `json:"tokenEndpointAuthMethod"` //nolint: tagliatelle
	Role      IAMRole    `json:"role"`
	PublicKey *PublicKey `json:"publicKey,omitempty"`
}

type AddUserRequest struct {
	Email string  `json:"email"`
	Role  IAMRole `json:"role"`
}

type CreateGroupRequest struct {
	Name    string   `json:"name"`
	Role    IAMRole  `json:"role"`
	Members []string `json:"members"`
}

type AddMembersToGroup struct {
	Members []string `json:"emails"` //nolint: tagliatelle
}

type RemoveMembersToGroup struct {
	Members []string `json:"memberIds"` //nolint: tagliatelle
}

type EditIAMRole struct {
	Role         IAMRole           `json:"role,omitempty"`
	ProjectsRole []EditProjectRole `json:"projectsRole,omitempty"`
}

type EditProjectRole struct {
	ProjectID        string                `json:"projectId"`
	Role             *IAMRole              `json:"role,omitempty"`
	EnvironmentsRole []EditEnvironmentRole `json:"environmentsRole,omitempty"`
}

type EditEnvironmentRole struct {
	EnvironmentID string  `json:"envId"` //nolint: tagliatelle
	Role          IAMRole `json:"role"`
}

type PublicKey struct {
	Type      string `json:"kty"` //nolint: tagliatelle
	Use       string `json:"use"` //nolint: tagliatelle
	Algorithm string `json:"alg"` //nolint: tagliatelle
	KeyID     string `json:"kid"` //nolint: tagliatelle
	Modulus   string `json:"n"`   //nolint: tagliatelle
	Exponent  string `json:"e"`   //nolint: tagliatelle
}

type IAMRole string

const (
	ServiceAccountBasic = "client_secret_basic"
	ServiceAccountJWT   = "private_key_jwt"

	IAMRoleGuest        = IAMRole("guest")
	IAMRoleReporter     = IAMRole("reporter")
	IAMRoleDeveloper    = IAMRole("developer")
	IAMRoleMaintainer   = IAMRole("maintainer")
	IAMRoleProjectAdmin = IAMRole("project-admin")
	IAMRoleCompanyOwner = IAMRole("company-owner")
)

type DeployProjectRequest struct {
	Environment string `json:"environment"`
	Revision    string `json:"revision"`
	Type        string `json:"deployType"`              //nolint: tagliatelle
	ForceDeploy bool   `json:"forceDeplpuWhenNoSemver"` //nolint: tagliatelle
}

type CreateJobRequest struct {
	From         string `json:"from"`
	ResourceName string `json:"resourceName"`
}
