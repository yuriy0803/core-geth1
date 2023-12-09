// Code generated by github.com/fjl/gencodec. DO NOT EDIT.

package types

import (
	"encoding/json"
	"errors"

	"github.com/yuriy0803/core-geth1/common/hexutil"
)

var _ = (*syncAggregateMarshaling)(nil)

// MarshalJSON marshals as JSON.
func (s SyncAggregate) MarshalJSON() ([]byte, error) {
	type SyncAggregate struct {
		Signers   hexutil.Bytes `gencodec:"required" json:"sync_committee_bits"`
		Signature hexutil.Bytes `gencodec:"required" json:"sync_committee_signature"`
	}
	var enc SyncAggregate
	enc.Signers = s.Signers[:]
	enc.Signature = s.Signature[:]
	return json.Marshal(&enc)
}

// UnmarshalJSON unmarshals from JSON.
func (s *SyncAggregate) UnmarshalJSON(input []byte) error {
	type SyncAggregate struct {
		Signers   *hexutil.Bytes `gencodec:"required" json:"sync_committee_bits"`
		Signature *hexutil.Bytes `gencodec:"required" json:"sync_committee_signature"`
	}
	var dec SyncAggregate
	if err := json.Unmarshal(input, &dec); err != nil {
		return err
	}
	if dec.Signers == nil {
		return errors.New("missing required field 'sync_committee_bits' for SyncAggregate")
	}
	if len(*dec.Signers) != len(s.Signers) {
		return errors.New("field 'sync_committee_bits' has wrong length, need 64 items")
	}
	copy(s.Signers[:], *dec.Signers)
	if dec.Signature == nil {
		return errors.New("missing required field 'sync_committee_signature' for SyncAggregate")
	}
	if len(*dec.Signature) != len(s.Signature) {
		return errors.New("field 'sync_committee_signature' has wrong length, need 96 items")
	}
	copy(s.Signature[:], *dec.Signature)
	return nil
}
