package v1

import (
	authenticationv1 "k8s.io/api/authentication/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

//GenerateRequest is a request to process generate rule
type GenerateRequest struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              GenerateRequestSpec   `json:"spec"`
	Status            GenerateRequestStatus `json:"status"`
}

//GenerateRequestSpec stores the request specification
type GenerateRequestSpec struct {
	Policy   string                 `json:"policy"`
	Resource ResourceSpec           `json:"resource"`
	Context  GenerateRequestContext `json:"context"`
}

//GenerateRequestContext stores the context to be shared
type GenerateRequestContext struct {
	UserRequestInfo RequestInfo `json:"userInfo,omitempty"`
}

// RequestInfo contains permission info carried in an admission request
type RequestInfo struct {
	// Roles is a list of possible role send the request
	Roles []string `json:"roles"`
	// ClusterRoles is a list of possible clusterRoles send the request
	ClusterRoles []string `json:"clusterRoles"`
	// UserInfo is the userInfo carried in the admission request
	AdmissionUserInfo authenticationv1.UserInfo `json:"userInfo"`
}

//GenerateRequestStatus stores the status of generated request
type GenerateRequestStatus struct {
	State   GenerateRequestState `json:"state"`
	Message string               `json:"message,omitempty"`
}

//GenerateRequestState defines the state of
type GenerateRequestState string

const (
	//Pending - the Request is yet to be processed or resource has not been created
	Pending GenerateRequestState = "Pending"
	//Failed - the Generate Request Controller failed to process the rules
	Failed GenerateRequestState = "Failed"
	//Completed - the Generate Request Controller created resources defined in the policy
	Completed GenerateRequestState = "Completed"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

//GenerateRequestList stores the list of generate requests
type GenerateRequestList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []GenerateRequest `json:"items"`
}

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ClusterPolicy ...
type ClusterPolicy Policy

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ClusterPolicyList ...
type ClusterPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []ClusterPolicy `json:"items"`
}

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ClusterPolicyViolation represents cluster-wide violations
type ClusterPolicyViolation PolicyViolationTemplate

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ClusterPolicyViolationList ...
type ClusterPolicyViolationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []ClusterPolicyViolation `json:"items"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PolicyViolation represents namespaced violations
type PolicyViolation PolicyViolationTemplate

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PolicyViolationList ...
type PolicyViolationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []PolicyViolation `json:"items"`
}

// Policy contains rules to be applied to created resources
type Policy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              Spec         `json:"spec"`
	Status            PolicyStatus `json:"status"`
}

// Spec describes policy behavior by its rules
type Spec struct {
	Rules                   []Rule `json:"rules"`
	ValidationFailureAction string `json:"validationFailureAction"`
	Background              bool   `json:"background,omitempty"`
}

// Rule is set of mutation, validation and generation actions
// for the single resource description
type Rule struct {
	Name             string           `json:"name"`
	MatchResources   MatchResources   `json:"match"`
	ExcludeResources ExcludeResources `json:"exclude,omitempty"`
	Mutation         Mutation         `json:"mutate"`
	Validation       Validation       `json:"validate"`
	Generation       Generation       `json:"generate"`
}

//MatchResources contains resource description of the resources that the rule is to apply on
type MatchResources struct {
	UserInfo
	ResourceDescription `json:"resources"`
}

//ExcludeResources container resource description of the resources that are to be excluded from the applying the policy rule
type ExcludeResources struct {
	UserInfo
	ResourceDescription `json:"resources"`
}

// UserInfo filter based on users
type UserInfo struct {
	Roles        []string         `json:"roles"`
	ClusterRoles []string         `json:"clusterRoles"`
	Subjects     []rbacv1.Subject `json:"subjects"`
}

// ResourceDescription describes the resource to which the PolicyRule will be applied.
type ResourceDescription struct {
	Kinds      []string              `json:"kinds"`
	Name       string                `json:"name"`
	Namespaces []string              `json:"namespaces,omitempty"`
	Selector   *metav1.LabelSelector `json:"selector"`
}

// Mutation describes the way how Mutating Webhook will react on resource creation
type Mutation struct {
	Overlay interface{} `json:"overlay"`
	Patches []Patch     `json:"patches"`
}

// +k8s:deepcopy-gen=false

// Patch declares patch operation for created object according to RFC 6902
type Patch struct {
	Path      string      `json:"path"`
	Operation string      `json:"op"`
	Value     interface{} `json:"value"`
}

// Validation describes the way how Validating Webhook will check the resource on creation
type Validation struct {
	Message    string        `json:"message"`
	Pattern    interface{}   `json:"pattern"`
	AnyPattern []interface{} `json:"anyPattern"`
}

// Generation describes which resources will be created when other resource is created
type Generation struct {
	Kind  string      `json:"kind"`
	Name  string      `json:"name"`
	Data  interface{} `json:"data"`
	Clone CloneFrom   `json:"clone"`
}

// CloneFrom - location of a Secret or a ConfigMap
// which will be used as source when applying 'generate'
type CloneFrom struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
}

//PolicyStatus provides status for violations
type PolicyStatus struct {
	ViolationCount int `json:"violationCount"`
	// Count of rules that were applied
	RulesAppliedCount int `json:"rulesAppliedCount"`
	// Count of resources for whom update/create api requests were blocked as the resoruce did not satisfy the policy rules
	ResourcesBlockedCount int `json:"resourcesBlockedCount"`
	// average time required to process the policy Mutation rules on a resource
	AvgExecutionTimeMutation string `json:"averageMutationRulesExecutionTime"`
	// average time required to process the policy Validation rules on a resource
	AvgExecutionTimeValidation string `json:"averageValidationRulesExecutionTime"`
	// average time required to process the policy Validation rules on a resource
	AvgExecutionTimeGeneration string `json:"averageGenerationRulesExecutionTime"`
	// statistics per rule
	Rules []RuleStats `json:"ruleStatus`
}

//RuleStats provides status per rule
type RuleStats struct {
	// Rule name
	Name string `json:"ruleName"`
	// average time require to process the rule
	ExecutionTime string `json:"averageExecutionTime"`
	// Count of rules that were applied
	AppliedCount int `json:"appliedCount"`
	// Count of rules that failed
	ViolationCount int `json:"violationCount"`
	// Count of mutations
	MutationCount int `json:"mutationsCount"`
}

// PolicyList is a list of Policy resources

// PolicyViolation stores the information regarinding the resources for which a policy failed to apply
type PolicyViolationTemplate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              PolicyViolationSpec   `json:"spec"`
	Status            PolicyViolationStatus `json:"status"`
}

// PolicyViolationSpec describes policy behavior by its rules
type PolicyViolationSpec struct {
	Policy        string `json:"policy"`
	ResourceSpec  `json:"resource"`
	ViolatedRules []ViolatedRule `json:"rules"`
}

// ResourceSpec information to identify the resource
type ResourceSpec struct {
	Kind      string `json:"kind"`
	Namespace string `json:"namespace,omitempty"`
	Name      string `json:"name"`
}

// ViolatedRule stores the information regarding the rule
type ViolatedRule struct {
	Name            string              `json:"name"`
	Type            string              `json:"type"`
	Message         string              `json:"message"`
	ManagedResource ManagedResourceSpec `json:"managedResource,omitempty"`
}

// ManagedResourceSpec is used when the violations is created on resource owner
// to determing the kind of child resource that caused the violation
type ManagedResourceSpec struct {
	Kind string `json:"kind,omitempty"`
	// Is not used in processing, but will is present for backward compatablitiy
	Namespace       string `json:"namespace,omitempty"`
	CreationBlocked bool   `json:"creationBlocked,omitempty"`
}

//PolicyViolationStatus provides information regarding policyviolation status
// status:
//		LastUpdateTime : the time the polivy violation was updated
type PolicyViolationStatus struct {
	LastUpdateTime metav1.Time `json:"lastUpdateTime,omitempty"`
}
