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

import (
	"encoding/json"
	"time"

	rulesentities "github.com/mia-platform/miactl/internal/resources/rules"
	"golang.org/x/oauth2"
)

type APIError struct {
	StatusCode int    `json:"statusCode"`
	Message    string `json:"message"`
}

type AuthProvider struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	Type  string `json:"type"`
}

type UserToken struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	ExpiresAt    int64  `json:"expiresAt"`
}

func (ut *UserToken) JWTToken() *oauth2.Token {
	return &oauth2.Token{
		AccessToken:  ut.AccessToken,
		RefreshToken: ut.RefreshToken,
		Expiry:       time.Unix(ut.ExpiresAt, 0),
	}
}

type ConfigurationManagement struct {
	SaveChangesRules []*rulesentities.SaveChangesRules `json:"saveChangesRules"`
}

type ProjectConfigurationManagement struct {
	SaveChangesRules []*rulesentities.ProjectSaveChangesRules `json:"saveChangesRules"`
}

type Company struct {
	ID                      string                  `json:"_id"` //nolint:tagliatelle
	Name                    string                  `json:"name"`
	TenantID                string                  `json:"tenantId"`
	Pipelines               Pipelines               `json:"pipelines"`
	Repository              Repository              `json:"repository"`
	ConfigurationManagement ConfigurationManagement `json:"configurationManagement"`
}

type MarketplaceItem struct {
	ID          string `json:"_id"` //nolint:tagliatelle
	ItemID      string `json:"itemId"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	SupportedBy string `json:"supportedBy"`
	TenantID    string `json:"tenantId"`
}

type Pipelines struct {
	Type string `json:"type"`
}

type Repository struct {
	Type string `json:"type"`
}

type Cluster struct {
	ID           string `json:"_id"`       //nolint:tagliatelle
	DisplayName  string `json:"clusterId"` //nolint:tagliatelle
	Description  string `json:"description"`
	Distribution string `json:"distribution"`
	Vendor       string `json:"vendor"`
}

type Project struct {
	ID                   string        `json:"_id"` //nolint:tagliatelle
	Name                 string        `json:"name"`
	ConfigurationGitPath string        `json:"configurationGitPath"`
	Environments         []Environment `json:"environments"`
	ProjectID            string        `json:"projectId"`
	Pipelines            Pipelines     `json:"pipelines"`
	CompanyID            string        `json:"tenantId"` //nolint:tagliatelle

	ConfigurationManagement ProjectConfigurationManagement `json:"configurationManagement"`
}

type Environment struct {
	DisplayName  string         `json:"label"` //nolint:tagliatelle
	EnvID        string         `json:"envId"`
	Cluster      ProjectCluster `json:"cluster"`
	IsProduction bool           `json:"isProduction"`
}

type ProjectCluster struct {
	ID        string `json:"clusterId"` //nolint:tagliatelle
	Namespace string `json:"namespace"`
}

type DeployProject struct {
	ID  int    `json:"id"`
	URL string `json:"url"`
}

type PipelineStatus struct {
	ID     int    `json:"id"`
	Status string `json:"status"`
}

type CreateJob struct {
	JobName string `json:"jobName"`
}

type ServiceAccount struct {
	ClientID         string `json:"clientId"`
	ClientSecret     string `json:"clientSecret"`
	ClientIDIssuedAt int64  `json:"clientIdIssuedAt"`
	Company          string `json:"company"`
}

type Pod struct {
	Name      string    `json:"name"`
	Phase     string    `json:"phase"`
	Status    string    `json:"status"`
	Age       time.Time `json:"startTime"` //nolint:tagliatelle
	Component []struct {
		Name    string `json:"name"`
		Version string `json:"version"`
	} `json:"component"`
	Containers []struct {
		Name         string `json:"name"`
		Ready        bool   `json:"ready"`
		RestartCount int    `json:"restartCount"`
		Status       string `json:"status"`
	} `json:"containers"`
}

type CronJob struct {
	Name         string    `json:"name"`
	Active       int       `json:"active"`
	Suspend      bool      `json:"suspend"`
	Schedule     string    `json:"schedule"`
	Age          time.Time `json:"creationTimestamp"` //nolint: tagliatelle
	LastSchedule time.Time `json:"lastScheduleTime"`  //nolint: tagliatelle
}

type Job struct {
	Name           string    `json:"name"`
	Active         int       `json:"active"`
	Failed         int       `json:"failed"`
	Succeeded      int       `json:"succeeded"`
	Age            time.Time `json:"creationTimestamp"` //nolint: tagliatelle
	StartTime      time.Time `json:"startTime"`
	CompletionTime time.Time `json:"completionTime"`
}

type Deployment struct {
	Name      string    `json:"name"`
	Available int       `json:"available"`
	Ready     int       `json:"ready"`
	Replicas  int       `json:"replicas"`
	Age       time.Time `json:"creationTimestamp"` //nolint: tagliatelle
}

type Service struct {
	Name      string    `json:"name"`
	Type      string    `json:"type"`
	ClusterIP string    `json:"clusterIP"` //nolint: tagliatelle
	Ports     []Port    `json:"ports"`
	Age       time.Time `json:"creationTimestamp"` //nolint: tagliatelle
}

type Port struct {
	Name       string `json:"name"`
	Port       int    `json:"port"`
	Protocol   string `json:"protocol"`
	TargetPort string `json:"targetPort"`
}

type RuntimeEvent struct {
	Type      string    `json:"type"`
	Object    string    `json:"subObjectPath"` //nolint: tagliatelle
	Message   string    `json:"message"`
	Reason    string    `json:"reason"`
	FirstSeen time.Time `json:"firstSeen"`
	LastSeen  time.Time `json:"lastSeen"`
}

func (re *RuntimeEvent) UnmarshalJSON(data []byte) error {
	type shadowEvent struct {
		Type          string `json:"type"`
		SubObjectPath string `json:"subObjectPath"`
		Message       string `json:"message"`
		Reason        string `json:"reason"`
		FirstSeen     string `json:"firstSeen"`
		LastSeen      string `json:"lastSeen"`
	}

	parseTime := func(timeString string) (time.Time, error) {
		if len(timeString) == 0 {
			return time.Time{}, nil
		}
		return time.Parse(time.RFC3339, timeString)
	}

	var se shadowEvent
	if err := json.Unmarshal(data, &se); err != nil {
		return err
	}

	re.Type = se.Type
	re.Object = se.SubObjectPath
	re.Message = se.Message
	re.Reason = se.Reason

	firstSeen, err := parseTime(se.FirstSeen)
	if err != nil {
		return err
	}

	lastSeen, err := parseTime(se.LastSeen)
	if err != nil {
		return err
	}

	re.FirstSeen = firstSeen
	re.LastSeen = lastSeen
	return nil
}

type IAMIdentity struct {
	ID           string        `json:"identityId"` //nolint: tagliatelle
	Name         string        `json:"name"`
	Type         string        `json:"identityType"` //nolint: tagliatelle
	Roles        []string      `json:"companyRoles"` //nolint: tagliatelle
	ProjectsRole []ProjectRole `json:"projects"`     //nolint: tagliatelle
}

type ProjectRole struct {
	ID           string            `json:"_id"` //nolint: tagliatelle
	Roles        []string          `json:"roles"`
	Environments []EnvironmentRole `json:"environments"`
}

type EnvironmentRole struct {
	ID    string   `json:"envId"` //nolint: tagliatelle
	Roles []string `json:"roles"`
}

type UserIdentity struct {
	ID        string          `json:"userId"` //nolint: tagliatelle
	Email     string          `json:"email"`
	FullName  string          `json:"fullname"` //nolint: tagliatelle
	Name      string          `json:"name"`
	Roles     []string        `json:"companyRoles"` //nolint: tagliatelle
	LastLogin time.Time       `json:"lastLogin"`
	Groups    []GroupIdentity `json:"groups"`
}

func (ui *UserIdentity) UnmarshalJSON(data []byte) error {
	type shadowIdentity struct {
		UserID       string          `json:"userId"`
		Email        string          `json:"email"`
		Fullname     string          `json:"fullname"`
		Name         string          `json:"name"`
		CompanyRoles []string        `json:"companyRoles"`
		LastLogin    string          `json:"lastLogin"`
		Groups       []GroupIdentity `json:"groups"`
	}

	parseTime := func(timeString string) (time.Time, error) {
		if len(timeString) == 0 {
			return time.Time{}, nil
		}
		return time.Parse(time.RFC3339, timeString)
	}

	var si shadowIdentity
	if err := json.Unmarshal(data, &si); err != nil {
		return err
	}

	ui.ID = si.UserID
	ui.Email = si.Email
	ui.FullName = si.Fullname
	ui.Name = si.Name
	ui.Roles = si.CompanyRoles
	ui.Groups = si.Groups

	lastLogin, err := parseTime(si.LastLogin)
	if err != nil {
		return err
	}

	ui.LastLogin = lastLogin
	return nil
}

type GroupIdentity struct {
	ID      string         `json:"_id"` //nolint: tagliatelle
	Name    string         `json:"name"`
	Role    string         `json:"role"`
	RoleID  string         `json:"roleId"`
	Members []UserIdentity `json:"members"`
}

type ServiceAccountIdentity struct {
	ID         string    `json:"clientId"` //nolint: tagliatelle
	Name       string    `json:"name"`
	AuthMethod string    `json:"authMethod"`
	Roles      []string  `json:"companyRoles"` //nolint: tagliatelle
	LastLogin  time.Time `json:"lastLogin"`
}

func (sai *ServiceAccountIdentity) UnmarshalJSON(data []byte) error {
	type shadowIdentity struct {
		ClientID     string   `json:"clientId"`
		Name         string   `json:"name"`
		AuthMethod   string   `json:"authMethod"`
		CompanyRoles []string `json:"companyRoles"`
		LastLogin    string   `json:"lastLogin"`
	}

	parseTime := func(timeString string) (time.Time, error) {
		if len(timeString) == 0 {
			return time.Time{}, nil
		}
		return time.Parse(time.RFC3339, timeString)
	}

	var si shadowIdentity
	if err := json.Unmarshal(data, &si); err != nil {
		return err
	}

	sai.ID = si.ClientID
	sai.Name = si.Name
	sai.AuthMethod = si.AuthMethod
	sai.Roles = si.CompanyRoles

	lastLogin, err := parseTime(si.LastLogin)
	if err != nil {
		return err
	}

	sai.LastLogin = lastLogin
	return nil
}
