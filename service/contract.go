package service

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"hedgex-server/config"
	"hedgex-server/contract/hedgex"
	"hedgex-server/gl"
	"log"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

var (
	//define the client to connect to the ethereum network
	EthHttpsClient *ethclient.Client
	EthWssClient   *ethclient.Client

	//define the contract's abi
	ContractAbi abi.ABI

	//definde the hash string of contract's event
	MintEvent         string
	BurnEvent         string
	RechargeEvent     string
	WithdrawEvent     string
	TradeEvent        string
	ExplosiveEvent    string
	TakeInterestEvent string
	EventNames        map[string]string

	//contract's instance
	Contracts map[string]*hedgex.Hedgex

	//
	privateKey    *ecdsa.PrivateKey
	publicAddress common.Address

	erc20TransferID []byte
	chainID         *big.Int
)

func init() {
	var err error

	EthHttpsClient, err = ethclient.Dial(config.ChainNode.Https)
	if err != nil {
		log.Panic(err)
	}

	Contracts = make(map[string]*hedgex.Hedgex)
	for i := range config.Contract {
		Contracts[config.Contract[i].Address], err = hedgex.NewHedgex(common.HexToAddress(config.Contract[i].Address), EthHttpsClient)
		if err != nil {
			log.Panic(err)
		}
	}

	EthWssClient, err = ethclient.Dial(config.ChainNode.Wss)
	if err != nil {
		log.Panic(err)
	}

	ContractAbi, err = abi.JSON(strings.NewReader(string(hedgex.HedgexABI)))
	if err != nil {
		log.Panic(err)
	}

	MintEvent = crypto.Keccak256Hash([]byte(ContractAbi.Events["Mint"].Sig)).Hex()
	BurnEvent = crypto.Keccak256Hash([]byte(ContractAbi.Events["Burn"].Sig)).Hex()
	RechargeEvent = crypto.Keccak256Hash([]byte(ContractAbi.Events["Recharge"].Sig)).Hex()
	WithdrawEvent = crypto.Keccak256Hash([]byte(ContractAbi.Events["Withdraw"].Sig)).Hex()
	TradeEvent = crypto.Keccak256Hash([]byte(ContractAbi.Events["Trade"].Sig)).Hex()
	ExplosiveEvent = crypto.Keccak256Hash([]byte(ContractAbi.Events["Explosive"].Sig)).Hex()
	TakeInterestEvent = crypto.Keccak256Hash([]byte(ContractAbi.Events["TakeInterest"].Sig)).Hex()

	EventNames = make(map[string]string)
	EventNames[MintEvent] = "Mint"
	EventNames[BurnEvent] = "Burn"
	EventNames[RechargeEvent] = "Recharge"
	EventNames[WithdrawEvent] = "Withdraw"
	EventNames[TradeEvent] = "Trade"
	EventNames[ExplosiveEvent] = "Explosive"
	EventNames[TakeInterestEvent] = "TakeInterest"

	erc20TransferID = []byte{0xa9, 0x05, 0x9c, 0xbb} //transfer(address,uint256)

	chainID, err = EthHttpsClient.NetworkID(context.Background())
	if err != nil {
		log.Panic(err)
	}
}

func getAccountAuth() (*bind.TransactOpts, error) {
	gasPrice, err := EthHttpsClient.SuggestGasPrice(context.Background())
	if err != nil {
		return nil, err
	}
	chainID, err := EthHttpsClient.NetworkID(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, chainID) // bind.NewKeyedTransactor(privateKey)
	if err != nil {
		return nil, err
	}
	auth.Value = big.NewInt(0)     // in wei
	auth.GasLimit = uint64(300000) // in units
	auth.GasPrice = gasPrice
	return auth, nil
}

func SendTestCoins(account string) error {
	if err := sendEth(account); err != nil {
		return err
	}
	if err := sendERC20(account); err != nil {
		return err
	}
	return nil
}

func sendEth(to string) error {
	value, _ := new(big.Int).SetString(config.TestCoin.CoinAnount, 10)
	return sendTransaction(common.HexToAddress(to), value, nil)
}

func sendERC20(to string) error { // in wei (0 eth)
	paddedAddress := common.LeftPadBytes(common.HexToAddress(to).Bytes(), 32) // 0x0000000000000000000000004592d8f8d7b001e72cb26a73e4fa1806a51ac79d
	amount, _ := new(big.Int).SetString(config.TestCoin.TokenAmount, 10)      // amount
	paddedAmount := common.LeftPadBytes(amount.Bytes(), 32)                   // 0x00000000000000000000000000000000000000000000003635c9adc5dea00000

	var data []byte
	data = append(data, erc20TransferID...)
	data = append(data, paddedAddress...)
	data = append(data, paddedAmount...)
	value := big.NewInt(0)
	return sendTransaction(common.HexToAddress(config.TestCoin.Token), value, data)
}

func sendTransaction(to common.Address, value *big.Int, data []byte) error {
	gasPrice, err := EthHttpsClient.SuggestGasPrice(context.Background())
	if err != nil {
		gl.OutLogger.Error("get gas price error. %v", err)
		return err
	}
	nonce, err := EthHttpsClient.PendingNonceAt(context.Background(), publicAddress)
	if err != nil {
		gl.OutLogger.Error("get nonce error address(%s). %v", publicAddress, err)
		return err
	}
	gasLimit := uint64(3000000)
	fmt.Println(nonce)
	tx := types.NewTransaction(nonce, to, value, gasLimit, gasPrice, data)
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
	if err != nil {
		gl.OutLogger.Error("create signedTx error. %v", err)
		return err
	}

	err = EthHttpsClient.SendTransaction(context.Background(), signedTx)
	if err != nil {
		gl.OutLogger.Error("send signedTx error. %v", err)
		return err
	}
	return nil
}
