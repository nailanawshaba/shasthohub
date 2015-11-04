package basex

import (
	"encoding/base64"
	"fmt"
	"testing"
)

const encodeURL = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_"

func newBase64URLEncoding() *Encoding {
	return NewEncoding(encodeURL, 3)
}

func TestHelloWorld(t *testing.T) {
	encoding := newBase64URLEncoding()
	input := []byte("Hello world! It is I, Bubba Karp!")
	for i := 0; i <= len(input); i++ {
		output := make([]byte, encoding.EncodedLen(i))
		encoding.Encode(output, input[0:i])
		ours := string(output)
		theirs := base64.RawURLEncoding.EncodeToString(input[0:i])
		if ours != theirs {
			t.Fatalf("Failed on input '%s': %s != %s", input[0:i], ours, theirs)
		}
		roundTrip := make([]byte, encoding.DecodedLen(len(output)))
		encoding.Decode(roundTrip, output)
		if string(roundTrip) != string(input[0:i]) {
			t.Fatalf("Decoding error: %s != %s", string(roundTrip), string(input[0:i]))
		}
		fmt.Printf("%s\n", roundTrip)
	}
}
