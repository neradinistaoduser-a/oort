package domain

import "context"

type RHABACRepo interface {
	CreateResource(ctx context.Context, req CreateResourceReq) AdministrationResp
	DeleteResource(ctx context.Context, req DeleteResourceReq) AdministrationResp
	GetResource(ctx context.Context, req GetResourceReq) GetResourceResp
	PutAttribute(ctx context.Context, req PutAttributeReq) AdministrationResp
	DeleteAttribute(ctx context.Context, req DeleteAttributeReq) AdministrationResp
	CreateInheritanceRel(ctx context.Context, req CreateInheritanceRelReq) AdministrationResp
	DeleteInheritanceRel(ctx context.Context, req DeleteInheritanceRelReq) AdministrationResp
	CreatePolicy(ctx context.Context, req CreatePolicyReq) AdministrationResp
	DeletePolicy(ctx context.Context, req DeletePolicyReq) AdministrationResp
	GetPermissionHierarchy(ctx context.Context, req GetPermissionHierarchyReq) GetPermissionHierarchyResp
	GetApplicablePolicies(ctx context.Context, req GetApplicablePoliciesReq) GetApplicablePoliciesResp
}

type CreateResourceReq struct {
	Resource Resource
}

type DeleteResourceReq struct {
	Resource Resource
}

type GetResourceReq struct {
	Resource Resource
}

type PutAttributeReq struct {
	Resource  Resource
	Attribute Attribute
}

type DeleteAttributeReq struct {
	Resource    Resource
	AttributeId AttributeId
}

type GetAttributeReq struct {
	Resource Resource
}

type CreateInheritanceRelReq struct {
	From Resource
	To   Resource
}

type DeleteInheritanceRelReq struct {
	From Resource
	To   Resource
}

type CreatePolicyReq struct {
	SubjectScope,
	ObjectScope Resource
	Permission Permission
}

type DeletePolicyReq struct {
	SubjectScope,
	ObjectScope Resource
	Permission Permission
}

type GetPermissionHierarchyReq struct {
	Subject,
	Object Resource
	PermissionName string
}

type AdministrationResp struct {
	Error error
}

type GetAttributeResp struct {
	Attributes []Attribute
	Error      error
}

type GetResourceResp struct {
	Resource *Resource
	Error    error
}

type GetPermissionHierarchyResp struct {
	Hierarchy PermissionHierarchy
	Error     error
}

type AuthorizationReq struct {
	Subject,
	Object Resource
	PermissionName string
	Env            []Attribute
}

type AuthorizationResp struct {
	Authorized bool
	Error      error
}

type GetApplicablePoliciesReq struct {
	Subject Resource
}

type GetApplicablePoliciesResp struct {
	Policies []Policy
	Error    error
}

type Policy struct {
	PermissionName string
	Subject,
	Object Resource
}

type GetGrantedPermissionsReq struct {
	Subject Resource
	Env     []Attribute
}

type GetGrantedPermissionsResp struct {
	Permissions []GrantedPermission
	Error       error
}

type GrantedPermission struct {
	PermissionName string
	Object         Resource
}
