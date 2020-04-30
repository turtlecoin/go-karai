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
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/gorilla/mux"
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
const appURL = "https://karai.io"

// File & folder constants
const credentialsFile = "private_credentials.karai"
const currentJSON = "./config/milestone.json"
const graphDir = "./graph"
const hashDat = graphDir + "/ipfs-hash-list.dat"
const p2pConfigDir = "./config/p2p"
const configPeerIDFile = p2pConfigDir + "/peer.id"

// Coordinator values
var isCoordinator bool = false
var karaiPort string = "4200"
var p2pPeerID string

// Version string
func semverInfo() string {
	var majorSemver, minorSemver, patchSemver, wholeString string
	majorSemver = "0"
	minorSemver = "4"
	patchSemver = "6"
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
}

// // SubGraph This is a struct for Tx wave construction
// type SubGraph struct {
// 	subGraphID       int
// 	timeStamp        int64
// 	milestone        int
// 	transactions     []byte
// 	subgraphChildren int
// 	supgraphOrder    int
// 	subgraphSize     int
// 	subgraphPeers    []byte
// 	waveTip          *GraphTx.Hash
// }

// Hello Karai
func main() {
	clearPeerID(configPeerIDFile)
	locateGraphDir()
	checkCreds()
	ascii()
	go restAPI()
	inputHandler()
}

func restAPI() {
	r := mux.NewRouter()

	api := r.PathPrefix("/api/v1").Subrouter()
	api.HandleFunc("/", returnPeerID).Methods(http.MethodGet)
	api.HandleFunc("/version", returnVersion).Methods(http.MethodGet)
	api.HandleFunc("/transactions", returnTransactions).Methods(http.MethodGet)
	// api.HandleFunc("", post).Methods(http.MethodPost)
	// api.HandleFunc("", put).Methods(http.MethodPut)
	// api.HandleFunc("", delete).Methods(http.MethodDelete)

	logrus.Error(http.ListenAndServe(":"+karaiPort, r))
}

func notFound(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte(`{"bruh": "lol"}`))
}

func returnPeerID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	peerFile, err := os.OpenFile(configPeerIDFile,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		handle("something went wrong creating a fresh peer.id file: ", err)
	}
	defer peerFile.Close()

	fileToRead, err := ioutil.ReadFile(configPeerIDFile)
	handle("nazi porn", err)
	logrus.Debug("Peer ID requested from API")
	// fmt.Print("\n", string(fileToRead))
	w.Write([]byte("{\"p2p_peer_ID\": \"" + string(fileToRead) + "\"}"))

}
func printFile(fileToPrint string) string {
	file, err := ioutil.ReadFile(configPeerIDFile)
	handle("There was a problem reading the peer file", err)
	fmt.Print(string(file))
	return string(file)
}
func returnVersion(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("{\"karai_version\": \"" + semverInfo() + "\"}"))
}

func returnTransactions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	matches, _ := filepath.Glob(graphDir + "/*.json")
	w.Write([]byte("[\n\t"))
	for _, match := range matches {
		w.Write([]byte(printTx(match)))
	}
	w.Write([]byte("{}"))
	w.Write([]byte("\n]"))
}

// func post(w http.ResponseWriter, r *http.Request) {
// 	w.Header().Set("Content-Type", "application/json")
// 	w.WriteHeader(http.StatusCreated)
// 	w.Write([]byte(`{"": ""}`))
// }

// func put(w http.ResponseWriter, r *http.Request) {
// 	w.Header().Set("Content-Type", "application/json")
// 	w.WriteHeader(http.StatusAccepted)
// 	w.Write([]byte(`{"": ""}`))
// }

// func delete(w http.ResponseWriter, r *http.Request) {
// 	w.Header().Set("Content-Type", "application/json")
// 	w.WriteHeader(http.StatusOK)
// 	w.Write([]byte(`{"": ""}`))
// }

// Splash logo
func ascii() {
	fmt.Printf("\n")
	color.Set(color.FgGreen, color.Bold)
	fmt.Printf("|   _   _  _  .\n")
	fmt.Printf("|( (_| |  (_| |\n")
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
	start := time.Now()

	matches, _ := filepath.Glob(graphDir + "/*.json")
	for _, match := range matches {
		pushTx(match)
	}

	end := time.Since(start)
	fmt.Println("Finished in: ", end)
}

func pushTx(file string) string {
	dat, _ := ioutil.ReadFile(file)
	color.Set(color.FgBlack, color.Bold)
	fmt.Print(string(dat) + "\n")
	sh := shell.NewShell("localhost:5001")
	cid, err := sh.Add(strings.NewReader(string(dat)))
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s", err)
		os.Exit(1)
	}
	fmt.Printf(color.GreenString("%v %v\n%v %v", color.YellowString("Tx:"), color.GreenString(file), color.YellowString("CID: "), color.GreenString(cid)))
	appendGraphCID(cid)
	return cid
}

func printTx(file string) string {
	dat, _ := ioutil.ReadFile(file)
	datString := string(dat) + ",\n"
	return datString
}

func appendGraphCID(cid string) {
	hashfile, err := os.OpenFile(hashDat,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		handle("something went wrong: ", err)
	}
	defer hashfile.Close()
	if isExist(cid, hashDat) {
		fmt.Printf("%v", color.RedString("\nDuplicate! Skipping...\n"))
	} else {
		hashfile.WriteString(cid + "\n")
	}

}

func isExist(str, filepath string) bool {
	accused, _ := ioutil.ReadFile(filepath)
	isExist, _ := regexp.Match(str, accused)
	return isExist
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
	var isCoordinator bool = true
	fmt.Printf("Coordinator status: %t", isCoordinator)
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
	logrus.Debug("Creating a new Karai <=> TRTL pointer")
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

// createPeer Create Libp2p Peer
func createPeer() peer.ID {
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

	return nodePeer.ID()
}

func clearPeerID(file string) {
	err := os.Remove(file)
	handle("Error deleting stale peer file: ", err)
}

func menuCreatePeer() {
	clearPeerID(configPeerIDFile)
	p2pPeerID := createPeer()
	openPeerIDFile, err := os.OpenFile(configPeerIDFile,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		handle("something went wrong: ", err)
	}
	defer openPeerIDFile.Close()

	openPeerIDFile.WriteString(p2pPeerID.Pretty())
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
	graph.addTx(2, "{\"tx_slot\": 3}")
	// go txHandler()
	// Report Txs
	fmt.Printf("\n\nTx Legend: %v %v %v\n", color.YellowString("Root"), color.GreenString("Milestone"), color.BlueString("Normal"))
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
			// fmt.Printf("\nTx(%x) %x\n", key, transaction.Hash)
			fmt.Printf("\nTx(%v) %x\n", color.YellowString(strconv.Itoa(key)), transaction.Hash)
		} else if len(prevHash) > 2 {
			dataString := "{\n\t\"tx_type\": " + strconv.Itoa(transaction.TxType) + ",\n\t\"tx_hash\": \"" + hash + "\",\n\t\"tx_prev\": \"" + prevHash + "\",\n\t\"tx_extra\": " + string(transaction.Extra) + "\n}"
			f, _ := os.Create(graphDir + "/" + "Tx_" + strconv.Itoa(key) + ".json")
			w := bufio.NewWriter(f)
			w.WriteString(dataString)
			w.Flush()
			// Indicate Tx type by color
			if transaction.TxType == 0 {
				// Root Tx
				fmt.Printf("Tx(%v) %x\n", color.YellowString(strconv.Itoa(key)), transaction.Hash)
			} else if transaction.TxType == 1 {
				// Milestone Tx
				fmt.Printf("Tx(%v) %x\n", color.GreenString(strconv.Itoa(key)), transaction.Hash)
			} else if transaction.TxType == 2 {
				// Normal Tx
				fmt.Printf("Tx(%v) %x\n", color.BlueString(strconv.Itoa(key)), transaction.Hash)
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
	fmt.Printf("\n\nTx Legend: %v %v %v\n", color.YellowString("Root"), color.GreenString("Milestone"), color.BlueString("Normal"))
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
			// fmt.Printf("\nTx(%x) %x\n", key, transaction.Hash)
			fmt.Printf("\nTx(%v) %x\n", color.YellowString(strconv.Itoa(key)), transaction.Hash)
		} else if len(prevHash) > 2 {
			dataString := "{\n\t\"tx_type\": " + strconv.Itoa(transaction.TxType) + ",\n\t\"tx_hash\": \"" + hash + "\",\n\t\"tx_prev\": \"" + prevHash + "\",\n\t\"tx_extra\": " + string(transaction.Extra) + "\n}"
			f, _ := os.Create(graphDir + "/" + "Tx_" + strconv.Itoa(key) + ".json")
			w := bufio.NewWriter(f)
			w.WriteString(dataString)
			w.Flush()
			// Indicate Tx type by color
			if transaction.TxType == 0 {
				// Root Tx
				fmt.Printf("Tx(%v) %x\n", color.YellowString(strconv.Itoa(key)), transaction.Hash)
			} else if transaction.TxType == 1 {
				// Milestone Tx
				fmt.Printf("Tx(%v) %x\n", color.GreenString(strconv.Itoa(key)), transaction.Hash)
			} else if transaction.TxType == 2 {
				// Normal Tx
				fmt.Printf("Tx(%v) %x\n", color.BlueString(strconv.Itoa(key)), transaction.Hash)
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
		fmt.Printf("\n%v%v%v\n", color.WhiteString("Type '"), color.GreenString("menu"), color.WhiteString("' to view a list of commands"))
		fmt.Print(color.GreenString("-> "))
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
	color.Set(color.FgGreen)
	fmt.Println("\nCHANNEL_OPTIONS")
	color.Set(color.FgWhite)
	fmt.Println("create-channel \t\t Create a karai transaction channel")
	fmt.Println("generate-pointer \t Generate a Karai <=> TRTL pointer")
	fmt.Println("benchmark \t\t Conducts timed benchmark")
	fmt.Println("push-graph \t\t Prints graph history")
	color.Set(color.FgGreen)
	fmt.Println("\nWALLET_API_OPTIONS")
	color.Set(color.FgWhite)
	fmt.Println("open-wallet \t\t Open a TRTL wallet")
	fmt.Println("open-wallet-info \t Show wallet and connection info")
	fmt.Println("create-wallet \t\t Create a TRTL wallet")
	color.Set(color.FgHiBlack)
	fmt.Println("wallet-balance \t\t Displays wallet balance")
	color.Set(color.FgGreen)
	fmt.Println("\nIPFS_OPTIONS")
	color.Set(color.FgWhite)
	fmt.Println("create-peer \t\t Creates IPFS peer")
	color.Set(color.FgHiBlack)
	fmt.Println("list-servers \t\t Lists pinning servers")
	color.Set(color.FgGreen)
	fmt.Println("\nGENERAL_OPTIONS")
	color.Set(color.FgWhite)
	fmt.Println("version \t\t Displays version")
	fmt.Println("license \t\t Displays license")
	fmt.Println("exit \t\t\t Quit immediately")
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
	fmt.Printf(color.GreenString("\n"+appName+" v"+semverInfo()) + color.WhiteString(" by "+appDev))
	color.Set(color.FgGreen)
	fmt.Println("\n" + appRepository + "\n" + appURL + "\n")

	color.Set(color.FgHiWhite)
	fmt.Println("\nMIT License\nCopyright (c) 2020-2021 RockSteady, TurtleCoin Developers")
	color.Set(color.FgHiBlack)
	fmt.Println("\nPermission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the 'Software'), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:\n\nThe above copyright notice and this permission notice shall be included in allcopies or substantial portions of the Software.\n\nTHE SOFTWARE IS PROVIDED 'AS IS', WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.")
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
