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
	"github.com/maestre3d/alexandria/media-service/internal/shared/domain/exception"
	"github.com/maestre3d/alexandria/media-service/pkg/media/action"
	"github.com/maestre3d/alexandria/media-service/pkg/media/service"
	"github.com/maestre3d/alexandria/media-service/pkg/transport/pb"
	stdopentracing "github.com/opentracing/opentracing-go"
	stdzipkin "github.com/openzipkin/zipkin-go"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"strings"
)

type MediaRPCServer struct {
	create grpctransport.Handler
	list   grpctransport.Handler
	get    grpctransport.Handler
	update grpctransport.Handler
	delete grpctransport.Handler
}

func NewMediaRPCServer(svc service.IMediaService, logger log.Logger, tracer stdopentracing.Tracer, zipkinTracer *stdzipkin.Tracer) pb.MediaServer {
	duration := kitprometheus.NewSummaryFrom(stdprometheus.SummaryOpts{
		Namespace:   "alexandria",
		Subsystem:   "rpc_media_service",
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

	return &MediaRPCServer{
		create: grpctransport.NewServer(
			action.MakeCreateMediaEndpoint(svc, logger, duration, tracer, zipkinTracer),
			decodeRPCCreateRequest,
			encodeRPCCreateResponse,
			append(options, grpctransport.ServerBefore(opentracing.GRPCToContext(tracer, "Create", logger)))...,
		),
		list: grpctransport.NewServer(
			action.MakeListMediaEndpoint(svc, logger, duration, tracer, zipkinTracer),
			decodeRPCListRequest,
			encodeRPCListResponse,
			append(options, grpctransport.ServerBefore(opentracing.GRPCToContext(tracer, "List", logger)))...,
		),
		get: grpctransport.NewServer(
			action.MakeGetMediaEndpoint(svc, logger, duration, tracer, zipkinTracer),
			decodeRPCGetRequest,
			encodeRPCGetResponse,
			append(options, grpctransport.ServerBefore(opentracing.GRPCToContext(tracer, "Get", logger)))...,
		),
		update: grpctransport.NewServer(
			action.MakeUpdateMediaEndpoint(svc, logger, duration, tracer, zipkinTracer),
			decodeRPCUpdateRequest,
			encodeRPCUpdateResponse,
			append(options, grpctransport.ServerBefore(opentracing.GRPCToContext(tracer, "Update", logger)))...,
		),
		delete: grpctransport.NewServer(
			action.MakeDeleteMediaEndpoint(svc, logger, duration, tracer, zipkinTracer),
			decodeRPCDeleteRequest,
			encodeRPCDeleteResponse,
			append(options, grpctransport.ServerBefore(opentracing.GRPCToContext(tracer, "Delete", logger)))...,
		),
	}
}


/* RPC Binding/Implementations */

func (a MediaRPCServer) Create(ctx context.Context, req *pb.CreateRequest) (*pb.MediaMessage, error) {
	_, rep, err := a.create.ServeGRPC(ctx, req)
	if err != nil {
		if errors.Is(err, exception.InvalidFieldFormat) || errors.Is(err, exception.InvalidFieldRange) || errors.Is(err, exception.RequiredField) {
			errDesc := strings.Split(err.Error(), ":")
			return nil, status.Error(codes.InvalidArgument, errDesc[len(errDesc)-1])
		} else if errors.Is(err, exception.EntityExists) {
			return nil, status.Error(codes.AlreadyExists, err.Error())
		}

		return nil, err
	}
	return rep.(*pb.MediaMessage), nil
}

func (a MediaRPCServer) List(ctx context.Context, req *pb.ListRequest) (*pb.ListResponse, error) {
	_, rep, err := a.list.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return rep.(*pb.ListResponse), nil
}

func (a MediaRPCServer) Get(ctx context.Context, req *pb.GetRequest) (*pb.MediaMessage, error) {
	_, rep, err := a.get.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return rep.(*pb.MediaMessage), nil
}

func (a MediaRPCServer) Update(ctx context.Context, req *pb.UpdateRequest) (*pb.MediaMessage, error) {
	_, rep, err := a.update.ServeGRPC(ctx, req)
	if err != nil {
		if errors.Is(err, exception.InvalidFieldFormat) || errors.Is(err, exception.InvalidFieldRange) || errors.Is(err, exception.RequiredField) {
			errDesc := strings.Split(err.Error(), ":")
			return nil, status.Error(codes.InvalidArgument, errDesc[len(errDesc)-1])
		} else if errors.Is(err, exception.EntityExists) {
			return nil, status.Error(codes.AlreadyExists, err.Error())
		}

		return nil, err
	}
	return rep.(*pb.MediaMessage), nil
}

func (a MediaRPCServer) Delete(ctx context.Context, req *pb.DeleteRequest) (*pb.Empty, error) {
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
		Title:       req.Title,
		DisplayName: req.DisplayName,
		Description: req.Description,
		UserID:      req.UserID,
		AuthorID:    req.AuthorID,
		PublishDate: req.PublishDate,
		MediaType:   req.MediaType,
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
		ID:          req.Id,
		Title:       req.Title,
		DisplayName: req.DisplayName,
		Description: req.Description,
		UserID:      req.UserID,
		AuthorID:    req.AuthorID,
		PublishDate: req.PublishDate,
		MediaType:   req.MediaType,
	}, nil
}

func decodeRPCDeleteRequest(_ context.Context, rpcReq interface{}) (interface{}, error) {
	req := rpcReq.(*action.DeleteRequest)
	return action.DeleteRequest{ID: req.ID}, nil
}

func encodeRPCCreateResponse(_ context.Context, response interface{}) (interface{}, error) {
	res := response.(action.CreateResponse)
	if res.Err != nil {
		return nil, res.Err
	}

	if res.Media == nil {
		return nil, exception.EmptyBody
	}

	return &pb.MediaMessage{
		Id:          res.Media.ExternalID,
		Title:       res.Media.Title,
		DisplayName: res.Media.DisplayName,
		Description: *res.Media.Description,
		UserID:      res.Media.UserID,
		AuthorID:    res.Media.AuthorID,
		PublishDate: res.Media.PublishDate.String(),
		MediaType:   res.Media.MediaType,
		CreateTime:  res.Media.CreateTime.String(),
		UpdateTime:  res.Media.UpdateTime.String(),
	}, nil
}

func encodeRPCListResponse(_ context.Context, response interface{}) (interface{}, error) {
	res := response.(action.ListResponse)
	if res.Err != nil {
		return nil, res.Err
	}

	if len(res.Media) == 0 {
		return nil, status.Error(codes.NotFound, exception.EntitiesNotFound.Error())
	}

	mediasRPC := make([]*pb.MediaMessage, 0)
	for _, media := range res.Media {
		mediaRPC := &pb.MediaMessage{
			Id:          media.ExternalID,
			Title:       media.Title,
			DisplayName: media.DisplayName,
			Description: *media.Description,
			UserID:      media.UserID,
			AuthorID:    media.AuthorID,
			PublishDate: media.PublishDate.String(),
			MediaType:   media.MediaType,
			CreateTime:  media.CreateTime.String(),
			UpdateTime:  media.UpdateTime.String(),
		}
		mediasRPC = append(mediasRPC, mediaRPC)
	}

	return &pb.ListResponse{
		Authors:       mediasRPC,
		NextPageToken: res.NextPageToken,
	}, nil
}

func encodeRPCGetResponse(_ context.Context, response interface{}) (interface{}, error) {
	res := response.(action.GetResponse)
	if res.Err != nil {
		return nil, res.Err
	}

	if res.Media == nil {
		return nil, status.Error(codes.NotFound, exception.EntityNotFound.Error())
	}

	return &pb.MediaMessage{
		Id:          res.Media.ExternalID,
		Title:       res.Media.Title,
		DisplayName: res.Media.DisplayName,
		Description: *res.Media.Description,
		UserID:      res.Media.UserID,
		AuthorID:    res.Media.AuthorID,
		PublishDate: res.Media.PublishDate.String(),
		MediaType:   res.Media.MediaType,
		CreateTime:  res.Media.CreateTime.String(),
		UpdateTime:  res.Media.UpdateTime.String(),
	}, nil
}

func encodeRPCUpdateResponse(_ context.Context, response interface{}) (interface{}, error) {
	res := response.(action.UpdateResponse)
	if res.Err != nil {
		return nil, res.Err
	}

	if res.Media == nil {
		return nil, exception.EmptyBody
	}

	return &pb.MediaMessage{
		Id:          res.Media.ExternalID,
		Title:       res.Media.Title,
		DisplayName: res.Media.DisplayName,
		Description: *res.Media.Description,
		UserID:      res.Media.UserID,
		AuthorID:    res.Media.AuthorID,
		PublishDate: res.Media.PublishDate.String(),
		MediaType:   res.Media.MediaType,
		CreateTime:  res.Media.CreateTime.String(),
		UpdateTime:  res.Media.UpdateTime.String(),
	}, nil
}

func encodeRPCDeleteResponse(_ context.Context, response interface{}) (interface{}, error) {
	res := response.(action.DeleteResponse)
	if res.Err != nil {
		return nil, res.Err
	}

	return nil, nil
}
