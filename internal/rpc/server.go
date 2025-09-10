package rpc

import (
	"context"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"
	"godatabase/internal/rpc/proto"
	"godatabase/internal/storage"
)

type Server struct {
	proto.UnimplementedStorageServer
	storage storage.Storage
	server  *grpc.Server
}

func NewServer(storage storage.Storage) *Server {
	return &Server{
		storage: storage,
		server:  grpc.NewServer(),
	}
}

func (s *Server) Start(addr string) error {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	proto.RegisterStorageServer(s.server, s)
	log.Printf("Starting gRPC server on %s", addr)
	return s.server.Serve(lis)
}

func (s *Server) Stop() {
	if s.server != nil {
		s.server.GracefulStop()
	}
}

// Put implements the Put RPC method
func (s *Server) Put(ctx context.Context, req *proto.PutRequest) (*proto.PutResponse, error) {
	err := s.storage.Put(req.Key, req.Value)
	if err != nil {
		return &proto.PutResponse{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	return &proto.PutResponse{
		Success: true,
	}, nil
}

// Get implements the Get RPC method
func (s *Server) Get(ctx context.Context, req *proto.GetRequest) (*proto.GetResponse, error) {
	value, err := s.storage.Get(req.Key)
	if err != nil {
		return &proto.GetResponse{
			Found: false,
			Error: err.Error(),
		}, nil
	}

	return &proto.GetResponse{
		Value: value,
		Found: true,
	}, nil
}

// Delete implements the Delete RPC method
func (s *Server) Delete(ctx context.Context, req *proto.DeleteRequest) (*proto.DeleteResponse, error) {
	err := s.storage.Delete(req.Key)
	if err != nil {
		return &proto.DeleteResponse{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	return &proto.DeleteResponse{
		Success: true,
	}, nil
}

// StreamOperations implements the StreamOperations RPC method
func (s *Server) StreamOperations(req *proto.StreamRequest, stream proto.Storage_StreamOperationsServer) error {
	// This would be implemented for replication
	// For now, we'll just return an error
	return fmt.Errorf("streaming not implemented yet")
} 