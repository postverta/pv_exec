package server

import (
	pb "github.com/postverta/pv_exec/proto/worktree"
	"github.com/postverta/pv_exec/worktree"
	context "golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

type WorktreeServer struct {
	manager *worktree.WorktreeManager
}

func NewWorktreeServer(manager *worktree.WorktreeManager) *WorktreeServer {
	return &WorktreeServer{
		manager: manager,
	}
}

func (s *WorktreeServer) Save(c context.Context, req *pb.SaveReq) (*pb.SaveResp, error) {
	err := s.manager.Save()
	if err != nil {
		return nil, grpc.Errorf(codes.Internal, "Cannot save worktree, err: "+err.Error())
	}
	return &pb.SaveResp{}, nil
}
