package neo4j

import (
	"context"
	"errors"

	"github.com/c12s/oort/internal/domain"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
	"go.opentelemetry.io/otel"
)

type RHABACRepo struct {
	manager *TransactionManager
	factory CypherFactory
}

func NewRHABACRepo(manager *TransactionManager, factory CypherFactory) domain.RHABACRepo {
	return RHABACRepo{
		manager: manager,
		factory: factory,
	}
}

func (store RHABACRepo) CreateResource(ctx context.Context, req domain.CreateResourceReq) domain.AdministrationResp {
	tracer := otel.Tracer("oort.neo4j.repo")
	ctx, span := tracer.Start(ctx, "RHABACRepo.CreateResource")
	defer span.End()
	cypher, params := store.factory.createResource(req)
	err := store.manager.WriteTransaction(ctx, cypher, params)
	return domain.AdministrationResp{Error: err}
}

func (store RHABACRepo) DeleteResource(ctx context.Context, req domain.DeleteResourceReq) domain.AdministrationResp {
	tracer := otel.Tracer("oort.neo4j.repo")
	ctx, span := tracer.Start(ctx, "RHABACRepo.DeleteResource")
	defer span.End()
	cypher, params := store.factory.deleteResource(req)
	err := store.manager.WriteTransaction(ctx, cypher, params)
	return domain.AdministrationResp{Error: err}
}

func (store RHABACRepo) GetResource(ctx context.Context, req domain.GetResourceReq) domain.GetResourceResp {
	tracer := otel.Tracer("oort.neo4j.repo")
	ctx, span := tracer.Start(ctx, "RHABACRepo.GetResource")
	defer span.End()
	cypher, params := store.factory.getResource(req)
	records, err := store.manager.ReadTransaction(ctx, cypher, params)
	if err != nil {
		return domain.GetResourceResp{Resource: nil, Error: err}
	}

	recordList, ok := records.([]*neo4j.Record)
	if !ok {
		return domain.GetResourceResp{Error: errors.New("invalid resp format")}
	}
	if len(recordList) == 0 {
		return domain.GetResourceResp{Error: errors.New("resource not found")}
	}
	return domain.GetResourceResp{Resource: getResource(records), Error: nil}
}

func (store RHABACRepo) PutAttribute(ctx context.Context, req domain.PutAttributeReq) domain.AdministrationResp {
	tracer := otel.Tracer("oort.neo4j.repo")
	ctx, span := tracer.Start(ctx, "RHABACRepo.PutAttribute")
	defer span.End()
	cypher, params := store.factory.putAttribute(req)
	err := store.manager.WriteTransaction(ctx, cypher, params)
	return domain.AdministrationResp{Error: err}
}

func (store RHABACRepo) DeleteAttribute(ctx context.Context, req domain.DeleteAttributeReq) domain.AdministrationResp {
	tracer := otel.Tracer("oort.neo4j.repo")
	ctx, span := tracer.Start(ctx, "RHABACRepo.DeleteAttribute")
	defer span.End()
	cypher, params := store.factory.deleteAttribute(req)
	err := store.manager.WriteTransaction(ctx, cypher, params)
	return domain.AdministrationResp{Error: err}
}

func (store RHABACRepo) CreateInheritanceRel(ctx context.Context, req domain.CreateInheritanceRelReq) domain.AdministrationResp {
	tracer := otel.Tracer("oort.neo4j.repo")
	ctx, span := tracer.Start(ctx, "RHABACRepo.CreateInheritanceRel")
	defer span.End()
	cypher, params := store.factory.createInheritanceRel(req)
	err := store.manager.WriteTransaction(ctx, cypher, params)
	return domain.AdministrationResp{Error: err}
}

func (store RHABACRepo) DeleteInheritanceRel(ctx context.Context, req domain.DeleteInheritanceRelReq) domain.AdministrationResp {
	tracer := otel.Tracer("oort.neo4j.repo")
	ctx, span := tracer.Start(ctx, "RHABACRepo.DeleteInheritanceRel")
	defer span.End()
	cypher, params := store.factory.deleteInheritanceRel(req)
	err := store.manager.WriteTransaction(ctx, cypher, params)
	return domain.AdministrationResp{Error: err}
}

func (store RHABACRepo) CreatePolicy(ctx context.Context, req domain.CreatePolicyReq) domain.AdministrationResp {
	tracer := otel.Tracer("oort.neo4j.repo")
	ctx, span := tracer.Start(ctx, "RHABACRepo.CreatePolicy")
	defer span.End()
	cypher, params := store.factory.createPolicy(req)
	err := store.manager.WriteTransaction(ctx, cypher, params)
	return domain.AdministrationResp{Error: err}
}

func (store RHABACRepo) DeletePolicy(ctx context.Context, req domain.DeletePolicyReq) domain.AdministrationResp {
	tracer := otel.Tracer("oort.neo4j.repo")
	ctx, span := tracer.Start(ctx, "RHABACRepo.DeletePolicy")
	defer span.End()
	cypher, params := store.factory.deletePolicy(req)
	err := store.manager.WriteTransaction(ctx, cypher, params)
	return domain.AdministrationResp{Error: err}
}

func (store RHABACRepo) GetPermissionHierarchy(ctx context.Context, req domain.GetPermissionHierarchyReq) domain.GetPermissionHierarchyResp {
	tracer := otel.Tracer("oort.neo4j.repo")
	ctx, span := tracer.Start(ctx, "RHABACRepo.GetPermissionHierarchy")
	defer span.End()
	cypher, params := store.factory.getEffectivePermissionsWithPriority(req)
	records, err := store.manager.ReadTransaction(ctx, cypher, params)
	if err != nil {
		return domain.GetPermissionHierarchyResp{Hierarchy: nil, Error: err}
	}

	hierarchy, err := getHierarchy(records)
	return domain.GetPermissionHierarchyResp{Hierarchy: hierarchy, Error: err}
}

func (store RHABACRepo) GetApplicablePolicies(ctx context.Context, req domain.GetApplicablePoliciesReq) domain.GetApplicablePoliciesResp {
	tracer := otel.Tracer("oort.neo4j.repo")
	ctx, span := tracer.Start(ctx, "RHABACRepo.GetApplicablePolicies")
	defer span.End()
	cypher, params := store.factory.getApplicablePolicies(req)
	records, err := store.manager.ReadTransaction(ctx, cypher, params)
	if err != nil {
		return domain.GetApplicablePoliciesResp{Policies: nil, Error: err}
	}
	policies, err := getPolicies(records)
	return domain.GetApplicablePoliciesResp{Policies: policies, Error: err}
}
