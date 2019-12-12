# Elrond Tx Sender

This tool sends *a lot* of transactions on Elrond's blockchain.

## Usage

All available options:
```
$ ./elrond-tx-sender --help
NAME:
   Elrond Tx Sender CLI App - This is the entry point for starting a new Elrond tx sender
USAGE:
   elrond-tx-sender [global options]
   
AUTHOR:
   Sebastian Johnsson <sebastian.johnsson@gmail.com>
   
GLOBAL OPTIONS:
   --api-endpoint value  Which API endpoint to use for API commands
   --tx-receiver value   Which address to send tokens to (disabled for now)
   --tx-count value      How many transactions to send (default: 1)
   --tx-data-file value  Which file to use for tx data (default: "./tx_data.txt")
   --proxy-file value    Which file to use for reading proxies (default: "./proxies.txt")
   --keys-path value     Path to keys (default: "./keys")
   --disable-ansi-color  This flag specifies that the log output should not use ANSI colors
   --logLevel value      This flag specifies the logger level (default: "*:INFO ")
   --help, -h            show help
   --version, -v         print the version
   
VERSION:
   go1.13.4/linux-amd64
```

### Example usage

Clone the repo and compile the binary:

```
git clone https://github.com/SebastianJ/elrond-tx-sender.git
go build
```

Create the folder `keys` and add all of the sending pem keys you want to use to send transactions (doesn't matter what they are named - the tool will look for *.pem in that folder).

#### Local nodes (using the nodes defined in utils/utilities.go, method HostForShard)

`./elrond-tx-sender --tx-count 1000`


#### Remote wallet API node

`./elrond-tx-sender --tx-count 1000 --api-endpoint https://wallet-api.elrond.com`

For the remote wallet there's also proxy support implemented to not get IP blocked by https://wallet-api.elrond.com. Create proxies.txt in the same folder as the elrond-tx-sender binary (or specify a custom path using --proxy-file PATH)

#### With tx payload

Create the file tx_data.txt in the same folder as the elrond-tx-sender binary (or supply --tx-data-file pointing to your custom tx data file).

Example files are included in examples: 4chan_marine_tx_data.txt (499,500 bytes) and mrbubz_tx_data.txt (126,896 bytes) - for my testing I used both, but primarily the larger 4chan_marine_tx_data.txt payload.

## Notes

Some of the code in the code base is hard coded - this will be refactored.
