package crc

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"io"

	"github.com/pkg/errors"
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

func validOffset(r io.ReaderAt, start int64) (bool, error) {
	b := make([]byte, 3)

	_, err := r.ReadAt(b, start)
	if err != nil {
		return false, errors.WithStack(err)
	}

	return b[0] == 0x81 && b[1] == 0x7b && b[2] == 0x20, nil
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
	startB, err := hex.DecodeString("817b20c2d13eba0f84")
	if err != nil {
		return nil, errors.WithStack(err)
	}
	// 817b20c2d13eba0f
	// 84
	// 50220000
	// 4531ed45
	// 31e4

	start, err := find(r, startB)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	if start < 0 {
		return nil, errors.New("start bytes not found")
	}

	start += 11

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

	ok, err := validOffset(r, start)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	if !ok {
		return nil, errors.New("found start bytes, but they were not a valid offset")
	}

	for ok {
		// fmt.Println("crc")
		crcAddr := start + 3

		_, err = r.ReadAt(dwordBuf, crcAddr)
		if err != nil {
			return nil, errors.WithStack(err)
		}

		// fmt.Println(hex.EncodeToString(dwordBuf))
		crc := binary.LittleEndian.Uint32(dwordBuf)
		// fmt.Println(crc)

		// fmt.Println("jump-target")
		baseAddr := start + 7
		jumpTargetOffsetAddr := baseAddr + 2

		_, err = r.ReadAt(dwordBuf, jumpTargetOffsetAddr)
		if err != nil {
			return nil, errors.WithStack(err)
		}

		// fmt.Println(hex.EncodeToString(dwordBuf))
		jumpTargetOffset := binary.LittleEndian.Uint32(dwordBuf)

		// fmt.Println("map-offset")
		mapOffsetAddr := baseAddr + int64(jumpTargetOffset) + 16
		// fmt.Println(mapOffsetAddr)
		_, err = r.ReadAt(dwordBuf, mapOffsetAddr)
		if err != nil {
			return nil, errors.WithStack(err)
		}

		// fmt.Println(hex.EncodeToString(dwordBuf))
		mapAddr := binary.LittleEndian.Uint32(dwordBuf) + 4 + uint32(mapOffsetAddr)

		// fmt.Println("map")

		const mapNameMaxLength = 64

		mapNameBuf := make([]byte, mapNameMaxLength)

		_, err = r.ReadAt(mapNameBuf, int64(mapAddr))
		if err != nil {
			return nil, errors.WithStack(err)
		}

		mapName := string(mapNameBuf[:clen(mapNameBuf)])
		// fmt.Println(crc, mapName)

		// fmt.Println("workshop-id")
		workshopIDAddr := baseAddr + int64(jumpTargetOffset) + 33

		_, err = r.ReadAt(dwordBuf, workshopIDAddr)
		if err != nil {
			return nil, errors.WithStack(err)
		}

		// fmt.Println(hex.EncodeToString(dwordBuf))
		workshopID := binary.LittleEndian.Uint32(dwordBuf)
		// fmt.Println(start, crc, mapName, workshopID)

		res = append(res, Entry{
			MapName:    mapName,
			MapCrc32:   crc,
			WorkshopID: workshopID,
		})

		start += 13
		ok, err = validOffset(r, start)
		if err != nil {
			return nil, errors.WithStack(err)
		}
	}

	return res, nil
}
