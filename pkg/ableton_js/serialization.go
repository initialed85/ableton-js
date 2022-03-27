package ableton_js

import (
    "bytes"
    "compress/zlib"
    "fmt"
    uuid2 "github.com/google/uuid"
    "io"
    "io/ioutil"
    "log"
    "os"
)

// GetCurrentSongTime is a big TODO
func GetCurrentSongTime() []byte {
    uuid, _ := uuid2.NewRandom()

    plainCommand := []byte(
        fmt.Sprintf(
            `{"uuid": "%v", "ns": "song", "name": "get_prop", "args": {"prop": "current_song_time"}}`,
            uuid.String(),
        ),
    )

    var compressedComand bytes.Buffer
    w := zlib.NewWriter(&compressedComand)
    _, err := w.Write(plainCommand)
    if err != nil {
        log.Fatal(err)
    }
    _ = w.Close()

    var final bytes.Buffer
    final.Write([]byte{255})
    final.Write(compressedComand.Bytes())

    return final.Bytes()
}

func HandleResponse(payload []byte) []byte {
    var compressedResponse bytes.Buffer

    compressedResponse.Write(payload[1:])

    r, err := zlib.NewReader(&compressedResponse)
    if err != nil {
        log.Fatal(err)
    }
    plainResponse, err := ioutil.ReadAll(r)
    if err != nil {
        log.Fatal(err)
    }

    _, err = io.Copy(os.Stdout, r)
    if err != nil {
        log.Fatal(err)
    }
    _ = r.Close()

    return plainResponse
}
