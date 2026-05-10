package main

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"
	"math/big"
	"os"
	"task02/contract"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/joho/godotenv"
)

func main() {
	// 加载配置
	godotenv.Load()
	client, err := ethclient.Dial(os.Getenv("SEPOLIA_RPC"))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	// 合约地址
	contractAddr := common.HexToAddress(os.Getenv("COUNTER_CONTRACT_ADDR"))
	counterContract, err := contract.NewContract(contractAddr, client)
	if err != nil {
		log.Fatal("合约绑定失败", err)
	}

	// ===================== 1. 查询当前计数 =====================
	fmt.Println("========== 查询计数 ==========")
	count, _ := counterContract.Get(&bind.CallOpts{})
	fmt.Println("当前 count =", count)

	// ===================== 2. 发送交易：increment() =====================
	fmt.Println("\n========== 执行 increment() ==========")
	privateKey, _ := crypto.HexToECDSA(os.Getenv("PRIVATE_KEY"))
	fromAddr := crypto.PubkeyToAddress(*privateKey.Public().(*ecdsa.PublicKey))

	// 获取 nonce & gas
	nonce, _ := client.PendingNonceAt(context.Background(), fromAddr)
	gasPrice, _ := client.SuggestGasPrice(context.Background())
	chainID, _ := client.ChainID(context.Background())

	// 签名授权
	auth, _ := bind.NewKeyedTransactorWithChainID(privateKey, chainID)
	auth.Nonce = big.NewInt(int64(nonce))
	auth.GasPrice = gasPrice
	auth.GasLimit = uint64(300000)

	// 调用合约方法
	tx, err := counterContract.Increment(auth)
	if err != nil {
		log.Fatal("调用失败", err)
	}
	fmt.Println("交易哈希:", tx.Hash().Hex())

	// 等待上链
	receipt, err := bind.WaitMined(context.Background(), client, tx)
	if err != nil {
		log.Fatal("上链失败", err)
	}
	fmt.Println("交易已打包，状态:", receipt.Status == types.ReceiptStatusSuccessful)

	// ===================== 3. 再次查询 =====================
	count, _ = counterContract.Get(&bind.CallOpts{})
	fmt.Println("执行后 count =", count)
}
