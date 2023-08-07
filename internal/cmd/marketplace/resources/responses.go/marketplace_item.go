package marketplace

import (
	"encoding/json"

	"gopkg.in/yaml.v2"
)

func UnmarshalMarketplaceItem(data []byte) (MarketplaceItem, error) {
	var r MarketplaceItem
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *MarketplaceItem) MarshalMarketplaceItem() ([]byte, error) {
	return json.Marshal(r)
}

func UnmarshalMarketplaceItemYaml(data []byte) (MarketplaceItem, error) {
	var r MarketplaceItem
	err := yaml.Unmarshal(data, &r)
	return r, err
}

func (r *MarketplaceItem) MarshalMarketplaceItemYaml() ([]byte, error) {
	return yaml.Marshal(r)
}

type MarketplaceItem struct {
	ID                        *string        `json:"_id,omitempty"  yaml:"_id,omitempty"`
	Category                  *Category      `json:"category,omitempty"  yaml:"category,omitempty"`
	CategoryID                *string        `json:"categoryId,omitempty"  yaml:"categoryId,omitempty"`
	ComingSoon                *bool          `json:"comingSoon,omitempty"  yaml:"comingSoon,omitempty"`
	ComponentsIDS             []string       `json:"componentsIds,omitempty"  yaml:"componentsIds,omitempty"`
	Description               *string        `json:"description,omitempty"  yaml:"description,omitempty"`
	Documentation             *Documentation `json:"documentation,omitempty"  yaml:"documentation,omitempty"`
	ImageURL                  *string        `json:"imageUrl,omitempty"  yaml:"imageUrl,omitempty"`
	Name                      *string        `json:"name,omitempty"  yaml:"name,omitempty"`
	PublishOnMiaDocumentation *bool          `json:"publishOnMiaDocumentation,omitempty"  yaml:"publishOnMiaDocumentation,omitempty"`
	ReleaseStage              *string        `json:"releaseStage,omitempty"  yaml:"releaseStage,omitempty"`
	RepositoryURL             *string        `json:"repositoryUrl,omitempty"  yaml:"repositoryUrl,omitempty"`
	Resources                 *Resources     `json:"resources,omitempty"  yaml:"resources,omitempty"`
	SupportedBy               *string        `json:"supportedBy,omitempty"  yaml:"supportedBy,omitempty"`
	SupportedByImageURL       *string        `json:"supportedByImageUrl,omitempty"  yaml:"supportedByImageUrl,omitempty"`
	Type                      *string        `json:"type,omitempty"  yaml:"type,omitempty"`
}

type Category struct {
	ID    string `json:"id"  yaml:"id"`
	Label string `json:"label"  yaml:"label"`
}

type Documentation struct {
	Type Type   `json:"type"  yaml:"type"`
	URL  string `json:"url"  yaml:"url"`
}

type Resources struct {
	Collections         map[string]Collection         `json:"collections,omitempty"  yaml:"collections,omitempty"`
	Endpoints           map[string]EndpointValue      `json:"endpoints,omitempty"  yaml:"endpoints,omitempty"`
	Services            map[string]Service            `json:"services,omitempty"  yaml:"services,omitempty"`
	UnsecretedVariables map[string]UnsecretedVariable `json:"unsecretedVariables,omitempty"  yaml:"unsecretedVariables,omitempty"`
}

type Collection struct {
	DefaultName *string `json:"defaultName,omitempty"  yaml:"defaultName,omitempty"`
}

type EndpointValue struct {
	DefaultBasePath    *string `json:"defaultBasePath,omitempty"  yaml:"defaultBasePath,omitempty"`
	DefaultPathRewrite *string `json:"defaultPathRewrite,omitempty"  yaml:"defaultPathRewrite,omitempty"`
}

type Service struct {
	AdditionalContainers                 []AdditionalContainer      `json:"additionalContainers,omitempty"  yaml:"additionalContainers,omitempty"`
	ArchiveURL                           *string                    `json:"archiveUrl,omitempty"  yaml:"archiveUrl,omitempty"`
	ComponentID                          *string                    `json:"componentId,omitempty"  yaml:"componentId,omitempty"`
	ContainerPorts                       []ServiceContainerPort     `json:"containerPorts,omitempty"  yaml:"containerPorts,omitempty"`
	CustomFilesConfig                    []ServiceCustomFilesConfig `json:"customFilesConfig,omitempty"  yaml:"customFilesConfig,omitempty"`
	DefaultAnnotations                   []ServiceDefaultAnnotation `json:"defaultAnnotations,omitempty"  yaml:"defaultAnnotations,omitempty"`
	DefaultArgs                          []string                   `json:"defaultArgs,omitempty"  yaml:"defaultArgs,omitempty"`
	DefaultConfigMaps                    []ServiceDefaultConfigMap  `json:"defaultConfigMaps,omitempty"  yaml:"defaultConfigMaps,omitempty"`
	DefaultDocumentationPath             *string                    `json:"defaultDocumentationPath,omitempty"  yaml:"defaultDocumentationPath,omitempty"`
	DefaultEnvironmentVariables          []map[string]interface{}   `json:"defaultEnvironmentVariables,omitempty"  yaml:"defaultEnvironmentVariables,omitempty"`
	DefaultHeaders                       []ServiceDefaultHeader     `json:"defaultHeaders,omitempty"  yaml:"defaultHeaders,omitempty"`
	DefaultLabels                        []ServiceDefaultLabel      `json:"defaultLabels,omitempty"  yaml:"defaultLabels,omitempty"`
	DefaultLogParser                     *string                    `json:"defaultLogParser,omitempty"  yaml:"defaultLogParser,omitempty"`
	DefaultMonitoring                    *ServiceDefaultMonitoring  `json:"defaultMonitoring,omitempty"  yaml:"defaultMonitoring,omitempty"`
	DefaultProbes                        *ServiceDefaultProbes      `json:"defaultProbes,omitempty"  yaml:"defaultProbes,omitempty"`
	DefaultResources                     *ServiceDefaultResources   `json:"defaultResources,omitempty"  yaml:"defaultResources,omitempty"`
	DefaultSecrets                       []ServiceDefaultSecret     `json:"defaultSecrets,omitempty"  yaml:"defaultSecrets,omitempty"`
	DefaultTerminationGracePeriodSeconds *float64                   `json:"defaultTerminationGracePeriodSeconds,omitempty"  yaml:"defaultTerminationGracePeriodSeconds,omitempty"`
	Type                                 *string                    `json:"type,omitempty"  yaml:"type,omitempty"`
}

type AdditionalContainer struct {
	ArchiveURL                           *string                                `json:"archiveUrl,omitempty"  yaml:"archiveUrl,omitempty"`
	ComponentID                          *string                                `json:"componentId,omitempty"  yaml:"componentId,omitempty"`
	ContainerPorts                       []AdditionalContainerContainerPort     `json:"containerPorts,omitempty"  yaml:"containerPorts,omitempty"`
	CustomFilesConfig                    []AdditionalContainerCustomFilesConfig `json:"customFilesConfig,omitempty"  yaml:"customFilesConfig,omitempty"`
	DefaultAnnotations                   []AdditionalContainerDefaultAnnotation `json:"defaultAnnotations,omitempty"  yaml:"defaultAnnotations,omitempty"`
	DefaultArgs                          []string                               `json:"defaultArgs,omitempty"  yaml:"defaultArgs,omitempty"`
	DefaultConfigMaps                    []AdditionalContainerDefaultConfigMap  `json:"defaultConfigMaps,omitempty"  yaml:"defaultConfigMaps,omitempty"`
	DefaultDocumentationPath             *string                                `json:"defaultDocumentationPath,omitempty"  yaml:"defaultDocumentationPath,omitempty"`
	DefaultEnvironmentVariables          []map[string]interface{}               `json:"defaultEnvironmentVariables,omitempty"  yaml:"defaultEnvironmentVariables,omitempty"`
	DefaultHeaders                       []AdditionalContainerDefaultHeader     `json:"defaultHeaders,omitempty"  yaml:"defaultHeaders,omitempty"`
	DefaultLabels                        []AdditionalContainerDefaultLabel      `json:"defaultLabels,omitempty"  yaml:"defaultLabels,omitempty"`
	DefaultLogParser                     *string                                `json:"defaultLogParser,omitempty"  yaml:"defaultLogParser,omitempty"`
	DefaultMonitoring                    *AdditionalContainerDefaultMonitoring  `json:"defaultMonitoring,omitempty"  yaml:"defaultMonitoring,omitempty"`
	DefaultProbes                        *AdditionalContainerDefaultProbes      `json:"defaultProbes,omitempty"  yaml:"defaultProbes,omitempty"`
	DefaultResources                     *AdditionalContainerDefaultResources   `json:"defaultResources,omitempty"  yaml:"defaultResources,omitempty"`
	DefaultSecrets                       []AdditionalContainerDefaultSecret     `json:"defaultSecrets,omitempty"  yaml:"defaultSecrets,omitempty"`
	DefaultTerminationGracePeriodSeconds *float64                               `json:"defaultTerminationGracePeriodSeconds,omitempty"  yaml:"defaultTerminationGracePeriodSeconds,omitempty"`
	Type                                 *string                                `json:"type,omitempty"  yaml:"type,omitempty"`
}

type AdditionalContainerContainerPort struct {
	From     interface{} `json:"from"  yaml:"from"`
	Name     string      `json:"name"  yaml:"name"`
	Protocol *string     `json:"protocol,omitempty"  yaml:"protocol,omitempty"`
	To       interface{} `json:"to"  yaml:"to"`
}

type AdditionalContainerCustomFilesConfig struct {
	FileName *string `json:"fileName,omitempty"  yaml:"fileName,omitempty"`
	FilePath *string `json:"filePath,omitempty"  yaml:"filePath,omitempty"`
	FileType *string `json:"fileType,omitempty"  yaml:"fileType,omitempty"`
	Ref      *string `json:"ref,omitempty"  yaml:"ref,omitempty"`
}

type AdditionalContainerDefaultAnnotation struct {
	Description *string `json:"description,omitempty"  yaml:"description,omitempty"`
	Name        string  `json:"name"  yaml:"name"`
	ReadOnly    *bool   `json:"readOnly,omitempty"  yaml:"readOnly,omitempty"`
	Value       string  `json:"value"  yaml:"value"`
}

type AdditionalContainerDefaultConfigMap struct {
	Files          []PurpleFile `json:"files,omitempty"  yaml:"files,omitempty"`
	Link           *PurpleLink  `json:"link,omitempty"  yaml:"link,omitempty"`
	MountPath      *string      `json:"mountPath,omitempty"  yaml:"mountPath,omitempty"`
	Name           *string      `json:"name,omitempty"  yaml:"name,omitempty"`
	SubPaths       []string     `json:"subPaths,omitempty"  yaml:"subPaths,omitempty"`
	ViewAsReadOnly *bool        `json:"viewAsReadOnly,omitempty"  yaml:"viewAsReadOnly,omitempty"`
}

type PurpleFile struct {
	Content *string `json:"content,omitempty"  yaml:"content,omitempty"`
	Name    *string `json:"name,omitempty"  yaml:"name,omitempty"`
}

type PurpleLink struct {
	TargetSection *string `json:"targetSection,omitempty"  yaml:"targetSection,omitempty"`
}

type AdditionalContainerDefaultHeader struct {
	Description string `json:"description"  yaml:"description"`
	Name        string `json:"name"  yaml:"name"`
	Value       string `json:"value"  yaml:"value"`
}

type AdditionalContainerDefaultLabel struct {
	Description *string `json:"description,omitempty"  yaml:"description,omitempty"`
	Name        string  `json:"name"  yaml:"name"`
	ReadOnly    *bool   `json:"readOnly,omitempty"  yaml:"readOnly,omitempty"`
	Value       string  `json:"value"  yaml:"value"`
}

type AdditionalContainerDefaultMonitoring struct {
	Endpoints []PurpleEndpoint `json:"endpoints,omitempty"  yaml:"endpoints,omitempty"`
}

type PurpleEndpoint struct {
	Interval *string `json:"interval,omitempty"  yaml:"interval,omitempty"`
	Path     *string `json:"path,omitempty"  yaml:"path,omitempty"`
	Port     *string `json:"port,omitempty"  yaml:"port,omitempty"`
}

type AdditionalContainerDefaultProbes struct {
	Liveness  *PurpleLiveness  `json:"liveness,omitempty"  yaml:"liveness,omitempty"`
	Readiness *PurpleReadiness `json:"readiness,omitempty"  yaml:"readiness,omitempty"`
}

type PurpleLiveness struct {
	Cmd                 []string `json:"cmd,omitempty"  yaml:"cmd,omitempty"`
	FailureThreshold    *float64 `json:"failureThreshold,omitempty"  yaml:"failureThreshold,omitempty"`
	InitialDelaySeconds *float64 `json:"initialDelaySeconds,omitempty"  yaml:"initialDelaySeconds,omitempty"`
	Path                *string  `json:"path,omitempty"  yaml:"path,omitempty"`
	PeriodSeconds       *float64 `json:"periodSeconds,omitempty"  yaml:"periodSeconds,omitempty"`
	Port                *int64   `json:"port,omitempty"  yaml:"port,omitempty"`
	SuccessThreshold    *float64 `json:"successThreshold,omitempty"  yaml:"successThreshold,omitempty"`
	TimeoutSeconds      *float64 `json:"timeoutSeconds,omitempty"  yaml:"timeoutSeconds,omitempty"`
}

type PurpleReadiness struct {
	Cmd                 []string `json:"cmd,omitempty"  yaml:"cmd,omitempty"`
	FailureThreshold    *float64 `json:"failureThreshold,omitempty"  yaml:"failureThreshold,omitempty"`
	InitialDelaySeconds *float64 `json:"initialDelaySeconds,omitempty"  yaml:"initialDelaySeconds,omitempty"`
	Path                *string  `json:"path,omitempty"  yaml:"path,omitempty"`
	PeriodSeconds       *float64 `json:"periodSeconds,omitempty"  yaml:"periodSeconds,omitempty"`
	Port                *int64   `json:"port,omitempty"  yaml:"port,omitempty"`
	SuccessThreshold    *float64 `json:"successThreshold,omitempty"  yaml:"successThreshold,omitempty"`
	TimeoutSeconds      *float64 `json:"timeoutSeconds,omitempty"  yaml:"timeoutSeconds,omitempty"`
}

type AdditionalContainerDefaultResources struct {
	CPULimits    *PurpleCPULimits    `json:"cpuLimits,omitempty"  yaml:"cpuLimits,omitempty"`
	MemoryLimits *PurpleMemoryLimits `json:"memoryLimits,omitempty"  yaml:"memoryLimits,omitempty"`
}

type PurpleCPULimits struct {
	Max *string `json:"max,omitempty"  yaml:"max,omitempty"`
	Min *string `json:"min,omitempty"  yaml:"min,omitempty"`
}

type PurpleMemoryLimits struct {
	Max *string `json:"max,omitempty"  yaml:"max,omitempty"`
	Min *string `json:"min,omitempty"  yaml:"min,omitempty"`
}

type AdditionalContainerDefaultSecret struct {
	MountPath *string `json:"mountPath,omitempty"  yaml:"mountPath,omitempty"`
	Name      *string `json:"name,omitempty"  yaml:"name,omitempty"`
}

type ServiceContainerPort struct {
	From     interface{} `json:"from"  yaml:"from"`
	Name     string      `json:"name"  yaml:"name"`
	Protocol *string     `json:"protocol,omitempty"  yaml:"protocol,omitempty"`
	To       interface{} `json:"to"  yaml:"to"`
}

type ServiceCustomFilesConfig struct {
	FileName *string `json:"fileName,omitempty"  yaml:"fileName,omitempty"`
	FilePath *string `json:"filePath,omitempty"  yaml:"filePath,omitempty"`
	FileType *string `json:"fileType,omitempty"  yaml:"fileType,omitempty"`
	Ref      *string `json:"ref,omitempty"  yaml:"ref,omitempty"`
}

type ServiceDefaultAnnotation struct {
	Description *string `json:"description,omitempty"  yaml:"description,omitempty"`
	Name        string  `json:"name"  yaml:"name"`
	ReadOnly    *bool   `json:"readOnly,omitempty"  yaml:"readOnly,omitempty"`
	Value       string  `json:"value"  yaml:"value"`
}

type ServiceDefaultConfigMap struct {
	Files          []FluffyFile `json:"files,omitempty"  yaml:"files,omitempty"`
	Link           *FluffyLink  `json:"link,omitempty"  yaml:"link,omitempty"`
	MountPath      *string      `json:"mountPath,omitempty"  yaml:"mountPath,omitempty"`
	Name           *string      `json:"name,omitempty"  yaml:"name,omitempty"`
	SubPaths       []string     `json:"subPaths,omitempty"  yaml:"subPaths,omitempty"`
	ViewAsReadOnly *bool        `json:"viewAsReadOnly,omitempty"  yaml:"viewAsReadOnly,omitempty"`
}

type FluffyFile struct {
	Content *string `json:"content,omitempty"  yaml:"content,omitempty"`
	Name    *string `json:"name,omitempty"  yaml:"name,omitempty"`
}

type FluffyLink struct {
	TargetSection *string `json:"targetSection,omitempty"  yaml:"targetSection,omitempty"`
}

type ServiceDefaultHeader struct {
	Description string `json:"description"  yaml:"description"`
	Name        string `json:"name"  yaml:"name"`
	Value       string `json:"value"  yaml:"value"`
}

type ServiceDefaultLabel struct {
	Description *string `json:"description,omitempty"  yaml:"description,omitempty"`
	Name        string  `json:"name"  yaml:"name"`
	ReadOnly    *bool   `json:"readOnly,omitempty"  yaml:"readOnly,omitempty"`
	Value       string  `json:"value"  yaml:"value"`
}

type ServiceDefaultMonitoring struct {
	Endpoints []FluffyEndpoint `json:"endpoints,omitempty"  yaml:"endpoints,omitempty"`
}

type FluffyEndpoint struct {
	Interval *string `json:"interval,omitempty"  yaml:"interval,omitempty"`
	Path     *string `json:"path,omitempty"  yaml:"path,omitempty"`
	Port     *string `json:"port,omitempty"  yaml:"port,omitempty"`
}

type ServiceDefaultProbes struct {
	Liveness  *FluffyLiveness  `json:"liveness,omitempty"  yaml:"liveness,omitempty"`
	Readiness *FluffyReadiness `json:"readiness,omitempty"  yaml:"readiness,omitempty"`
}

type FluffyLiveness struct {
	Cmd                 []string `json:"cmd,omitempty"  yaml:"cmd,omitempty"`
	FailureThreshold    *float64 `json:"failureThreshold,omitempty"  yaml:"failureThreshold,omitempty"`
	InitialDelaySeconds *float64 `json:"initialDelaySeconds,omitempty"  yaml:"initialDelaySeconds,omitempty"`
	Path                *string  `json:"path,omitempty"  yaml:"path,omitempty"`
	PeriodSeconds       *float64 `json:"periodSeconds,omitempty"  yaml:"periodSeconds,omitempty"`
	Port                *int64   `json:"port,omitempty"  yaml:"port,omitempty"`
	SuccessThreshold    *float64 `json:"successThreshold,omitempty"  yaml:"successThreshold,omitempty"`
	TimeoutSeconds      *float64 `json:"timeoutSeconds,omitempty"  yaml:"timeoutSeconds,omitempty"`
}

type FluffyReadiness struct {
	Cmd                 []string `json:"cmd,omitempty"  yaml:"cmd,omitempty"`
	FailureThreshold    *float64 `json:"failureThreshold,omitempty"  yaml:"failureThreshold,omitempty"`
	InitialDelaySeconds *float64 `json:"initialDelaySeconds,omitempty"  yaml:"initialDelaySeconds,omitempty"`
	Path                *string  `json:"path,omitempty"  yaml:"path,omitempty"`
	PeriodSeconds       *float64 `json:"periodSeconds,omitempty"  yaml:"periodSeconds,omitempty"`
	Port                *int64   `json:"port,omitempty"  yaml:"port,omitempty"`
	SuccessThreshold    *float64 `json:"successThreshold,omitempty"  yaml:"successThreshold,omitempty"`
	TimeoutSeconds      *float64 `json:"timeoutSeconds,omitempty"  yaml:"timeoutSeconds,omitempty"`
}

type ServiceDefaultResources struct {
	CPULimits    *FluffyCPULimits    `json:"cpuLimits,omitempty"  yaml:"cpuLimits,omitempty"`
	MemoryLimits *FluffyMemoryLimits `json:"memoryLimits,omitempty"  yaml:"memoryLimits,omitempty"`
}

type FluffyCPULimits struct {
	Max *string `json:"max,omitempty"  yaml:"max,omitempty"`
	Min *string `json:"min,omitempty"  yaml:"min,omitempty"`
}

type FluffyMemoryLimits struct {
	Max *string `json:"max,omitempty"  yaml:"max,omitempty"`
	Min *string `json:"min,omitempty"  yaml:"min,omitempty"`
}

type ServiceDefaultSecret struct {
	MountPath *string `json:"mountPath,omitempty"  yaml:"mountPath,omitempty"`
	Name      *string `json:"name,omitempty"  yaml:"name,omitempty"`
}

type UnsecretedVariable struct {
	NoProductionEnv string `json:"noProductionEnv"  yaml:"noProductionEnv"`
	ProductionEnv   string `json:"productionEnv"  yaml:"productionEnv"`
}

type Type string

const (
	ExternalLink Type = "externalLink"
	Markdown     Type = "markdown"
)
