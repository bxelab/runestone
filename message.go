package go_runestone

import (
	"errors"

	"github.com/btcsuite/btcd/wire"
	"lukechampine.com/uint128"
)

type Message struct {
	Flaw   *Flaw
	Edicts []Edict
	Fields map[uint128.Uint128][]uint128.Uint128
}

func MessageFromIntegers(tx *wire.MsgTx, payload []uint128.Uint128) (*Message, error) {
	edicts := []Edict{}
	fields := make(map[uint128.Uint128][]uint128.Uint128)
	var flaw *Flaw

	for i := 0; i < len(payload); i += 2 {
		tag := Tag(payload[i].Lo)

		if tag == TagBody {
			id := RuneId{} // Initialize with default values
			for _, chunk := range payload[i+1:] {
				if len(chunk) != 4 {
					flaw = &Flaw{} // Initialize with appropriate flaw
					break
				}

				next, err := id.Next(chunk[0], chunk[1])
				if err != nil {
					flaw = &Flaw{} // Initialize with appropriate flaw
					break
				}

				edict, err := EdictFromIntegers(tx, next, chunk[2], chunk[3])
				if err != nil {
					flaw = &Flaw{} // Initialize with appropriate flaw
					break
				}

				id = next
				edicts = append(edicts, edict)
			}
			break
		}

		if i+1 >= len(payload) {
			flaw = &Flaw{} // Initialize with appropriate flaw
			break
		}

		value := payload[i+1]
		fields[tag] = append(fields[tag], value)
	}

	if flaw != nil {
		return nil, errors.New("error parsing payload")
	}

	return &Message{
		Flaw:   flaw,
		Edicts: edicts,
		Fields: fields,
	}, nil
}
