package service

import (
	"context"
	"crypto/x509"
	"github.com/zoenion/common/errors"
	authpb "github.com/zoenion/common/proto/auth"
	"github.com/zoenion/common/service/jwt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"log"
	"path"
	"strings"
	"time"
)

type MethodMappedAuthRules map[string]*MethodAuthenticationRule

type MethodAuthenticationRule struct {
	Secure bool
	Jwt    bool
}

type GRPCInterceptor interface {
	InterceptUnary(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error)
	InterceptStream(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error
}

type GRPCMethodAuthenticator interface {
	AuthenticateContext(ctx context.Context) error
	ValidateCertificate(ctx context.Context, method string) error
}

type gRPCInterceptor struct {
	caCert        *x509.Certificate
	jwtVerifier   jwt.Verifier
	gatewaySecret string
	mappedRules   MethodMappedAuthRules
}

func (interceptor *gRPCInterceptor) InterceptUnary(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	var (
		rsp interface{}
		err error
	)
	start := time.Now()

	method := path.Base(info.FullMethod)
	log.Printf("gRPC request - Method:%s at %s\t\n", method, start)
	defer log.Printf("gRPC request - Method:%s\tDuration:%s\tError:%v\n", method, time.Since(start), err)

	rule := interceptor.mappedRules[method]
	if rule != nil {
		authorization := interceptor.getMeta(ctx, "authorization")
		if strings.HasPrefix(authorization, AuthGateway) {
			parts := strings.Split(authorization, ":")
			secret := parts[0]
			authorization = parts[1]

			if len(parts) == 2 {
				err = interceptor.validateGatewaySecret(secret)
				if err == nil {
					err = interceptor.validateJwt(ctx, authorization)
				}
			} else {
				err = errors.Forbidden
			}
		} else if strings.HasPrefix(authorization, AuthBearer) {
			err = interceptor.validateJwt(ctx, authorization)
		} else {
			err = errors.Forbidden
		}
	}

	if err == nil {
		rsp, err = handler(ctx, req)
	}

	return rsp, err
}

func (interceptor *gRPCInterceptor) InterceptStream(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	var err error

	start := time.Now()
	ctx := ss.Context()

	method := path.Base(info.FullMethod)
	log.Printf("gRPC request - Method:%s at %s\t\n", method, start)
	defer log.Printf("gRPC request - Method:%s\tDuration:%s\tError:%v\n", method, time.Since(start), err)

	rule := interceptor.mappedRules[method]
	if rule != nil {
		authorization := interceptor.getMeta(ctx, "authorization")
		if strings.HasPrefix(authorization, AuthGateway) {
			parts := strings.Split(authorization, ":")
			secret := parts[0]
			authorization = parts[1]

			if len(parts) == 2 {
				err = interceptor.validateGatewaySecret(secret)
				if err == nil {
					err = interceptor.validateJwt(ctx, authorization)
				}
			} else {
				err = errors.Forbidden
			}
		} else if strings.HasPrefix(authorization, AuthBearer) {
			err = interceptor.validateJwt(ctx, authorization)
		} else {
			err = errors.Forbidden
		}
	}

	if err == nil {
		err = handler(srv, ss)
	}
	return err
}

func (interceptor *gRPCInterceptor) validateClientCertificate(cert *x509.Certificate) error {
	return nil
}

func (interceptor *gRPCInterceptor) validateJwt(ctx context.Context, jwt string) error {
	return interceptor.jwtVerifier.Verify(ctx, jwt)
}

func (interceptor *gRPCInterceptor) validateGatewaySecret(secret string) error {
	if interceptor.gatewaySecret != secret {
		return errors.Forbidden
	}
	return nil
}

func (interceptor *gRPCInterceptor) getMeta(ctx context.Context, name string) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}

	meta := md.Get(name)
	if len(meta) == 0 {
		return ""
	}

	return meta[0]
}

func NewInterceptor(v *Box, rules MethodMappedAuthRules) GRPCInterceptor {
	return &gRPCInterceptor{
		jwtVerifier:   authpb.NewStringTokenVerifier(newVerifier(v)),
		gatewaySecret: "",
		mappedRules:   rules,
		caCert:        v.authorityCert,
	}
}
