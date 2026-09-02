package v1

import (
	"context"
	"errors"
	"io"
	"testing"

	gogoproto "github.com/gogo/protobuf/proto"
	"github.com/kubescape/backend/pkg/client/v1/proto"
	"github.com/kubescape/storage/pkg/apis/softwarecomposition/v1beta1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// fakeCPStream serves a scripted chunk sequence and records how many times
// the client actually pulled from the stream, so tests can prove the
// unchanged path never reads past the metadata chunk. The embedded nil
// grpc.ClientStream panics if the client touches any method beyond the two
// overridden here, which is itself an assertion.
type fakeCPStream struct {
	grpc.ClientStream
	ctx       context.Context
	chunks    []*proto.GetContainerProfileStreamChunk
	idx       int
	recvCalls int
}

func (f *fakeCPStream) Recv() (*proto.GetContainerProfileStreamChunk, error) {
	f.recvCalls++
	if err := f.ctx.Err(); err != nil {
		return nil, err
	}
	if f.idx >= len(f.chunks) {
		return nil, io.EOF
	}
	c := f.chunks[f.idx]
	f.idx++
	return c, nil
}

func (f *fakeCPStream) Context() context.Context { return f.ctx }

// cpStreamRecorder captures the request the client constructed and the
// stream it was handed, so a test can inspect both after the call returns.
type cpStreamRecorder struct {
	req    *proto.GetContainerProfileStreamRequest
	stream *fakeCPStream
}

// newCPStreamClient wires a StorageClient to a mock that serves chunks and
// records the outbound request.
func newCPStreamClient(t *testing.T, chunks ...*proto.GetContainerProfileStreamChunk) (*StorageClient, *cpStreamRecorder) {
	t.Helper()
	rec := &cpStreamRecorder{}
	client, err := NewStorageClient("grpc://example.com:50051", "acct", "key", "cluster")
	require.NoError(t, err)
	client.protoClient = &mockStorageServiceClient{
		getContainerProfileStreamFunc: func(ctx context.Context, in *proto.GetContainerProfileStreamRequest, _ ...grpc.CallOption) (grpc.ServerStreamingClient[proto.GetContainerProfileStreamChunk], error) {
			rec.req = in
			rec.stream = &fakeCPStream{ctx: ctx, chunks: chunks}
			return rec.stream, nil
		},
	}
	return client, rec
}

func cpMetaChunk(md *proto.GetContainerProfileStreamChunkMetadata) *proto.GetContainerProfileStreamChunk {
	return &proto.GetContainerProfileStreamChunk{Metadata: md}
}

// marshaledSampleCP returns a valid ContainerProfile payload for the happy path.
func marshaledSampleCP(t *testing.T) ([]byte, *v1beta1.ContainerProfile) {
	t.Helper()
	cp := &v1beta1.ContainerProfile{
		ObjectMeta: metav1.ObjectMeta{Name: "nginx-abc", Namespace: "default"},
	}
	b, err := cp.Marshal()
	require.NoError(t, err)
	return b, cp
}

// TestGetContainerProfileStream_KnownChecksumOption pins the request shape:
// the option must reach the wire, and its absence must leave the request
// byte-identical to what the pre-change code produced.
func TestGetContainerProfileStream_KnownChecksumOption(t *testing.T) {
	payload, _ := marshaledSampleCP(t)

	t.Run("option sets KnownChecksum on the request", func(t *testing.T) {
		client, rec := newCPStreamClient(t,
			cpMetaChunk(&proto.GetContainerProfileStreamChunkMetadata{Success: true, Exists: true}),
			&proto.GetContainerProfileStreamChunk{BlobChunk: payload},
		)
		_, err := client.GetContainerProfileStream(context.Background(), "default", "nginx-abc",
			WithProfileKnownChecksum("abc"))
		require.NoError(t, err)
		require.NotNil(t, rec.req)
		assert.Equal(t, "abc", rec.req.KnownChecksum)
	})

	t.Run("without the option the request is byte-identical to pre-change output", func(t *testing.T) {
		// Golden captured by marshaling GetContainerProfileStreamRequest with the
		// four pre-existing fields using the generated code as it stood BEFORE
		// known_checksum was added (commit b127e3d, gogoproto.Marshal). Proto3
		// omits zero-valued scalars, so an unset known_checksum must reproduce
		// these bytes exactly — the assertion is real, not tautological.
		golden := []byte{
			0x0a, 0x07, 0x64, 0x65, 0x66, 0x61, 0x75, 0x6c, 0x74, 0x12, 0x14, 0x6e,
			0x67, 0x69, 0x6e, 0x78, 0x2d, 0x37, 0x64, 0x34, 0x66, 0x38, 0x62, 0x39,
			0x63, 0x2d, 0x61, 0x62, 0x63, 0x31, 0x32, 0x1a, 0x09, 0x75, 0x73, 0x2d,
			0x65, 0x61, 0x73, 0x74, 0x2d, 0x31, 0x22, 0x0c, 0x31, 0x32, 0x33, 0x34,
			0x35, 0x36, 0x37, 0x38, 0x39, 0x30, 0x31, 0x32,
		}

		client, rec := newCPStreamClient(t,
			cpMetaChunk(&proto.GetContainerProfileStreamChunkMetadata{Success: true, Exists: true}),
			&proto.GetContainerProfileStreamChunk{BlobChunk: payload},
		)
		_, err := client.GetContainerProfileStream(context.Background(), "default", "nginx-7d4f8b9c-abc12",
			WithProfileRegion("us-east-1"), WithProfileCloudAccountIdentifier("123456789012"))
		require.NoError(t, err)
		require.NotNil(t, rec.req)
		assert.Empty(t, rec.req.KnownChecksum)

		got, err := gogoproto.Marshal(rec.req)
		require.NoError(t, err)
		assert.Equal(t, golden, got, "unset known_checksum must not perturb the wire format")
	})

	t.Run("profileOptionsWithDefaults defaults KnownChecksum to empty", func(t *testing.T) {
		assert.Empty(t, profileOptionsWithDefaults(nil).KnownChecksum)
		assert.Equal(t, "xyz", profileOptionsWithDefaults([]ProfileOption{WithProfileKnownChecksum("xyz")}).KnownChecksum)
	})
}

// TestGetContainerProfileStream_Unchanged covers the sentinel path, the D5
// protocol-violation guard, and the stream teardown that the early return
// must not skip.
func TestGetContainerProfileStream_Unchanged(t *testing.T) {
	t.Run("returns the sentinel without unmarshaling", func(t *testing.T) {
		// The trailing blob is deliberately un-unmarshalable: if the client ever
		// read and decoded it, the call would fail with an unmarshal error
		// instead of returning the sentinel.
		client, rec := newCPStreamClient(t,
			cpMetaChunk(&proto.GetContainerProfileStreamChunkMetadata{Success: true, Unchanged: true, Checksum: "abc"}),
			&proto.GetContainerProfileStreamChunk{BlobChunk: []byte{0xff, 0xff, 0xff}},
		)
		profile, err := client.GetContainerProfileStream(context.Background(), "default", "nginx-abc",
			WithProfileKnownChecksum("abc"))

		require.Error(t, err)
		assert.ErrorIs(t, err, ErrProfileUnchanged)
		assert.Nil(t, profile)
		assert.NotContains(t, err.Error(), "not found")
		assert.NotContains(t, err.Error(), "unmarshal")
		assert.Equal(t, 1, rec.stream.recvCalls, "must not read past the metadata chunk")
	})

	t.Run("unchanged wins over a default-false exists", func(t *testing.T) {
		// A server implementing only unchanged may leave exists at its proto3
		// default; that must not become a spurious not-found.
		client, _ := newCPStreamClient(t,
			cpMetaChunk(&proto.GetContainerProfileStreamChunkMetadata{Success: true, Unchanged: true, Exists: false}),
		)
		_, err := client.GetContainerProfileStream(context.Background(), "default", "nginx-abc",
			WithProfileKnownChecksum("abc"))
		assert.ErrorIs(t, err, ErrProfileUnchanged)
	})

	t.Run("D5: unchanged without a sent checksum is a loud non-sentinel error", func(t *testing.T) {
		client, _ := newCPStreamClient(t,
			cpMetaChunk(&proto.GetContainerProfileStreamChunkMetadata{Success: true, Unchanged: true, Exists: true}),
		)
		profile, err := client.GetContainerProfileStream(context.Background(), "default", "nginx-abc")

		require.Error(t, err)
		assert.Nil(t, profile)
		assert.False(t, errors.Is(err, ErrProfileUnchanged),
			"a protocol violation must never be reported as the sentinel")
		assert.Contains(t, err.Error(), "unconditional request")
		assert.Contains(t, err.Error(), "nginx-abc")
	})

	t.Run("half-read stream is torn down on the early return", func(t *testing.T) {
		client, rec := newCPStreamClient(t,
			cpMetaChunk(&proto.GetContainerProfileStreamChunkMetadata{Success: true, Unchanged: true, Checksum: "abc"}),
			&proto.GetContainerProfileStreamChunk{BlobChunk: []byte("chunk-1")},
			&proto.GetContainerProfileStreamChunk{BlobChunk: []byte("chunk-2")},
		)
		_, err := client.GetContainerProfileStream(context.Background(), "default", "nginx-abc",
			WithProfileKnownChecksum("abc"))
		assert.ErrorIs(t, err, ErrProfileUnchanged)

		assert.Equal(t, 1, rec.stream.recvCalls, "the queued blob chunks must not be read")
		require.Error(t, rec.stream.ctx.Err(), "the RPC context must be cancelled, not leaked")
		assert.ErrorIs(t, rec.stream.ctx.Err(), context.Canceled)
	})
}

// TestGetContainerProfileStream_NormalPath proves the body-fetching path is
// unaffected and that the checksum annotation is stamped only when present.
func TestGetContainerProfileStream_NormalPath(t *testing.T) {
	payload, want := marshaledSampleCP(t)

	t.Run("unchanged absent decodes as before", func(t *testing.T) {
		client, _ := newCPStreamClient(t,
			cpMetaChunk(&proto.GetContainerProfileStreamChunkMetadata{Success: true, Exists: true}),
			&proto.GetContainerProfileStreamChunk{BlobChunk: payload},
		)
		got, err := client.GetContainerProfileStream(context.Background(), "default", "nginx-abc")
		require.NoError(t, err)
		assert.Equal(t, want.Name, got.Name)
		assert.Equal(t, want.Namespace, got.Namespace)
		assert.Nil(t, got.Annotations, "no checksum means no annotation map is created")
	})

	t.Run("unchanged explicitly false decodes as before", func(t *testing.T) {
		client, _ := newCPStreamClient(t,
			cpMetaChunk(&proto.GetContainerProfileStreamChunkMetadata{Success: true, Exists: true, Unchanged: false}),
			&proto.GetContainerProfileStreamChunk{BlobChunk: payload},
		)
		got, err := client.GetContainerProfileStream(context.Background(), "default", "nginx-abc",
			WithProfileKnownChecksum("stale"))
		require.NoError(t, err)
		assert.Equal(t, want.Name, got.Name)
	})

	t.Run("checksum is stamped as an annotation", func(t *testing.T) {
		client, _ := newCPStreamClient(t,
			cpMetaChunk(&proto.GetContainerProfileStreamChunkMetadata{Success: true, Exists: true, Checksum: "abc"}),
			&proto.GetContainerProfileStreamChunk{BlobChunk: payload},
		)
		got, err := client.GetContainerProfileStream(context.Background(), "default", "nginx-abc")
		require.NoError(t, err)
		require.NotNil(t, got.Annotations)
		assert.Equal(t, "abc", got.Annotations[ContainerProfileChecksumAnnotationKey])
	})

	t.Run("empty checksum leaves existing annotations untouched", func(t *testing.T) {
		cp := &v1beta1.ContainerProfile{ObjectMeta: metav1.ObjectMeta{
			Name:        "nginx-abc",
			Namespace:   "default",
			Annotations: map[string]string{"kubescape.io/status": "completed"},
		}}
		withAnnotations, err := cp.Marshal()
		require.NoError(t, err)

		client, _ := newCPStreamClient(t,
			cpMetaChunk(&proto.GetContainerProfileStreamChunkMetadata{Success: true, Exists: true, Checksum: ""}),
			&proto.GetContainerProfileStreamChunk{BlobChunk: withAnnotations},
		)
		got, err := client.GetContainerProfileStream(context.Background(), "default", "nginx-abc")
		require.NoError(t, err)
		assert.Equal(t, map[string]string{"kubescape.io/status": "completed"}, got.Annotations)
		assert.NotContains(t, got.Annotations, ContainerProfileChecksumAnnotationKey)
	})

	t.Run("not-found is still reported when the row is absent", func(t *testing.T) {
		client, _ := newCPStreamClient(t,
			cpMetaChunk(&proto.GetContainerProfileStreamChunkMetadata{Success: true, Exists: false}),
		)
		_, err := client.GetContainerProfileStream(context.Background(), "default", "missing")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
		assert.False(t, errors.Is(err, ErrProfileUnchanged))
	})
}

// oldGetContainerProfileStreamRequest and oldGetContainerProfileStreamChunkMetadata
// reproduce the generated struct shapes exactly as they stood before
// known_checksum / unchanged / checksum were added (struct tags copied verbatim
// from commit b127e3d's storage_service.pb.go). They stand in for an
// old-binary peer in the wire-compatibility round-trips below.
type oldGetContainerProfileStreamRequest struct {
	Namespace              string `protobuf:"bytes,1,opt,name=namespace,proto3" json:"namespace,omitempty"`
	Name                   string `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
	Region                 string `protobuf:"bytes,3,opt,name=region,proto3" json:"region,omitempty"`
	CloudAccountIdentifier string `protobuf:"bytes,4,opt,name=cloud_account_identifier,json=cloudAccountIdentifier,proto3" json:"cloud_account_identifier,omitempty"`
	XXX_unrecognized       []byte `json:"-"`
}

func (m *oldGetContainerProfileStreamRequest) Reset()         { *m = oldGetContainerProfileStreamRequest{} }
func (m *oldGetContainerProfileStreamRequest) String() string { return gogoproto.CompactTextString(m) }
func (*oldGetContainerProfileStreamRequest) ProtoMessage()    {}

type oldGetContainerProfileStreamChunkMetadata struct {
	Success          bool            `protobuf:"varint,1,opt,name=success,proto3" json:"success,omitempty"`
	ErrorMessage     string          `protobuf:"bytes,2,opt,name=error_message,json=errorMessage,proto3" json:"error_message,omitempty"`
	ErrorCode        proto.ErrorCode `protobuf:"varint,3,opt,name=error_code,json=errorCode,proto3,enum=storage.ErrorCode" json:"error_code,omitempty"`
	Exists           bool            `protobuf:"varint,4,opt,name=exists,proto3" json:"exists,omitempty"`
	XXX_unrecognized []byte          `json:"-"`
}

func (m *oldGetContainerProfileStreamChunkMetadata) Reset() {
	*m = oldGetContainerProfileStreamChunkMetadata{}
}
func (m *oldGetContainerProfileStreamChunkMetadata) String() string {
	return gogoproto.CompactTextString(m)
}
func (*oldGetContainerProfileStreamChunkMetadata) ProtoMessage() {}

// TestProtoWireCompatibility runs the round-trip in both directions, so proto3
// default behavior is re-checked on every future change to these messages
// rather than resting on a one-time argument.
func TestProtoWireCompatibility(t *testing.T) {
	t.Run("old request bytes decode into the new struct with zero-valued new fields", func(t *testing.T) {
		oldReq := &oldGetContainerProfileStreamRequest{
			Namespace:              "default",
			Name:                   "nginx-abc",
			Region:                 "us-east-1",
			CloudAccountIdentifier: "123456789012",
		}
		wire, err := gogoproto.Marshal(oldReq)
		require.NoError(t, err)

		newReq := &proto.GetContainerProfileStreamRequest{}
		require.NoError(t, gogoproto.Unmarshal(wire, newReq))
		assert.Equal(t, "default", newReq.Namespace)
		assert.Equal(t, "nginx-abc", newReq.Name)
		assert.Equal(t, "us-east-1", newReq.Region)
		assert.Equal(t, "123456789012", newReq.CloudAccountIdentifier)
		assert.Equal(t, "", newReq.KnownChecksum, "absent field must read as the proto3 default")
	})

	t.Run("new request bytes survive an old-shaped decoder", func(t *testing.T) {
		newReq := &proto.GetContainerProfileStreamRequest{
			Namespace:              "default",
			Name:                   "nginx-abc",
			Region:                 "us-east-1",
			CloudAccountIdentifier: "123456789012",
			KnownChecksum:          "abc",
		}
		wire, err := gogoproto.Marshal(newReq)
		require.NoError(t, err)

		oldReq := &oldGetContainerProfileStreamRequest{}
		require.NoError(t, gogoproto.Unmarshal(wire, oldReq), "an old peer must tolerate the new field, not error")
		assert.Equal(t, "default", oldReq.Namespace)
		assert.Equal(t, "nginx-abc", oldReq.Name)
		assert.Equal(t, "us-east-1", oldReq.Region)
		assert.Equal(t, "123456789012", oldReq.CloudAccountIdentifier)
	})

	t.Run("old metadata bytes decode into the new struct with zero-valued new fields", func(t *testing.T) {
		oldMD := &oldGetContainerProfileStreamChunkMetadata{Success: true, Exists: true}
		wire, err := gogoproto.Marshal(oldMD)
		require.NoError(t, err)

		newMD := &proto.GetContainerProfileStreamChunkMetadata{}
		require.NoError(t, gogoproto.Unmarshal(wire, newMD))
		assert.True(t, newMD.Success)
		assert.True(t, newMD.Exists)
		assert.False(t, newMD.Unchanged, "absent unchanged must read as false")
		assert.Equal(t, "", newMD.Checksum, "absent checksum must read as empty")
	})

	t.Run("new metadata bytes survive an old-shaped decoder", func(t *testing.T) {
		newMD := &proto.GetContainerProfileStreamChunkMetadata{
			Success:   true,
			Exists:    true,
			Unchanged: true,
			Checksum:  "abc",
		}
		wire, err := gogoproto.Marshal(newMD)
		require.NoError(t, err)

		oldMD := &oldGetContainerProfileStreamChunkMetadata{}
		require.NoError(t, gogoproto.Unmarshal(wire, oldMD), "an old peer must tolerate the new fields, not error")
		assert.True(t, oldMD.Success)
		assert.True(t, oldMD.Exists)
	})
}
