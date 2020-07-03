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
	"github.com/maestre3d/alexandria/media-service/pb"
	"github.com/maestre3d/alexandria/media-service/pkg/media/action"
	"github.com/maestre3d/alexandria/media-service/pkg/media/usecase"
	stdopentracing "github.com/opentracing/opentracing-go"
	stdzipkin "github.com/openzipkin/zipkin-go"
	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type MediaRPCServer struct {
	srv pb.MediaServer
}

// Compile-time RPC implementation
type mediaRPCImp struct {
	create     grpctransport.Handler
	list       grpctransport.Handler
	get        grpctransport.Handler
	update     grpctransport.Handler
	delete     grpctransport.Handler
	restore    grpctransport.Handler
	hardDelete grpctransport.Handler
}

func NewMediaRPC(svc usecase.MediaInteractor, logger log.Logger, tracer stdopentracing.Tracer, zipkinTracer *stdzipkin.Tracer) *MediaRPCServer {
	duration := kitprometheus.NewSummaryFrom(prometheus.SummaryOpts{
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
		// provided operation name or a global tracing usecase can be instantiated
		// without an operation name and fed to each Go kit endpoint as ServerOption.
		// In the latter case, the operation name will be the endpoint's http method.
		// We demonstrate a global tracing usecase here.
		options = append(options, zipkin.GRPCServerTrace(zipkinTracer, zipkin.Logger(logger), zipkin.Name("media_service"),
			zipkin.AllowPropagation(true)))
	}

	srv := mediaRPCImp{
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
		restore: grpctransport.NewServer(
			action.MakeRestoreMediaEndpoint(svc, logger, duration, tracer, zipkinTracer),
			decodeRPCRestoreRequest,
			encodeRPCRestoreResponse,
			append(options, grpctransport.ServerBefore(opentracing.GRPCToContext(tracer, "Restore", logger)))...,
		),
		hardDelete: grpctransport.NewServer(
			action.MakeHardDeleteMediaEndpoint(svc, logger, duration, tracer, zipkinTracer),
			decodeRPCHardDeleteRequest,
			encodeRPCHardDeleteResponse,
			append(options, grpctransport.ServerBefore(opentracing.GRPCToContext(tracer, "HardDelete", logger)))...,
		),
	}

	return &MediaRPCServer{srv}
}

func (a MediaRPCServer) SetRoutes(srv *grpc.Server) {
	pb.RegisterMediaServer(srv, a.srv)
}

/* RPC Action Binding/Implementations */

func (a mediaRPCImp) Create(ctx context.Context, req *pb.MediaCreateRequest) (*pb.MediaMessage, error) {
	_, rep, err := a.create.ServeGRPC(ctx, req)
	if err != nil {
		return nil, grpcutil.ResponseErr(err)
	}
	return rep.(*pb.MediaMessage), nil
}

func (a mediaRPCImp) List(ctx context.Context, req *pb.ListRequest) (*pb.MediaListResponse, error) {
	_, rep, err := a.list.ServeGRPC(ctx, req)
	if err != nil {
		return nil, grpcutil.ResponseErr(err)
	}
	return rep.(*pb.MediaListResponse), nil
}

func (a mediaRPCImp) Get(ctx context.Context, req *pb.IDRequest) (*pb.MediaMessage, error) {
	_, rep, err := a.get.ServeGRPC(ctx, req)
	if err != nil {
		return nil, grpcutil.ResponseErr(err)
	}
	return rep.(*pb.MediaMessage), nil
}

func (a mediaRPCImp) Update(ctx context.Context, req *pb.MediaUpdateRequest) (*pb.MediaMessage, error) {
	_, rep, err := a.update.ServeGRPC(ctx, req)
	if err != nil {
		return nil, grpcutil.ResponseErr(err)
	}
	return rep.(*pb.MediaMessage), nil
}

func (a mediaRPCImp) Delete(ctx context.Context, req *pb.IDRequest) (*pb.Empty, error) {
	_, rep, err := a.delete.ServeGRPC(ctx, req)
	if err != nil {
		return nil, grpcutil.ResponseErr(err)
	}
	return rep.(*pb.Empty), nil
}

func (a mediaRPCImp) Restore(ctx context.Context, req *pb.IDRequest) (*pb.Empty, error) {
	_, rep, err := a.restore.ServeGRPC(ctx, req)
	if err != nil {
		return nil, grpcutil.ResponseErr(err)
	}
	return rep.(*pb.Empty), nil
}

func (a mediaRPCImp) HardDelete(ctx context.Context, req *pb.IDRequest) (*pb.Empty, error) {
	_, rep, err := a.hardDelete.ServeGRPC(ctx, req)
	if err != nil {
		return nil, grpcutil.ResponseErr(err)
	}
	return rep.(*pb.Empty), nil
}

/* Decoders */
func decodeRPCCreateRequest(_ context.Context, rpcReq interface{}) (interface{}, error) {
	req := rpcReq.(*pb.MediaCreateRequest)
	return action.CreateRequest{
		Title:        req.Title,
		DisplayName:  req.DisplayName,
		Description:  req.Description,
		LanguageCode: req.LanguageCode,
		PublisherID:  req.PublisherID,
		AuthorID:     req.AuthorID,
		PublishDate:  req.PublishDate,
		MediaType:    req.MediaType,
	}, nil
}

func decodeRPCListRequest(_ context.Context, rpcReq interface{}) (interface{}, error) {
	req := rpcReq.(*pb.ListRequest)
	return action.ListRequest{
		PageToken:    req.PageToken,
		PageSize:     req.PageSize,
		FilterParams: req.Filter,
	}, nil
}

func decodeRPCGetRequest(_ context.Context, rpcReq interface{}) (interface{}, error) {
	req := rpcReq.(*pb.GetRequest)
	return action.GetRequest{ID: req.Id}, nil
}

func decodeRPCUpdateRequest(_ context.Context, rpcReq interface{}) (interface{}, error) {
	req := rpcReq.(*pb.MediaUpdateRequest)
	return action.UpdateRequest{
		ID:           req.Id,
		Title:        req.Title,
		DisplayName:  req.DisplayName,
		Description:  req.Description,
		LanguageCode: req.LanguageCode,
		PublisherID:  req.PublisherID,
		AuthorID:     req.AuthorID,
		PublishDate:  req.PublishDate,
		MediaType:    req.MediaType,
		URL:          req.ContentURL,
	}, nil
}

func decodeRPCDeleteRequest(_ context.Context, rpcReq interface{}) (interface{}, error) {
	req := rpcReq.(*action.DeleteRequest)
	return action.DeleteRequest{ID: req.ID}, nil
}

func decodeRPCRestoreRequest(_ context.Context, rpcReq interface{}) (interface{}, error) {
	req := rpcReq.(*action.RestoreRequest)
	return action.RestoreRequest{ID: req.ID}, nil
}

func decodeRPCHardDeleteRequest(_ context.Context, rpcReq interface{}) (interface{}, error) {
	req := rpcReq.(*action.HardDeleteRequest)
	return action.HardDeleteRequest{ID: req.ID}, nil
}

/* Encoders */

func encodeRPCCreateResponse(_ context.Context, response interface{}) (interface{}, error) {
	res := response.(action.CreateResponse)
	if res.Err != nil {
		return nil, res.Err
	}

	if res.Media == nil {
		return nil, exception.EmptyBody
	}

	return &pb.MediaMessage{
		Id:           res.Media.ExternalID,
		Title:        res.Media.Title,
		DisplayName:  res.Media.DisplayName,
		Description:  res.Media.Description,
		LanguageCode: res.Media.LanguageCode,
		PublisherID:  res.Media.PublisherID,
		AuthorID:     res.Media.AuthorID,
		PublishDate:  res.Media.PublishDate.String(),
		MediaType:    res.Media.MediaType,
		CreateTime:   res.Media.CreateTime.String(),
		UpdateTime:   res.Media.UpdateTime.String(),
		DeleteTime:   res.Media.DeleteTime.String(),
		Active:       res.Media.Active,
		ContentURL:   *res.Media.ContentURL,
		TotalViews:   res.Media.TotalViews,
		Status:       res.Media.Status,
	}, nil
}

func encodeRPCListResponse(_ context.Context, response interface{}) (interface{}, error) {
	res := response.(action.ListResponse)
	if res.Err != nil {
		return nil, res.Err
	}

	if len(res.Medias) == 0 {
		return nil, status.Error(codes.NotFound, exception.EntitiesNotFound.Error())
	}

	MediasRPC := make([]*pb.MediaMessage, 0)
	for _, Media := range res.Medias {
		MediaRPC := &pb.MediaMessage{
			Id:           Media.ExternalID,
			Title:        Media.Title,
			DisplayName:  Media.DisplayName,
			Description:  Media.Description,
			LanguageCode: Media.LanguageCode,
			PublisherID:  Media.PublisherID,
			AuthorID:     Media.AuthorID,
			PublishDate:  Media.PublishDate.String(),
			MediaType:    Media.MediaType,
			CreateTime:   Media.CreateTime.String(),
			UpdateTime:   Media.UpdateTime.String(),
			DeleteTime:   Media.DeleteTime.String(),
			Active:       Media.Active,
			ContentURL:   *Media.ContentURL,
			TotalViews:   Media.TotalViews,
			Status:       Media.Status,
		}
		MediasRPC = append(MediasRPC, MediaRPC)
	}

	return &pb.MediaListResponse{
		Media:         MediasRPC,
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
		Id:           res.Media.ExternalID,
		Title:        res.Media.Title,
		DisplayName:  res.Media.DisplayName,
		Description:  res.Media.Description,
		LanguageCode: res.Media.LanguageCode,
		PublisherID:  res.Media.PublisherID,
		AuthorID:     res.Media.AuthorID,
		PublishDate:  res.Media.PublishDate.String(),
		MediaType:    res.Media.MediaType,
		CreateTime:   res.Media.CreateTime.String(),
		UpdateTime:   res.Media.UpdateTime.String(),
		DeleteTime:   res.Media.DeleteTime.String(),
		Active:       res.Media.Active,
		ContentURL:   *res.Media.ContentURL,
		TotalViews:   res.Media.TotalViews,
		Status:       res.Media.Status,
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
		Id:           res.Media.ExternalID,
		Title:        res.Media.Title,
		DisplayName:  res.Media.DisplayName,
		Description:  res.Media.Description,
		LanguageCode: res.Media.LanguageCode,
		PublisherID:  res.Media.PublisherID,
		AuthorID:     res.Media.AuthorID,
		PublishDate:  res.Media.PublishDate.String(),
		MediaType:    res.Media.MediaType,
		CreateTime:   res.Media.CreateTime.String(),
		UpdateTime:   res.Media.UpdateTime.String(),
		DeleteTime:   res.Media.DeleteTime.String(),
		Active:       res.Media.Active,
		ContentURL:   *res.Media.ContentURL,
		TotalViews:   res.Media.TotalViews,
		Status:       res.Media.Status,
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
