package bind

import (
	"context"
	"github.com/alexandria-oss/core/exception"
	"github.com/alexandria-oss/core/grpcutil"
	"github.com/go-kit/kit/log"
	kitprometheus "github.com/go-kit/kit/metrics/prometheus"
	"github.com/go-kit/kit/tracing/opentracing"
	"github.com/go-kit/kit/tracing/zipkin"
	"github.com/go-kit/kit/transport"
	grpctransport "github.com/go-kit/kit/transport/grpc"
	"github.com/maestre3d/alexandria/author-service/internal/domain"
	"github.com/maestre3d/alexandria/author-service/pkg/author/action"
	"github.com/maestre3d/alexandria/author-service/pkg/author/usecase"
	"github.com/maestre3d/alexandria/author-service/pkg/transport/pb"
	stdopentracing "github.com/opentracing/opentracing-go"
	stdzipkin "github.com/openzipkin/zipkin-go"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuthorRPCServer struct {
	create     grpctransport.Handler
	list       grpctransport.Handler
	get        grpctransport.Handler
	update     grpctransport.Handler
	delete     grpctransport.Handler
	restore    grpctransport.Handler
	hardDelete grpctransport.Handler
}

func NewAuthorRPC(svc usecase.AuthorInteractor, logger log.Logger, tracer stdopentracing.Tracer, zipkinTracer *stdzipkin.Tracer) pb.AuthorServer {
	duration := kitprometheus.NewSummaryFrom(stdprometheus.SummaryOpts{
		Namespace:   "alexandria",
		Subsystem:   "rpc_author_service",
		Name:        "request_duration_seconds",
		Help:        "total duration of requests in microseconds",
		ConstLabels: nil,
		Objectives:  nil,
		MaxAge:      0,
		AgeBuckets:  0,
		BufCap:      0,
	}, []string{"method", "success"})

	options := []grpctransport.ServerOption{
		grpctransport.ServerErrorHandler(transport.NewLogErrorHandler(logger)),
	}

	if zipkinTracer != nil {
		// Zipkin HTTP Server Trace can either be instantiated per endpoint with a
		// provided operation name or a global tracing usecase can be instantiated
		// without an operation name and fed to each Go kit endpoint as ServerOption.
		// In the latter case, the operation name will be the endpoint's http method.
		// We demonstrate a global tracing usecase here.
		options = append(options, zipkin.GRPCServerTrace(zipkinTracer))
	}

	return &AuthorRPCServer{
		create: grpctransport.NewServer(
			action.MakeCreateAuthorEndpoint(svc, logger, duration, tracer, zipkinTracer),
			decodeRPCCreateRequest,
			encodeRPCCreateResponse,
			append(options, grpctransport.ServerBefore(opentracing.GRPCToContext(tracer, "Create", logger)))...,
		),
		list: grpctransport.NewServer(
			action.MakeListAuthorEndpoint(svc, logger, duration, tracer, zipkinTracer),
			decodeRPCListRequest,
			encodeRPCListResponse,
			append(options, grpctransport.ServerBefore(opentracing.GRPCToContext(tracer, "List", logger)))...,
		),
		get: grpctransport.NewServer(
			action.MakeGetAuthorEndpoint(svc, logger, duration, tracer, zipkinTracer),
			decodeRPCGetRequest,
			encodeRPCGetResponse,
			append(options, grpctransport.ServerBefore(opentracing.GRPCToContext(tracer, "Get", logger)))...,
		),
		update: grpctransport.NewServer(
			action.MakeUpdateAuthorEndpoint(svc, logger, duration, tracer, zipkinTracer),
			decodeRPCUpdateRequest,
			encodeRPCUpdateResponse,
			append(options, grpctransport.ServerBefore(opentracing.GRPCToContext(tracer, "Update", logger)))...,
		),
		delete: grpctransport.NewServer(
			action.MakeDeleteAuthorEndpoint(svc, logger, duration, tracer, zipkinTracer),
			decodeRPCDeleteRequest,
			encodeRPCDeleteResponse,
			append(options, grpctransport.ServerBefore(opentracing.GRPCToContext(tracer, "Delete", logger)))...,
		),
		restore: grpctransport.NewServer(
			action.MakeRestoreAuthorEndpoint(svc, logger, duration, tracer, zipkinTracer),
			decodeRPCRestoreRequest,
			encodeRPCRestoreResponse,
			append(options, grpctransport.ServerBefore(opentracing.GRPCToContext(tracer, "Restore", logger)))...,
		),
		hardDelete: grpctransport.NewServer(
			action.MakeHardDeleteAuthorEndpoint(svc, logger, duration, tracer, zipkinTracer),
			decodeRPCHardDeleteRequest,
			encodeRPCHardDeleteResponse,
			append(options, grpctransport.ServerBefore(opentracing.GRPCToContext(tracer, "HardDelete", logger)))...,
		),
	}
}

/* RPC Action Binding/Implementations */

func (a AuthorRPCServer) Create(ctx context.Context, req *pb.CreateRequest) (*pb.AuthorMessage, error) {
	_, rep, err := a.create.ServeGRPC(ctx, req)
	if err != nil {
		return nil, grpcutil.ResponseErr(err)
	}
	return rep.(*pb.AuthorMessage), nil
}

func (a AuthorRPCServer) List(ctx context.Context, req *pb.ListRequest) (*pb.ListResponse, error) {
	_, rep, err := a.list.ServeGRPC(ctx, req)
	if err != nil {
		return nil, grpcutil.ResponseErr(err)
	}
	return rep.(*pb.ListResponse), nil
}

func (a AuthorRPCServer) Get(ctx context.Context, req *pb.GetRequest) (*pb.AuthorMessage, error) {
	_, rep, err := a.get.ServeGRPC(ctx, req)
	if err != nil {
		return nil, grpcutil.ResponseErr(err)
	}
	return rep.(*pb.AuthorMessage), nil
}

func (a AuthorRPCServer) Update(ctx context.Context, req *pb.UpdateRequest) (*pb.AuthorMessage, error) {
	_, rep, err := a.update.ServeGRPC(ctx, req)
	if err != nil {
		return nil, grpcutil.ResponseErr(err)
	}
	return rep.(*pb.AuthorMessage), nil
}

func (a AuthorRPCServer) Delete(ctx context.Context, req *pb.DeleteRequest) (*pb.Empty, error) {
	_, rep, err := a.delete.ServeGRPC(ctx, req)
	if err != nil {
		return nil, grpcutil.ResponseErr(err)
	}
	return rep.(*pb.Empty), nil
}

func (a AuthorRPCServer) Restore(ctx context.Context, req *pb.RestoreRequest) (*pb.Empty, error) {
	_, rep, err := a.restore.ServeGRPC(ctx, req)
	if err != nil {
		return nil, grpcutil.ResponseErr(err)
	}
	return rep.(*pb.Empty), nil
}

func (a AuthorRPCServer) HardDelete(ctx context.Context, req *pb.HardDeleteRequest) (*pb.Empty, error) {
	_, rep, err := a.hardDelete.ServeGRPC(ctx, req)
	if err != nil {
		return nil, grpcutil.ResponseErr(err)
	}
	return rep.(*pb.Empty), nil
}

/* Decoders */
func decodeRPCCreateRequest(_ context.Context, rpcReq interface{}) (interface{}, error) {
	req := rpcReq.(*pb.CreateRequest)
	return action.CreateRequest{
		FirstName:     req.FirstName,
		LastName:      req.LastName,
		DisplayName:   req.DisplayName,
		OwnershipType: req.OwnershipType,
		OwnerID:       req.OwnerID,
	}, nil
}

func decodeRPCListRequest(_ context.Context, rpcReq interface{}) (interface{}, error) {
	req := rpcReq.(*pb.ListRequest)
	return action.ListRequest{
		PageToken:    req.PageToken,
		PageSize:     req.PageSize,
		FilterParams: req.FilterParams,
	}, nil
}

func decodeRPCGetRequest(_ context.Context, rpcReq interface{}) (interface{}, error) {
	req := rpcReq.(*pb.GetRequest)
	return action.GetRequest{ID: req.Id}, nil
}

func decodeRPCUpdateRequest(_ context.Context, rpcReq interface{}) (interface{}, error) {
	req := rpcReq.(*pb.UpdateRequest)
	return action.UpdateRequest{
		ID:            req.Id,
		FirstName:     req.FirstName,
		LastName:      req.LastName,
		DisplayName:   req.DisplayName,
		OwnershipType: req.OwnershipType,
		OwnerID:       req.OwnerID,
		Verified:      req.Verified,
		Picture:       req.Picture,
		Owners:        parseOwnerMessages(req.Owners),
	}, nil
}

func decodeRPCDeleteRequest(_ context.Context, rpcReq interface{}) (interface{}, error) {
	req := rpcReq.(*action.DeleteRequest)
	return action.DeleteRequest{ID: req.ID}, nil
}

func decodeRPCRestoreRequest(_ context.Context, rpcReq interface{}) (interface{}, error) {
	req := rpcReq.(*action.RestoreRequest)
	return action.RestoreRequest{req.ID}, nil
}

func decodeRPCHardDeleteRequest(_ context.Context, rpcReq interface{}) (interface{}, error) {
	req := rpcReq.(*action.HardDeleteRequest)
	return action.HardDeleteRequest{req.ID}, nil
}

/* Encoders */

func encodeRPCCreateResponse(_ context.Context, response interface{}) (interface{}, error) {
	res := response.(action.CreateResponse)
	if res.Err != nil {
		return nil, res.Err
	}

	if res.Author == nil {
		return nil, exception.EmptyBody
	}

	return &pb.AuthorMessage{
		Id:            res.Author.ExternalID,
		FirstName:     res.Author.FirstName,
		LastName:      res.Author.LastName,
		DisplayName:   res.Author.DisplayName,
		OwnershipType: res.Author.OwnershipType,
		CreateTime:    res.Author.CreateTime.String(),
		UpdateTime:    res.Author.UpdateTime.String(),
		DeleteTime:    res.Author.DeleteTime.String(),
		Active:        res.Author.Active,
		Verified:      res.Author.Verified,
		Picture:       *res.Author.Picture,
		TotalViews:    res.Author.TotalViews,
		Owners:        parseOwners(res.Author.Owners),
	}, nil
}

func encodeRPCListResponse(_ context.Context, response interface{}) (interface{}, error) {
	res := response.(action.ListResponse)
	if res.Err != nil {
		return nil, res.Err
	}

	if len(res.Authors) == 0 {
		return nil, status.Error(codes.NotFound, exception.EntitiesNotFound.Error())
	}

	authorsRPC := make([]*pb.AuthorMessage, 0)
	for _, author := range res.Authors {
		authorRPC := &pb.AuthorMessage{
			Id:            author.ExternalID,
			FirstName:     author.FirstName,
			LastName:      author.LastName,
			DisplayName:   author.DisplayName,
			OwnershipType: author.OwnershipType,
			CreateTime:    author.CreateTime.String(),
			UpdateTime:    author.UpdateTime.String(),
			DeleteTime:    author.DeleteTime.String(),
			Active:        author.Active,
			Verified:      author.Verified,
			Picture:       *author.Picture,
			TotalViews:    author.TotalViews,
			Owners:        parseOwners(author.Owners),
		}
		authorsRPC = append(authorsRPC, authorRPC)
	}

	return &pb.ListResponse{
		Authors:       authorsRPC,
		NextPageToken: res.NextPageToken,
	}, nil
}

func encodeRPCGetResponse(_ context.Context, response interface{}) (interface{}, error) {
	res := response.(action.GetResponse)
	if res.Err != nil {
		return nil, res.Err
	}

	if res.Author == nil {
		return nil, status.Error(codes.NotFound, exception.EntityNotFound.Error())
	}

	return &pb.AuthorMessage{
		Id:            res.Author.ExternalID,
		FirstName:     res.Author.FirstName,
		LastName:      res.Author.LastName,
		DisplayName:   res.Author.DisplayName,
		OwnershipType: res.Author.OwnershipType,
		CreateTime:    res.Author.CreateTime.String(),
		UpdateTime:    res.Author.UpdateTime.String(),
		DeleteTime:    res.Author.DeleteTime.String(),
		Active:        res.Author.Active,
		Verified:      res.Author.Verified,
		Picture:       *res.Author.Picture,
		TotalViews:    res.Author.TotalViews,
		Owners:        parseOwners(res.Author.Owners),
	}, nil
}

func encodeRPCUpdateResponse(_ context.Context, response interface{}) (interface{}, error) {
	res := response.(action.UpdateResponse)
	if res.Err != nil {
		return nil, res.Err
	}

	if res.Author == nil {
		return nil, exception.EmptyBody
	}

	return &pb.AuthorMessage{
		Id:            res.Author.ExternalID,
		FirstName:     res.Author.FirstName,
		LastName:      res.Author.LastName,
		DisplayName:   res.Author.DisplayName,
		OwnershipType: res.Author.OwnershipType,
		CreateTime:    res.Author.CreateTime.String(),
		UpdateTime:    res.Author.UpdateTime.String(),
		DeleteTime:    res.Author.DeleteTime.String(),
		Active:        res.Author.Active,
		Verified:      res.Author.Verified,
		Picture:       *res.Author.Picture,
		TotalViews:    res.Author.TotalViews,
		Owners:        parseOwners(res.Author.Owners),
	}, nil
}

func encodeRPCDeleteResponse(_ context.Context, response interface{}) (interface{}, error) {
	res := response.(action.DeleteResponse)
	if res.Err != nil {
		return nil, res.Err
	}

	return nil, nil
}

func encodeRPCRestoreResponse(_ context.Context, response interface{}) (interface{}, error) {
	res := response.(action.RestoreResponse)
	if res.Err != nil {
		return nil, res.Err
	}
	return nil, nil
}

func encodeRPCHardDeleteResponse(_ context.Context, response interface{}) (interface{}, error) {
	res := response.(action.HardDeleteResponse)
	if res.Err != nil {
		return nil, res.Err
	}
	return nil, nil
}

// Parse domain owners to rpc message
func parseOwners(owners []*domain.Owner) []*pb.OwnerMessage {
	// Parse domain to rpc
	ownersRPC := make([]*pb.OwnerMessage, 0)
	for _, owner := range owners {
		ownerMsg := &pb.OwnerMessage{
			Id:   owner.ID,
			Role: owner.Role,
		}

		ownersRPC = append(ownersRPC, ownerMsg)
	}

	return ownersRPC
}

// Parse rpc owners to domain message
func parseOwnerMessages(owners []*pb.OwnerMessage) []*domain.Owner {
	// Parse rpc to domain
	ownersDomain := make([]*domain.Owner, 0)
	for _, owner := range owners {
		ownerMsg := &domain.Owner{
			ID:   owner.Id,
			Role: owner.Role,
		}

		ownersDomain = append(ownersDomain, ownerMsg)
	}

	return ownersDomain
}
