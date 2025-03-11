package processor

import (
	"context"
	"encoding/json"
	"io"
	"strings"

	core_v3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	service_ext_proc_v3 "github.com/envoyproxy/go-control-plane/envoy/service/ext_proc/v3"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"
)

type ProcessingServer struct {
	Log *zap.Logger
}

//type AddHeaderInstruction struct {
//	// add headers from the request or response.
//	AddHeaders map[string]string
//}
//
//type RemoveHeaderInstruction struct {
//	// remove headers from the request or response.
//	RemoveHeaders []string
//}

type Instructions struct {
	AddHeaders    map[string]string `json:"addHeaders"`
	RemoveHeaders []string          `json:"removeHeaders"`
}

type HealthServer struct {
	Log *zap.Logger
}

func New(log *zap.Logger) *ProcessingServer {
	ps := &ProcessingServer{Log: log}
	return ps
}

func (s *HealthServer) Check(ctx context.Context, in *grpc_health_v1.HealthCheckRequest) (*grpc_health_v1.HealthCheckResponse, error) {
	s.Log.Debug("handling health check request", zap.String("service", in.String()))
	return &grpc_health_v1.HealthCheckResponse{Status: grpc_health_v1.HealthCheckResponse_SERVING}, nil
}

func (s *HealthServer) Watch(in *grpc_health_v1.HealthCheckRequest, srv grpc_health_v1.Health_WatchServer) error {
	return status.Error(codes.Unimplemented, "watch is not implemented")
}

func (s *ProcessingServer) Process(srv service_ext_proc_v3.ExternalProcessor_ProcessServer) error {
	ctx := srv.Context()
	for {
		select {
		case <-ctx.Done():
			s.Log.Debug("context done")
			return ctx.Err()
		default:
		}

		req, err := srv.Recv()
		if err == io.EOF {
			// envoy has closed the stream. Don't return anything and close this stream entirely
			return nil
		}
		if err != nil {
			return status.Errorf(codes.Unknown, "cannot receive stream request: %v", err)
		}

		// build response based on request type
		resp := &service_ext_proc_v3.ProcessingResponse{}
		switch v := req.Request.(type) {
		case *service_ext_proc_v3.ProcessingRequest_RequestHeaders:
			s.Log.Debug("got RequestHeaders", zap.Any("headers", v.RequestHeaders))
			h := req.Request.(*service_ext_proc_v3.ProcessingRequest_RequestHeaders)
			headersResp, err := s.processHeaderInstructions(h.RequestHeaders)
			if err != nil {
				return err
			}
			resp = &service_ext_proc_v3.ProcessingResponse{
				Response: &service_ext_proc_v3.ProcessingResponse_RequestHeaders{
					RequestHeaders: headersResp,
				},
			}

		case *service_ext_proc_v3.ProcessingRequest_RequestBody:
			s.Log.Debug("got RequestBody (not currently implemented)")

		case *service_ext_proc_v3.ProcessingRequest_RequestTrailers:
			s.Log.Debug("got RequestTrailers (not currently implemented)")

		case *service_ext_proc_v3.ProcessingRequest_ResponseHeaders:
			h := req.Request.(*service_ext_proc_v3.ProcessingRequest_ResponseHeaders)
			headersResp, err := s.processHeaderInstructions(h.ResponseHeaders)
			if err != nil {
				return err
			}
			resp = &service_ext_proc_v3.ProcessingResponse{
				Response: &service_ext_proc_v3.ProcessingResponse_ResponseHeaders{
					ResponseHeaders: headersResp,
				},
			}

		case *service_ext_proc_v3.ProcessingRequest_ResponseBody:
			s.Log.Debug("got ResponseBody (not currently implemented)")

		case *service_ext_proc_v3.ProcessingRequest_ResponseTrailers:
			s.Log.Debug("got ResponseTrailers (not currently handled)")

		default:
			s.Log.Error("unknown Request type", zap.Any("v", v))
		}

		// At this point we believe we have created a valid response...
		// note that this is sometimes not the case
		// anyways for now just send it
		s.Log.Debug("sending ProcessingResponse")
		if err := srv.Send(resp); err != nil {
			s.Log.Error("send error", zap.Error(err))
			return err
		}
	}
}

func (s *ProcessingServer) getHeaderInstructions(in *service_ext_proc_v3.HttpHeaders) (*Instructions, error) {
	instructions := &Instructions{}
	for _, n := range in.Headers.Headers {
		if strings.EqualFold(n.Key, "instructions") {
			val := string(n.GetRawValue())
			err := json.Unmarshal([]byte(val), instructions)
			if err != nil {
				s.Log.Error("error unmarshalling instructions", zap.Error(err))
				return nil, err
			}
		}
	}
	return instructions, nil
}

func (s *ProcessingServer) processHeaderInstructions(in *service_ext_proc_v3.HttpHeaders) (*service_ext_proc_v3.HeadersResponse, error) {
	instructions, err := s.getHeaderInstructions(in)

	// no instructions were sent, so don't modify anything
	if err != nil || (instructions == nil) {
		return &service_ext_proc_v3.HeadersResponse{}, nil
	}

	// build the response
	resp := &service_ext_proc_v3.HeadersResponse{
		Response: &service_ext_proc_v3.CommonResponse{},
	}

	// headers
	if len(instructions.AddHeaders) > 0 || len(instructions.RemoveHeaders) > 0 {
		var addHeaders []*core_v3.HeaderValueOption
		for k, v := range instructions.AddHeaders {
			s.Log.Info("adding headers", zap.String("key", k), zap.String("value", v))
			addHeaders = append(addHeaders, &core_v3.HeaderValueOption{
				Header: &core_v3.HeaderValue{Key: k, RawValue: []byte(v)},
			})
		}
		resp.Response.HeaderMutation = &service_ext_proc_v3.HeaderMutation{
			SetHeaders:    addHeaders,
			RemoveHeaders: instructions.RemoveHeaders,
		}
	}

	return resp, nil
}
