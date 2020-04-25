package main

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/common-nighthawk/go-figure"
	shell "github.com/ipfs/go-ipfs-api"
	"github.com/libp2p/go-libp2p"
	autonat "github.com/libp2p/go-libp2p-autonat-svc"
	connmgr "github.com/libp2p/go-libp2p-connmgr"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	libp2pquic "github.com/libp2p/go-libp2p-quic-transport"
	routing "github.com/libp2p/go-libp2p-routing"
	secio "github.com/libp2p/go-libp2p-secio"
	libp2ptls "github.com/libp2p/go-libp2p-tls"
	"github.com/sirupsen/logrus"
	rashedCrypto "github.com/turtlecoin/go-turtlecoin/crypto"
	rashedMnemonic "github.com/turtlecoin/go-turtlecoin/walletbackend/mnemonics"
)

// Attribution constants
const appName = "go-karai"
const appDev = "The TurtleCoin Developers"
const appDescription = appName + " - Karai Transaction Channels"
const appLicense = "https://choosealicense.com/licenses/mit/"
const appRepository = "https://github.com/turtlecoin/go-karai"

// Location constants
const credentialsFile = "private_credentials.karai"
const currentJSON = "./config/milestone.json"
const graphDir = "./graph"
const hashDat = graphDir + "/ipfs-hash-list.dat"

// const paramFile = "./config/milestone.json"

// Version string
func semverInfo() string {
	var majorSemver, minorSemver, patchSemver, wholeString string
	majorSemver = "0"
	minorSemver = "3"
	patchSemver = "1"
	wholeString = majorSemver + "." + minorSemver + "." + patchSemver
	return wholeString
}

// Graph This is the structure of the Graph
type Graph struct {
	transactions []*GraphTx
}

// GraphTx This is the structure of the transaction
type GraphTx struct {
	TxType   int
	Hash     []byte
	Extra    []byte
	PrevHash []byte
	// TxVer int
	// WavePosition int
}

// // SubGraph This is a struct for Tx wave construction
// type SubGraph struct {
// 	subGraphID   int
// 	timeStamp    int64
// 	milestone   int
// 	transactions []byte
// 	// waveTip    GraphTx.Hash
// }

// Hello Karai
func main() {
	locateGraphDir()
	checkCreds()
	ascii()
	inputHandler()
}

// Splash logo
func ascii() {
	fmt.Println("\033[1;32m")
	splash := figure.NewFigure("karai", "straight", true)
	splash.Print()
	fmt.Println("\x1b[0m")
}

// checkCreds locate or create Karai credentials
func checkCreds() {
	if _, err := os.Stat(credentialsFile); err == nil {
		logrus.Debug("Karai Credentials Found!")
	} else {
		logrus.Debug("No Credentials Found! Generating Credentials...")
		generateEd25519()
	}
}

// generateEd25519 use TRTL Crypto to generate credentials
func generateEd25519() {
	logrus.Debug("Generating credentials")
	priv, pub, err := rashedCrypto.GenerateKeys()
	seed := rashedMnemonic.PrivateKeyToMnemonic(priv)
	timeUnixNow := strconv.FormatInt(time.Now().Unix(), 10)
	// TODO: Replace manually entered JSON
	logrus.Debug("Writing credentials to file")
	writeFile := []byte("{\n\t\"date_generated\": " + timeUnixNow + ",\n\t\"key_priv\": \"" + hex.EncodeToString(priv[:]) + "\",\n\t\"key_pub\": \"" + hex.EncodeToString(pub[:]) + "\",\n\t\"seed\": \"" + seed + "\"\n}")
	logrus.Debug("Writing main file")
	errWriteFile := ioutil.WriteFile("./"+credentialsFile, writeFile, 0644)
	logrus.Debug(errWriteFile)
	handle("Error writing file: ", err)
	logrus.Debug("Writing backup credential file")
	errWriteBackupFile := ioutil.WriteFile("./."+credentialsFile+"."+timeUnixNow+".backup", writeFile, 0644)
	handle("Error writing file backup: ", err)
	logrus.Debug(errWriteBackupFile)
}

// hashTx This will compute the tx hash using sha256
func (graphTx *GraphTx) hashTx() {
	// logrus.Debug("Hashing a Tx ", graphTx.Hash)
	data := bytes.Join([][]byte{graphTx.Extra, graphTx.PrevHash}, []byte{})
	hash := sha256.Sum256(data)
	graphTx.Hash = hash[:]
}

// addTx This will add a transaction to the graph
func (graph *Graph) addTx(txType int, data string) {
	// logrus.Debug("Adding a Tx")
	prevTx := graph.transactions[len(graph.transactions)-1]
	new := txConstructor(txType, data, prevTx.Hash)
	graph.transactions = append(graph.transactions, new)
}

func pushIPFS() {
	searchDir := graphDir
	// sh := shell.NewShell("localhost:5001")
	fileList := []string{}
	filepath.Walk(searchDir, func(path string, f os.FileInfo, err error) error {
		fileList = append(fileList, path)
		return nil
	})
	for _, file := range fileList {
		pushTx(file)
	}
}

func pushTx(file string) string {
	dat, _ := ioutil.ReadFile(file)
	fmt.Print("\033[0;90m" + string(dat) + "\n")
	sh := shell.NewShell("localhost:5001")
	cid, err := sh.Add(strings.NewReader(string(dat)))
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s", err)
		os.Exit(1)
	}
	fmt.Printf("\033[1;33madded \033[1;32m%s\033[1;33m for transaction \033[1;32m%s\033[1;33m", cid, file)
	appendGraphCID(cid)
	return cid
}

func appendGraphCID(cid string) {
	f, err := os.OpenFile(hashDat,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		handle("something went wrong: ", err)
	}
	defer f.Close()
	if _, err := f.WriteString(cid + "\n"); err != nil {
		handle("something went wrong: ", err)
	}
}

// addMilestone This will add a milestone to the graph
func (graph *Graph) addMilestone(data string) {
	prevTransaction := graph.transactions[len(graph.transactions)-1]
	// paramFile, _ = os.Open("./config/milestone.json")
	new := txConstructor(1, data, prevTransaction.Hash)
	graph.transactions = append(graph.transactions, new)
}

// txConstructor This will construct a tx
func txConstructor(txType int, data string, prevHash []byte) *GraphTx {
	transaction := &GraphTx{txType, []byte{}, []byte(data), prevHash}
	transaction.hashTx()
	return transaction
}

// rootTx Transaction channels start with a rootTx transaction always
func rootTx() *GraphTx {
	return txConstructor(0, "Karai Transaction Channel - Root", []byte{})
}

// spawnGraph starts a new transaction channel with Root Tx
func spawnGraph() *Graph {
	return &Graph{[]*GraphTx{rootTx()}}
}

// v4ToHex convert an ip4 to hex
func v4ToHex(addr string) string {
	ip := net.ParseIP(addr).To4()
	buffer := new(bytes.Buffer)
	for _, s := range ip {
		binary.Write(buffer, binary.BigEndian, uint8(s))
	}
	var dec uint32
	binary.Read(buffer, binary.BigEndian, &dec)
	return fmt.Sprintf("%08x", dec)
}

// portToHex convert a port to hex
func portToHex(port string) string {
	portNum, _ := strconv.ParseUint(port, 10, 16)
	return fmt.Sprintf("%04x", portNum)
}

// generatePointer create the TRTL <=> Karai pointer
func generatePointer() {
	logrus.Info("Creating a new Karai <=> TRTL pointer")
	readerKtxIP := bufio.NewReader(os.Stdin)
	fmt.Print("Enter Karai Coordinator IP: ")
	ktxIP, _ := readerKtxIP.ReadString('\n')
	readerKtxPort := bufio.NewReader(os.Stdin)
	fmt.Print("Enter Karai Coordinator Port: ")
	ktxPort, _ := readerKtxPort.ReadString('\n')
	ip := v4ToHex(strings.TrimRight(ktxIP, "\n"))
	port := portToHex(strings.TrimRight(ktxPort, "\n"))
	fmt.Printf("\nGenerating pointer for %s:%s\n", strings.TrimRight(ktxIP, "\n"), ktxPort)
	fmt.Println("Your pointer is: ")
	fmt.Printf("Hex:\t6b747828%s%s29", ip, port)
	fmt.Println("\nAscii:\tktx(" + strings.TrimRight(ktxIP, "\n") + ":" + strings.TrimRight(ktxPort, "\n") + ")")
}

// loadMilestoneJSON Read pending milestone Tx JSON
func loadMilestoneJSON() string {
	// TODO: Check if milestone is ready first, avoid re-use
	dat, _ := ioutil.ReadFile(currentJSON)
	datMilestone := string(dat)
	return datMilestone
	// Kek
}

// // txHandler Wait for Tx, assemble subgraph
// func txHandler() {
// 	var txListenTime time.Duration = 10
// 	var txPoolDepth int = 0
// 	if txPoolDepth > 0 {
// 		// if a tx is received, start the interval, listen for Tx, assemble subgraph
// 		// var int64 SubGraph.timeStamp = time.Now().Unix()
// 		// fmt.Println("Transaction Wave Forming...\nTimestamp: " + string(SubGraph.timeStamp))
// 		time.Sleep(txListenTime * time.Second)
// 		fmt.Println("Listening for " + string(txListenTime) + " seconds")
// 	}
// 	// order the transactions
// 	// assign positions on graph
// }

// menuCreatePeer Create Libp2p Peer
func menuCreatePeer() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	priv, _, err := crypto.GenerateKeyPair(
		crypto.Ed25519, -1,
	)
	handle("Error generating libp2p keypair: ", err)
	var idht *dht.IpfsDHT
	nodePeer, err := libp2p.New(ctx,
		libp2p.Identity(priv),
		libp2p.ListenAddrStrings(
			"/ip4/0.0.0.0/tcp/9000",
			"/ip4/0.0.0.0/udp/9000/quic",
		),
		libp2p.Security(libp2ptls.ID, libp2ptls.New),
		libp2p.Security(secio.ID, secio.New),
		libp2p.Transport(libp2pquic.NewTransport),
		libp2p.DefaultTransports,
		libp2p.ConnectionManager(connmgr.NewConnManager(
			100,         // Lowwater
			400,         // HighWater,
			time.Minute, // GracePeriod
		)),
		libp2p.NATPortMap(),
		libp2p.Routing(func(h host.Host) (routing.PeerRouting, error) {
			idht, err = dht.New(ctx, h)
			return idht, err
		}),
		libp2p.EnableAutoRelay(),
	)
	handle("Error connecting as libp2p peer: ", err)
	_, err = autonat.NewAutoNATService(ctx, nodePeer,
		libp2p.Security(libp2ptls.ID, libp2ptls.New),
		libp2p.Security(secio.ID, secio.New),
		libp2p.Transport(libp2pquic.NewTransport),
		libp2p.DefaultTransports,
	)
	for _, addr := range dht.DefaultBootstrapPeers {
		pi, _ := peer.AddrInfoFromP2pAddr(addr)
		nodePeer.Connect(ctx, *pi)
	}
	fmt.Printf("Peer ID is %s\n", nodePeer.ID())
}

// P2P stream open r/w
// func handleStream(s network.Stream) {
// 	logrus.Debug("New stream")
// 	rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))
// 	go readData(rw)
// 	go writeData(rw)
// }

// P2P read buffer, consume graph, verify integrity
// func readData(rw *bufio.ReadWriter) {
//  // TODO: Consume graph
//  // TODO: When Tx is received, increment TxPoolDepth
// }

// P2P write buffer
// func writeData(rw *bufio.ReadWriter) {
// 	stdReader := bufio.NewReader(os.Stdin)
// }

// spawnChannel Create a Tx Channel, Root Tx and Milestone, listen for Tx
func spawnChannel() {
	// Generate Root Tx
	graph := spawnGraph()
	// Add the current milestone.json in config
	graph.addMilestone(loadMilestoneJSON())
	// go txHandler()
	// Report Txs
	fmt.Println("\nTx Legend: \033[1;33mRoot\x1b[0m \033[1;32mMilestone\x1b[0m \033[1;34mNormal\x1b[0m")
	for key, transaction := range graph.transactions {
		var hash string = fmt.Sprintf("%x", transaction.Hash)
		var prevHash string = fmt.Sprintf("%x", transaction.PrevHash)
		// Root Tx will not have a previous hash
		if prevHash == "" {
			dataString := "{\n\t\"tx_type\": " + strconv.Itoa(transaction.TxType) + ",\n\t\"tx_hash\": \"" + hash + "\",\n\t\"tx_extra\": \"" + string(transaction.Extra) + "\"\n}"
			f, _ := os.Create(graphDir + "/" + "Tx_" + strconv.Itoa(key) + ".json")
			w := bufio.NewWriter(f)
			w.WriteString(dataString)
			w.Flush()
			fmt.Printf("\033[1;36mTx(\033[1;33m%x\033[1;36m)\x1b[0m %x\n", key, transaction.Hash)
		} else if len(prevHash) > 2 {
			dataString := "{\n\t\"tx_type\": " + strconv.Itoa(transaction.TxType) + ",\n\t\"tx_hash\": \"" + hash + "\",\n\t\"tx_prev\": \"" + prevHash + "\",\n\t\"tx_extra\": " + string(transaction.Extra) + "\n}"
			f, _ := os.Create(graphDir + "/" + "Tx_" + strconv.Itoa(key) + ".json")
			w := bufio.NewWriter(f)
			w.WriteString(dataString)
			w.Flush()
			// Indicate Tx type by color
			if transaction.TxType == 0 {
				// Root Tx
				fmt.Printf("\033[1;36mTx(\033[1;33m%x\033[1;36m)\x1b[0m %x\n", key, transaction.Hash)
			} else if transaction.TxType == 1 {
				// Milestone Tx
				fmt.Printf("\033[1;36mTx(\033[1;32m%x\033[1;36m)\x1b[0m %x\n", key, transaction.Hash)
			} else if transaction.TxType == 2 {
				// Normal Tx
				fmt.Printf("\033[1;36mTx(\033[1;34m%x\033[1;36m)\x1b[0m %x\n", key, transaction.Hash)
			}
		}
	}
	fmt.Println()
}

// benchmark Add a number of transactions and time the execution
func benchmark() {
	benchTxCount := 1000000
	graph := spawnGraph()
	graph.addMilestone(loadMilestoneJSON())
	count := 0
	ascii()
	fmt.Printf("Benchmark: %d transactions\n", benchTxCount)
	fmt.Println("Starting in 5 seconds. Press CTRL C to interrupt.")
	time.Sleep(5 * time.Second)
	start := time.Now()
	for i := 1; i < benchTxCount; i++ {
		count += i
		dataString := "{\"tx_slot\": " + strconv.Itoa(i+1) + "}"
		graph.addTx(2, dataString)
	}
	end := time.Since(start)
	fmt.Println("\nTx Legend: \033[1;33mRoot\x1b[0m \033[1;32mMilestone\x1b[0m \033[1;34mNormal\x1b[0m")
	for key, transaction := range graph.transactions {
		var hash string = fmt.Sprintf("%x", transaction.Hash)
		var prevHash string = fmt.Sprintf("%x", transaction.PrevHash)
		if prevHash == "" {
			dataString := "{\n\t\"tx_type\": " + strconv.Itoa(transaction.TxType) + ",\n\t\"tx_hash\": \"" + hash + "\",\n\t\"tx_extra\": \"" + string(transaction.Extra) + "\"\n}"
			// Write the Tx to disk in JSON format
			f, _ := os.Create(graphDir + "/" + "Tx_" + strconv.Itoa(key) + ".json")
			w := bufio.NewWriter(f)
			w.WriteString(dataString)
			w.Flush()
			fmt.Printf("\033[1;36mTx(\033[1;33m%x\033[1;36m)\x1b[0m %x\n", key, transaction.Hash)
		} else if len(prevHash) > 2 {
			dataString := "{\n\t\"tx_type\": " + strconv.Itoa(transaction.TxType) + ",\n\t\"tx_hash\": \"" + hash + "\",\n\t\"tx_prev\": \"" + prevHash + "\",\n\t\"tx_extra\": " + string(transaction.Extra) + "\n}"
			// Write the Tx to disk in JSON format
			f, _ := os.Create(graphDir + "/" + "Tx_" + strconv.Itoa(key) + ".json")
			w := bufio.NewWriter(f)
			w.WriteString(dataString)
			w.Flush()
			if transaction.TxType == 0 {
				// Root Tx
				fmt.Printf("\033[1;36mTx(\033[1;33m%x\033[1;36m)\x1b[0m %x\n", key, transaction.Hash)
			} else if transaction.TxType == 1 {
				// Milestone Tx
				fmt.Printf("\033[1;36mTx(\033[1;32m%x\033[1;36m)\x1b[0m %x\n", key, transaction.Hash)
			} else if transaction.TxType == 2 {
				// Normal Tx
				fmt.Printf("\033[1;36mTx(\033[1;34m%x\033[1;36m)\x1b[0m %x\n", key, transaction.Hash)
			}
		}
	}
	fmt.Println()
	fmt.Printf("%d Transactions in %s", benchTxCount, end)
}

// locateGraphDir find graph storage, create if missing.
func locateGraphDir() {
	if _, err := os.Stat(graphDir); os.IsNotExist(err) {
		logrus.Debug("Graph directory does not exist.")
		err = os.MkdirAll("./graph", 0755)
		handle("Error locating graph directory: ", err)
	}
}

// inputHandler present menu, accept user input
func inputHandler() {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Println("\n\033[0;37mType \033[1;32m'menu'\033[0;37m to view a list of commands\033[1;37m")
		fmt.Print("-> ")
		text, _ := reader.ReadString('\n')
		text = strings.Replace(text, "\n", "", -1)
		if strings.Compare("help", text) == 0 {
			menu()
		} else if strings.Compare("?", text) == 0 {
			menu()
		} else if strings.Compare("menu", text) == 0 {
			menu()
		} else if strings.Compare("version", text) == 0 {
			logrus.Debug("Displaying version")
			menuVersion()
		} else if strings.Compare("license", text) == 0 {
			logrus.Debug("Displaying license")
			printLicense()
		} else if strings.Compare("create-wallet", text) == 0 {
			logrus.Debug("Creating Wallet")
			menuCreateWallet()
		} else if strings.Compare("open-wallet", text) == 0 {
			logrus.Debug("Opening Wallet")
			menuOpenWallet()
		} else if strings.Compare("transaction-history", text) == 0 {
			logrus.Debug("Opening Transaction History")
			menuGetContainerTransactions()
		} else if strings.Compare("push-graph", text) == 0 {
			logrus.Debug("Opening Graph History")
			pushIPFS()
		} else if strings.Compare("open-wallet-info", text) == 0 {
			logrus.Debug("Opening Wallet Info")
			menuOpenWalletInfo()
		} else if strings.Compare("benchmark", text) == 0 {
			logrus.Debug("Benchmark")
			benchmark()
		} else if strings.Compare("create-peer", text) == 0 {
			menuCreatePeer()
		} else if strings.Compare("exit", text) == 0 {
			logrus.Warning("Exiting")
			menuExit()
		} else if strings.Compare("create-channel", text) == 0 {
			logrus.Debug("Creating Karai Transaction Channel")
			spawnChannel()
		} else if strings.Compare("generate-pointer", text) == 0 {
			generatePointer()
		} else if strings.Compare("quit", text) == 0 {
			logrus.Warning("Exiting")
			menuExit()
		} else if strings.Compare("close", text) == 0 {
			logrus.Warning("Exiting")
			menuExit()
		} else if strings.Compare("\n", text) == 0 {
			fmt.Println("")
		} else {
			fmt.Println("\nChoose an option from the menu")
			menu()
		}
	}
}

// provide list of commands
func menu() {
	fmt.Println("\n\033[1;32mCHANNEL_OPTIONS\033[1;37m\x1b[0m")
	fmt.Println("\033[1;37mcreate-channel \t\t \033[0;37mCreate a karai transaction channel\x1b[0m")
	fmt.Println("\033[1;37mgenerate-pointer \t \033[0;37mGenerate a Karai <=> TRTL pointer\x1b[0m")
	fmt.Println("\033[1;37mbenchmark \t\t \033[0;37mConducts timed benchmark\033[0m")
	fmt.Println("\033[1;37mpush-graph \t\t \033[0;37mPrints graph history\033[0m")
	fmt.Println("\n\033[1;32mWALLET_API_OPTIONS\033[1;37m\x1b[0m")
	fmt.Println("\033[1;37mopen-wallet \t\t \033[0;37mOpen a TRTL wallet\x1b[0m")
	fmt.Println("\033[1;37mopen-wallet-info \t \033[0;37mShow wallet and connection info\x1b[0m")
	fmt.Println("\033[1;37mcreate-wallet \t\t \033[0;37mCreate a TRTL wallet\x1b[0m")
	fmt.Println("\033[1;30mwallet-balance \t\t Displays wallet balance\x1b[0m")
	fmt.Println("\n\033[1;32mIPFS_OPTIONS\033[1;37m\x1b[0m")
	fmt.Println("\033[1;37mcreate-peer \t\t \033[0;37mCreates IPFS peer\x1b[0m")
	fmt.Println("\033[1;30mlist-servers \t\t Lists pinning servers\x1b[0m")
	fmt.Println("\n\033[1;32mGENERAL_OPTIONS\033[1;37m\x1b[0m")
	fmt.Println("\033[1;37mversion \t\t \033[0;37mDisplays version\033[0m")
	fmt.Println("\033[1;37mlicense \t\t \033[0;37mDisplays license\033[0m")
	fmt.Println("\033[1;37mexit \t\t\t \033[0;37mQuit immediately\x1b[0m")
	fmt.Println("")
}

// Some basic TRTL API stats
func menuOpenWalletInfo() {
	walletInfoPrimaryAddressBalance()
	getNodeInfo()
	getWalletAPIStatus()
}

// Get Wallet-API transactions
func menuGetContainerTransactions() {
	req, err := http.NewRequest("GET", "http://127.0.0.1:8070/transactions", nil)
	handle("Error getting container transactions: ", err)
	req.Header.Set("X-API-KEY", "pineapples")
	client := &http.Client{Timeout: time.Second * 10}
	resp, err := client.Do(req)
	handle("Error getting container transactions: ", err)
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	handle("Error getting container transactions: ", err)
	fmt.Printf("%s\n", body)
}

// Get Wallet-API status
func getWalletAPIStatus() {
	logrus.Info("[Wallet-API Status]")
	req, err := http.NewRequest("GET", "http://127.0.0.1:8070/status", nil)
	handle("Error getting Wallet-API status: ", err)
	req.Header.Set("X-API-KEY", "pineapples")
	client := &http.Client{Timeout: time.Second * 10}
	resp, err := client.Do(req)
	handle("Error getting Wallet-API status: ", err)
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	handle("Error getting Wallet-API status: ", err)
	fmt.Printf("%s\n", body)
}

// Get TRTL Node Info
func getNodeInfo() {
	logrus.Info("[Node Info]")
	req, err := http.NewRequest("GET", "http://127.0.0.1:8070/node", nil)
	handle("Error getting node info: ", err)
	req.Header.Set("X-API-KEY", "pineapples")
	client := &http.Client{Timeout: time.Second * 10}
	resp, err := client.Do(req)
	handle("Error getting node info: ", err)
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	handle("Error getting node info: ", err)
	fmt.Printf("%s\n", body)
}

// Get primary TRTL address balance
func walletInfoPrimaryAddressBalance() {
	logrus.Info("[Primary Address]")
	req, err := http.NewRequest("GET", "http://127.0.0.1:8070/balances", nil)
	handle("Error getting wallet info primary address: ", err)
	req.Header.Set("X-API-KEY", "pineapples")
	client := &http.Client{Timeout: time.Second * 10}
	resp, err := client.Do(req)
	handle("Error getting wallet info primary address: ", err)
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	handle("Error getting wallet info primary address: ", err)
	fmt.Printf("%s\n", body)
}

// Print the license for the user
func printLicense() {
	fmt.Println("\n\033[1;32m" + appName + " \033[0;32mv" + semverInfo() + "\033[0;37m by \033[1;37m" + appDev)
	fmt.Println("\033[0;32m" + appRepository + "\n")
	fmt.Println("\033[1;37mMIT License\n\nCopyright (c) 2020-2021 RockSteady, TurtleCoin Developers\n\033[1;30mPermission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the 'Software'), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:\n\nThe above copyright notice and this permission notice shall be included in allcopies or substantial portions of the Software.\n\nTHE SOFTWARE IS PROVIDED 'AS IS', WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.")
	fmt.Println()
}

// Create a wallet in the wallet-api container
func menuCreateWallet() {
	logrus.Debug("Creating Wallet")
	url := "http://127.0.0.1:8070/wallet/create"
	data := []byte(`{"daemonHost": "127.0.0.1",	"daemonPort": 11898, "filename": "karai-wallet.wallet", "password": "supersecretpassword"}`)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	handle("Error creating wallet: ", err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-KEY", "pineapples")
	client := &http.Client{Timeout: time.Second * 10}
	logrus.Info(req.Header)
	resp, err := client.Do(req)
	handle("Error creating wallet: ", err)
	defer resp.Body.Close()
	logrus.Info("response Status:", resp.Status)
	logrus.Info("response Headers:", resp.Header)
	body, err := ioutil.ReadAll(resp.Body)
	handle("Error creating wallet: ", err)
	fmt.Printf("%s\n", body)
}

// Open a wallet file
func menuOpenWallet() {
	logrus.Debug("Opening Wallet")
	url := "http://127.0.0.1:8070/wallet/open"
	data := []byte(`{"daemonHost": "127.0.0.1",	"daemonPort": 11898, "filename": "karai-wallet.wallet", "password": "supersecretpassword"}`)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	handle("Error opening wallet: ", err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-KEY", "pineapples")
	client := &http.Client{Timeout: time.Second * 10}
	logrus.Info(req.Header)
	resp, err := client.Do(req)
	handle("Error opening wallet: ", err)
	defer resp.Body.Close()
	logrus.Info("response Status:", resp.Status)
	logrus.Info("response Headers:", resp.Header)
	body, err := ioutil.ReadAll(resp.Body)
	handle("Error opening wallet: ", err)
	fmt.Printf("%s\n", body)
}

// Print the version string for the user
func menuVersion() {
	fmt.Println(appName + " - v" + semverInfo())
}

// Exit the program
func menuExit() {
	os.Exit(0)
}

func handle(msg string, err error) {
	if err != nil {
		logrus.Error(msg, err)
	}
}
