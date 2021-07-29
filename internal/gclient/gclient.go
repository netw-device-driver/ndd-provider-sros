/*
Copyright 2021 Wim Henderickx.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package gclient

import (
	"context"
	"crypto/tls"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/netw-device-driver/ndd-runtime/pkg/logging"
	"github.com/netw-device-driver/netwdevpb"
)

const (
	defaultTimeout = 30 * time.Second
	maxMsgSize     = 512 * 1024 * 1024
)

// An Applicator brings up or down a set of resources in the API server
type Applicator interface {
	Register(ctx context.Context, target string, req *netwdevpb.RegistrationRequest) (*netwdevpb.RegistrationReply, error)

	DeRegister(ctx context.Context, target string, req *netwdevpb.RegistrationRequest) (*netwdevpb.RegistrationReply, error)

	CacheStatus(ctx context.Context, target string, req *netwdevpb.CacheStatusRequest) (*netwdevpb.CacheStatusReply, error)
}

type GrpcApplicator struct {
	Username   string
	Password   string
	Proxy      bool
	NoTLS      bool
	TLSCA      string
	TLSCert    string
	TLSKey     string
	SkipVerify bool
	Insecure   bool
	Target     string
	MaxMsgSize int
	log        logging.Logger
}

// ClientApplicatorOption is used to configure the ClientApplicator.
type GrpcApplicatorOption func(*GrpcApplicator)

// WithUsername specifies the username parameters
func WithUsername(username string) GrpcApplicatorOption {
	return func(c *GrpcApplicator) {
		c.Username = username
	}
}

// WithPassword specifies the password parameters
func WithPassword(password string) GrpcApplicatorOption {
	return func(c *GrpcApplicator) {
		c.Password = password
	}
}

// WithProxy specifies the proxy parameters
func WithProxy(proxy bool) GrpcApplicatorOption {
	return func(c *GrpcApplicator) {
		c.Proxy = proxy
	}
}

// WithTLS specifies the tls parameters
func WithTLS(noTLS bool, tlsca, tlscert, tlskey string) GrpcApplicatorOption {
	return func(c *GrpcApplicator) {
		c.NoTLS = noTLS
		c.TLSCA = tlsca
		c.TLSCert = tlscert
		c.TLSKey = tlskey
	}
}

// WithSkipVerify specifies skip verify grpc transport
func WithSkipVerify(b bool) GrpcApplicatorOption {
	return func(c *GrpcApplicator) {
		c.SkipVerify = b
	}
}

// WithInsecure specifies insecure grpc transport
func WithInsecure(b bool) GrpcApplicatorOption {
	return func(c *GrpcApplicator) {
		c.Insecure = b
	}
}

// WithLogger specifies how the Reconciler should log messages.
func WithTarget(target string) GrpcApplicatorOption {
	return func(c *GrpcApplicator) {
		c.Target = target
	}
}

// WithLogger specifies how the ClientApplicator should log messages.
func WithLogger(log logging.Logger) GrpcApplicatorOption {
	return func(c *GrpcApplicator) {
		c.log = log
	}
}

// NewAPIEstablisher creates a new APIEstablisher.
func NewClientApplicator(opts ...GrpcApplicatorOption) *GrpcApplicator {
	c := &GrpcApplicator{
		MaxMsgSize: maxMsgSize,
	}

	for _, f := range opts {
		f(c)
	}

	return c
}

func (c *GrpcApplicator) Register(ctx context.Context, target string, req *netwdevpb.RegistrationRequest) (*netwdevpb.RegistrationReply, error) {
	// Connect Options.
	c.log.Debug("client flags", "insecure", c.Insecure, "skipverify", c.SkipVerify)
	var opts []grpc.DialOption
	if c.Insecure {
		opts = append(opts, grpc.WithInsecure())
	} else {
		tlsConfig, err := c.newTLS()
		if err != nil {
			return nil, err
		}
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
	}
	timeoutCtx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	conn, err := grpc.DialContext(timeoutCtx, target, opts...)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	client := netwdevpb.NewRegistrationClient(conn)

	return client.Register(timeoutCtx, req)
}

func (c *GrpcApplicator) DeRegister(ctx context.Context, target string, req *netwdevpb.RegistrationRequest) (*netwdevpb.RegistrationReply, error) {
	// Connect Options.
	c.log.Debug("client flags", "insecure", c.Insecure, "skipverify", c.SkipVerify)
	var opts []grpc.DialOption
	if c.Insecure {
		opts = append(opts, grpc.WithInsecure())
	} else {
		tlsConfig, err := c.newTLS()
		if err != nil {
			return nil, err
		}
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
	}
	timeoutCtx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	conn, err := grpc.DialContext(timeoutCtx, target, opts...)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	client := netwdevpb.NewRegistrationClient(conn)

	return client.DeRegister(timeoutCtx, req)
}

func (c *GrpcApplicator) CacheStatus(ctx context.Context, target string, req *netwdevpb.CacheStatusRequest) (*netwdevpb.CacheStatusReply, error) {
	// Connect Options.
	c.log.Debug("client flags", "insecure", c.Insecure, "skipverify", c.SkipVerify)
	var opts []grpc.DialOption
	if c.Insecure {
		opts = append(opts, grpc.WithInsecure())
	} else {
		tlsConfig, err := c.newTLS()
		if err != nil {
			return nil, err
		}
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
	}
	timeoutCtx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	conn, err := grpc.DialContext(timeoutCtx, target, opts...)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	client := netwdevpb.NewCacheStatusClient(conn)

	return client.Request(timeoutCtx, req)
}

// newTLS sets up a new TLS profile
func (c *GrpcApplicator) newTLS() (*tls.Config, error) {
	tlsConfig := &tls.Config{
		Renegotiation:      tls.RenegotiateNever,
		InsecureSkipVerify: c.SkipVerify,
	}
	err := c.loadCerts(tlsConfig)
	if err != nil {
		return nil, err
	}
	return tlsConfig, nil
}

func (c *GrpcApplicator) loadCerts(tlscfg *tls.Config) error {
	/*
		if *c.TLSCert != "" && *c.TLSKey != "" {
			certificate, err := tls.LoadX509KeyPair(*c.TLSCert, *c.TLSKey)
			if err != nil {
				return err
			}
			tlscfg.Certificates = []tls.Certificate{certificate}
			tlscfg.BuildNameToCertificate()
		}
		if c.TLSCA != nil && *c.TLSCA != "" {
			certPool := x509.NewCertPool()
			caFile, err := ioutil.ReadFile(*c.TLSCA)
			if err != nil {
				return err
			}
			if ok := certPool.AppendCertsFromPEM(caFile); !ok {
				return errors.New("failed to append certificate")
			}
			tlscfg.RootCAs = certPool
		}
	*/
	return nil
}
