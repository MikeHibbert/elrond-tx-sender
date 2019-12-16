package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"sync"
	"syscall"

	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/crypto/signing/kyber"
	"github.com/ElrondNetwork/elrond-go/crypto/signing/kyber/singlesig"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/display"
	"github.com/ElrondNetwork/elrond-go/logger"
	"github.com/SebastianJ/elrond-tx-sender/api"
	senderUtils "github.com/SebastianJ/elrond-tx-sender/utils"
	"github.com/urfave/cli"
)

var (
	helpTemplate = `NAME:
   {{.Name}} - {{.Usage}}
USAGE:
   {{.HelpName}} {{if .VisibleFlags}}[global options]{{end}}
   {{if len .Authors}}
AUTHOR:
   {{range .Authors}}{{ . }}{{end}}
   {{end}}{{if .Commands}}
GLOBAL OPTIONS:
   {{range .VisibleFlags}}{{.}}
   {{end}}
VERSION:
   {{.Version}}
   {{end}}
`
	// endpoint defines the API endpoint to use.
	apiEndpoint = cli.StringFlag{
		Name:  "api-endpoint",
		Usage: "Which API endpoint to use for API commands",
		Value: "",
	}

	// txReceiver defines the address to send tokens to - not used right now
	txReceiver = cli.StringFlag{
		Name:  "tx-receiver",
		Usage: "Which address to send tokens to (disabled for now)",
		Value: "",
	}

	// txCount defines the number of transactions to send
	txCount = cli.IntFlag{
		Name:  "tx-count",
		Usage: "How many transactions to send",
		Value: 1,
	}

	// txDataFile defines the file to use for tx data
	txDataFile = cli.StringFlag{
		Name:  "tx-data-file",
		Usage: "Which file to use for tx data",
		Value: "./tx_data.txt",
	}

	// txDataFile defines the file to use for tx data
	proxyFile = cli.StringFlag{
		Name:  "proxy-file",
		Usage: "Which file to use for reading proxies",
		Value: "./proxies.txt",
	}

	// keysPath defines the path to the initialBalances.pem files to use for signing transactions
	keysPath = cli.StringFlag{
		Name:  "keys-path",
		Usage: "Path to keys",
		Value: "./keys",
	}
	// logLevel defines the logger level
	logLevel = cli.StringFlag{
		Name:  "logLevel",
		Usage: "This flag specifies the logger level",
		Value: "*:" + logger.LogInfo.String(),
	}
	// disableAnsiColor defines if the logger subsystem should prevent displaying ANSI colors
	disableAnsiColor = cli.BoolFlag{
		Name:  "disable-ansi-color",
		Usage: "This flag specifies that the log output should not use ANSI colors",
	}
)

func main() {
	_ = display.SetDisplayByteSlice(display.ToHexShort)
	log := logger.GetOrCreate("main")

	app := cli.NewApp()
	cli.AppHelpTemplate = helpTemplate
	app.Name = "Elrond Tx Sender CLI App"
	app.Version = fmt.Sprintf("%s/%s-%s", runtime.Version(), runtime.GOOS, runtime.GOARCH)
	app.Usage = "This is the entry point for starting a new Elrond tx sender"
	app.Flags = []cli.Flag{
		apiEndpoint,
		txReceiver,
		txCount,
		txDataFile,
		proxyFile,
		keysPath,
		disableAnsiColor,
		logLevel,
	}
	app.Authors = []cli.Author{
		{
			Name:  "Mike Hibbert",
			Email: "mike@hibbertitsolutions.co.uk",
		},
	}

	app.Action = func(c *cli.Context) error {
		return startSender(c, log, app.Version)
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
}

func startSender(ctx *cli.Context, log logger.Logger, version string) error {
	log.Trace("startSender called")
	logLevel := ctx.GlobalString(logLevel.Name)
	err := logger.SetLogLevel(logLevel)
	if err != nil {
		return err
	}
	noAnsiColor := ctx.GlobalBool(disableAnsiColor.Name)
	if noAnsiColor {
		err = logger.RemoveLogObserver(os.Stdout)
		if err != nil {
			//we need to print this manually as we do not have console log observer
			fmt.Println("error removing log observer: " + err.Error())
			return err
		}

		err = logger.AddLogObserver(os.Stdout, &logger.PlainFormatter{})
		if err != nil {
			//we need to print this manually as we do not have console log observer
			fmt.Println("error setting log observer: " + err.Error())
			return err
		}
	}
	log.Trace("logger updated", "level", logLevel, "disable ANSI color", noAnsiColor)

	log.Info("starting tx sender", "version", version, "pid", os.Getpid())

	log.Info("application is now running")
	txCount := ctx.GlobalInt(txCount.Name)
	txDataFilePath := ctx.GlobalString(txDataFile.Name)
	txData, err := senderUtils.ReadFileToString(txDataFilePath)

	if txData != "" && err == nil {
		log.Info("Found tx data from file ", txDataFilePath)
	}

	proxyPath, _ := filepath.Abs(ctx.GlobalString(proxyFile.Name))
	proxies, _ := senderUtils.FetchProxies(proxyPath)

	if len(proxies) > 0 {
		log.Info(fmt.Sprintf("Found %d proxies in file: %s", len(proxies), proxyPath))
	}

	certKeyPath, err := filepath.Abs(ctx.GlobalString(keysPath.Name))
	pemCerts, err := senderUtils.IdentifyPemFiles(certKeyPath)

	apiHost := ctx.GlobalString(apiEndpoint.Name)

	log.Info("Looking for pem certs!")

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	for _, pemCert := range pemCerts {
		log.Info(fmt.Sprintf("Found pem certificate: %s", pemCert))
		go bulkSendTxs(pemCert, apiHost, txCount, txData, proxies, log)
		//bulkSendTxs(pemCert, txCount, txData, proxies, log)
	}

	<-sigs

	log.Info("terminating at user's signal...")

	return nil
}

func bulkSendTxs(pemCert string, apiHost string, txCount int, txData string, proxies []string, log logger.Logger) error {
	_, privKey, pubKey, err := senderUtils.GenerateCryptoSuite(pemCert, 0, kyber.NewBlakeSHA256Ed25519())

	pubKeyBytes, _ := pubKey.ToByteArray()
	sender := hex.EncodeToString(pubKeyBytes)
	hexSender, _ := hex.DecodeString(sender)
	senderShard := senderUtils.IdentifyAddressShard(sender)

	if apiHost == "" {
		apiHost = senderUtils.HostForShard(senderShard)
	}

	receiver := senderUtils.RandomReceiverTarget(senderShard)
	hexReceiver, _ := hex.DecodeString(receiver)
	receiverShard := senderUtils.IdentifyAddressShard(receiver)

	log.Info(fmt.Sprintf("Sender: %s (shard %d)", sender, senderShard))
	log.Info(fmt.Sprintf("Will use API Host: %s", apiHost))

	proxy := senderUtils.RandomProxy(proxies)
	accountData, err := api.GetAccount(apiHost, sender, proxy, log)

	if err != nil {
		log.Error("Failed to retrieve account data", err.Error())
		return err
	}

	fmt.Println("Nonce is now:", accountData.Nonce)
	fmt.Println("Balance is now:", accountData.Balance)

	nonce := accountData.Nonce
	amount := big.NewInt(1) // 10000 = 1 token
	gasPrice := uint64(1000000000000000)
	gasLimit := 100000 + uint64(len(txData)) // 10 ERD fee when sending 0 amount and no data

	respond := make(chan error, txCount)
	var wg sync.WaitGroup
	wg.Add(txCount)

	for {
		for i := 0; i < txCount; i++ {
			proxy := senderUtils.RandomProxy(proxies)
			go sendTx(respond, &wg, apiHost, privKey, sender, hexSender, senderShard, receiver, hexReceiver, receiverShard, amount, gasPrice, gasLimit, txData, nonce, proxy, log)

			if err != nil {
				return err
			}

			nonce++
		}
		wg.Wait()
		close(respond)
	}

	return nil
}

func sendTx(respond chan<- error, wg *sync.WaitGroup, apiHost string, privKey crypto.PrivateKey, sender string, hexSender []byte, senderShard int, receiver string, hexReceiver []byte, receiverShard int, amount *big.Int, gasPrice uint64, gasLimit uint64, txData string, nonce uint64, proxy string, log logger.Logger) error {
	defer wg.Done()

	log.Info("")
	log.Info(fmt.Sprintf("Sender: %s (shard %d)", sender, senderShard))
	log.Info(fmt.Sprintf("Receiver: %s (shard %d)", receiver, receiverShard))
	log.Info(fmt.Sprintf("Amount: %s", amount.String()))
	log.Info(fmt.Sprintf("Nonce: %d", nonce))

	txSingleSigner := &singlesig.SchnorrSigner{}

	tx := transaction.Transaction{
		Nonce:    nonce,
		SndAddr:  hexSender,
		RcvAddr:  hexReceiver,
		Value:    amount,
		Data:     txData,
		GasPrice: gasPrice,
		GasLimit: gasLimit,
	}

	txBuff, _ := json.Marshal(&tx)
	signature, _ := txSingleSigner.Sign(privKey, txBuff)

	txHexHash, txError := api.SendTransaction(apiHost, nonce, sender, receiver, amount.String(), gasPrice, gasLimit, txData, signature, proxy, log)

	if txError != nil {
		log.Error("Failed to send transaction", txError.Error())
		return txError
	}

	log.Info(fmt.Sprintf("Successfully sent transaction: %s", txHexHash))

	log.Info("")

	return nil
}
