package main

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"
	"math/big"
	"os"
	"strconv"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/joho/godotenv"
)

func main() {
	//加载.env配置文件
	err := godotenv.Load()
	if err != nil {
		log.Fatal("❌ 加载 .env 文件失败，请检查文件是否存在")
	}
	fmt.Println("✅ 成功加载 .env 配置")

	// 从环境变量读取配置
	sepoliaRPC := os.Getenv("SEPOLIA_RPC")
	privateKey := os.Getenv("PRIVATE_KEY")
	toAddress := os.Getenv("TO_ADDRESS")
	queryBlockStr := os.Getenv("QUERY_BLOCK_NUMBER")

	//转换区块号
	queryBlockNumber, err := strconv.ParseInt(queryBlockStr, 10, 64)
	if err != nil {
		log.Fatalf("❌ 区块号格式错误: %v", err)
	}

	//连接 Sepolia 测试网
	client, err := ethclient.Dial(sepoliaRPC)
	if err != nil {
		log.Fatalf("❌ 连接 Sepolia 失败: %v", err)
	}
	defer client.Close()
	fmt.Println("✅ 成功连接 Sepolia 测试网\n")

	// ===================== 查询区块 =====================
	fmt.Println("========== 查询区块信息 ==========")
	getBlockInfo(client, uint64(queryBlockNumber))

	// ===================== 发送交易 =====================
	fmt.Println("\n========== 发送转账交易 ==========")
	txHash := sendTransaction(client, privateKey, toAddress)
	fmt.Printf("✅ 交易发送成功！\n交易哈希: %s\n", txHash)
	fmt.Printf("🔍 区块浏览器查看: https://sepolia.etherscan.io/tx/%s\n", txHash)
}

// getBlockInfo 查询指定区块信息
func getBlockInfo(client *ethclient.Client, blockNumber uint64) {
	block, err := client.BlockByNumber(context.Background(), big.NewInt(int64(blockNumber)))
	if err != nil {
		log.Fatalf("❌ 获取区块失败: %v", err)
	}

	fmt.Printf("区块号: %d\n", block.Number().Uint64())
	fmt.Printf("区块哈希: %s\n", block.Hash().Hex())
	fmt.Printf("时间: %s\n", time.Unix(int64(block.Time()), 0).Format("2006-01-02 15:04:05"))
	fmt.Printf("交易数量: %d 笔\n", len(block.Transactions()))
	fmt.Printf("Gas已使用: %d\n", block.GasUsed())
}

// sendTransaction 签名并发送交易
func sendTransaction(client *ethclient.Client, privateKey, toAddress string) string {
	// 解析私钥
	privateECDSA, err := crypto.HexToECDSA(privateKey)
	if err != nil {
		log.Fatalf("❌ 私钥解析失败: %v", err)
	}

	//获取发送方地址
	publicKey := privateECDSA.Public().(*ecdsa.PublicKey)
	fromAddress := crypto.PubkeyToAddress(*publicKey)
	fmt.Printf("发送方: %s\n", fromAddress.Hex())
	fmt.Printf("接收方: %s\n", toAddress)

	//获取nonce
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		log.Fatalf("❌ 获取 nonce 失败: %v", err)
	}

	//交易参数
	gasLimit := uint64(21000)
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		log.Fatalf("❌ 获取 gasPrice 失败: %v", err)
	}

	toAddr := common.HexToAddress(toAddress)
	value := new(big.Int).Mul(big.NewInt(1e15), big.NewInt(1)) // 0.001 ETH

	//构造交易
	tx := types.NewTx(&types.LegacyTx{
		Nonce:    nonce,
		GasPrice: gasPrice,
		Gas:      gasLimit,
		To:       &toAddr,
		Value:    value,
		Data:     []byte{},
	})

	//签名
	chanID, err := client.ChainID(context.Background())
	if err != nil {
		log.Fatalf("❌ 获取 chainID 失败: %v", err)
	}
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chanID), privateECDSA)
	if err != nil {
		log.Fatalf("❌ 交易签名失败: %v", err)
	}

	// 发送交易
	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		log.Fatalf("❌ 发送交易失败: %v", err)
	}
	return signedTx.Hash().Hex()
}
