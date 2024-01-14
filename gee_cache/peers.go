package gee_cache

import "mycache/gee_cachepb/pb"

type PeerPicker interface {
	PickPeer(key string) (peer PeerGetter, ok bool)
}

type PeerGetter interface {
	Get(in *pb.Request, response *pb.Response) ([]byte, error)
}
