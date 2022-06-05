package bzlremotecache

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

// BazelRemoteCache is a client of a Bazel remote cache.
type BazelRemoteCache struct {
	client       *grpc.ClientConn
	instanceName string

	ac  remoteexecution.ActionCacheClient
	cas remoteexecution.ContentAddressableStorageClient
}

// New creates a new client to access of a Bazel remote cache.
func New(ctx context.Context, remote string, instanceName string) (*BazelRemoteCache, error) {
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

// GetCacheResult returns the given ActionCache stored in the Bazel remote cache.
func (brc *BazelRemoteCache) GetCacheResult(ctx context.Context, digest string) (*remoteexecution.ActionResult, error) {
	return brc.ac.GetActionResult(ctx, &remoteexecution.GetActionResultRequest{
		InstanceName: brc.instanceName,
		ActionDigest: &remoteexecution.Digest{
			Hash:      digest,
			SizeBytes: 1,
		},
	})
}

// GetBlob returns the content of a Bazel remote cache blob.
func (brc *BazelRemoteCache) GetBlob(ctx context.Context, digest *Digest) ([]byte, error) {
	resp, err := brc.cas.BatchReadBlobs(ctx, &remoteexecution.BatchReadBlobsRequest{
		InstanceName: brc.instanceName,
		Digests: []*remoteexecution.Digest{
			{
				Hash:      digest.Hash,
				SizeBytes: digest.Size,
			},
		},
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

// ErrorMsg returns the error message of the given error.
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

// Close closed the client of a Bazel remote cache.
func (brc *BazelRemoteCache) Close() {
	if err := brc.client.Close(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Can't close gRPC client: %v\n", err)
	}
}
