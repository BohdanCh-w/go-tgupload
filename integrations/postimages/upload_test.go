package postimages_test

import (
	"os"
	"testing"

	"github.com/bohdanch-w/go-tgupload/integrations/postimages"
	"github.com/stretchr/testify/require"
)

func TestParseResponse(t *testing.T) {
	file, err := os.ReadFile("testdata/resp.xml")
	require.NoError(t, err)

	link, err := postimages.ParseResponse(file)
	require.NoError(t, err)
	require.Equal(t, "https://i.postimg.cc/KY92bYbL/02.png", link)
}
