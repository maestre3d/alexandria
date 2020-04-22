package handler

import (
	"context"
	"errors"
	"github.com/go-kit/kit/log"
	kitprometheus "github.com/go-kit/kit/metrics/prometheus"
	"github.com/go-kit/kit/tracing/opentracing"
	"github.com/go-kit/kit/tracing/zipkin"
	"github.com/go-kit/kit/transport"
	grpctransport "github.com/go-kit/kit/transport/grpc"
	"github.com/maestre3d/alexandria/author-service/internal/shared/domain/exception"
	"github.com/maestre3d/alexandria/author-service/pkg/author/action"
	"github.com/maestre3d/alexandria/author-service/pkg/author/service"
	"github.com/maestre3d/alexandria/author-service/pkg/transport/pb"
	stdopentracing "github.com/opentracing/opentracing-go"
	stdzipkin "github.com/openzipkin/zipkin-go"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
	"strings"
)

type AuthorRPCServer struct {
	create grpctransport.Handler
	list   grpctransport.Handler
	get    grpctransport.Handler
	update grpctransport.Handler
	delete grpctransport.Handler
}

func NewAuthorRPCServer(svc service.IAuthorService, logger log.Logger, tracer stdopentracing.Tracer, zipkinTracer *stdzipkin.Tracer) pb.AuthorServer {
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
		// provided operation name or a global tracing service can be instantiated
		// without an operation name and fed to each Go kit endpoint as ServerOption.
		// In the latter case, the operation name will be the endpoint's http method.
		// We demonstrate a global tracing service here.
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
	}
}

/* RPC Implementations */

func (a AuthorRPCServer) Create(ctx context.Context, req *pb.CreateRequest) (*pb.AuthorMessage, error) {
	_, rep, err := a.create.ServeGRPC(ctx, req)
	if err != nil {
		if errors.Is(err, exception.InvalidFieldFormat) || errors.Is(err, exception.InvalidFieldRange) || errors.Is(err, exception.RequiredField) {
			errDesc := strings.Split(err.Error(), ":")
			return nil, errors.New(errDesc[len(errDesc)-1])
		}

		return nil, err
	}
	return rep.(*pb.AuthorMessage), nil
}

func (a AuthorRPCServer) List(ctx context.Context, req *pb.ListRequest) (*pb.ListResponse, error) {
	_, rep, err := a.list.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return rep.(*pb.ListResponse), nil
}

func (a AuthorRPCServer) Get(ctx context.Context, req *pb.GetRequest) (*pb.AuthorMessage, error) {
	_, rep, err := a.get.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return rep.(*pb.AuthorMessage), nil
}

func (a AuthorRPCServer) Update(ctx context.Context, req *pb.UpdateRequest) (*pb.AuthorMessage, error) {
	_, rep, err := a.update.ServeGRPC(ctx, req)
	if err != nil {
		if errors.Is(err, exception.InvalidFieldFormat) || errors.Is(err, exception.InvalidFieldRange) || errors.Is(err, exception.RequiredField) {
			errDesc := strings.Split(err.Error(), ":")
			return nil, errors.New(errDesc[len(errDesc)-1])
		}

		return nil, err
	}
	return rep.(*pb.AuthorMessage), nil
}

func (a AuthorRPCServer) Delete(ctx context.Context, req *pb.DeleteRequest) (*pb.Empty, error) {
	_, rep, err := a.delete.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return rep.(*pb.Empty), nil
}

/* Encoders/Decoders */
func decodeRPCCreateRequest(_ context.Context, rpcReq interface{}) (interface{}, error) {
	req := rpcReq.(*pb.CreateRequest)
	return action.CreateRequest{
		FirstName:   req.FirstName,
		LastName:    req.LastName,
		DisplayName: req.DisplayName,
		BirthDate:   req.BirthDate,
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
	return action.GetRequest{req.Id}, nil
}

func decodeRPCUpdateRequest(_ context.Context, rpcReq interface{}) (interface{}, error) {
	req := rpcReq.(*pb.UpdateRequest)
	return action.UpdateRequest{
		ID:          req.Id,
		FirstName:   req.FirstName,
		LastName:    req.LastName,
		DisplayName: req.DisplayName,
		BirthDate:   req.BirthDate,
	}, nil
}

func decodeRPCDeleteRequest(_ context.Context, rpcReq interface{}) (interface{}, error) {
	req := rpcReq.(*action.DeleteRequest)
	return action.DeleteRequest{req.ID}, nil
}

func encodeRPCCreateResponse(_ context.Context, response interface{}) (interface{}, error) {
	r := response.(action.CreateResponse)
	if r.Err != nil {
		if errors.Is(r.Err, exception.InvalidFieldFormat) || errors.Is(r.Err, exception.InvalidFieldRange) || errors.Is(r.Err, exception.RequiredField) {
			errDesc := strings.Split(r.Err.Error(), ":")
			return nil, errors.New(errDesc[len(errDesc)-1])
		}

		return nil, r.Err
	}

	if r.Author == nil {
		return nil, exception.EmptyBody
	}

	return &pb.AuthorMessage{
		Id:          r.Author.ExternalID,
		FirstName:   r.Author.FirstName,
		LastName:    r.Author.LastName,
		DisplayName: r.Author.DisplayName,
		BirthDate:   r.Author.BirthDate.String(),
		CreateTime:  r.Author.CreateTime.String(),
		UpdateTime:  r.Author.UpdateTime.String(),
	}, nil
}

func encodeRPCListResponse(_ context.Context, response interface{}) (interface{}, error) {
	resp := response.(action.ListResponse)
	if resp.Err != nil {
		return nil, resp.Err
	}

	if len(resp.Authors) == 0 {
		return nil, exception.EntitiesNotFound
	}

	authorsRPC := make([]*pb.AuthorMessage, 0)
	for _, author := range resp.Authors {
		authorRPC := &pb.AuthorMessage{
			Id:          author.ExternalID,
			FirstName:   author.FirstName,
			LastName:    author.LastName,
			DisplayName: author.DisplayName,
			BirthDate:   author.BirthDate.String(),
			CreateTime:  author.CreateTime.String(),
			UpdateTime:  author.UpdateTime.String(),
		}
		authorsRPC = append(authorsRPC, authorRPC)
	}

	return &pb.ListResponse{
		Authors:       authorsRPC,
		NextPageToken: resp.NextPageToken,
	}, nil
}

func encodeRPCGetResponse(_ context.Context, response interface{}) (interface{}, error) {
	resp := response.(action.GetResponse)
	if resp.Err != nil {
		return nil, resp.Err
	}

	if resp.Author == nil {
		return nil, exception.EntityNotFound
	}

	return &pb.AuthorMessage{
		Id:          resp.Author.ExternalID,
		FirstName:   resp.Author.FirstName,
		LastName:    resp.Author.LastName,
		DisplayName: resp.Author.DisplayName,
		BirthDate:   resp.Author.BirthDate.String(),
		CreateTime:  resp.Author.CreateTime.String(),
		UpdateTime:  resp.Author.UpdateTime.String(),
	}, nil
}

func encodeRPCUpdateResponse(_ context.Context, response interface{}) (interface{}, error) {
	resp := response.(action.UpdateResponse)
	if resp.Err != nil {
		if errors.Is(resp.Err, exception.InvalidFieldFormat) || errors.Is(resp.Err, exception.InvalidFieldRange) || errors.Is(resp.Err, exception.RequiredField) {
			errDesc := strings.Split(resp.Err.Error(), ":")
			return nil, errors.New(errDesc[len(errDesc)-1])
		}

		return nil, resp.Err
	}

	if resp.Author == nil {
		return nil, exception.EmptyBody
	}

	return &pb.AuthorMessage{
		Id:          resp.Author.ExternalID,
		FirstName:   resp.Author.FirstName,
		LastName:    resp.Author.LastName,
		DisplayName: resp.Author.DisplayName,
		BirthDate:   resp.Author.BirthDate.String(),
		CreateTime:  resp.Author.CreateTime.String(),
		UpdateTime:  resp.Author.UpdateTime.String(),
	}, nil
}

func encodeRPCDeleteResponse(_ context.Context, response interface{}) (interface{}, error) {
	resp := response.(action.DeleteResponse)
	if resp.Err != nil {
		return nil, resp.Err
	}

	return nil, nil
}
