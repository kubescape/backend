package v1

import (
	"context"
	"strings"
	"testing"

	"github.com/kubescape/backend/pkg/client/v1/proto"
	"github.com/kubescape/storage/pkg/apis/softwarecomposition/v1beta1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ---------------------------------------------------------------------------
// Connect / Close / IsConnected / GetAddress lifecycle
// ---------------------------------------------------------------------------

// TestStorageClient_ConnectLifecycle exercises the full dial lifecycle for the
// insecure (grpc://) path: a fresh client is not connected; Connect wires up
// the conn (IsConnected true, GetAddress reflects the parsed host:port); a
// second Connect is rejected as already-connected; Close tears the conn down
// (IsConnected false); and a second Close is a no-op.
func TestStorageClient_ConnectLifecycle(t *testing.T) {
	client, err := NewStorageClient("grpc://storage.example.com:50051", "acct", "key", "cluster")
	require.NoError(t, err)

	// Before Connect.
	assert.False(t, client.IsConnected(), "new client must not report connected")
	assert.Equal(t, "storage.example.com:50051", client.GetAddress())

	// Connect wires up the conn.
	require.NoError(t, client.Connect())
	assert.True(t, client.IsConnected(), "client must report connected after Connect")
	assert.NotNil(t, client.protoClient)

	// Second Connect is rejected.
	err = client.Connect()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already connected")
	assert.True(t, client.IsConnected(), "failed re-Connect must not drop the existing conn")

	// Close tears it down.
	require.NoError(t, client.Close())
	assert.False(t, client.IsConnected(), "client must not report connected after Close")
	assert.Nil(t, client.protoClient)

	// Second Close is a no-op (conn already nil).
	assert.NoError(t, client.Close())
}

// TestStorageClient_ConnectSecure covers the TLS branch of Connect: a grpcs://
// URL takes the credentials.NewTLS dial path. grpc.NewClient is lazy so no
// real handshake happens; we only assert the conn is established and torn down.
func TestStorageClient_ConnectSecure(t *testing.T) {
	client, err := NewStorageClient("grpcs://storage.example.com:443", "acct", "key", "cluster")
	require.NoError(t, err)

	require.NoError(t, client.Connect())
	assert.True(t, client.IsConnected())
	assert.Equal(t, "storage.example.com:443", client.GetAddress())
	require.NoError(t, client.Close())
	assert.False(t, client.IsConnected())
}

// ---------------------------------------------------------------------------
// Unary GetContainerProfile / SendContainerProfile — real bufconn round-trips
// ---------------------------------------------------------------------------

// TestStorageClient_SendContainerProfile_RoundTrip drives the unary
// SendContainerProfile over bufconn (not a mock): the server records the
// request it received and the client surfaces the server's response.
// Covers both the success and the !resp.Success paths — note the unary
// SendContainerProfile has no !Success branch in the client itself (it returns
// the response verbatim), so the second case simply asserts the failure fields
// propagate to the caller unchanged.
func TestStorageClient_SendContainerProfile_RoundTrip(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		rtSrv, client, cleanup := startBufconnStorageServer(t)
		defer cleanup()
		rtSrv.unarySendSuccess = true

		resp, err := client.SendContainerProfile(context.Background(), sampleContainerProfile())
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.True(t, resp.Success)

		require.NotNil(t, rtSrv.unarySendReceived)
		require.NotNil(t, rtSrv.unarySendReceived.ContainerProfile)
		assert.Equal(t, "nginx-7d4f8b9c-abc12", rtSrv.unarySendReceived.ContainerProfile.Name)
	})

	t.Run("server reports failure", func(t *testing.T) {
		rtSrv, client, cleanup := startBufconnStorageServer(t)
		defer cleanup()
		rtSrv.unarySendSuccess = false
		rtSrv.unarySendErrMsg = "pulsar unavailable"
		rtSrv.unarySendErrCode = proto.ErrorCode_ERROR_CODE_PULSAR_ERROR

		resp, err := client.SendContainerProfile(context.Background(), sampleContainerProfile())
		require.NoError(t, err) // unary SendContainerProfile returns the response verbatim
		require.NotNil(t, resp)
		assert.False(t, resp.Success)
		assert.Equal(t, "pulsar unavailable", resp.ErrorMessage)
		assert.Equal(t, proto.ErrorCode_ERROR_CODE_PULSAR_ERROR, resp.ErrorCode)
	})
}

// TestStorageClient_GetContainerProfile_RoundTrip drives the unary
// GetContainerProfile (GetProfile RPC) over bufconn. The success case returns a
// profile; the failure case sets resp.Success=false and asserts the client's
// !resp.Success branch converts it into a non-nil error carrying the server's
// message and code.
func TestStorageClient_GetContainerProfile_RoundTrip(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		rtSrv, client, cleanup := startBufconnStorageServer(t)
		defer cleanup()
		rtSrv.unaryGetSuccess = true
		rtSrv.unaryGetProfile = sampleContainerProfile()

		got, err := client.GetContainerProfile(context.Background(), "default", "nginx-7d4f8b9c-abc12")
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, "nginx-7d4f8b9c-abc12", got.Name)

		require.NotNil(t, rtSrv.unaryGetReceived)
		assert.Equal(t, "ContainerProfile", rtSrv.unaryGetReceived.Kind)
		assert.Equal(t, "default", rtSrv.unaryGetReceived.Namespace)
		assert.Equal(t, "nginx-7d4f8b9c-abc12", rtSrv.unaryGetReceived.Name)
	})

	t.Run("server reports failure", func(t *testing.T) {
		rtSrv, client, cleanup := startBufconnStorageServer(t)
		defer cleanup()
		rtSrv.unaryGetSuccess = false
		rtSrv.unaryGetErrMsg = "profile not found"
		rtSrv.unaryGetErrCode = proto.ErrorCode_ERROR_CODE_PROFILE_NOT_FOUND

		got, err := client.GetContainerProfile(context.Background(), "default", "missing")
		require.Error(t, err)
		assert.Nil(t, got)
		assert.Contains(t, err.Error(), "profile not found")
		assert.Contains(t, err.Error(), "ERROR_CODE_PROFILE_NOT_FOUND")
	})
}

// ---------------------------------------------------------------------------
// GetContainerProfileStream — error branches
// ---------------------------------------------------------------------------

func TestStorageClient_GetContainerProfileStream_NotConnected(t *testing.T) {
	client, err := NewStorageClient("grpc://storage.example.com:50051", "acct", "key", "cluster")
	require.NoError(t, err)

	got, err := client.GetContainerProfileStream(context.Background(), "default", "x")
	require.Error(t, err)
	assert.Nil(t, got)
	assert.Contains(t, err.Error(), "not connected")
}

// TestStorageClient_GetContainerProfileStream_OpenError forces the RPC open
// (c.protoClient.GetContainerProfileStream) to fail by closing the underlying
// connection first, so NewStream is rejected before any chunk flows.
func TestStorageClient_GetContainerProfileStream_OpenError(t *testing.T) {
	_, client, cleanup := startBufconnStorageServer(t)
	defer cleanup()

	require.NoError(t, client.conn.Close()) // break the conn so NewStream fails

	got, err := client.GetContainerProfileStream(context.Background(), "default", "x")
	require.Error(t, err)
	assert.Nil(t, got)
	assert.Contains(t, err.Error(), "failed to open GetContainerProfileStream")
}

// TestStorageClient_GetContainerProfileStream_FirstRecvError covers the branch
// where the very first stream.Recv returns an error (server refused before
// sending any chunk).
func TestStorageClient_GetContainerProfileStream_FirstRecvError(t *testing.T) {
	rtSrv, client, cleanup := startBufconnStorageServer(t)
	defer cleanup()
	rtSrv.cpGetFirstRecvErr = true

	got, err := client.GetContainerProfileStream(context.Background(), "default", "x")
	require.Error(t, err)
	assert.Nil(t, got)
	assert.Contains(t, err.Error(), "failed to receive first chunk")
}

// TestStorageClient_GetContainerProfileStream_NilMetadata covers the branch
// where the first chunk arrives but carries no metadata header.
func TestStorageClient_GetContainerProfileStream_NilMetadata(t *testing.T) {
	rtSrv, client, cleanup := startBufconnStorageServer(t)
	defer cleanup()
	rtSrv.cpGetNilMetadata = true

	got, err := client.GetContainerProfileStream(context.Background(), "default", "x")
	require.Error(t, err)
	assert.Nil(t, got)
	assert.Contains(t, err.Error(), "missing metadata")
}

// TestStorageClient_GetContainerProfileStream_ServerError covers the
// !md.Success branch: the server returns a metadata header reporting a
// server-side failure, and the client must convert it into an error that
// carries the server's message and error code.
func TestStorageClient_GetContainerProfileStream_ServerError(t *testing.T) {
	rtSrv, client, cleanup := startBufconnStorageServer(t)
	defer cleanup()
	rtSrv.cpGetNotSuccess = true

	got, err := client.GetContainerProfileStream(context.Background(), "default", "x")
	require.Error(t, err)
	assert.Nil(t, got)
	assert.Contains(t, err.Error(), "failed to get container profile")
	assert.Contains(t, err.Error(), "boom on the server")
	assert.Contains(t, err.Error(), "ERROR_CODE_INTERNAL_ERROR")
}

// TestStorageClient_GetContainerProfileStream_MidStreamRecvError covers the
// error branch inside the reassembly loop: metadata + one blob chunk arrive,
// then the server aborts, so a subsequent stream.Recv returns a non-EOF error.
func TestStorageClient_GetContainerProfileStream_MidStreamRecvError(t *testing.T) {
	rtSrv, client, cleanup := startBufconnStorageServer(t)
	defer cleanup()
	rtSrv.cpGetMidStreamErr = true

	got, err := client.GetContainerProfileStream(context.Background(), "default", "x")
	require.Error(t, err)
	assert.Nil(t, got)
	assert.Contains(t, err.Error(), "failed to receive chunk")
}

// TestStorageClient_GetContainerProfileStream_UnmarshalError covers the final
// branch: the reassembled bytes are not a valid ContainerProfile proto, so
// profile.Unmarshal fails. The server serves exists=true with garbage bytes.
func TestStorageClient_GetContainerProfileStream_UnmarshalError(t *testing.T) {
	rtSrv, client, cleanup := startBufconnStorageServer(t)
	defer cleanup()
	rtSrv.cpServeExists = true
	// A truncated / invalid protobuf: a field-0 tag is illegal, so Unmarshal
	// rejects it. Splitting across chunks also exercises the reassembly path.
	rtSrv.cpServeBytes = []byte{0x00, 0x01, 0x02, 0xff, 0xff, 0xff, 0xff, 0xff}
	rtSrv.cpServeChunkSize = 3

	got, err := client.GetContainerProfileStream(context.Background(), "default", "x")
	require.Error(t, err)
	assert.Nil(t, got)
	assert.Contains(t, err.Error(), "failed to unmarshal ContainerProfile")
}

// ---------------------------------------------------------------------------
// SendContainerProfileStream — error branches + chunk-split proof
// ---------------------------------------------------------------------------

func TestStorageClient_SendContainerProfileStream_NotConnected(t *testing.T) {
	client, err := NewStorageClient("grpc://storage.example.com:50051", "acct", "key", "cluster")
	require.NoError(t, err)

	resp, err := client.SendContainerProfileStream(context.Background(), sampleContainerProfile())
	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "not connected")
}

// TestStorageClient_SendContainerProfileStream_OpenError forces the RPC open to
// fail by closing the underlying connection before the call.
func TestStorageClient_SendContainerProfileStream_OpenError(t *testing.T) {
	_, client, cleanup := startBufconnStorageServer(t)
	defer cleanup()

	require.NoError(t, client.conn.Close())

	resp, err := client.SendContainerProfileStream(context.Background(), sampleContainerProfile())
	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "failed to open SendContainerProfileStream")
}

// TestStorageClient_SendContainerProfileStream_ServerAbort covers the Send
// error branch(es): the server accepts the first chunk then aborts, so the
// client's subsequent Send calls in the chunk loop observe the broken stream
// and fail. A multi-chunk (>1 MiB) payload guarantees the loop runs and hits
// the "failed to send chunk" branch.
//
// Note: the sibling "failed to send first chunk" branch is the same failure
// mode but is NOT deterministically reachable over a healthy bufconn — the
// first Send is buffered and almost always succeeds before any server RST is
// observed, so the error surfaces in the loop (or at CloseAndRecv) instead.
// We therefore pin the deterministic loop branch here.
func TestStorageClient_SendContainerProfileStream_ServerAbort(t *testing.T) {
	rtSrv, client, cleanup := startBufconnStorageServer(t)
	defer cleanup()
	rtSrv.cpSendAbort = true

	profile := largeContainerProfile(t, 4*sbomStreamChunkSize) // several chunks

	resp, err := client.SendContainerProfileStream(context.Background(), profile)
	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "failed to send")
}

// TestStorageClient_SendContainerProfileStream_ChunkSplit is the load-bearing
// proof that the streaming API actually splits a large ContainerProfile into
// multiple gRPC messages rather than shipping one oversized frame. It builds a
// profile whose marshaled size exceeds gRPC's default 4 MiB unary limit, sends
// it, and asserts the round-trip server received MORE THAN ONE chunk — exactly
// ceil(payloadSize / sbomStreamChunkSize) of them. This is the whole reason
// SendContainerProfileStream exists over the deprecated unary form.
func TestStorageClient_SendContainerProfileStream_ChunkSplit(t *testing.T) {
	rtSrv, client, cleanup := startBufconnStorageServer(t)
	defer cleanup()

	const fourMiB = 4 << 20
	profile := largeContainerProfile(t, fourMiB+1) // force the payload over 4 MiB

	payload, err := profile.Marshal()
	require.NoError(t, err)
	require.Greater(t, len(payload), fourMiB,
		"payload must exceed the 4 MiB unary limit for this test to prove anything")

	resp, err := client.SendContainerProfileStream(context.Background(), profile)
	require.NoError(t, err)
	assert.True(t, resp.Success)

	// The client cuts the payload into sbomStreamChunkSize (1 MiB) pieces:
	// the first chunk plus one per remaining piece.
	expectedChunks := (len(payload) + sbomStreamChunkSize - 1) / sbomStreamChunkSize
	assert.Greater(t, expectedChunks, 1, "sanity: payload should span multiple chunks")
	assert.Equal(t, expectedChunks, rtSrv.cpReceivedChunks,
		"server must receive exactly ceil(payload/chunkSize) chunks — proves the client SPLIT the payload")

	// And the reassembled bytes must equal the originally marshaled payload.
	assert.Equal(t, payload, rtSrv.cpReceivedBytes,
		"reassembled chunks must reconstruct the exact marshaled ContainerProfile")
}

// largeContainerProfile builds a ContainerProfile whose marshaled size is at
// least approxBytes, by stuffing a single large annotation value. Used to drive
// the chunking loop and the >4 MiB split proof.
func largeContainerProfile(t *testing.T, approxBytes int) *v1beta1.ContainerProfile {
	t.Helper()
	return &v1beta1.ContainerProfile{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ContainerProfile",
			APIVersion: "spdx.softwarecomposition.kubescape.io/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "big-profile",
			Namespace: "default",
			Annotations: map[string]string{
				"kubescape.io/blob": strings.Repeat("a", approxBytes),
			},
		},
	}
}
