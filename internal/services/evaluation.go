package services

import (
	"context"
	"log"

	"github.com/c12s/oort/internal/domain"
	"go.opentelemetry.io/otel"
)

type EvaluationService struct {
	repo domain.RHABACRepo
}

func NewEvaluationService(repo domain.RHABACRepo) (*EvaluationService, error) {
	return &EvaluationService{
		repo: repo,
	}, nil
}

func (h EvaluationService) Authorize(ctx context.Context, req domain.AuthorizationReq) domain.AuthorizationResp {
	tracer := otel.Tracer("oort.service.evaluation")
	ctx, span := tracer.Start(ctx, "EvaluationService.Authorize")
	defer span.End()

	resp := h.repo.GetPermissionHierarchy(ctx, domain.GetPermissionHierarchyReq{
		Subject:        req.Subject,
		Object:         req.Object,
		PermissionName: req.PermissionName,
	})
	if resp.Error != nil {
		return domain.AuthorizationResp{
			Authorized: false,
			Error:      resp.Error,
		}
	}

	subAttrs, err := h.getAttributes(ctx, req.Subject)
	if err != nil {
		return domain.AuthorizationResp{
			Authorized: false,
			Error:      err,
		}
	}
	objAttrs, err := h.getAttributes(ctx, req.Object)
	if err != nil {
		return domain.AuthorizationResp{
			Authorized: false,
			Error:      err,
		}
	}

	evalReq := domain.PermissionEvalRequest{
		Subject: subAttrs,
		Object:  objAttrs,
		Env:     req.Env,
	}
	evalResult := resp.Hierarchy.Eval(evalReq)

	checkResp := domain.AuthorizationResp{
		Authorized: authorized(evalResult),
		Error:      nil,
	}

	return checkResp
}

func (h EvaluationService) GetGrantedPermissions(ctx context.Context, req domain.GetGrantedPermissionsReq) domain.GetGrantedPermissionsResp {
	// dobavi sve politike koje su subjektno direktno dodeljene ili ih je nasledio
	// svaka ukljucuje naziv dozvole i objekat nad kojim vazi
	tracer := otel.Tracer("oort.service.evaluation")
	ctx, span := tracer.Start(ctx, "EvaluationService.GetGrantedPermissions")
	defer span.End()

	resp := h.repo.GetApplicablePolicies(ctx, domain.GetApplicablePoliciesReq{
		Subject: req.Subject,
	})
	if resp.Error != nil {
		return domain.GetGrantedPermissionsResp{Error: resp.Error}
	}

	granted := make([]domain.GrantedPermission, 0)

	subAttrs, err := h.getAttributes(ctx, req.Subject)
	if err != nil {
		return domain.GetGrantedPermissionsResp{Error: resp.Error}
	}
	// proveravamo nad vise objekata, svaki objekat je element u mapi
	objAttrMap := make(map[string][]domain.Attribute)

	// za svaki policy proveri da li trenutno daje dozvolu subjektu
	for _, policy := range resp.Policies {
		objAttrs, ok := objAttrMap[policy.Object.Name()]
		if !ok {
			objAttrs, err = h.getAttributes(ctx, policy.Object)
			if err != nil {
				log.Println(err)
				continue
			}
			objAttrMap[policy.Object.Name()] = objAttrs
		}

		hierarchyResp := h.repo.GetPermissionHierarchy(ctx, domain.GetPermissionHierarchyReq{
			Subject:        req.Subject,
			Object:         policy.Object,
			PermissionName: policy.PermissionName,
		})
		if hierarchyResp.Error != nil {
			log.Println(hierarchyResp.Error)
			continue
		}

		evalReq := domain.PermissionEvalRequest{
			Subject: subAttrs,
			Object:  objAttrs,
			Env:     req.Env,
		}
		evalResp := hierarchyResp.Hierarchy.Eval(evalReq)
		if authorized(evalResp) {
			granted = append(granted, domain.GrantedPermission{
				PermissionName: policy.PermissionName,
				Object:         policy.Object,
			})
		}
	}

	return domain.GetGrantedPermissionsResp{
		Permissions: granted,
		Error:       nil,
	}
}

func (h EvaluationService) getAttributes(ctx context.Context, resource domain.Resource) ([]domain.Attribute, error) {
	tracer := otel.Tracer("oort.service.evaluation")
	ctx, span := tracer.Start(ctx, "EvaluationService.getAttributes")
	defer span.End()

	res := h.repo.GetResource(ctx, domain.GetResourceReq{Resource: resource})
	if res.Error != nil {
		return nil, res.Error
	}
	return res.Resource.Attributes, nil
}

func authorized(result domain.EvalResult) bool {
	return result == domain.EvalResultAllowed
}
