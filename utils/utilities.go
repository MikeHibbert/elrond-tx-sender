package utils

import (
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/crypto/signing"
)

// HostForShard - selects a random api host to use based on the supplied shard id
func HostForShard(shard int) string {
	hostMap := map[int][]string{
		0: []string{"localhost:8083", "localhost:8087"},
		1: []string{"localhost:8084", "localhost:8088"},
		2: []string{"localhost:8085", "localhost:8089"},
		3: []string{"localhost:8080", "localhost:8090"},
		4: []string{"localhost:8081", "localhost:8091"},
	}

	rand.Seed(time.Now().Unix())

	hostsForShard := hostMap[shard]
	hostForShard := hostsForShard[rand.Intn(len(hostsForShard))]

	if hostForShard != "" {
		hostForShard = fmt.Sprintf("http://%s", hostForShard)
	}

	return hostForShard
}

// FetchProxies - fetch a list of proxies from a specified file
func FetchProxies(filePath string) ([]string, error) {
	data, err := ReadFileToString(filePath)

	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(data), "\n")

	return lines, nil
}

// RandomProxy - fetches a random proxy from the specified file
func RandomProxy(proxies []string) string {
	rand.Seed(time.Now().Unix())
	proxy := proxies[rand.Intn(len(proxies))]

	return proxy
}

// RandomReceiverTarget - generate a random receiver based on a specific shard
func RandomReceiverTarget(shard int) string {
	targetMap := map[int][]string{
		0: []string{"c427aa79c2545ba8b00663ee1d02a785a7712ba2bba6eb4aa940ebc9eada7250", "39dd77813657fbd5bd0d3ebc7389c3fb55468b9b68510cf390e006fb44542628", "8124fe7889db00ad338f1f1fa23f87c6a77f272929d01bc12426c450ee1eb2b0", "f3026a61cd465c929d099a6ba867b08d53aa60e65e49439f1cbd9b39f8a0d930", "7f08ce0f2d17bdefa33e911a81aa1ffcdd088caea97db72d9981261ff25f0468", "33ac512e383e17ae8d76875f0b88e1567b33d39864c0ebb7f971d6b356aea9c0", "57db553b5b27b694c0960c7baecc8b41b4709fb5c574cfe48637ce6a409bee18", "2a0873f7775c60a7e0d352fdeea5837ab564a8cc3b808874ca1564c680896270", "07a58ff196bdc4bd69f77f03007971773594d69d76ce617fd2269b347f2d6c30", "9d6c3bb12f8c9505108765ca18d4ce49ee132f94a8ba9803d73a7191c7c7c480"},
		1: []string{"fea250241b2b56c4f6997af21576b607d09edfc68f859924179acfaae2c25701", "693517527bd120f25d490689e1b6c0e041c157b1079b2c84d0271473c33cf559", "f24225740905683725e53b6282ce913975ae445d659ce432a4cda8afa77a5ae5", "231845b11c8b5c3a9195e6d0b6658e7b1b0dabb3159248d8750a4ebdeed13c5d", "b8eb95eab28372ef3e546089a0e88f4f3b13f033a4be120eaf0ea6ef9dbbd135", "15ff0fc938d841ec37d8fc877f491f9b313d6ec647391999e18cf42e6b6a709d", "033b506b958aa4684a30b63b9d183ef8773dcc2cbd4dd8219d1653e027784059"},
		2: []string{"14abebdb4960d68716217d0828bfb58e6c99cc515e351f3e8a90f1968b8a82fa", "80688df42a15465529602dcb79386a78b133e9e71e1169015a3e12ed696e9826", "f98a23085ead87a965b2d35ec1251796d9019849ef04da6dbb314158c2faaf36", "12ae508ad1a5216ca70cde4a892018e3cfa55020d14a51fb4ab685d84d9e1506", "b85416ae13cbbe4f6073bd3233802a855d4bdf4f4d31c10bfe20ff88d06a7ad6", "a1b7fbacb7512a8a70913330193f832c75142f882ba34349e1698302a44372f6", "353b9f2e47da61e0db3462f5dfb06e14a82737493b6cae101b2e02a2fbba1be2", "4431da1e65723c567c96401ed1d38075b05a034b6dc7392f527e2eaa4bde76de", "5310cc393701935233dbb32b0fda32442710facd3bc5c1f88d2e40e9dc0af48e", "e6bf2405c46c8f9b96b187ac381f281a7884f8c2645d0207c7645076a7a9a21e"},
		3: []string{"b0b206a9dcde4ae0052f349e0af79a1138b72b9db15d149da3505a37d59ee353", "630c1bdd9ee77649eed05c8c0933185896d51d27f7046ace7e91b4c99922facb", "6da79cc2731a3ff0aaac34063869f51e9fa03cb4105e4e4894b20a38142817b3", "94a563cfd3945d1874c076416b7b4595502de6efe0e33e18e52681c09d7e16cf", "4fa540ef7d79d61837ff8936bfb9c5f1401c3c23912d0fc1bfc26c42f884c5a7", "6079fe3795ac40064c846f4e7f6f30079d8c6d673750fec0044b4cda028a6d03", "98738afafe146aa6ff16aa00fca789b599e3216ade9feedd78bffee97b91aa87", "472659c4e186f2523dbf41701ef79a17fbd5b0f5724690bd0e8ffee7e169633f"},
		4: []string{"fdb671804ea5e9f77f4d6168c76f16ee4caae8ce10ae8ac06fcb09a6ba831da4", "8189adffea72ba50e79e923dc01c5f92470c2592c20e70ff75e2aea2890e7294", "b2d0c676b5270dd2e97552598e1a1c211295e5b427804c9331d0f152b515b634", "e0128cbcc2b7386c560b0dbc78af3cf7dfa5bc436cecfe6255963c4c3e041f24", "c301de7b075d191dea696cafd7b054d72cc4db1d5931176d22c1e1bb1373684c", "0989863eec936d4248ade93eb6da36f0275ffbb5fda9805847f0e86f06503b04", "f9c3666d0f9e3aa8180c2fbc36ec3388f0e2bb4e302224db68f77cf21ca3ddbc", "ab16bd554c5fc56dc45793ce90fcd751d7bdb5074b4ac1b122c4fc68c306fb9c", "80d5495a9e4c6da7e8bb1bb842e3cb723c9390d17565ef96d4b24afb59aa7ccc", "adfd7a1d32c082f03913a8b00903b814abf2097ecfa4e755bde33a25f2efb65c", "5ca6eb6c6feafa58658de8bc8a5af28617048204f523d39e4f98c1747f24ddcc", "06d73c2a496127050a0cd1e89a19725013e94fd6839e64aba9b21b91785dabec"},
	}

	rand.Seed(time.Now().Unix())

	targetsForShard := targetMap[shard]
	targetForShard := targetsForShard[rand.Intn(len(targetsForShard))]

	return targetForShard
}

// GenerateCryptoSuite - generate all required objects for signing
func GenerateCryptoSuite(
	pemFilePath string,
	pemIndex int,
	suite crypto.Suite,
) (keyGen crypto.KeyGenerator, privKey crypto.PrivateKey, pubKey crypto.PublicKey, err error) {
	encodedSk, err := core.LoadSkFromPemFile(pemFilePath, pemIndex)

	if err != nil {
		return nil, nil, nil, err
	}

	decodedSk, err := hex.DecodeString(string(encodedSk))

	keyGen = signing.NewKeyGenerator(suite)

	privKey, err = keyGen.PrivateKeyFromByteArray(decodedSk)
	if err != nil {
		return nil, nil, nil, err
	}

	pubKey = privKey.GeneratePublic()

	return keyGen, privKey, pubKey, err
}

// IdentifyAddressShard - identifies what shard an address belongs to
func IdentifyAddressShard(address string) int {
	lastAddressCharacter := strings.ToLower(string(address[len(address)-1]))

	var shard int

	switch lastAddressCharacter {
	case "0", "8":
		shard = 0
	case "1", "5", "9", "d":
		shard = 1
	case "2", "6", "a", "e":
		shard = 2
	case "3", "7", "b", "f":
		shard = 3
	case "4", "c":
		shard = 4
	}

	return shard
}

// IdentifyPemFiles - identify pem files from a specified path
func IdentifyPemFiles(path string) ([]string, error) {
	pattern := fmt.Sprintf("%s/*.pem", path)

	fmt.Println("Pem pattern is now: ", pattern)

	keys, err := globFiles(pattern)

	if err != nil {
		return nil, err
	}

	return keys, err
}

func globFiles(pattern string) ([]string, error) {
	files, err := filepath.Glob(pattern)

	if err != nil {
		return nil, err
	}

	return files, nil
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// ReadFileToString - check if a file exists, proceed to read it to memory if it does
func ReadFileToString(filePath string) (string, error) {
	if fileExists(filePath) {
		data, err := ioutil.ReadFile(filePath)

		if err != nil {
			return "", err
		}

		return string(data), nil
	} else {
		return "", nil
	}
}
