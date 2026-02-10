package services

import (
	"context"

	"github.com/c12s/oort/internal/domain"
)

type AdministrationService struct {
	repo domain.RHABACRepo
}

func NewAdministrationService(repo domain.RHABACRepo) (*AdministrationService, error) {
	return &AdministrationService{
		repo: repo,
	}, nil
}

func (h AdministrationService) CreateResource(ctx context.Context, req domain.CreateResourceReq) domain.AdministrationResp {
	return h.repo.CreateResource(ctx, req)
}

func (h AdministrationService) DeleteResource(ctx context.Context, req domain.DeleteResourceReq) domain.AdministrationResp {
	return h.repo.DeleteResource(ctx, req)
}

func (h AdministrationService) PutAttribute(ctx context.Context, req domain.PutAttributeReq) domain.AdministrationResp {
	return h.repo.PutAttribute(ctx, req)
}

func (h AdministrationService) DeleteAttribute(ctx context.Context, req domain.DeleteAttributeReq) domain.AdministrationResp {
	return h.repo.DeleteAttribute(ctx, req)
}

func (h AdministrationService) CreateInheritanceRel(ctx context.Context, req domain.CreateInheritanceRelReq) domain.AdministrationResp {
	return h.repo.CreateInheritanceRel(ctx, req)
}

func (h AdministrationService) DeleteInheritanceRel(ctx context.Context, req domain.DeleteInheritanceRelReq) domain.AdministrationResp {
	return h.repo.DeleteInheritanceRel(ctx, req)
}

func (h AdministrationService) CreatePolicy(ctx context.Context, req domain.CreatePolicyReq) domain.AdministrationResp {
	if req.SubjectScope.Name() == "" {
		req.SubjectScope = domain.RootResource
	}
	if req.ObjectScope.Name() == "" {
		req.ObjectScope = domain.RootResource
	}
	return h.repo.CreatePolicy(ctx, req)
}

func (h AdministrationService) DeletePolicy(ctx context.Context, req domain.DeletePolicyReq) domain.AdministrationResp {
	if req.SubjectScope.Name() == "" {
		req.SubjectScope = domain.RootResource
	}
	if req.ObjectScope.Name() == "" {
		req.ObjectScope = domain.RootResource
	}
	return h.repo.DeletePolicy(ctx, req)
}
