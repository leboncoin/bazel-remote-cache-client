package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	remoteexecution "github.com/bazelbuild/remote-apis/build/bazel/remote/execution/v2"
	gcode "google.golang.org/genproto/googleapis/rpc/code"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

type BazelRemoteCache struct {
	client       *grpc.ClientConn
	instanceName string

	ac  remoteexecution.ActionCacheClient
	cas remoteexecution.ContentAddressableStorageClient
}

func NewBazelRemoteCache(ctx context.Context, remote string, instanceName string) (*BazelRemoteCache, error) {
	client, err := grpc.DialContext(
		ctx, remote,
		grpc.WithTransportCredentials(
			insecure.NewCredentials(),
		),
		grpc.WithUserAgent("bazel-remote-cache-client"),
	)

	if err != nil {
		return nil, fmt.Errorf("can't connect to the remote cache: %v", err)
	}

	return &BazelRemoteCache{
		client:       client,
		instanceName: instanceName,

		ac:  remoteexecution.NewActionCacheClient(client),
		cas: remoteexecution.NewContentAddressableStorageClient(client),
	}, nil
}

func (brc *BazelRemoteCache) GetCacheResult(ctx context.Context, digest string) (*remoteexecution.ActionResult, error) {
	return brc.ac.GetActionResult(ctx, &remoteexecution.GetActionResultRequest{
		InstanceName: brc.instanceName,
		ActionDigest: &remoteexecution.Digest{
			Hash:      digest,
			SizeBytes: 1,
		},
	})
}

func (brc *BazelRemoteCache) GetBlob(ctx context.Context, digest *remoteexecution.Digest) ([]byte, error) {
	resp, err := brc.cas.BatchReadBlobs(ctx, &remoteexecution.BatchReadBlobsRequest{
		InstanceName: brc.instanceName,
		Digests:      []*remoteexecution.Digest{digest},
	})

	if err != nil {
		return nil, err
	}

	if len(resp.Responses) != 1 {
		return nil, errors.New("no reponses from the remote cache")
	}

	r := resp.Responses[0]

	switch gcode.Code(r.Status.GetCode()) {
	case gcode.Code_OK:
	case gcode.Code_NOT_FOUND:
		return nil, errors.New("blob not found")
	default:
		return nil, fmt.Errorf("error %v", r.Status)
	}

	return r.Data, nil
}

func (brc *BazelRemoteCache) ErrorMsg(err error) string {
	var errMsg string
	if st, ok := status.FromError(err); ok {
		switch st.Code() {
		case codes.NotFound:
			errMsg = "Not found"
		default:
			errMsg = st.Message()
		}
	} else {
		errMsg = err.Error()
	}

	return errMsg
}

func (brc *BazelRemoteCache) Close() {
	if err := brc.client.Close(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Can't close gRPC client: %v\n", err)
	}
}
