package blocksutil

import (
	"context"
	"fmt"

	"github.com/ipfs/go-cid"
	cbornode "github.com/ipfs/go-ipld-cbor"
	format "github.com/ipfs/go-ipld-format"
	mh "github.com/multiformats/go-multihash"
)

// NewDAGGenerator returns an object capable of
// producing IPLD DAGs.
func NewDAGGenerator() *DAGGenerator {
	return &DAGGenerator{}
}

// DAGGenerator generates BasicBlocks on demand.
// For each instance of DAGGenerator, each new DAG is different from the
// previous, although two different instances will produce the same, given the
// same parameters.
type DAGGenerator struct {
	seq int
}

func (dg *DAGGenerator) MakeDag(adder format.NodeAdder, fanout uint, depth uint) (c cid.Cid, allCids []cid.Cid, err error) {
	if depth == 1 {
		c, err = dg.encodeBlock(adder)
		if err != nil {
			return cid.Undef, nil, err
		}
		return c, []cid.Cid{c}, nil
	}
	links := make([]cid.Cid, fanout)
	for i := uint(0); i < fanout; i++ {
		var children []cid.Cid
		links[i], children, err = dg.MakeDag(adder, fanout, depth-1)
		if err != nil {
			return cid.Undef, nil, err
		}
		allCids = append(allCids, children...)
	}
	c, err = dg.encodeBlock(adder, links...)
	if err != nil {
		return cid.Undef, nil, err
	}
	return c, append([]cid.Cid{c}, allCids...), nil
}

func (dg *DAGGenerator) encodeBlock(adder format.NodeAdder, links ...cid.Cid) (cid.Cid, error) {
	dg.seq++
	obj := map[string]interface{}{
		"seq": fmt.Sprint(dg.seq),
	}
	if len(links) > 0 {
		obj["links"] = links
	}
	node, err := cbornode.WrapObject(obj, mh.SHA2_256, -1)
	if err != nil {
		return cid.Undef, err
	}
	err = adder.Add(context.Background(), node)
	if err != nil {
		return cid.Undef, err
	}
	return node.Cid(), nil
}
