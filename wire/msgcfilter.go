// Copyright (c) 2018 The ExchangeCoin team
// Copyright (c) 2017 The btcsuite developers
// Copyright (c) 2017 The Lightning Network Developers
// Copyright (c) 2018 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package wire

import (
	"fmt"
	"io"

	"github.com/EXCCoin/exccd/chaincfg/chainhash"
)

const (
	// MaxCFilterDataSize is the maximum byte size of a committed filter.
	// The maximum size is currently defined as 256KiB.
	MaxCFilterDataSize = 256 * 1024
)

// MsgCFilter implements the Message interface and represents a cfilter message.
// It is used to deliver a committed filter in response to a getcfilter
// (MsgGetCFilter) message.
type MsgCFilter struct {
	BlockHash  chainhash.Hash
	FilterType FilterType
	Data       []byte
}

// BtcDecode decodes r using the wire protocol encoding into the receiver.
// This is part of the Message interface implementation.
func (msg *MsgCFilter) BtcDecode(r io.Reader, pver uint32) error {
	if pver < NodeCFVersion {
		str := fmt.Sprintf("cfilter message invalid for protocol "+
			"version %d", pver)
		return messageError("MsgCFilter.BtcDecode", str)
	}

	// Read the hash of the filter's block
	err := readElement(r, &msg.BlockHash)
	if err != nil {
		return err
	}

	// Read filter type
	err = readElement(r, (*uint8)(&msg.FilterType))
	if err != nil {
		return err
	}

	// Read filter data
	msg.Data, err = ReadVarBytes(r, pver, MaxCFilterDataSize,
		"cfilter data")
	return err
}

// BtcEncode encodes the receiver to w using the wire protocol encoding. This is
// part of the Message interface implementation.
func (msg *MsgCFilter) BtcEncode(w io.Writer, pver uint32) error {
	if pver < NodeCFVersion {
		str := fmt.Sprintf("cfilter message invalid for protocol "+
			"version %d", pver)
		return messageError("MsgCFHeaders.BtcEncode", str)
	}

	size := len(msg.Data)
	if size > MaxCFilterDataSize {
		str := fmt.Sprintf("cfilter size too large for message "+
			"[size %v, max %v]", size, MaxCFilterDataSize)
		return messageError("MsgCFilter.BtcEncode", str)
	}

	err := writeElement(w, &msg.BlockHash)
	if err != nil {
		return err
	}

	err = binarySerializer.PutUint8(w, uint8(msg.FilterType))
	if err != nil {
		return err
	}

	return WriteVarBytes(w, pver, msg.Data)
}

// Deserialize decodes a filter from r into the receiver using a format that is
// suitable for long-term storage such as a database. This function differs from
// BtcDecode in that BtcDecode decodes from the wire protocol as it was sent
// across the network.  The wire encoding can technically differ depending on
// the protocol version and doesn't even really need to match the format of a
// stored filter at all. As of the time this comment was written, the encoded
// filter is the same in both instances, but there is a distinct difference and
// separating the two allows the API to be flexible enough to deal with changes.
func (msg *MsgCFilter) Deserialize(r io.Reader) error {
	// At the current time, there is no difference between the wire encoding
	// and the stable long-term storage format.  As a result, make use of
	// BtcDecode.
	return msg.BtcDecode(r, 0)
}

// Command returns the protocol command string for the message.  This is part
// of the Message interface implementation.
func (msg *MsgCFilter) Command() string {
	return CmdCFilter
}

// MaxPayloadLength returns the maximum length the payload can be for the
// receiver.  This is part of the Message interface implementation.
func (msg *MsgCFilter) MaxPayloadLength(pver uint32) uint32 {
	return uint32(VarIntSerializeSize(MaxCFilterDataSize)) +
		MaxCFilterDataSize + chainhash.HashSize + 1
}

// NewMsgCFilter returns a new cfilter message that conforms to the Message
// interface. See MsgCFilter for details.
func NewMsgCFilter(blockHash *chainhash.Hash, filterType FilterType,
	data []byte) *MsgCFilter {
	return &MsgCFilter{
		BlockHash:  *blockHash,
		FilterType: filterType,
		Data:       data,
	}
}
