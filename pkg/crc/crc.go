package crc

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"io"
)

const chunkSize = 4096

// see https://forum.golangbridge.org/t/how-to-find-the-offset-of-a-byte-in-a-large-binary-file/16457
func find(r io.Reader, search []byte) (int64, error) {
	var offset int64
	tailLen := len(search) - 1
	chunk := make([]byte, chunkSize+tailLen)
	n, err := r.Read(chunk[tailLen:])
	idx := bytes.Index(chunk[tailLen:n+tailLen], search)
	for {
		if idx >= 0 {
			return offset + int64(idx), nil
		}

		if err == io.EOF {
			return -1, nil
		} else if err != nil {
			return -1, err
		}

		copy(chunk, chunk[chunkSize:])

		offset += chunkSize
		n, err = r.Read(chunk[tailLen:])
		idx = bytes.Index(chunk[:n+tailLen], search)
	}
}

func clen(n []byte) int {
	for i := 0; i < len(n); i++ {
		if n[i] == 0 {
			return i
		}
	}

	return len(n)
}

func validOffset(r io.ReaderAt, start int64) bool {
	b := make([]byte, 3)

	r.ReadAt(b, start)

	return b[0] == 0x81 && b[1] == 0x7b && b[2] == 0x20
}

type Entry struct {
	MapName    string `json:"map_name"`
	MapCrc32   uint32 `json:"map_crc"`
	WorkshopID uint32 `json:"workshop_id"`
}

type Reader interface {
	io.Reader
	io.ReaderAt
}

func ExtractCRCTable(r Reader) ([]Entry, error) {
	startB, err := hex.DecodeString("817b20c2d13eba0f84502200004531ed4531e4")
	if err != nil {
		return nil, err
	}

	start, err := find(r, startB)
	if err != nil {
		return nil, err
	}

	start++

	var (
		dwordBuf = make([]byte, 4)
		res      = []Entry{
			{ // first entry is hardcoded as it doesn't match the pattern - but this is always the same entry
				MapName:    "de_nuke",
				MapCrc32:   3124679106,
				WorkshopID: 157233767,
			},
		}
	)

	for ; validOffset(r, start); start += 13 {
		//fmt.Println("crc")
		crcAddr := start + 3
		r.ReadAt(dwordBuf, crcAddr)
		//fmt.Println(hex.EncodeToString(dwordBuf))
		crc := binary.LittleEndian.Uint32(dwordBuf)
		//fmt.Println(crc)

		//fmt.Println("jump-target")
		baseAddr := start + 7
		jumpTargetOffsetAddr := baseAddr + 2
		r.ReadAt(dwordBuf, jumpTargetOffsetAddr)
		//fmt.Println(hex.EncodeToString(dwordBuf))
		jumpTargetOffset := binary.LittleEndian.Uint32(dwordBuf)

		//fmt.Println("map-offset")
		mapOffsetAddr := baseAddr + int64(jumpTargetOffset) + 16
		//fmt.Println(mapOffsetAddr)
		r.ReadAt(dwordBuf, mapOffsetAddr)
		//fmt.Println(hex.EncodeToString(dwordBuf))
		mapAddr := binary.LittleEndian.Uint32(dwordBuf) + 4 + uint32(mapOffsetAddr)

		//fmt.Println("map")

		const mapNameMaxLength = 64

		mapNameBuf := make([]byte, mapNameMaxLength)
		r.ReadAt(mapNameBuf, int64(mapAddr))

		mapName := string(mapNameBuf[:clen(mapNameBuf)])
		//fmt.Println(crc, mapName)

		//fmt.Println("workshop-id")
		workshopIDAddr := baseAddr + int64(jumpTargetOffset) + 33
		r.ReadAt(dwordBuf, workshopIDAddr)
		//fmt.Println(hex.EncodeToString(dwordBuf))
		workshopID := binary.LittleEndian.Uint32(dwordBuf)
		//fmt.Println(start, crc, mapName, workshopID)

		res = append(res, Entry{
			MapName:    mapName,
			MapCrc32:   crc,
			WorkshopID: workshopID,
		})
	}

	return res, nil
}
