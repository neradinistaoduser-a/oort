package servers

import (
	"context"

	"github.com/c12s/oort/internal/mappers/proto"
	"github.com/c12s/oort/internal/services"
	"github.com/c12s/oort/pkg/api"
)

type oortAdministratorGrpcServer struct {
	api.UnimplementedOortAdministratorServer
	service services.AdministrationService
}

func NewOortAdministratorGrpcServer(service services.AdministrationService) (api.OortAdministratorServer, error) {
	return &oortAdministratorGrpcServer{
		service: service,
	}, nil
}

func (o *oortAdministratorGrpcServer) CreateResource(ctx context.Context, req *api.CreateResourceReq) (*api.AdministrationResp, error) {
	request, err := proto.CreateResourceReqToDomain(req)
	if err != nil {
		return nil, err
	}
	resp := o.service.CreateResource(ctx, *request)
	return &api.AdministrationResp{}, resp.Error
}

func (o *oortAdministratorGrpcServer) DeleteResource(ctx context.Context, req *api.DeleteResourceReq) (*api.AdministrationResp, error) {
	request, err := proto.DeleteResourceReqToDomain(req)
	if err != nil {
		return nil, err
	}
	resp := o.service.DeleteResource(ctx, *request)
	return &api.AdministrationResp{}, resp.Error
}

func (o *oortAdministratorGrpcServer) CreateInheritanceRel(ctx context.Context, req *api.CreateInheritanceRelReq) (*api.AdministrationResp, error) {
	request, err := proto.CreateInheritanceRelReqToDomain(req)
	if err != nil {
		return nil, err
	}
	resp := o.service.CreateInheritanceRel(ctx, *request)
	return &api.AdministrationResp{}, resp.Error
}

func (o *oortAdministratorGrpcServer) DeleteInheritanceRel(ctx context.Context, req *api.DeleteInheritanceRelReq) (*api.AdministrationResp, error) {
	request, err := proto.DeleteInheritanceRelReqToDomain(req)
	if err != nil {
		return nil, err
	}
	resp := o.service.DeleteInheritanceRel(ctx, *request)
	return &api.AdministrationResp{}, resp.Error
}

func (o *oortAdministratorGrpcServer) PutAttribute(ctx context.Context, req *api.PutAttributeReq) (*api.AdministrationResp, error) {
	request, err := proto.PutAttributeReqToDomain(req)
	if err != nil {
		return nil, err
	}
	resp := o.service.PutAttribute(ctx, *request)
	return &api.AdministrationResp{}, resp.Error
}

func (o *oortAdministratorGrpcServer) DeleteAttribute(ctx context.Context, req *api.DeleteAttributeReq) (*api.AdministrationResp, error) {
	request, err := proto.DeleteAttributeReqToDomain(req)
	if err != nil {
		return nil, err
	}
	resp := o.service.DeleteAttribute(ctx, *request)
	return &api.AdministrationResp{}, resp.Error
}

func (o *oortAdministratorGrpcServer) CreatePolicy(ctx context.Context, req *api.CreatePolicyReq) (*api.AdministrationResp, error) {
	request, err := proto.CreatePolicyReqToDomain(req)
	if err != nil {
		return nil, err
	}
	resp := o.service.CreatePolicy(ctx, *request)
	return &api.AdministrationResp{}, resp.Error
}

func (o *oortAdministratorGrpcServer) DeletePolicy(ctx context.Context, req *api.DeletePolicyReq) (*api.AdministrationResp, error) {
	request, err := proto.DeletePolicyReqToDomain(req)
	if err != nil {
		return nil, err
	}
	resp := o.service.DeletePolicy(ctx, *request)
	return &api.AdministrationResp{}, resp.Error
}
