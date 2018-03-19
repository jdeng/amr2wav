package amr2wav

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"unsafe"
)

// #cgo CXXFLAGS: -I . -I ./include
// #cgo CFLAGS: -I . -I ./include
//
//#include "sp_dec.h"
//#include "amrdecode.h"
//
//void* Decoder_Interface_init(void) {
//    void* ptr = NULL;
//    GSMInitDecode(&ptr, (int8*)"Decoder");
//    return ptr;
//}
//
//void Decoder_Interface_exit(void* state) {
//    GSMDecodeFrameExit(&state);
//}
//
//short Decoder_Interface_Decode(void* state, void* _in, void* out, int bfi) {
//    unsigned char *in = (unsigned char *)_in;
//    unsigned char type = (in[0] >> 3) & 0x0f;
//    in++;
//    return AMRDecode(state, (enum Frame_Type_3GPP) type, (UWord8*) in, (short *)out, MIME_IETF);
//}
import "C"

func decode(in []byte) ([]byte, error) {
	const PCM_FRAME_SIZE = 160 // 8khz 8000 * 0.02 = 160

	d := C.Decoder_Interface_init()
	if d == nil {
		return nil, fmt.Errorf("Failed to init decoder")
	}

	var b bytes.Buffer
	var err error
	buf := make([]byte, PCM_FRAME_SIZE*2)
	for {
		if len(in) == 0 {
			break
		}
		ret := C.Decoder_Interface_Decode(d, unsafe.Pointer(&in[0]), unsafe.Pointer(&buf[0]), C.int(0))
		if ret < 0 {
			err = fmt.Errorf("Invalid data")
			break
		}
		in = in[ret+1:]
		b.Write(buf)
	}
	C.Decoder_Interface_exit(d)

	return b.Bytes(), err
}

type wavHeader struct {
	ChunkID       [4]byte
	ChunkSize     uint32
	Format        [4]byte
	Subchunk1ID   [4]byte
	Subchunk1Size uint32
	AudioFormat   uint16
	NumChannels   uint16
	SampleRate    uint32
	ByteRate      uint32
	BlockAlign    uint16
	BitsPerSample uint16
	Subchunk2ID   [4]byte
	Subchunk2Size uint32
}

const SAMPLE_RATE = 8000

func Convert(buf []byte) ([]byte, error) {
	const TAG = "#!AMR\n"
	if len(buf) < len(TAG) || !bytes.HasPrefix(buf, []byte(TAG)) {
		return nil, fmt.Errorf("Invalid header")
	}

	out, err := decode(buf[len(TAG):])
	if err != nil {
		return nil, err
	}

	datalen := uint32(len(out))
	hdr := wavHeader{
		ChunkID:       [4]byte{'R', 'I', 'F', 'F'},
		ChunkSize:     36 + datalen,
		Format:        [4]byte{'W', 'A', 'V', 'E'},
		Subchunk1ID:   [4]byte{'f', 'm', 't', ' '},
		Subchunk1Size: 16,
		AudioFormat:   1, //1 = PCM not compressed
		NumChannels:   1,
		SampleRate:    SAMPLE_RATE,
		ByteRate:      2 * SAMPLE_RATE,
		BlockAlign:    2,
		BitsPerSample: 16,
		Subchunk2ID:   [4]byte{'d', 'a', 't', 'a'},
		Subchunk2Size: datalen,
	}

	var b bytes.Buffer
	binary.Write(&b, binary.LittleEndian, hdr)
	b.Write(out)
	return b.Bytes(), nil
}

func main() {
	if len(os.Args) != 3 {
		fmt.Printf("Usage: %s inputfile outputfile\n", os.Args[0])
		return
	}

	data, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		log.Println(err)
		return
	}

	out, err := Convert(data)
	if err != nil {
		log.Println(err)
		return
	}

	ioutil.WriteFile(os.Args[2], out, 0755)
	log.Printf("Converted %s to %s\n", os.Args[1], os.Args[2])
}
