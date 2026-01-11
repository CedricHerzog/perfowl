package mcp

import (
	"context"
	"testing"

	"github.com/CedricHerzog/perfowl/internal/testutil"
	"github.com/mark3labs/mcp-go/mcp"
)

func BenchmarkHandleGetSummary(b *testing.B) {
	server := NewServer()
	profile := testutil.MediumProfile()
	path := testutil.TempProfileFile(b, profile)

	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]interface{}{
		"path": path,
	}

	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = server.handleGetSummary(ctx, request)
	}
}

func BenchmarkHandleGetBottlenecks(b *testing.B) {
	server := NewServer()
	profile := testutil.ProfileWithNMarkers(500)
	path := testutil.TempProfileFile(b, profile)

	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]interface{}{
		"path": path,
	}

	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = server.handleGetBottlenecks(ctx, request)
	}
}

func BenchmarkHandleGetCallTree(b *testing.B) {
	server := NewServer()
	profile := testutil.MediumProfile()
	path := testutil.TempProfileFile(b, profile)

	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]interface{}{
		"path": path,
	}

	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = server.handleGetCallTree(ctx, request)
	}
}

func BenchmarkHandleAnalyzeWorkers(b *testing.B) {
	server := NewServer()
	profile := testutil.ProfileWithNWorkers(4, 500)
	path := testutil.TempProfileFile(b, profile)

	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]interface{}{
		"path": path,
	}

	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = server.handleAnalyzeWorkers(ctx, request)
	}
}

func BenchmarkHandleAnalyzeCrypto(b *testing.B) {
	server := NewServer()
	profile := testutil.ProfileWithCryptoOperations(1000)
	path := testutil.TempProfileFile(b, profile)

	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]interface{}{
		"path": path,
	}

	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = server.handleAnalyzeCrypto(ctx, request)
	}
}

func BenchmarkHandleGetMarkers(b *testing.B) {
	server := NewServer()
	profile := testutil.ProfileWithNMarkers(1000)
	path := testutil.TempProfileFile(b, profile)

	b.Run("NoFilter", func(b *testing.B) {
		request := mcp.CallToolRequest{}
		request.Params.Arguments = map[string]interface{}{
			"path": path,
		}
		ctx := context.Background()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = server.handleGetMarkers(ctx, request)
		}
	})

	b.Run("WithTypeFilter", func(b *testing.B) {
		request := mcp.CallToolRequest{}
		request.Params.Arguments = map[string]interface{}{
			"path": path,
			"type": "GCMajor",
		}
		ctx := context.Background()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = server.handleGetMarkers(ctx, request)
		}
	})

	b.Run("WithLimit", func(b *testing.B) {
		request := mcp.CallToolRequest{}
		request.Params.Arguments = map[string]interface{}{
			"path":  path,
			"limit": float64(100),
		}
		ctx := context.Background()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = server.handleGetMarkers(ctx, request)
		}
	})
}

func BenchmarkHandleAnalyzeProfile(b *testing.B) {
	server := NewServer()
	profile := testutil.MediumProfile()
	path := testutil.TempProfileFile(b, profile)

	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]interface{}{
		"path": path,
	}

	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = server.handleAnalyzeProfile(ctx, request)
	}
}

func BenchmarkHandleCompareProfiles(b *testing.B) {
	server := NewServer()
	baseline := testutil.MediumProfile()
	comparison := testutil.MediumProfile()
	baselinePath := testutil.TempProfileFile(b, baseline)
	comparisonPath := testutil.TempProfileFile(b, comparison)

	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]interface{}{
		"baseline":   baselinePath,
		"comparison": comparisonPath,
	}

	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = server.handleCompareProfiles(ctx, request)
	}
}
