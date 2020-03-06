// Code generated by protoc-gen-grpc-gateway. DO NOT EDIT.
// source: auth.proto

/*
Package authpb is a reverse proxy.

It translates gRPC into RESTful JSON APIs.
*/
package authpb

import (
	"context"
	"io"
	"net/http"

	"github.com/golang/protobuf/descriptor"
	"github.com/golang/protobuf/proto"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/grpc-ecosystem/grpc-gateway/utilities"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/status"
)

// Suppress "imported and not used" errors
var _ codes.Code
var _ io.Reader
var _ status.Status
var _ = runtime.String
var _ = utilities.NewDoubleArray
var _ = descriptor.ForMessage

var (
	filter_TokenStoreService_FindToken_0 = &utilities.DoubleArray{Encoding: map[string]int{"jti": 0}, Base: []int{1, 1, 0}, Check: []int{0, 1, 2}}
)

func request_TokenStoreService_FindToken_0(ctx context.Context, marshaler runtime.Marshaler, client TokenStoreServiceClient, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	var protoReq FindJWTRequest
	var metadata runtime.ServerMetadata

	var (
		val string
		ok  bool
		err error
		_   = err
	)

	val, ok = pathParams["jti"]
	if !ok {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "missing parameter %s", "jti")
	}

	protoReq.Jti, err = runtime.String(val)

	if err != nil {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "type mismatch, parameter: %s, error: %v", "jti", err)
	}

	if err := req.ParseForm(); err != nil {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	}
	if err := runtime.PopulateQueryParameters(&protoReq, req.Form, filter_TokenStoreService_FindToken_0); err != nil {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	}

	msg, err := client.FindToken(ctx, &protoReq, grpc.Header(&metadata.HeaderMD), grpc.Trailer(&metadata.TrailerMD))
	return msg, metadata, err

}

func local_request_TokenStoreService_FindToken_0(ctx context.Context, marshaler runtime.Marshaler, server TokenStoreServiceServer, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	var protoReq FindJWTRequest
	var metadata runtime.ServerMetadata

	var (
		val string
		ok  bool
		err error
		_   = err
	)

	val, ok = pathParams["jti"]
	if !ok {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "missing parameter %s", "jti")
	}

	protoReq.Jti, err = runtime.String(val)

	if err != nil {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "type mismatch, parameter: %s, error: %v", "jti", err)
	}

	if err := runtime.PopulateQueryParameters(&protoReq, req.URL.Query(), filter_TokenStoreService_FindToken_0); err != nil {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	}

	msg, err := server.FindToken(ctx, &protoReq)
	return msg, metadata, err

}

// RegisterTokenStoreServiceHandlerServer registers the http handlers for service TokenStoreService to "mux".
// UnaryRPC     :call TokenStoreServiceServer directly.
// StreamingRPC :currently unsupported pending https://github.com/grpc/grpc-go/issues/906.
func RegisterTokenStoreServiceHandlerServer(ctx context.Context, mux *runtime.ServeMux, server TokenStoreServiceServer) error {

	mux.Handle("GET", pattern_TokenStoreService_FindToken_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateIncomingContext(ctx, mux, req)
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}
		resp, md, err := local_request_TokenStoreService_FindToken_0(rctx, inboundMarshaler, server, req, pathParams)
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}

		forward_TokenStoreService_FindToken_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})

	return nil
}

// RegisterTokenStoreServiceHandlerFromEndpoint is same as RegisterTokenStoreServiceHandler but
// automatically dials to "endpoint" and closes the connection when "ctx" gets done.
func RegisterTokenStoreServiceHandlerFromEndpoint(ctx context.Context, mux *runtime.ServeMux, endpoint string, opts []grpc.DialOption) (err error) {
	conn, err := grpc.Dial(endpoint, opts...)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			if cerr := conn.Close(); cerr != nil {
				grpclog.Infof("Failed to close conn to %s: %v", endpoint, cerr)
			}
			return
		}
		go func() {
			<-ctx.Done()
			if cerr := conn.Close(); cerr != nil {
				grpclog.Infof("Failed to close conn to %s: %v", endpoint, cerr)
			}
		}()
	}()

	return RegisterTokenStoreServiceHandler(ctx, mux, conn)
}

// RegisterTokenStoreServiceHandler registers the http handlers for service TokenStoreService to "mux".
// The handlers forward requests to the grpc endpoint over "conn".
func RegisterTokenStoreServiceHandler(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error {
	return RegisterTokenStoreServiceHandlerClient(ctx, mux, NewTokenStoreServiceClient(conn))
}

// RegisterTokenStoreServiceHandlerClient registers the http handlers for service TokenStoreService
// to "mux". The handlers forward requests to the grpc endpoint over the given implementation of "TokenStoreServiceClient".
// Note: the gRPC framework executes interceptors within the gRPC handler. If the passed in "TokenStoreServiceClient"
// doesn't go through the normal gRPC flow (creating a gRPC client etc.) then it will be up to the passed in
// "TokenStoreServiceClient" to call the correct interceptors.
func RegisterTokenStoreServiceHandlerClient(ctx context.Context, mux *runtime.ServeMux, client TokenStoreServiceClient) error {

	mux.Handle("GET", pattern_TokenStoreService_FindToken_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateContext(ctx, mux, req)
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}
		resp, md, err := request_TokenStoreService_FindToken_0(rctx, inboundMarshaler, client, req, pathParams)
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}

		forward_TokenStoreService_FindToken_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})

	return nil
}

var (
	pattern_TokenStoreService_FindToken_0 = runtime.MustPattern(runtime.NewPattern(1, []int{2, 0, 2, 1, 1, 0, 4, 1, 5, 2}, []string{"jwt", "find", "jti"}, "", runtime.AssumeColonVerbOpt(true)))
)

var (
	forward_TokenStoreService_FindToken_0 = runtime.ForwardResponseMessage
)
